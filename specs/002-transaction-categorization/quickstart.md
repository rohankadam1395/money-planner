# Quickstart: Transaction Categorization

**Feature**: Transaction Categorization (Phase 2) | **Created**: 2026-07-12

Validation scenarios that prove the transaction categorization feature works end-to-end. Each scenario covers one user story and includes prerequisites, setup, test commands, and expected outcomes.

---

## Scenario 1: Rule-Based Categorization (US1)

**User Story**: A user uploads a bank statement and transactions are automatically categorized by merchant dictionary.

### Prerequisites

1. **Backend running**: `go run ./backend/cmd/statement-import-api/main.go`
2. **Merchant dictionary seeded**: Run migration `db/migrations/xxx_seed_merchant_dictionary.sql`
3. **Sample statement file**: `test-data/sample_hdfc_statement.csv` (provided in repo)
4. **Valid JWT token**: Generate via `/auth/login` endpoint

### Setup Steps

```bash
# 1. Create test user and get JWT token
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' \
  > response.json

JWT_TOKEN=$(jq -r '.token' response.json)
echo "JWT_TOKEN=$JWT_TOKEN"

# 2. Upload statement (uses Phase 1 import feature)
curl -X POST http://localhost:8080/api/v1/statements/upload \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "file=@test-data/sample_hdfc_statement.csv" \
  > upload_response.json

STATEMENT_ID=$(jq -r '.statement_id' upload_response.json)
echo "STATEMENT_ID=$STATEMENT_ID"

# 3. Get preview with categorization
curl -X GET http://localhost:8080/api/v1/statements/$STATEMENT_ID/preview \
  -H "Authorization: Bearer $JWT_TOKEN" \
  > preview_response.json

cat preview_response.json | jq '.'
```

### Test Commands

#### Test 1.1: Known Merchant Categorization

```bash
# Check that Swiggy transaction is categorized as "Food"
cat preview_response.json | jq '.transactions[] | select(.merchant_name | contains("Swiggy")) | {merchant_name, category_name, method, confidence}'

# Expected output:
# {
#   "merchant_name": "Swiggy Food Delivery",
#   "category_name": "Food",
#   "method": "rule_based",
#   "confidence": 1.0
# }
```

#### Test 1.2: Multiple Categories

```bash
# Show all categories assigned in preview
cat preview_response.json | jq '.transactions[] | {merchant_name, category_name}' | head -20

# Expected: Mix of Food, Shopping, Transport, Housing, etc.
```

#### Test 1.3: Persist Categorization

```bash
# Confirm import (stores categorization to database)
curl -X POST http://localhost:8080/api/v1/statements/$STATEMENT_ID/confirm \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"confirmed":true}' \
  > confirm_response.json

# Verify transactions now have persistent categories
TRANSACTION_ID=$(jq -r '.transactions[0].id' confirm_response.json)

curl -X GET http://localhost:8080/api/v1/transactions/$TRANSACTION_ID \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.category_id, .category_name, .categorization_method'

# Expected: category_id, category_name, categorization_method populated
```

### Expected Outcomes

- ✅ All transactions with known merchants (Swiggy, Amazon, Uber) display correct categories
- ✅ Categorization confidence is 1.0 for rule-based matches
- ✅ Categories persist to database after import confirmation
- ✅ Preview displays categories before user confirms (allows review)

---

## Scenario 2: LLM Categorization for Unknown Merchants (US2)

**User Story**: A transaction with an unknown merchant shows LLM-suggested category (via Ollama or configured provider) during preview.

### Prerequisites

1. **Backend running with LLM support**: `go run ./backend/cmd/statement-import-api/main.go`
2. **Ollama running locally** (default provider):
   ```bash
   # Install Ollama: https://ollama.ai
   # Download Mistral 7B model (first time, ~4GB download):
   ollama pull mistral:7b
   # Start Ollama server (runs on http://localhost:11434):
   ollama serve
   ```
   **Alternative: Use Claude API** by setting `LLM_PROVIDER=claude` and `ANTHROPIC_API_KEY=sk-...`

