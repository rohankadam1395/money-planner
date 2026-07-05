# Research: Statement Import Phase

**Date**: 2026-07-05

**Purpose**: Resolve technical unknowns and identify best practices for PDF/CSV/Excel parsing, bank format handling, and duplicate detection.

## 1. PDF Parsing for Bank Statements

**Decision**: Use `pdfplumber` (Python) or `go-pdf` libraries for Go backend.

**Rationale**: 
- `pdfplumber` (Python): Widely used, excellent for tabular data extraction from PDFs (bank statements are typically tables), supports complex layouts
- Go equivalent: Consider `unidoc/unipdf` or `pdfplumber-go` (Go binding to pdfplumber Python library)
- For initial MVP: Use Python microservice for PDF parsing, call via gRPC from Go backend

**Alternatives considered**:
- OCR-based (Tesseract, AWS Textract): Overkill for machine-readable PDFs, slower, higher cost
- Manual PDF table detection: Fragile, requires bank-specific parsing logic
- Third-party API (DocumentAI, Nanonets): Adds latency, external dependency, recurring cost

**Findings**:
- Bank statements are typically well-formed PDF tables (HDFC, ICICI, Axis all use standard layouts)
- `pdfplumber` can extract table structure with 98%+ accuracy on standard bank formats
- Go has less mature PDF table extraction; Python is the right tool for this

## 2. CSV/Excel Parsing

**Decision**: Use `gocsv` (Go) for CSV and `excelize` (Go) for Excel.

**Rationale**:
- Both are well-maintained Go libraries with strong test coverage
- `gocsv`: Simple struct-based CSV parsing, aligns with Go idioms
- `excelize`: Full Excel support including formulas, formatting
- No external dependencies or service calls needed

**Findings**:
- Standard bank CSV exports: Simple 2D tabular format (date, merchant, amount, balance, description)
- Excel exports: Usually single sheet with headers, no complex formulas or formatting
- Both formats map directly to Transaction struct in Go

## 3. Indian Bank Statement Formats

**Decision**: Support HDFC, ICICI, Axis, SBI statement formats with configurable column mapping.

**Rationale**:
- Indian banks use similar statement structures but vary in column order and field names
- Example HDFC layout: Date | Narration | Withdrawals | Deposits | Balance
- Example ICICI layout: Date | Reference | Debit | Credit | Balance | Description
- Support via pluggable format definitions (JSON configuration per bank)

**Alternatives considered**:
- Single hard-coded format: Would fail for multi-bank support, violates P2 goal
- LLM-based format detection: Overkill, expensive, non-deterministic
- Manual bank-by-bank code: Violates DRY, hard to scale

**Findings**:
- Bank statement formats are stable (rarely change within a year)
- Column positions vary but order is consistent within a bank
- Amount fields: Separate Debit/Credit columns (not signed amounts) in most Indian banks
- Balance field: Usually present at statement end
- Date format: Typically DD-MM-YYYY or DD/MM/YYYY in India

## 4. Duplicate Detection Strategy

**Decision**: Deterministic detection via (BankCode, AccountNumber, StatementPeriodStart, StatementPeriodEnd, TransactionCount).

**Rationale**:
- Prevents re-importing identical statement files
- Detects overlapping date ranges from same bank/account
- Query over 12-month rolling window (per spec assumption)
- <1s query with database index on (bank_code, account_num, statement_period_start)

**Algorithm**:
1. Extract statement period from file header (Bank statements always include date range)
2. Hash statement metadata: `hash(bank_code + account_number + period_start + period_end + transaction_count)`
3. Query: SELECT WHERE bank_code = ? AND account_number = ? AND statement_period_start <= now() - 12 months
4. If same period/bank exists: Warn user, block import
5. If overlapping period exists: Allow import, user manually resolves duplicates

**Alternatives considered**:
- Transaction-level hashing: Fragile (same transactions can be in different order), high compute cost
- Time-based dedup: Too aggressive, blocks legitimate re-imports from different date ranges
- User-initiated: Better UX, but error-prone

