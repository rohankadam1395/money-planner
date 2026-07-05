# Data Model: Statement Import

**Date**: 2026-07-05

**Purpose**: Define entities, attributes, relationships, and validation rules for the Statement Import feature.

---

## Entities

### 1. Transaction

Represents a single financial transaction extracted from a bank statement.

**Fields**:

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `transaction_id` | UUID | PK, NOT NULL | Unique identifier |
| `user_id` | UUID | FK (users table), NOT NULL | Statement owner |
| `statement_id` | UUID | FK (statements table), NOT NULL | Source statement |
| `transaction_date` | DATE | NOT NULL | When transaction occurred (YYYY-MM-DD) |
| `merchant` | VARCHAR(256) | NOT NULL | Payee/merchant name (trimmed, no HTML) |
| `amount` | DECIMAL(12, 2) | NOT NULL, > 0 | Transaction amount |
| `type` | ENUM (DEBIT, CREDIT) | NOT NULL | Transaction direction |
| `balance` | DECIMAL(12, 2) | NULL | Account balance after transaction (optional per FR-009) |
| `description` | VARCHAR(512) | NULL | Memo/notes field |
| `currency` | CHAR(3) | DEFAULT 'INR' | ISO 4217 code (e.g., INR, USD, EUR) |
| `imported_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | When record was imported |
| `bank_code` | CHAR(4) | NOT NULL | Bank identifier (e.g., HDFC, ICIC, AXIS, SBI) |
| `account_number_hash` | VARCHAR(64) | NOT NULL | SHA-256 hash of account number (never store plaintext) |
| `raw_data` | JSONB | NULL | Original parsed fields (for debugging/re-processing) |
| `created_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row update time |

**Indexes**:
- PK: `transaction_id`
- `(user_id, transaction_date DESC)` — Query user's transactions by date
- `(statement_id)` — Query transactions from a specific statement
- `(user_id, bank_code, account_number_hash, transaction_date)` — Duplicate detection query

**Validation Rules**:
- `transaction_date` must be within statement's period (statement.period_start ≤ transaction_date ≤ statement.period_end)
- `amount` must be positive decimal number (no scientific notation)
- `type` must be valid ENUM value
- `merchant` must be non-empty, ≤256 chars, no null bytes, no HTML/script tags
- `currency` must be valid ISO 4217 code
- `account_number_hash` must be SHA-256 hex string (64 chars)

**Privacy Notes**:
- Never store plaintext account numbers (use hash)
- Transactions are sensitive PII; encrypt at rest (PostgreSQL encryption), in transit (TLS)
- Log access to transactions for audit trail

---

### 2. Statement

Represents a bank statement file and its import metadata.

**Fields**:

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `statement_id` | UUID | PK, NOT NULL | Unique identifier |
| `user_id` | UUID | FK (users table), NOT NULL | Statement owner |
| `file_name` | VARCHAR(256) | NOT NULL | Original uploaded file name |
| `file_format` | ENUM (PDF, CSV, XLSX) | NOT NULL | File type |
| `file_size_bytes` | INT | NOT NULL, > 0 | Size of uploaded file |
| `file_hash` | VARCHAR(64) | NOT NULL | SHA-256 of file (duplicate detection) |
| `bank_code` | CHAR(4) | NOT NULL | Bank identifier (HDFC, ICIC, AXIS, SBI, etc.) |
| `account_number_hash` | VARCHAR(64) | NOT NULL | SHA-256 hash of account number |
| `statement_period_start` | DATE | NOT NULL | First transaction date in statement |
| `statement_period_end` | DATE | NOT NULL | Last transaction date in statement |
| `transaction_count` | INT | NOT NULL, ≥ 0 | Total rows extracted from file |
| `status` | ENUM (SUCCESS, FAILED, PENDING) | NOT NULL | Import result |
| `error_log` | TEXT | NULL | Error messages if status=FAILED |
| `uploaded_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | When file was uploaded |
| `imported_at` | TIMESTAMP | NULL | When import completed |
| `created_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row update time |

**Indexes**:
- PK: `statement_id`
- `(user_id, uploaded_at DESC)` — List user's statements by upload date
- `(file_hash)` — Fast duplicate file detection
- `(bank_code, account_number_hash, statement_period_start)` — Overlapping period detection

**Validation Rules**:
- `statement_period_start` ≤ `statement_period_end`
- `file_size_bytes` ≤ 50,000,000 (50MB limit)
- `transaction_count` ≥ 0
- `file_hash` must be SHA-256 hex string (64 chars)
- `account_number_hash` must be SHA-256 hex string (64 chars)

