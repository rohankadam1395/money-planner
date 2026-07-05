# Quickstart: Statement Import Validation

**Date**: 2026-07-05

**Purpose**: Runnable validation scenarios that prove Statement Import feature works end-to-end.

---

## Prerequisites

### Backend Setup

1. **Start PostgreSQL**:
   ```bash
   docker run -d \
     -e POSTGRES_PASSWORD=testpass \
     -e POSTGRES_DB=moneyplan \
     -p 5432:5432 \
     postgres:14
   ```

2. **Apply migrations**:
   ```bash
   cd backend
   go run ./cmd/migrate statements
   ```
   Migrations create: `statements`, `transactions`, `import_jobs` tables.

3. **Start backend API server**:
   ```bash
   cd backend
   go run ./cmd/statement-import-api \
     -port 8080 \
     -db "postgres://user:testpass@localhost/moneyplan"
   ```
   API listens on `http://localhost:8080`.

### Frontend Setup

1. **Install dependencies**:
   ```bash
   cd frontend
   npm install
   ```

2. **Start dev server**:
   ```bash
   npm run dev
   ```
   Frontend available on `http://localhost:3000`.

### Test Data

1. **Get test bank statements**:
   - Sample HDFC PDF: `test-data/hdfc-sample.pdf`
   - Sample ICICI CSV: `test-data/icici-sample.csv`
   - Sample Excel file: `test-data/axis-sample.xlsx`

2. **Create test user** (if needed):
   ```bash
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "email": "test@example.com",
       "password": "testpass123"
     }'
   ```

3. **Get JWT token**:
   ```bash
   export JWT=$(curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "testpass123"}' \
     | jq -r '.token')
   echo $JWT
   ```

---

## Validation Scenario 1: Upload & Preview HDFC PDF

**Goal**: Upload a valid HDFC bank statement PDF, preview extracted transactions.

### Steps

1. **Upload PDF file**:
   ```bash
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@test-data/hdfc-sample.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=d4a2b7f4c1e3a8d2b5f7c9e1a3d5b7f9" \
     | jq .
   ```

2. **Expected response**:
   - Status: `202 Accepted`
   - Contains: `statement_id`, `import_job_id`, `status=PENDING`

3. **Wait for extraction** (2-5 seconds):
   ```bash
   sleep 3
   
   export STATEMENT_ID="<statement_id from response>"
   export IMPORT_JOB_ID="<import_job_id from response>"
   ```

4. **Get preview**:
   ```bash
   curl -X GET http://localhost:8080/api/statements/$STATEMENT_ID/preview \
     -H "Authorization: Bearer $JWT" \
     | jq .
   ```

5. **Expected preview response**:
   - `status`: PENDING (not yet confirmed) or PROCESSING (extraction in progress)
   - `transactions[]`: Array of extracted transactions with fields: date, merchant, amount, type, balance
   - `validation_summary`: Shows valid_rows, invalid_rows, error_rate_percent
   - All transactions should have `validation_status: VALID` for test file

6. **Verify extracted data**:
   - Check transaction dates are within statement period
   - Verify merchant names are non-empty
   - Confirm amounts are positive numbers
   - Check balance progression is logical (increasing with credits, decreasing with debits)

### Expected Outcome

- Preview shows ≥40 transactions (typical HDFC 1-month statement)
- 100% valid transactions (0 errors)
- First transaction date ≈ 2026-06-01
- Last transaction date ≈ 2026-06-30
- Validation summary shows error_rate_percent = 0

---

## Validation Scenario 2: Preview and Confirm Import

**Goal**: Confirm preview and persist transactions to database.

### Steps

1. **Confirm import** (from Scenario 1):
   ```bash
   curl -X POST http://localhost:8080/api/statements/$STATEMENT_ID/confirm \
     -H "Authorization: Bearer $JWT" \
     -H "Content-Type: application/json" \
     -d '{"auto_categorize": false}' \
     | jq .
   ```

