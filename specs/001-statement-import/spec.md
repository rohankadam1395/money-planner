# Feature Specification: Statement Import

**Feature Branch**: `001-statement-import`

**Created**: 2026-07-05

**Status**: Draft

**Input**: User description: "phase 1: statement import"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload and Parse Bank Statement (Priority: P1)

A user clicks "Upload Statement" and selects a PDF file from their computer. The system validates the file, extracts transaction data (date, merchant, amount, debit/credit indicator, balance, description), and displays a preview of extracted transactions for confirmation before saving.

**Why this priority**: Core MVP functionality; without this, no transactions exist to analyze, budget, or categorize.

**Independent Test**: User can upload a valid bank statement file and preview the extracted transactions independently of any other feature.

**Acceptance Scenarios**:

1. **Given** a valid PDF bank statement, **When** user uploads it, **Then** system extracts and displays all transaction fields (date, merchant, amount, debit/credit, balance, description)
2. **Given** a CSV or Excel file with standard bank export format, **When** user uploads it, **Then** system extracts and displays transactions correctly
3. **Given** user clicks "Confirm", **When** transactions are imported, **Then** system persists them and displays success confirmation

---

### User Story 2 - Multi-Bank Support (Priority: P2)

A user with accounts at multiple banks (HDFC, ICICI, Axis, etc.) can upload statements from each bank. The system normalizes the different formats and merges them into a unified transaction view.

**Why this priority**: Essential for real users who typically have multiple accounts; enables complete financial picture.

**Independent Test**: User can upload statements from 2+ different banks and view all transactions together in a single list.

**Acceptance Scenarios**:

1. **Given** user uploads statements from Bank A and Bank B (different formats), **When** both are processed, **Then** transactions from both are merged and queryable together
2. **Given** overlapping date ranges from multiple banks, **When** user uploads a second statement, **Then** system displays all transactions chronologically without duplication errors

---

### User Story 3 - Upload History & Duplicate Detection (Priority: P3)

User can view previously uploaded statements and re-upload the same statement. System detects if a statement has already been imported (same file or same transaction date range from same bank) to prevent duplicates.

**Why this priority**: Improves user experience by preventing accidental re-imports and provides visibility into data history; secondary value compared to initial import.

**Independent Test**: User can upload the same statement twice and system prevents duplicate transactions in the database.

**Acceptance Scenarios**:

1. **Given** user uploads a statement already in the system, **When** upload is attempted, **Then** system warns user and prevents duplicate import
2. **Given** user views "Upload History", **When** page loads, **Then** all previously imported statements are listed with upload date and transaction count

---

### Edge Cases

- What happens when a file is corrupted or unreadable (PDF encrypted, Excel formula errors)?
- How does system handle statements with missing fields (e.g., no balance column in CSV)?
- How does system handle transactions with unusual characters or special symbols in merchant name?
- What if user uploads the same file before the first upload completes?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept PDF bank statements from major Indian banks (HDFC, ICICI, Axis, SBI, etc.)
- **FR-002**: System MUST accept CSV and Excel (.xlsx) files in standard bank export format
- **FR-003**: System MUST extract from each transaction: date, merchant/payee name, amount, debit/credit indicator, account balance, description/notes
- **FR-004**: System MUST validate extracted transactions before persisting (date format valid, amount is numeric, debit/credit is valid indicator)
- **FR-005**: System MUST display a preview of extracted transactions to user before confirming import
- **FR-006**: System MUST persist imported transactions to database with transaction date, merchant, amount, type (debit/credit), balance, and description
- **FR-007**: System MUST detect and prevent duplicate imports (same statement file or same date range from same bank)
- **FR-008**: System MUST store transaction currency in the database for future multi-currency support. For MVP Phase 1, system assumes all transactions are in Indian Rupees (INR); currency field defaults to "INR" on import. Future phases (US2+) will add currency conversion and multi-currency display.
- **FR-009**: System MUST handle statements with missing optional fields gracefully (e.g., balance column not always present in all bank formats)
- **FR-010**: System MUST log all import operations with file name, row count, success/failure status, timestamp

### Key Entities

- **Transaction**: Date, Merchant, Amount, Type (Debit/Credit), Balance, Description, Source Bank, Currency, Import Timestamp
- **Statement**: File Name, File Format (PDF/CSV/Excel), Source Bank, Statement Period (from/to date), Row Count, Upload Timestamp, Status (Success/Failed)
- **ImportJob**: User ID, Statement ID, Transaction Count, Import Timestamp, Status, Error Log (if any)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: User can upload a bank statement and view extracted transactions within 10 seconds
- **SC-002**: System extracts at least 95% of transactions correctly from test bank statements (matches manual count)
- **SC-003**: System rejects corrupted/unreadable files with clear error message in under 5 seconds
- **SC-004**: User can import statements from 3 different banks and retrieve all transactions in unified view
- **SC-005**: Duplicate transactions are detected and prevented with 100% accuracy across 12-month statement windows

## Assumptions

- Bank statements follow standard CSV/Excel/PDF layouts used by major Indian banks (date in column A, merchant in column B, amount in column C, etc.)
- User has stable internet connectivity for file upload (no multi-part resume required for MVP)
- File size limit: statements under 50MB (typical bank exports are 1-5MB for 12 months)
- Transactions are stored with currency field; MVP assumes Indian Rupees (INR) unless explicitly stated in statement header. Multi-currency conversion deferred to future phases (US2+).
- PDF parsing will use standard library (e.g., PyPDF2, pdfplumber in Python or equivalent in Go); no OCR required for machine-readable PDFs
- Duplicate detection scope: within same bank account, past 12 months of transactions
- User authentication already exists (out of scope for this feature; statement import assumes authenticated user context)