**Privacy Notes**:
- Never store plaintext account numbers or file contents (hash only)
- Limit error_log to non-sensitive info (don't expose raw transaction data in errors)

---

### 3. ImportJob

Represents a user-triggered import operation (may import multiple statements in sequence).

**Fields**:

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `import_job_id` | UUID | PK, NOT NULL | Unique identifier |
| `user_id` | UUID | FK (users table), NOT NULL | Job initiator |
| `statement_id` | UUID | FK (statements table), NOT NULL | Associated statement |
| `status` | ENUM (QUEUED, IN_PROGRESS, SUCCESS, FAILED) | NOT NULL | Job status |
| `transaction_count` | INT | NULL | Transactions successfully imported |
| `error_count` | INT | DEFAULT 0 | Rows that failed validation |
| `started_at` | TIMESTAMP | NULL | When processing began |
| `completed_at` | TIMESTAMP | NULL | When processing finished |
| `duration_ms` | INT | NULL | Total processing time |
| `error_log` | TEXT | NULL | Detailed error messages |
| `created_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW(), NOT NULL | Row update time |

**Indexes**:
- PK: `import_job_id`
- `(user_id, created_at DESC)` — List user's import history

**Validation Rules**:
- `transaction_count` ≥ 0
- `error_count` ≥ 0
- If `status = SUCCESS`: `error_count` ≤ 0.05 * transaction_count (≤5% failure tolerance)
- `duration_ms` > 0 if completed

**Privacy Notes**:
- error_log must not contain transaction data; only summarize issues (e.g., "Row 5: Invalid date format")

---

## Relationships

```
User (external)
  ├── 1:N → Statement (user_id)
  └── 1:N → ImportJob (user_id)

Statement
  ├── 1:N → Transaction (statement_id)
  └── 1:N → ImportJob (statement_id)

ImportJob
  └── N:1 ← Transaction (via Statement)
```

**Cascade Rules**:
- DELETE User → CASCADE delete Statements, ImportJobs, Transactions
- DELETE Statement → CASCADE delete Transactions, ImportJobs (keep for audit)

---

## State Transitions

### Statement States

```
PENDING → IN_PROGRESS → SUCCESS
       ↘ FAILED ↗
```

- **PENDING**: File uploaded, validation queued
- **IN_PROGRESS**: Parsing and validation in progress
- **SUCCESS**: All transactions persisted successfully (error_count ≤5%)
- **FAILED**: Parsing/validation error; no transactions persisted

### ImportJob States

```
QUEUED → IN_PROGRESS → SUCCESS
      ↘ FAILED ↗
```

---

## Duplicate Detection Schema

**Query**: Find existing statements with overlapping date ranges for the same bank/account.

```sql
SELECT * FROM statements 
WHERE bank_code = :bank_code
  AND account_number_hash = :account_hash
  AND statement_period_end >= (now() - interval '12 months')
  AND status = 'SUCCESS'
ORDER BY statement_period_start DESC;
```

**Logic**:
1. If `file_hash` matches existing: Reject (same file)
2. If `bank_code + account_number_hash + statement_period` matches: Warn user (exact duplicate)
3. If date ranges overlap: Warn user (overlapping statements)
4. If no conflicts: Proceed with import

---

## Encryption & Security

**At Rest**:
- PostgreSQL transparent encryption (TDE) or per-column encryption for sensitive fields
- account_number_hash: Salted SHA-256 (salt = user_id to prevent pre-computed attacks)
- Transactions: Consider PII encryption given sensitive nature

**In Transit**:
- All API calls over TLS 1.3+
- file_hash and account_number_hash computed on client (if sensitive) or server with HTTPS

**Audit Trail**:
- Log all import operations: user_id, statement_id, status, timestamp, transaction_count
- Retention: Keep for 2 years (compliance requirement TBD)

---

## Migration Strategy

**Phase 1 (Initial)**:
```sql
CREATE TABLE statements (
  statement_id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  file_name VARCHAR(256) NOT NULL,
  file_format CHAR(4) NOT NULL CHECK (file_format IN ('PDF', 'CSV', 'XLSX')),
  file_size_bytes INT NOT NULL CHECK (file_size_bytes > 0 AND file_size_bytes <= 50000000),
  file_hash VARCHAR(64) NOT NULL UNIQUE,
  bank_code CHAR(4) NOT NULL,
  account_number_hash VARCHAR(64) NOT NULL,
  statement_period_start DATE NOT NULL,
  statement_period_end DATE NOT NULL,
  transaction_count INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING')),
  error_log TEXT,
  uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
  imported_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(file_hash),
  INDEX idx_user_uploaded (user_id, uploaded_at DESC),
  INDEX idx_dedup (bank_code, account_number_hash, statement_period_start)
);

CREATE TABLE transactions (
  transaction_id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  statement_id UUID NOT NULL REFERENCES statements(statement_id) ON DELETE CASCADE,
  transaction_date DATE NOT NULL,
  merchant VARCHAR(256) NOT NULL,
  amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
  type VARCHAR(8) NOT NULL CHECK (type IN ('DEBIT', 'CREDIT')),
  balance DECIMAL(12, 2),
  description VARCHAR(512),
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  imported_at TIMESTAMP NOT NULL DEFAULT NOW(),
  bank_code CHAR(4) NOT NULL,
  account_number_hash VARCHAR(64) NOT NULL,
  raw_data JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  INDEX idx_user_date (user_id, transaction_date DESC),
  INDEX idx_statement (statement_id),
  INDEX idx_dedup (user_id, bank_code, account_number_hash, transaction_date)
);

CREATE TABLE import_jobs (
  import_job_id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  statement_id UUID NOT NULL REFERENCES statements(statement_id),
  status VARCHAR(32) NOT NULL DEFAULT 'QUEUED' CHECK (status IN ('QUEUED', 'IN_PROGRESS', 'SUCCESS', 'FAILED')),
  transaction_count INT,
  error_count INT NOT NULL DEFAULT 0,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  duration_ms INT,
  error_log TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  INDEX idx_user_history (user_id, created_at DESC)
);
```