2. **Expected response**:
   - Status: `200 OK`
   - `status`: SUCCESS
   - `transactions_imported`: ≥40
   - `transactions_failed`: 0
   - `message`: "Statement import completed successfully. X transactions saved."

3. **Query imported transactions** (verify in database):
   ```bash
   curl -X GET "http://localhost:8080/api/transactions?statement_id=$STATEMENT_ID&limit=10" \
     -H "Authorization: Bearer $JWT" \
     | jq .
   ```

4. **Expected transaction response**:
   ```json
   {
     "total": 45,
     "transactions": [
       {
         "transaction_id": "uuid",
         "transaction_date": "2026-06-01",
         "merchant": "SALARY CREDIT",
         "amount": 150000.00,
         "type": "CREDIT",
         "balance": 250000.00,
         "imported_at": "2026-07-05T10:35:00Z"
       },
       ...
     ]
   }
   ```

### Expected Outcome

- Transactions persisted to database
- Can retrieve all transactions via transactions API
- Transaction count matches import_job response
- No duplicate transactions in database

---

## Validation Scenario 3: Duplicate Detection

**Goal**: Verify system prevents re-importing the same statement.

### Steps

1. **Try uploading the same file again**:
   ```bash
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@test-data/hdfc-sample.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=d4a2b7f4c1e3a8d2b5f7c9e1a3d5b7f9" \
     | jq .
   ```

2. **Expected response**:
   - Status: `409 Conflict`
   - `error`: DUPLICATE_FILE
   - `existing_statement_id`: <id from first upload>
   - `message`: "Statement file already imported on 2026-07-05"

### Expected Outcome

- System detects duplicate file
- User prevented from importing same statement twice
- Original statement remains unchanged

---

## Validation Scenario 4: Multi-Bank Import (ICICI CSV)

**Goal**: Upload statement from different bank (CSV format), verify multi-bank support.

### Steps

1. **Upload ICICI CSV**:
   ```bash
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@test-data/icici-sample.csv" \
     -F "bank_code=ICIC" \
     -F "account_number_hash=a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6" \
     | jq .
   ```

2. **Expected response**:
   - Status: `202 Accepted`
   - Different `statement_id` from HDFC import

3. **Preview ICICI statement**:
   ```bash
   export ICICI_STATEMENT_ID="<statement_id from response>"
   curl -X GET http://localhost:8080/api/statements/$ICICI_STATEMENT_ID/preview \
     -H "Authorization: Bearer $JWT" \
     | jq .
   ```

4. **Verify ICICI data**:
   - Transactions extracted from CSV (different format than PDF)
   - Field mapping correct for ICICI layout

5. **Confirm ICICI import**:
   ```bash
   curl -X POST http://localhost:8080/api/statements/$ICICI_STATEMENT_ID/confirm \
     -H "Authorization: Bearer $JWT" \
     -H "Content-Type: application/json" \
     -d '{"auto_categorize": false}'
   ```

6. **Query all transactions across banks**:
   ```bash
   curl -X GET "http://localhost:8080/api/transactions?limit=100" \
     -H "Authorization: Bearer $JWT" \
     | jq '.transactions | length'
   ```
   Should show: HDFC transactions + ICICI transactions

### Expected Outcome

- ICICI statement imported successfully from CSV
- CSV parsing works (different field order than HDFC PDF)
- Multi-bank queries work (can retrieve transactions from both banks together)
- No cross-bank duplicate issues

---

## Validation Scenario 5: Error Handling - Corrupted File

**Goal**: Verify system gracefully rejects invalid files.

### Steps

1. **Upload corrupted file**:
   ```bash
   echo "This is not a PDF" > /tmp/fake.pdf
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@/tmp/fake.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=d4a2b7f4c1e3a8d2b5f7c9e1a3d5b7f9" \
     | jq .
   ```

2. **Expected response**:
   - Status: `400 Bad Request`
   - `error`: INVALID_FILE_FORMAT
   - `message`: "File must be PDF, CSV, or XLSX"