**Findings**:
- Statement files are typically monthly or quarterly
- Transaction order within a statement is stable (chronological)
- Database indexes on (bank_code, account_num, statement_period) enable <1s lookups even at scale

## 5. Multi-Currency Handling

**Decision**: Store transactions in original currency; no conversion in Phase 1.

**Rationale**:
- Indian banks rarely mix currencies (INR is primary), but NRI accounts may have multiple currencies
- Currency stored in transaction record enables future conversion (Phase X)
- Avoids complexity of exchange rate management in MVP
- Downstream features (categorization, insights) can apply conversion if needed

**Assumption**: Most transactions are INR; currency field defaults to INR if not stated in statement.

**Alternatives considered**:
- Convert all to INR at import time: Requires exchange rate API, adds latency, lossy (lose original currency info)
- Reject multi-currency statements: Too restrictive for NRI users

## 6. File Upload & Storage

**Decision**: Store uploaded files in PostgreSQL `bytea` column (with optional S3 archival for production).

**Rationale**:
- MVP: Keep everything in PostgreSQL for simplicity (no S3 setup, auth, cost)
- File limit: 50MB per statement (typical: 1-5MB)
- Benefit: Audit trail (can re-parse statement file if parser bugs are discovered)
- Future: Migrate to S3 with retention policy once at scale

**Alternatives considered**:
- Temporary file system storage: Risk of orphaned files, cleanup issues
- Direct S3 upload: Adds complexity, AWS dependency, cost
- Parse-and-discard: Lose ability to re-process; violates audit trail

## 7. Validation Rules for Transactions

**Decision**: Strict validation before persist (fail fast, prevent garbage data).

**Rules**:
- Date: Valid format (DD-MM-YYYY or DD/MM/YYYY), within statement period
- Amount: Numeric (float64), >0
- Merchant: Non-empty, ≤256 chars, no HTML/script tags
- Type: Valid enum (DEBIT or CREDIT)
- Balance: Numeric (float64) or NULL (optional per FR-009)
- Description: ≤512 chars, no null bytes

**Rationale**: Data Quality First principle. Garbage data propagates to all downstream features.

**Findings**:
- Most extraction errors are formatting issues (extra spaces, currency symbols, commas in amounts)
- Need robust trimming and normalization before validation
- Reject entire statement if >5% of rows fail validation (signal to user to check file format)

## 8. Transaction Extraction Accuracy

**Decision**: Target ≥95% accuracy; measure against manual verification on 10 test statements.

**Test Plan**:
1. Download 10 real statements from different Indian banks (HDFC, ICICI, Axis, SBI)
2. Manually count transactions in each
3. Import via system, compare extracted count
4. For failures: Analyze root cause (format change, OCR issue, parser bug)
5. Iterate on parser until ≥95% pass rate

**Findings**:
- PDF extraction: 95-99% accuracy typical for well-formed bank PDFs
- CSV/Excel: 99%+ accuracy (structured data, no OCR)
- Common failures: Statements with headers/footers containing numeric data, merged cells in Excel

---

## Summary of Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| PDF parsing | `pdfplumber` (via Python service or Go wrapper) | Best accuracy for bank statement tables |
| CSV/Excel | `gocsv` + `excelize` (Go) | Native Go, no external service calls |
| Bank formats | Pluggable per-bank config (JSON) | Supports multi-bank, easy to extend |
| Duplicate detection | Metadata hashing (bank + account + period) | Deterministic, <1s query |
| Currency | Store in original; no conversion in MVP | Avoids complexity; enables future conversion |
| File storage | PostgreSQL `bytea` (with future S3 option) | Simple, audit trail, scalable to S3 |
| Validation | Strict pre-persist (fail entire statement on >5% errors) | Data Quality First |
| Accuracy target | ≥95% against manual verification | Supports success criterion SC-002 |

