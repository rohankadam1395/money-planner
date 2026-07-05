# Statement Import Feature - Testing Guide

This guide provides comprehensive instructions for testing the Statement Import feature (US1) using the provided sample data.

## Prerequisites

- Backend running on `http://localhost:8080`
- Frontend running on `http://localhost:3000`
- PostgreSQL database configured and migrations run
- Valid authentication token

## Quick Start

### 1. Start the Backend

```bash
cd backend
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=money_planner
export DB_USER=postgres
export DB_PASSWORD=your_password

go run cmd/statement-import-api/main.go
# Server will run on http://localhost:8080
```

### 2. Start the Frontend

```bash
cd frontend
npm run dev
# Frontend will run on http://localhost:3000
```

### 3. Authenticate

- Navigate to `http://localhost:3000/auth` (implement login as needed)
- Or use a valid JWT token in the `Authorization: Bearer <token>` header

---

## Test Scenarios

### Test 1: Upload Single Bank Statement (HDFC)

**Objective**: Verify upload and parsing of HDFC bank statement

**Steps**:
1. Navigate to `http://localhost:3000/statements/upload`
2. Select file: `backend/tests/testdata/hdfc_sample.csv`
3. Select bank: **HDFC Bank**
4. Click **Upload Statement**

**Expected Results**:
- ✅ Upload progress indicator shows 0-100%
- ✅ Request returns `202 Accepted` with statement_id
- ✅ Response includes: `status: PENDING`, `bank_code: HDFC`
- ✅ Redirect to preview page

**Verify Preview**:
- ✅ Shows 28 valid transactions (excluding opening/closing)
- ✅ Validation Summary shows:
  - Total rows: 29
  - Valid transactions: 28
  - Invalid transactions: 0
  - Extraction quality: ≥95%
- ✅ Period dates detected: July 1-29, 2026
- ✅ Transactions table displays all fields correctly:
  - Dates in DD Mon YYYY format
  - Amounts with ₹ symbol and thousand separators
  - Transaction types (DEBIT/CREDIT) color-coded
  - Descriptions truncated with merchant info

**Confirm Import**:
- ✅ Click "Confirm & Import"
- ✅ Response returns `200 OK` with `transactions_imported: 28`
- ✅ Redirect to success page
- ✅ Database contains 28 transaction records

**Test Duration**: ~15 seconds

---

### Test 2: Upload Different Bank Format (ICICI)

**Objective**: Verify parser flexibility with different column names

**Steps**:
1. Navigate to `http://localhost:3000/statements/upload`
2. Select file: `backend/tests/testdata/icici_sample.csv`
3. Select bank: **ICICI Bank**
4. Click **Upload Statement**

**Expected Results**:
- ✅ Same upload flow as HDFC
- ✅ Parser correctly identifies columns: "Txn Date", "Particulars", "Withdrawal", "Deposit"
- ✅ 28 transactions extracted despite different column naming

**Verify Preview**:
- ✅ All transactions show correctly mapped data
- ✅ Withdrawal amounts show as DEBIT
- ✅ Deposit amounts show as CREDIT
- ✅ Validation quality ≥95%

**Test Duration**: ~15 seconds

---

### Test 3: Upload Alternative Format (Axis)

**Objective**: Verify support for additional bank format variations

**Steps**:
1. Navigate to `http://localhost:3000/statements/upload`
2. Select file: `backend/tests/testdata/axis_sample.csv`
3. Select bank: **Axis Bank**
4. Click **Upload Statement**

**Expected Results**:
- ✅ Parser identifies "Value Date" column as date
- ✅ "Running Balance" column is processed correctly
- ✅ 28 transactions extracted

**Verify Preview**:
- ✅ Transactions display correctly
- ✅ Column name variations don't affect parsing
- ✅ Validation quality ≥95%

**Test Duration**: ~15 seconds

---

### Test 4: Duplicate Detection

**Objective**: Verify system prevents duplicate imports

**Steps**:
1. Upload `hdfc_sample.csv` first time (follow Test 1)
2. Confirm import (transactions persisted)
3. Upload same `hdfc_sample.csv` file again