### Expected Outcome

- System rejects invalid file format
- Clear error message to user
- No partial statement created in database

---

## Validation Scenario 6: File Size Limit

**Goal**: Verify 50MB file size limit.

### Steps

1. **Create a file >50MB**:
   ```bash
   dd if=/dev/zero of=/tmp/large.pdf bs=1M count=51
   ```

2. **Try to upload**:
   ```bash
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@/tmp/large.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=d4a2b7f4c1e3a8d2b5f7c9e1a3d5b7f9" \
     | jq .
   ```

3. **Expected response**:
   - Status: `413 Payload Too Large`
   - `error`: FILE_TOO_LARGE
   - `message`: "Maximum file size is 50MB"

### Expected Outcome

- File size check enforced
- User prevented from uploading oversized statements
- Clear error message

---

## Validation Scenario 7: Extraction Accuracy (95% Target)

**Goal**: Verify extraction accuracy meets ≥95% threshold.

### Steps

1. **Upload test statement with known row count**:
   ```bash
   # test-data/accuracy-test.pdf has 50 transactions (manually verified)
   curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@test-data/accuracy-test.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=testaccounthash123" \
     | jq -r '.statement_id' > /tmp/accuracy_statement_id.txt
   
   export ACCURACY_STATEMENT_ID=$(cat /tmp/accuracy_statement_id.txt)
   ```

2. **Preview and check extraction count**:
   ```bash
   curl -X GET http://localhost:8080/api/statements/$ACCURACY_STATEMENT_ID/preview \
     -H "Authorization: Bearer $JWT" \
     | jq '.total_transactions'
   ```

3. **Expected**:
   - `total_transactions`: 50 (matches manual count)
   - `validation_summary.valid_rows`: ≥47.5 (95% of 50)
   - `validation_summary.error_rate_percent`: ≤5

### Expected Outcome

- Extraction accuracy ≥95%
- Meets success criterion SC-002

---

## Performance Validation

### Scenario 8: Upload-to-Preview Latency (<10 seconds)

1. **Time the upload and preview**:
   ```bash
   time curl -X POST http://localhost:8080/api/statements/upload \
     -H "Authorization: Bearer $JWT" \
     -F "file=@test-data/hdfc-sample.pdf" \
     -F "bank_code=HDFC" \
     -F "account_number_hash=d4a2b7f4c1e3a8d2b5f7c9e1a3d5b7f9" \
     2>&1 | tee /tmp/upload_response.json
   
   export STATEMENT_ID=$(jq -r '.statement_id' /tmp/upload_response.json)
   
   # Wait for processing
   sleep 2
   
   time curl -X GET http://localhost:8080/api/statements/$STATEMENT_ID/preview \
     -H "Authorization: Bearer $JWT" > /dev/null
   ```

2. **Expected**:
   - Upload response: <1 second (202 Accepted)
   - Preview ready: <10 seconds total from upload
   - Meets success criterion SC-001

---

## Cleanup

After validation:

```bash
# Stop containers
docker-compose down

# Clean up test databases
rm -rf /tmp/moneyplan_test_db
```

---

## Success Criteria Checklist

After completing all scenarios, verify:

| Scenario | Criterion | Status |
|----------|-----------|--------|
| Scenario 1 | SC-001: Upload to preview <10s | ✓ |
| Scenario 2 | SC-002: ≥95% extraction accuracy | ✓ |
| Scenario 3 | FR-007: Duplicate detection works | ✓ |
| Scenario 4 | FR-002: Multi-bank support (CSV) | ✓ |
| Scenario 5 | FR-004: Validation rejects bad files | ✓ |
| Scenario 6 | FR-002: File size limit enforced | ✓ |
| Scenario 7 | SC-002: 95% accuracy verified | ✓ |
| Scenario 8 | SC-001: Performance target met | ✓ |

All scenarios must pass before feature is production-ready.

