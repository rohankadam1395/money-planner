# API Contract: Preview Extracted Transactions

**Endpoint**: `GET /api/statements/{statement_id}/preview`

**Purpose**: Retrieve extracted transactions from an uploaded statement before user confirms import. Shows transaction data for preview/confirmation.

**Status**: REST API (JSON response)

---

## Request

### Path Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `statement_id` | UUID | YES | Statement ID returned from upload endpoint |

### Query Parameters

| Parameter | Type | Default | Notes |
|-----------|------|---------|-------|
| `limit` | Integer | 100 | Max rows to return (1-1000) |
| `offset` | Integer | 0 | Pagination offset |

### Headers

```
Authorization: Bearer {jwt_token}
```

**Example Request**:
```bash
curl -X GET \
  -H "Authorization: Bearer {jwt_token}" \
  https://api.moneyplan.ai/api/statements/550e8400-e29b-41d4-a716-446655440000/preview?limit=50&offset=0
```

---

## Response (Success: 200 OK)

```json
{
  "statement_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING",
  "file_name": "HDFC_Statement_Jun2026.pdf",
  "file_format": "PDF",
  "bank_code": "HDFC",
  "account_number_hash": "abc123def456...",
  "statement_period_start": "2026-06-01",
  "statement_period_end": "2026-06-30",
  "total_transactions": 45,
  "transactions": [
    {
      "row_number": 1,
      "transaction_date": "2026-06-01",
      "merchant": "SALARY CREDIT",
      "amount": 150000.00,
      "type": "CREDIT",
      "balance": 250000.00,
      "description": "Monthly Salary",
      "validation_status": "VALID"
    },
    {
      "row_number": 2,
      "transaction_date": "2026-06-02",
      "merchant": "AMAZON SELLER",
      "amount": 5499.50,
      "type": "DEBIT",
      "balance": 244500.50,
      "description": "Online Purchase",
      "validation_status": "VALID"
    },
    {
      "row_number": 3,
      "transaction_date": "2026-06-03",
      "merchant": "ELECTRICITY BILL",
      "amount": 3500.00,
      "type": "DEBIT",
      "balance": 241000.50,
      "description": "",
      "validation_status": "VALID"
    }
  ],
  "validation_summary": {
    "total_rows": 45,
    "valid_rows": 43,
    "invalid_rows": 2,
    "error_rate_percent": 4.4,
    "errors": [
      {
        "row_number": 15,
        "field": "amount",
        "error": "Invalid decimal format: '5,000.00 Rs'"
      },
      {
        "row_number": 28,
        "field": "transaction_date",
        "error": "Date '32-06-2026' is outside statement period"
      }
    ]
  },
  "import_job": {
    "import_job_id": "660e8400-e29b-41d4-a716-446655440001",
    "status": "IN_PROGRESS",
    "started_at": "2026-07-05T10:30:15Z",
    "progress_percent": 75
  },
  "links": {
    "confirm": "/api/statements/550e8400-e29b-41d4-a716-446655440000/confirm",
    "cancel": "/api/statements/550e8400-e29b-41d4-a716-446655440000/cancel"
  }
}
```

**Field Descriptions**:

| Field | Type | Notes |
|-------|------|-------|
| `statement_id` | UUID | Statement identifier |
| `status` | ENUM | PENDING (import not confirmed), PROCESSING, SUCCESS, FAILED |
| `file_name` | String | Original file name |
| `file_format` | String | PDF, CSV, or XLSX |
| `bank_code` | String | Bank identifier |
| `statement_period_start` | Date | First transaction date extracted from statement |
| `statement_period_end` | Date | Last transaction date extracted from statement |
| `total_transactions` | Integer | Total rows in file |
| `transactions[]` | Array | Paginated list of extracted transactions |
| `transactions[].row_number` | Integer | 1-based row number in statement file |
| `transactions[].transaction_date` | Date | YYYY-MM-DD |
| `transactions[].merchant` | String | Payee/merchant name |
| `transactions[].amount` | Decimal | Absolute value |
| `transactions[].type` | ENUM | DEBIT or CREDIT |
| `transactions[].balance` | Decimal | Account balance after transaction (may be null) |
| `transactions[].description` | String | Memo/notes (may be empty) |
| `transactions[].validation_status` | ENUM | VALID or INVALID |
| `validation_summary.total_rows` | Integer | Total rows parsed |
| `validation_summary.valid_rows` | Integer | Rows that passed validation |
| `validation_summary.invalid_rows` | Integer | Rows that failed validation |
| `validation_summary.error_rate_percent` | Float | Percentage of invalid rows |
| `validation_summary.errors[]` | Array | Specific validation errors |
| `import_job.status` | ENUM | Current processing status |
| `import_job.progress_percent` | Integer | 0-100 progress indicator |
| `links.confirm` | String | POST endpoint to confirm import |
| `links.cancel` | String | DELETE endpoint to cancel/discard |

---

## Response (Error Cases)

### 404 Not Found
```json
{
  "error": "STATEMENT_NOT_FOUND",
  "message": "Statement with ID 550e8400-... not found"
}
```

### 400 Bad Request (Invalid parameters)
```json
{
  "error": "INVALID_LIMIT",
  "message": "Limit must be between 1 and 1000"
}
```

### 401 Unauthorized
```json
{
  "error": "UNAUTHORIZED",
  "message": "JWT token invalid or expired"
}
```

### 403 Forbidden (Not owner)
```json
{
  "error": "FORBIDDEN",
  "message": "You do not have access to this statement"
}
```

### 410 Gone (Import already confirmed)
```json
{
  "error": "STATEMENT_ALREADY_CONFIRMED",
  "message": "This statement has already been imported and cannot be previewed",
  "import_completed_at": "2026-07-05T10:35:00Z"
}
```

---

## State Transitions

**Statement States During Preview**:
- `PENDING`: Upload complete, preview available, awaiting user confirmation
- `PROCESSING`: User confirmed, transactions being persisted
- `SUCCESS`: All transactions persisted successfully
- `FAILED`: Validation or persistence error; no transactions saved

**At Preview**:
- If `status = PENDING`: Full preview available with validation_summary
- If `status = PROCESSING`: Snapshot preview available (transactions extracted but not yet persisted)
- If `status = SUCCESS`: Preview not available (410 Gone)
- If `status = FAILED`: Error details available for user understanding

---

## Contract Tests

**Test 1**: Valid preview request (PENDING status)
- Request: GET /api/statements/{valid_id}/preview
- Assert: 200 OK, transactions array populated, validation_summary shows errors (if any)

**Test 2**: Preview with pagination
- Request: GET /api/statements/{valid_id}/preview?limit=10&offset=20
- Assert: 200 OK, returns exactly 10 transactions starting from offset 20

**Test 3**: Preview during processing (IN_PROGRESS)
- Request: GET /api/statements/{id_in_progress}/preview
- Assert: 200 OK, snapshot of extracted data, progress_percent > 0, status=IN_PROGRESS

**Test 4**: Preview after confirmation (SUCCESS)
- Request: GET /api/statements/{confirmed_id}/preview
- Assert: 410 Gone, error=STATEMENT_ALREADY_CONFIRMED

**Test 5**: Invalid statement ID
- Request: GET /api/statements/invalid-uuid/preview
- Assert: 404 Not Found

**Test 6**: Unauthorized access (different user's statement)
- Request: GET /api/statements/{other_users_statement}/preview (as different user)
- Assert: 403 Forbidden

**Test 7**: Invalid pagination parameters
- Request: GET /api/statements/{id}/preview?limit=2000&offset=-1
- Assert: 400 Bad Request, error=INVALID_LIMIT