**Expected Results**:
- ✅ First upload: `202 Accepted`
- ✅ Second upload: `409 Conflict` response with error message
- ✅ Error indicates: "Statement already imported" or similar
- ✅ File hash matching prevents duplicate

**Verify**:
- ✅ No duplicate transactions in database
- ✅ Only 28 transactions from first upload exist
- ✅ Second upload is rejected before processing

**Test Duration**: ~30 seconds

---

### Test 5: Multi-Bank Consolidation (US2 Preview)

**Objective**: Verify ability to upload and query multiple bank statements

**Steps**:
1. Upload all three sample files in sequence:
   - HDFC: `hdfc_sample.csv`
   - ICICI: `icici_sample.csv`
   - Axis: `axis_sample.csv`
2. Confirm each import
3. Navigate to `http://localhost:3000/transactions` (when implemented)

**Expected Results**:
- ✅ Three separate statements uploaded successfully
- ✅ Total 84 transactions in database (28 × 3)
- ✅ Each transaction correctly tagged with bank_code:
  - HDFC transactions: 28
  - ICICI transactions: 28
  - Axis transactions: 28

**Verify via API**:
```bash
# Get all transactions
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/transactions

# Filter by bank
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/transactions?bank_code=HDFC"
```

**Expected Response**:
```json
[
  {
    "id": "uuid",
    "statement_id": "statement-uuid",
    "date": "2026-07-02",
    "description": "Salary Deposit - ACME Corp",
    "amount": 75000.00,
    "type": "CREDIT",
    "merchant": "ACME Corp",
    "currency": "INR",
    "bank_code": "HDFC"
  },
  ...
]
```

**Test Duration**: ~1 minute

---

### Test 6: Date Range Filtering

**Objective**: Verify ability to filter transactions by date

**Steps**:
1. Ensure all three sample files are uploaded
2. Query API with date range filter:
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/transactions?date_start=2026-07-10&date_end=2026-07-20"
```

**Expected Results**:
- ✅ Returns only transactions between July 10-20
- ✅ Approximately 8-10 transactions per bank
- ✅ Approximately 24-30 total transactions across 3 banks

**Verify**:
- ✅ All returned transactions have dates within range
- ✅ No transactions outside range are included
- ✅ Bank filter can be combined: `?bank_code=HDFC&date_start=...&date_end=...`

**Test Duration**: ~10 seconds

---

### Test 7: Validation Error Handling

**Objective**: Verify system handles invalid data correctly

**Steps**:
1. Attempt upload with invalid file:
   - Wrong file extension (e.g., .txt)
   - Corrupted CSV (missing columns)
   - File exceeds size limit (>50MB)
2. Verify error responses

**Expected Results**:
- ✅ Invalid format: `400 Bad Request`
- ✅ Missing columns: Parser error, `400 Bad Request`
- ✅ File too large: `413 Payload Too Large`
- ✅ Error messages clearly indicate problem

**Test Duration**: ~20 seconds

---

### Test 8: Latency Verification (SC-001)

**Objective**: Verify upload-to-preview latency <10 seconds

**Steps**:
1. Record start time before upload
2. Upload HDFC sample
3. Record time when preview page loads
4. Calculate total time

**Measurement**:
- Start: Click "Upload Statement"
- End: Preview page fully loaded with all transactions
- Expected: <10 seconds

**Verify**:
- ✅ Upload + parsing + redirect: <10s
- ✅ Network latency acceptable
- ✅ No performance degradation with repeated uploads

**Test Duration**: ~2 minutes (with measurements)

---

### Test 9: Extraction Accuracy Verification (SC-002)

**Objective**: Verify extraction accuracy ≥95%

**Steps**:
1. Upload each sample file
2. Note extraction accuracy in validation summary
3. Compare extracted count vs. expected count:
   - HDFC: 28 valid / 29 total = 96.6% ✓
   - ICICI: 28 valid / 29 total = 96.6% ✓
   - Axis: 28 valid / 29 total = 96.6% ✓

**Expected Results**:
- ✅ All samples show ≥95% extraction accuracy
- ✅ Validation summary displays percentage
- ✅ Only opening/closing balance lines excluded

**Test Duration**: ~30 seconds

---

### Test 10: Full User Journey

**Objective**: End-to-end user workflow test

**Steps**:
1. Fresh session (new user)
2. Authenticate
3. Navigate to upload page
4. Upload HDFC sample
5. Review preview
6. Confirm import
7. Verify transaction list shows imported data
8. Apply filters (date range, bank)
9. Verify transaction details

**Expected Results**:
- ✅ All steps complete without errors
- ✅ No console errors or warnings
- ✅ Data consistency maintained throughout
- ✅ Responsive UI (no hanging/loading)

**Test Duration**: ~3 minutes

---

## API Testing with cURL

### Upload Statement

```bash
curl -X POST http://localhost:8080/api/statements/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@backend/tests/testdata/hdfc_sample.csv" \
  -F "bank_code=HDFC"

