# Contract: Categorize Transactions Endpoint

**Feature**: Transaction Categorization | **Phase**: Phase 1 (US1: Rule-Based Categorization)

## Endpoint

```
POST /api/v1/transactions/categorize
```

## Purpose

Categorize transactions during statement import preview. Applies rule-based and optionally LLM categorization (via configurable provider: Ollama, Claude, OpenAI, etc.) to determine category assignments before user confirms import.

**LLM Provider Selection**: The active LLM provider is determined at server startup via configuration (`config/llm-config.yaml` + `LLM_PROVIDER` env var). Default: Ollama (local). No per-request provider selection; all transactions use the configured provider.

## Request

### Headers
```
Content-Type: application/json
Authorization: Bearer {jwt_token}
```

### Body

```json
{
  "transactions": [
    {
      "id": "txn_001",
      "merchant_name": "Swiggy Food Deliv",
      "amount": 450.50,
      "date": "2026-07-10T14:30:00Z",
      "description": "SWIGGY FOOD DELIV"
    },
    {
      "id": "txn_002",
      "merchant_name": "Unknown Small Vendor",
      "amount": 250.00,
      "date": "2026-07-11T10:15:00Z",
      "description": null
    }
  ],
  "include_llm_categorization": true  // optional, default: false for MVP
}
```

### Validation
- `transactions` array required, 1-1000 items
- Each transaction must have: `id`, `merchant_name`, `amount`
- `date` and `description` optional
- `merchant_name` required, non-empty string
- `amount` must be > 0

## Response

### Success (200 OK)

```json
{
  "categorizations": [
    {
      "transaction_id": "txn_001",
      "category_id": "cat_food",
      "category_name": "Food",
      "method": "rule_based",
      "llm_provider": null,
      "confidence": 1.0,
      "explanation": "Matched known merchant 'Swiggy'"
    },
    {
      "transaction_id": "txn_002",
      "category_id": "cat_food",
      "category_name": "Food",
      "method": "llm",
      "llm_provider": "ollama",
      "confidence": 0.80,
      "explanation": "Ollama (Mistral 7B) inferred category based on merchant name"
    }
  ],
  "stats": {
    "total": 2,
    "rule_based": 1,
    "llm": 1,
    "uncategorized": 0,
    "llm_providers": {
      "ollama": 1,
      "claude": 0
    },
    "processing_time_ms": 280
  }
}
```

### Partial Failure (206 Partial Content)

If some transactions fail to categorize (e.g., LLM timeout on subset):

```json
{
  "categorizations": [
    {
      "transaction_id": "txn_001",
      "category_id": "cat_food",
      "category_name": "Food",
      "method": "rule_based",
      "confidence": 1.0,
      "explanation": "Matched known merchant 'Swiggy'"
    },
    {
      "transaction_id": "txn_002",
      "category_id": "cat_uncategorized",
      "category_name": "Uncategorized",
      "method": "none",
      "confidence": 0.0,
      "explanation": "LLM categorization timeout; defaulting to uncategorized"
    }
  ],
  "errors": [
    {
      "transaction_id": "txn_002",
      "error": "LLM_TIMEOUT",
      "message": "Claude API timeout after 3 retries"
    }
  ],
  "stats": {
    "total": 2,
    "rule_based": 1,
    "llm": 0,
    "uncategorized": 1,
    "processing_time_ms": 2100,
    "errors": 1
  }
}
```

### Error (400 Bad Request)

```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid request",
  "details": [
    {
      "field": "transactions[1].amount",
      "message": "amount must be > 0"
    }
  ]
}
```

### Error (401 Unauthorized)

```json
{
  "error": "UNAUTHORIZED",
  "message": "Invalid or missing authentication token"
}
```

### Error (500 Internal Server Error)

```json
{
  "error": "INTERNAL_ERROR",
  "message": "An unexpected error occurred during categorization"
}
```

## Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | string | ID of the transaction (from request) |
| `category_id` | string | ID of assigned category |
| `category_name` | string | Display name (e.g., "Food") |
| `method` | string | "rule_based", "fuzzy", "llm", "none" |
| `llm_provider` | string \| null | Which LLM provider was used: "ollama", "claude", "openai", or null if not LLM |
| `confidence` | float | 0.0-1.0; 1.0 for rule-based, provider-specific score for LLM, 0.0 for uncategorized |
| `explanation` | string | Human-readable reason for categorization |

## Categories

All responses use these predefined category IDs:

```
cat_food → "Food"
cat_shopping → "Shopping"
cat_transport → "Transport"
cat_housing → "Housing"
cat_utilities → "Utilities"
cat_entertainment → "Entertainment"
cat_income → "Income"
cat_healthcare → "Healthcare"
cat_education → "Education"
cat_miscellaneous → "Miscellaneous"
cat_uncategorized → "Uncategorized"
```

## Error Handling

### Transient Failures (LLM Timeout, Database Temporarily Unavailable)
- Status: 206 Partial Content
- Include successfully categorized transactions
- Flag failed transactions in `errors` array
- Suggest client retry later

### Validation Errors
- Status: 400 Bad Request
- Return detailed validation errors in `details` array

### Authentication Failures
- Status: 401 Unauthorized

### Server Errors
- Status: 500 Internal Server Error
- Do not expose internal error details to client

## Performance Targets

- **Rule-based only** (MVP): <100ms for 1000 transactions
- **With Ollama LLM** (Phase 2): 200-400ms per transaction (local inference), <10s for 100 transactions
- **With Claude LLM** (fallback): <2s per batch with 10 concurrent API calls
- P99 latency: <10000ms

**Configuration Note**: Default provider is Ollama. Switch to Claude/OpenAI by setting `LLM_PROVIDER` env var or updating `config/llm-config.yaml`.

## Testing Contract

### Contract Tests (from `/speckit-tasks`)

```bash
# Test 1: Rule-based categorization with known merchants
POST /api/v1/transactions/categorize
Body: { transactions: [{ merchant_name: "Swiggy", ... }] }
Expect: 200 OK, category_name: "Food", method: "rule_based"

# Test 2: Unknown merchant defaults to uncategorized
POST /api/v1/transactions/categorize
Body: { transactions: [{ merchant_name: "Unknown Vendor XYZ", ... }] }
Expect: 200 OK, category_name: "Uncategorized", method: "none"

# Test 3: Fuzzy matching for abbreviations
POST /api/v1/transactions/categorize
Body: { transactions: [{ merchant_name: "SWIGGY FD", ... }] }
Expect: 200 OK, category_name: "Food", method: "fuzzy", confidence: 0.85-0.99

# Test 4: Invalid request
POST /api/v1/transactions/categorize
Body: { transactions: [{ merchant_name: "", amount: -100 }] }
Expect: 400 Bad Request, error: "VALIDATION_ERROR"

# Test 5: Batch processing with mixed results
POST /api/v1/transactions/categorize
Body: { transactions: [ known_merchant, unknown_merchant, invalid_merchant ] }
Expect: 206 Partial Content, mixed results with errors array
```