3. **Sample statement with unknown merchant**: `test-data/sample_unknown_merchants.csv`
4. **LLM config** at `backend/config/llm-config.yaml` (created during setup):
   ```yaml
   llm:
     default_provider: "ollama"
     providers:
       ollama:
         enabled: true
         model: "mistral:7b"
         base_url: "http://localhost:11434"
         timeout_seconds: 60
       claude:
         enabled: false
   ```

### Setup Steps

```bash
# 1. Verify Ollama is running and model is available
curl http://localhost:11434/api/generate -d '{"model":"mistral:7b","prompt":"test"}' 2>/dev/null | head -20
# Expected: JSON response with model response

# 2. Verify backend can access Ollama
grep -r "OLLAMA_BASE_URL\|LLM_PROVIDER" backend/config/ || echo "Using defaults"

# 3. Reuse JWT token from Scenario 1
JWT_TOKEN="<from scenario 1>"

# 4. Upload statement with unknown merchants
curl -X POST http://localhost:8080/api/v1/statements/upload \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "file=@test-data/sample_unknown_merchants.csv" \
  > upload_response.json

STATEMENT_ID=$(jq -r '.statement_id' upload_response.json)

# 5. Get preview with LLM categorization enabled
# (Ollama will be used automatically since it's the default provider)
curl -X GET http://localhost:8080/api/v1/statements/$STATEMENT_ID/preview \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Include-LLM-Categorization: true" \
  > preview_response.json

cat preview_response.json | jq '.'
```

### Test Commands

#### Test 2.1: LLM-Suggested Categories (Ollama)

```bash
# Find transactions with LLM method and provider
cat preview_response.json | jq '.transactions[] | select(.method == "llm") | {merchant_name, category_name, llm_provider, confidence, explanation}'

# Expected output (with Ollama):
# {
#   "merchant_name": "Aashish Restaurant Pvt Ltd",
#   "category_name": "Food",
#   "llm_provider": "ollama",
#   "confidence": 0.80,
#   "explanation": "Ollama (Mistral 7B) inferred category based on merchant name"
# }

# Note: Confidence with Ollama is typically 0.65-0.80 (fixed score);
#       with Claude would be 0.85-0.95 (higher accuracy)
```

#### Test 2.2: Low Confidence Flagging

```bash
# Find transactions with low LLM confidence (< 75%)
cat preview_response.json | jq '.transactions[] | select(.method == "llm" and .confidence < 0.75) | {merchant_name, category_name, confidence}'

# Expected: Some transactions flagged for review if confidence < 75%
```

#### Test 2.3: User Override Before Confirm

```bash
# User can override LLM suggestion before confirming
curl -X POST http://localhost:8080/api/v1/statements/$STATEMENT_ID/transactions/override \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "<unknown_merchant_txn_id>",
    "category_id": "cat_food",
    "notes": "Actually confirmed as Food"
  }' \
  > override_response.json

# Verify override applied in new preview
curl -X GET http://localhost:8080/api/v1/statements/$STATEMENT_ID/preview \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.transactions[] | select(.id == "<unknown_merchant_txn_id>") | {category_name, method}'

# Expected: category updated, method might show "manual" after override
```

#### Test 2.4: Confirm Import with LLM Corrections

```bash
# Confirm import (saves LLM categorizations and user overrides)
curl -X POST http://localhost:8080/api/v1/statements/$STATEMENT_ID/confirm \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"confirmed":true}' \
  > confirm_response.json

jq '.transactions_count, .categorized_count, .llm_categorized_count' confirm_response.json

# Expected: All transactions categorized, some via LLM
```

### Expected Outcomes

