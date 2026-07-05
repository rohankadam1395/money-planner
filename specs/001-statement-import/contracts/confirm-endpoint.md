# API Contract: Confirm Statement Import

**Endpoint**: `POST /api/statements/{statement_id}/confirm`

**Purpose**: User confirms preview and commits extracted transactions to database.

**Status**: REST API (JSON response)

---

## Request

### Path Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `statement_id` | UUID | YES | Statement ID from upload |

### Headers

```
Authorization: Bearer {jwt_token}
Content-Type: application/json
```

### Body (optional)

```json
{
  "auto_categorize": false
}
```

**Fields**:
- `auto_categorize` (Boolean, optional, default=false): If true, trigger automatic transaction categorization after import (future feature)

**Example Request**:
```bash
curl -X POST \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{"auto_categorize": false}' \
  https://api.moneyplan.ai/api/statements/550e8400-e29b-41d4-a716-446655440000/confirm
```

---

## Response (Success: 200 OK)

```json
{
  "statement_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "SUCCESS",
  "import_job_id": "660e8400-e29b-41d4-a716-446655440001",
  "transactions_imported": 43,
  "transactions_failed": 2,
  "completion_time_ms": 2450,
  "message": "Statement import completed successfully. 43 transactions saved.",
  "failed_rows": [
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
}
```

**Field Descriptions**:

| Field | Type | Notes |
|-------|------|-------|
| `statement_id` | UUID | Statement identifier |
| `status` | ENUM | SUCCESS (≥95% imported) or FAILED (too many errors) |
| `import_job_id` | UUID | Job ID for audit trail |
| `transactions_imported` | Integer | Count of successfully persisted transactions |
| `transactions_failed` | Integer | Count of rows that failed validation (discarded) |
| `completion_time_ms` | Integer | Total time to parse, validate, and persist |
| `message` | String | Human-readable summary |
| `failed_rows[]` | Array | Details of validation failures (if any) |

---

## Response (Error Cases)

### 400 Bad Request
```json
{
  "error": "INVALID_STATE_FOR_CONFIRM",
  "message": "Statement must be in PENDING state to confirm (current: SUCCESS)"
}
```

**Possible errors**:
- `INVALID_STATE_FOR_CONFIRM` — Statement status is not PENDING
- `STATEMENT_VALIDATION_FAILED` — >5% of rows failed validation
- `INVALID_REQUEST_BODY` — auto_categorize parameter invalid

### 404 Not Found
```json
{
  "error": "STATEMENT_NOT_FOUND",
  "message": "Statement with ID 550e8400-... not found"
}
```

### 401 Unauthorized
```json
{
  "error": "UNAUTHORIZED",
  "message": "JWT token invalid or expired"
}
```

### 403 Forbidden
```json
{
  "error": "FORBIDDEN",
  "message": "You do not have access to this statement"
}
```

### 409 Conflict (Duplicate detected)
```json
{
  "error": "DUPLICATE_STATEMENT_DETECTED",
  "message": "A statement for this bank/account and date range already exists",
  "existing_statement_id": "440e8400-e29b-41d4-a716-446655440000",
  "existing_import_date": "2026-06-20"
}
```

### 422 Unprocessable Entity (Validation failed)
```json
{
  "error": "IMPORT_VALIDATION_FAILED",
  "message": "6 of 100 rows failed validation (6% failure rate > 5% threshold)",
  "transactions_imported": 94,
  "transactions_failed": 6,
  "error_summary": {
    "invalid_amount": 3,
    "invalid_date": 2,
    "empty_merchant": 1
  }
}
```

---

## Processing Flow

1. **User calls confirm** after reviewing preview
2. **Backend validates**:
   - Statement status is PENDING
   - No newer overlapping statement exists for same bank/account
   - Overall failure rate ≤5%
3. **If valid**: Update Statement status → PROCESSING, ImportJob status → IN_PROGRESS
4. **Database transaction**:
   - Persist all valid transactions (INVALID rows discarded)
   - Update Statement status → SUCCESS, ImportJob status → SUCCESS
   - Log completion time
5. **Response**: 200 OK with import summary
6. **If >5% failed**: 422 Unprocessable Entity, no transactions persisted

---

## Transaction Safety

**ACID Guarantees**:
- All valid transactions persisted atomically (either all or none)
- Statement and ImportJob records updated together
- If persistence fails at any point: full rollback, Statement status = FAILED

**Idempotency**:
- Calling confirm twice with same statement_id:
  - First call: Persists transactions, returns 200 OK
  - Second call: Statement already SUCCESS, returns 409 Conflict (invalid state)

---

## Contract Tests

**Test 1**: Successful confirm with all valid transactions
- Setup: Upload statement, verify preview shows 0 errors
- Request: POST /api/statements/{id}/confirm
- Assert: 200 OK, status=SUCCESS, transactions_imported > 0, transactions_failed=0

**Test 2**: Confirm with some validation failures (≤5%)
- Setup: Upload statement with 100 rows, 3 fail validation
- Request: POST /api/statements/{id}/confirm
- Assert: 200 OK, status=SUCCESS, transactions_imported=97, transactions_failed=3

**Test 3**: Confirm with >5% failures (exceeds threshold)
- Setup: Upload statement with 100 rows, 6 fail validation
- Request: POST /api/statements/{id}/confirm
- Assert: 422 Unprocessable Entity, status=FAILED, no transactions persisted

**Test 4**: Duplicate statement conflict
- Setup: Import statement once, then try to confirm same bank/account/period again
- Request: POST /api/statements/{new_id}/confirm
- Assert: 409 Conflict, error=DUPLICATE_STATEMENT_DETECTED

**Test 5**: Invalid state (already confirmed)
- Setup: Confirm a statement once
- Request: POST /api/statements/{same_id}/confirm again
- Assert: 400 Bad Request, error=INVALID_STATE_FOR_CONFIRM

**Test 6**: Unauthorized (missing JWT)
- Request: POST /api/statements/{id}/confirm without Authorization header
- Assert: 401 Unauthorized

