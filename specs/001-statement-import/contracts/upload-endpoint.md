# API Contract: Statement Upload

**Endpoint**: `POST /api/statements/upload`

**Purpose**: Accept a bank statement file (PDF/CSV/Excel) and initiate import processing.

**Status**: REST API (JSON request/response)

---

## Request

### Headers

```
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}
X-Idempotency-Key: {uuid} (optional, for retry safety)
```

### Body (multipart/form-data)

| Field | Type | Required | Constraints | Notes |
|-------|------|----------|-------------|-------|
| `file` | Binary | YES | ≤50MB, type=PDF/CSV/XLSX | Bank statement file |
| `bank_code` | String | YES | ENUM (HDFC, ICIC, AXIS, SBI, YESB, KOTAK, etc.) | Bank identifier |
| `account_number_hash` | String | YES | SHA-256 hex, 64 chars | Hashed account number (client-side) |

**Example Request** (curl):
```bash
curl -X POST \
  -H "Authorization: Bearer {jwt_token}" \
  -F "file=@statement.pdf" \
  -F "bank_code=HDFC" \
  -F "account_number_hash=abc123def456..." \
  https://api.moneyplan.ai/api/statements/upload
```

### Validation

- File must be valid PDF/CSV/XLSX (magic bytes check)
- File size ≤50MB
- `bank_code` must be in approved list (case-insensitive, converted to uppercase)
- `account_number_hash` must be 64-character hex string

---

## Response (Success: 202 Accepted)

```json
{
  "statement_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING",
  "file_name": "statement.pdf",
  "file_format": "PDF",
  "file_size_bytes": 1245670,
  "bank_code": "HDFC",
  "account_number_hash": "abc123def456...",
  "uploaded_at": "2026-07-05T10:30:00Z",
  "import_job_id": "660e8400-e29b-41d4-a716-446655440001",
  "location": "/api/statements/550e8400-e29b-41d4-a716-446655440000"
}
```

**Field Descriptions**:

| Field | Type | Notes |
|-------|------|-------|
| `statement_id` | UUID | Unique identifier for this statement |
| `status` | ENUM | Always PENDING on initial upload; client polls or listens for updates |
| `file_name` | String | Original uploaded file name |
| `file_format` | String | Detected format (PDF, CSV, XLSX) |
| `file_size_bytes` | Integer | File size |
| `bank_code` | String | Normalized bank code |
| `account_number_hash` | String | Account hash provided in request |
| `uploaded_at` | ISO 8601 | Timestamp of upload |
| `import_job_id` | UUID | Job ID for tracking import progress |
| `location` | String | Link to statement detail endpoint |

---

## Response (Error Cases)

### 400 Bad Request
```json
{
  "error": "INVALID_FILE_FORMAT",
  "message": "File must be PDF, CSV, or XLSX",
  "request_id": "req-xyz123"
}
```

**Possible error codes**:
- `INVALID_FILE_FORMAT` — File is not PDF/CSV/XLSX
- `FILE_TOO_LARGE` — File exceeds 50MB
- `INVALID_BANK_CODE` — Bank code not in approved list
- `INVALID_ACCOUNT_HASH` — Account hash is not 64-char hex string

### 401 Unauthorized
```json
{
  "error": "UNAUTHORIZED",
  "message": "JWT token invalid or expired"
}
```

### 409 Conflict (Duplicate)
```json
{
  "error": "DUPLICATE_FILE",
  "message": "Statement file already imported on 2026-06-20",
  "existing_statement_id": "440e8400-e29b-41d4-a716-446655440000",
  "request_id": "req-xyz123"
}
```

### 413 Payload Too Large
```json
{
  "error": "FILE_TOO_LARGE",
  "message": "Maximum file size is 50MB"
}
```

### 429 Too Many Requests
```json
{
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Maximum 10 concurrent uploads per user",
  "retry_after_seconds": 300
}
```

### 500 Internal Server Error
```json
{
  "error": "INTERNAL_ERROR",
  "message": "Failed to process statement",
  "request_id": "req-xyz123"
}
```

---

## Processing Flow

1. **Upload received** (202 Accepted) → `status = PENDING`
2. **Backend validates** file format, size, bank code
3. **If valid**: Extract metadata (period_start, period_end, transaction_count from header), check for duplicates
4. **If no duplicates**: Create Statement (status=PENDING) and ImportJob (status=QUEUED)
5. **Async job**: Parse file, extract transactions, validate, persist
6. **Update**: Statement status → SUCCESS or FAILED; ImportJob status → SUCCESS or FAILED
7. **Client polls** `/api/statements/{statement_id}` to check import_job status

---

## Contract Tests

**Test 1**: Valid PDF upload
- Request: PDF file, valid bank_code, valid account_hash
- Assert: 202 Accepted, statement_id returned, status=PENDING

**Test 2**: Duplicate file detection
- Request: Same file uploaded twice
- First upload: 202 Accepted
- Second upload: 409 Conflict with existing_statement_id

**Test 3**: Invalid file format
- Request: .txt or .zip file
- Assert: 400 Bad Request, error=INVALID_FILE_FORMAT

**Test 4**: File too large
- Request: 51MB PDF
- Assert: 413 Payload Too Large

**Test 5**: Unauthorized (missing JWT)
- Request: No Authorization header
- Assert: 401 Unauthorized

**Test 6**: Rate limit
- Request: 11 concurrent uploads
- Assert: 10th succeeds, 11th gets 429 with retry_after_seconds