- ✅ Unknown merchants trigger LLM categorization (via Ollama or configured provider)
- ✅ LLM returns category with confidence score (Ollama: 0.65-0.80, Claude: 0.85-0.95)
- ✅ LLM provider tracked in response (`llm_provider`: "ollama", "claude", etc.)
- ✅ Low confidence (<75%) transactions are flagged for review
- ✅ User can override LLM suggestions before confirming
- ✅ LLM suggestions persist after import confirmation
- ✅ **Switching providers**: Change `LLM_PROVIDER=claude` env var and restart backend to use Claude instead (no other code changes needed)

---

## Scenario 3: Category Analytics & Recategorization (US3)

**User Story**: User views spending by category and can recategorize transactions post-import.

### Prerequisites

1. **Transactions imported**: From Scenarios 1 & 2 (or fresh import)
2. **Category tables populated**: categories, transaction_categories, category_stats
3. **Valid JWT token**

### Setup Steps

```bash
# 1. Ensure transactions from earlier scenarios are confirmed and categorized
JWT_TOKEN="<from scenario 1>"

# 2. Query category dashboard
curl -X GET "http://localhost:8080/api/v1/categories?include_stats=true&period=2026-07" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  > categories_response.json

cat categories_response.json | jq '.'
```

### Test Commands

#### Test 3.1: View Categories with Spending Totals

```bash
# Get all categories with spending stats for July 2026
curl -X GET "http://localhost:8080/api/v1/categories?include_stats=true&period=2026-07" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.categories[] | {name, stats: {total_spent, transaction_count, average_transaction}}'

# Expected output:
# {
#   "name": "Food",
#   "stats": {
#     "total_spent": 4500.50,
#     "transaction_count": 24,
#     "average_transaction": 187.52
#   }
# }
```

#### Test 3.2: Drill Down to Category Detail

```bash
# Get all transactions in "Food" category for July
curl -X GET "http://localhost:8080/api/v1/categories/cat_food/transactions?period_start=2026-07-01&period_end=2026-07-31&limit=10" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.transactions[] | {merchant_name, amount, date, categorization_method}'

# Expected: List of Food transactions with amounts and dates
```

#### Test 3.3: Recategorize a Transaction

```bash
# User recategorizes a transaction (e.g., misclassified Starbucks from Miscellaneous to Food)
TRANSACTION_ID=$(curl -X GET "http://localhost:8080/api/v1/categories/cat_miscellaneous/transactions?limit=1" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq -r '.transactions[0].transaction_id')

curl -X POST "http://localhost:8080/api/v1/transactions/$TRANSACTION_ID/recategorize" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "cat_food",
    "learn_correction": true,
    "notes": "Coffee shop should be Food, not Miscellaneous"
  }' \
  > recategorize_response.json

cat recategorize_response.json | jq '{old_category: .old_category.category_name, new_category: .new_category.category_name, learned: .learned}'

# Expected:
# {
#   "old_category": "Miscellaneous",
#   "new_category": "Food",
#   "learned": true
# }
```

#### Test 3.4: Verify Category Stats Updated

```bash
# Check that category stats updated after recategorization
# Miscellaneous total_spent decreased, Food total_spent increased

curl -X GET "http://localhost:8080/api/v1/categories/cat_food/stats?period=2026-07" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.stats | {total_spent, transaction_count}'

# Expected: totals updated to reflect recategorized transaction
```

#### Test 3.5: Filter Transactions by Confidence

```bash
# Show low-confidence categorizations for manual review
curl -X GET "http://localhost:8080/api/v1/categories/cat_uncategorized/transactions?min_confidence=0&sort_by=confidence_asc" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  | jq '.transactions[] | {merchant_name, categorization_method, confidence}'

# Expected: Transactions sorted by low confidence first
```

### Expected Outcomes

- ✅ Category dashboard displays total spending by category
- ✅ User can drill down to see all transactions in a category
- ✅ Transactions can be recategorized with instant feedback
- ✅ Category statistics update after recategorization
- ✅ Learning flag enables merchant dictionary updates (see new merchant mappings in future)