# Response: 202 Accepted
# {
#   "statement_id": "550e8400-e29b-41d4-a716-446655440000",
#   "status": "PENDING",
#   "bank_code": "HDFC",
#   "file_name": "hdfc_sample.csv",
#   "file_format": "CSV",
#   "uploaded_at": "2026-07-05T12:00:00Z"
# }
```

### Get Preview

```bash
curl http://localhost:8080/api/statements/550e8400-e29b-41d4-a716-446655440000/preview \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response: 200 OK
# {
#   "statement_id": "550e8400-e29b-41d4-a716-446655440000",
#   "transactions": [...],
#   "validation_summary": {
#     "total_rows": 29,
#     "valid_transactions": 28,
#     "invalid_transactions": 0,
#     "errors": [],
#     "period_start": "2026-07-01",
#     "period_end": "2026-07-29"
#   }
# }
```

### Confirm Import

```bash
curl -X POST http://localhost:8080/api/statements/550e8400-e29b-41d4-a716-446655440000/confirm \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"confirmed": true}'

# Response: 200 OK
# {
#   "statement_id": "550e8400-e29b-41d4-a716-446655440000",
#   "status": "SUCCESS",
#   "transactions_imported": 28,
#   "message": "Statement imported successfully"
# }
```

### List Transactions

```bash
curl "http://localhost:8080/api/transactions?limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response: 200 OK with transaction array
```

---

## Test Data Summary

| Sample | Transactions | Period | Amount Range | Coverage |
|--------|--------------|--------|--------------|----------|
| HDFC | 28 | Jul 1-29 | ₹150 - ₹75,000 | 100% |
| ICICI | 28 | Jul 1-29 | ₹250 - ₹80,000 | 100% |
| Axis | 28 | Jul 1-29 | ₹200 - ₹70,000 | 100% |

---

## Troubleshooting

### Issue: "Invalid file format" error
- **Cause**: File extension not CSV
- **Solution**: Rename to `.csv` or verify MIME type

### Issue: "Column not found" error
- **Cause**: CSV doesn't match bank format
- **Solution**: Verify column names match expected format

### Issue: Preview shows 0 transactions
- **Cause**: Parser couldn't extract data
- **Solution**: Check CSV structure, verify date format

### Issue: "Statement already imported" on first upload
- **Cause**: File hash matches existing import
- **Solution**: Verify file is new, check database for duplicates

### Issue: Latency >10 seconds
- **Cause**: Database/network performance
- **Solution**: Check database indexes, network latency

---

## Success Criteria

- ✅ All 10 test scenarios pass
- ✅ Extraction accuracy ≥95% on all samples
- ✅ Upload-to-preview latency <10 seconds
- ✅ No duplicate imports
- ✅ Date range filtering works correctly
- ✅ Multi-bank consolidation functional
- ✅ Error handling appropriate for edge cases
- ✅ No console errors or warnings
- ✅ Database consistency maintained

---

## Next Steps

1. Document any issues or unexpected behavior
2. Update test data as new formats discovered
3. Create PDF sample files for PDF parser testing
4. Add Excel format samples for ExcelParser testing
5. Create edge case samples (large files, special characters, etc.)

---

## Contacts & Support

- Code: `backend/tests/testdata/`
- Documentation: `backend/tests/testdata/README.md`
- Issues: Check backend logs and test output
- Contributors: Add new sample files and update this guide
