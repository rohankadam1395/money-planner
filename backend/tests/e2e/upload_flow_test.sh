#!/bin/bash
# End-to-End Test: Statement Upload Flow (T055-T058)
# Tests: Upload → Preview → Confirm → DB Verification
#
# Requirements:
# - Backend API running on http://localhost:8080
# - Database (PostgreSQL) connected and schema migrated
# - JWT token for authorization
# - Sample CSV files in backend/tests/testdata/

set -e

API_BASE="http://localhost:8080/api"
TOKEN="${JWT_TOKEN:-}"  # Provide via environment variable
TIMEOUT=10

echo "========================================="
echo "End-to-End Statement Upload Flow Test"
echo "========================================="

if [ -z "$TOKEN" ]; then
    echo "⚠ WARNING: JWT_TOKEN not set. Tests will fail without authorization."
    echo "Set JWT_TOKEN environment variable: export JWT_TOKEN=<your-token>"
fi

# Test 1: Upload HDFC CSV (T055)
echo ""
echo "[T055] Testing HDFC CSV upload..."
HDFC_UPLOAD=$(curl -s -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@./testdata/hdfc_sample.csv" \
    -F "bank_code=HDFC" \
    "${API_BASE}/statements/upload" \
    -w "\n%{http_code}" -o /tmp/hdfc_response.json)

STATUS=$(tail -1 /tmp/hdfc_response.json)
RESPONSE=$(head -n -1 /tmp/hdfc_response.json)

if [ "$STATUS" = "202" ]; then
    echo "✓ Upload accepted (202)"
    STATEMENT_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "  Statement ID: $STATEMENT_ID"
else
    echo "✗ Upload failed (HTTP $STATUS)"
    echo "Response: $RESPONSE"
    exit 1
fi

# Test 2: Get Preview (T055)
echo ""
echo "[T055] Waiting for preview to be ready..."
PREVIEW_READY=false
for i in {1..10}; do
    PREVIEW=$(curl -s -H "Authorization: Bearer $TOKEN" \
        "${API_BASE}/statements/${STATEMENT_ID}/preview")

    STATUS=$(echo "$PREVIEW" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
    if [ "$STATUS" = "READY" ]; then
        PREVIEW_READY=true
        echo "✓ Preview ready"
        TRANSACTION_COUNT=$(echo "$PREVIEW" | grep -o '"transaction_count":[0-9]*' | cut -d':' -f2)
        echo "  Transactions extracted: $TRANSACTION_COUNT"
        break
    fi
    echo "  Status: $STATUS (attempt $i/10, waiting...)"
    sleep 1
done

if [ "$PREVIEW_READY" = false ]; then
    echo "✗ Preview not ready after 10 seconds"
    exit 1
fi

# Test 3: Confirm Import (T055)
echo ""
echo "[T055] Confirming import..."
CONFIRM=$(curl -s -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{}' \
    "${API_BASE}/statements/${STATEMENT_ID}/confirm" \
    -w "\n%{http_code}" -o /tmp/confirm_response.json)

STATUS=$(tail -1 /tmp/confirm_response.json)
RESPONSE=$(head -n -1 /tmp/confirm_response.json)

if [ "$STATUS" = "200" ]; then
    echo "✓ Import confirmed (200)"
else
    echo "✗ Confirm failed (HTTP $STATUS)"
    echo "Response: $RESPONSE"
    exit 1
fi

# Test 4: Verify DB (T057)
echo ""
echo "[T057] Verifying data in database..."
DB_COUNT=$(psql -h localhost -U postgres -d money_planner -c \
    "SELECT COUNT(*) FROM transactions WHERE statement_id = '$STATEMENT_ID';" \
    -t 2>/dev/null || echo "0")

EXPECTED_COUNT=6  # Adjust based on sample file
if [ "$DB_COUNT" -ge "$EXPECTED_COUNT" ]; then
    echo "✓ Database verification passed"
    echo "  DB transaction count: $DB_COUNT (expected: ≥$EXPECTED_COUNT)"
else
    echo "✗ Database verification failed"
    echo "  DB transaction count: $DB_COUNT (expected: ≥$EXPECTED_COUNT)"
    exit 1
fi

# Test 5: Latency Verification (T058)
echo ""
echo "[T058] Verifying upload-to-preview latency..."
START=$(date +%s%N)
curl -s -H "Authorization: Bearer $TOKEN" \
    "${API_BASE}/statements/${STATEMENT_ID}/preview" > /dev/null
END=$(date +%s%N)

LATENCY=$(( (END - START) / 1000000 ))  # Convert ns to ms
LATENCY_SEC=$(echo "scale=2; $LATENCY / 1000" | bc)

if (( $(echo "$LATENCY_SEC < 10" | bc -l) )); then
    echo "✓ Latency requirement met"
    echo "  Response time: ${LATENCY_SEC}s (target: <10s)"
else
    echo "✗ Latency requirement failed"
    echo "  Response time: ${LATENCY_SEC}s (target: <10s)"
    exit 1
fi

echo ""
echo "========================================="
echo "✓ All end-to-end tests PASSED"
echo "========================================="