---

## Validation Checklist

Use this checklist to verify the feature is complete:

- [ ] **US1 Complete**: Rule-based categorization works for known merchants
  - [ ] Transactions display categories during preview
  - [ ] Categories persist after import confirmation
  - [ ] Confidence = 1.0 for exact matches

- [ ] **US2 Complete**: LLM categorization works for unknown merchants
  - [ ] LLM is called only for unknown merchants (rule-based first)
  - [ ] LLM confidence scores returned (0.6-0.95 range)
  - [ ] User can override LLM suggestions before confirming
  - [ ] Low confidence (<75%) flagged for review

- [ ] **US3 Complete**: Category analytics and recategorization work
  - [ ] Category dashboard shows spending totals per category
  - [ ] Users can drill down to category transactions
  - [ ] Recategorization updates are reflected instantly
  - [ ] Category stats update after recategorization
  - [ ] Learning flag captures user corrections for dictionary

---

## Troubleshooting

### Categorization returning "Uncategorized" for known merchants

**Cause**: Merchant dictionary not seeded or merchant name doesn't match exactly.

**Solution**:
1. Check merchant_dictionary table: `SELECT * FROM merchant_dictionary LIMIT 10;`
2. Add missing merchants: `INSERT INTO merchant_dictionary (merchant_name, category_id) VALUES ('Swiggy', cat_id);`
3. Re-import statement

### Ollama Timeout or Connection Error

**Cause**: Ollama not running, model not loaded, or misconfigured base URL.

**Solution**:
1. Check if Ollama is running: `curl http://localhost:11434/api/tags 2>/dev/null | jq '.models[] | .name'`
2. If not running: `ollama serve` in another terminal
3. Check if Mistral model is loaded: `ollama pull mistral:7b`
4. Verify config: `cat backend/config/llm-config.yaml | grep -A3 "ollama:"`
5. Override base URL if needed: `export OLLAMA_BASE_URL=http://your-ollama-host:11434`
6. Check backend logs: `grep -i "ollama\|llm_provider" ./backend/logs/*.log`
7. Transactions should default to "Uncategorized" without blocking import (graceful degradation)

### LLM API Timeout (Claude)

**Cause**: Claude API slow or network issue.

**Solution**:
1. Set provider: `export LLM_PROVIDER=claude`
2. Check API key: `echo $ANTHROPIC_API_KEY`
3. Test API directly: `curl -X GET https://api.anthropic.com/v1/models -H "Authorization: Bearer $ANTHROPIC_API_KEY"`
4. Check backend logs for retry attempts
5. Transactions should default to "Uncategorized" without blocking import

### Switching LLM Providers

**To switch from Ollama to Claude**:
```bash
# 1. Set environment variable
export LLM_PROVIDER=claude
export ANTHROPIC_API_KEY=sk-...

# 2. Restart backend
# (Config file llm-config.yaml already has Claude provider defined; env var selects it at runtime)

# 3. Test: Upload statement, verify response includes "llm_provider": "claude"
```

**To switch back to Ollama**:
```bash
export LLM_PROVIDER=ollama
# Restart backend
```

### Category stats not updating after recategorization

**Cause**: Category stats update is async or database trigger failed.

**Solution**:
1. Check category_stats table: `SELECT * FROM category_stats WHERE user_id=? AND period='2026-07';`
2. Manual recalc: Run `backend/scripts/recalc_category_stats.sh`
3. Verify transaction_categories record exists: `SELECT * FROM transaction_categories WHERE transaction_id=?;`

---

## Next Steps

After validating this quickstart:

1. **Run full test suite**: `go test ./... && npm run test:frontend`
2. **Load test**: Categorize 10,000 transactions, measure latency
3. **User acceptance test**: Have domain expert review categorization accuracy
4. **Deploy to staging**: Run feature with real bank data
5. **Monitor**: Track categorization accuracy, LLM cost, cache hit rates
