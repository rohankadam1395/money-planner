# Contract: Category Analytics Endpoint

**Feature**: Transaction Categorization | **Phase**: Phase 1 (US3: Category Management & Analytics)

## Endpoints

```
GET /api/v1/categories
GET /api/v1/categories/{category_id}/stats
GET /api/v1/categories/{category_id}/transactions
```

## Purpose

Provide analytics views of transactions grouped by category, enabling spending analysis, budget tracking, and recategorization workflows.

---

## 1. List Categories

```
GET /api/v1/categories
```

### Request

#### Headers
```
Authorization: Bearer {jwt_token}
```

#### Query Parameters (optional)
- `include_stats` (boolean, default: false): Include spending stats in response
- `period` (string, YYYY-MM): Filter stats to specific month if `include_stats=true`

### Response

#### Success (200 OK)

```json
{
  "categories": [
    {
      "id": "cat_food",
      "name": "Food",
      "description": "Restaurants, food delivery, groceries",
      "color": "#FF6B6B",
      "icon": "utensils",
      "is_predefined": true
    },
    {
      "id": "cat_shopping",
      "name": "Shopping",
      "description": "Online and retail purchases",
      "color": "#4ECDC4",
      "icon": "shopping-bag",
      "is_predefined": true
    }
  ],
  "total": 10
}
```

#### With Stats (include_stats=true, period=2026-07)

```json
{
  "categories": [
    {
      "id": "cat_food",
      "name": "Food",
      "description": "Restaurants, food delivery, groceries",
      "color": "#FF6B6B",
      "icon": "utensils",
      "is_predefined": true,
      "stats": {
        "total_spent": 4500.50,
        "transaction_count": 24,
        "average_transaction": 187.52,
        "min_transaction": 50.00,
        "max_transaction": 450.00,
        "last_transaction_at": "2026-07-11T20:30:00Z"
      }
    }
  ],
  "total": 10
}
```

---

## 2. Get Category Stats

```
GET /api/v1/categories/{category_id}/stats
```

### Request

#### Path Parameters
- `category_id` (string): ID of the category (e.g., "cat_food")

#### Query Parameters (optional)
- `period` (string, YYYY-MM): Filter to specific month (default: current month)
- `include_transactions` (boolean, default: false): Include transaction list in response
- `sort_by` (string, default: "date"): Sort transactions by "date", "amount", "merchant"
- `limit` (integer, default: 100): Max transactions to include (if include_transactions=true)

### Response

#### Success (200 OK)

```json
{
  "category_id": "cat_food",
  "category_name": "Food",
  "period": "2026-07",
  "stats": {
    "total_spent": 4500.50,
    "transaction_count": 24,
    "average_transaction": 187.52,
    "min_transaction": 50.00,
    "max_transaction": 450.00,
    "last_transaction_at": "2026-07-11T20:30:00Z"
  }
}
```

#### With Transactions (include_transactions=true)

```json
{
  "category_id": "cat_food",
  "category_name": "Food",
  "period": "2026-07",
  "stats": {
    "total_spent": 4500.50,
    "transaction_count": 24,
    "average_transaction": 187.52,
    "min_transaction": 50.00,
    "max_transaction": 450.00,
    "last_transaction_at": "2026-07-11T20:30:00Z"
  },
  "transactions": [
    {
      "transaction_id": "txn_001",
      "merchant_name": "Swiggy Food Delivery",
      "amount": 450.00,
      "date": "2026-07-11T20:30:00Z",
      "category_id": "cat_food",
      "categorization_method": "rule_based",
      "confidence": 1.0
    }
  ],
  "pagination": {
    "total": 24,
    "returned": 24,
    "offset": 0,
    "limit": 100
  }
}
```

---

## 3. Get Transactions in Category

```
GET /api/v1/categories/{category_id}/transactions
```

### Request

#### Path Parameters
- `category_id` (string): ID of the category

#### Query Parameters
- `period_start` (ISO8601 date, optional): Filter by date range start
- `period_end` (ISO8601 date, optional): Filter by date range end
- `min_amount` (decimal, optional): Filter by minimum transaction amount
- `max_amount` (decimal, optional): Filter by maximum transaction amount
- `min_confidence` (float, 0.0-1.0, optional): Filter by categorization confidence (e.g., show only high-confidence categorizations)
- `sort_by` (string, default: "date_desc"): "date_asc", "date_desc", "amount_asc", "amount_desc"
- `offset` (integer, default: 0): Pagination offset
- `limit` (integer, default: 100): Max 1000

### Response

#### Success (200 OK)

```json
{
  "category_id": "cat_food",
  "category_name": "Food",
  "filters_applied": {
    "period_start": "2026-07-01",
    "period_end": "2026-07-31",
    "min_confidence": 0.75
  },
  "transactions": [
    {
      "transaction_id": "txn_001",
      "statement_id": "stmt_001",
      "date": "2026-07-11T20:30:00Z",
      "merchant_name": "Swiggy Food Delivery",
      "description": "SWIGGY FOOD DELIV",
      "amount": 450.00,
      "transaction_type": "debit",
      "balance_after": 5000.00,
      "category_id": "cat_food",
      "category_name": "Food",
      "categorization_method": "rule_based",
      "confidence": 1.0,
      "categorization_timestamp": "2026-07-11T20:30:30Z",
      "can_recategorize": true
    }
  ],
  "pagination": {
    "total": 24,
    "returned": 10,
    "offset": 0,
    "limit": 100
  },
  "summary": {
    "total_amount": 4500.50,
    "transaction_count": 24,
    "average_amount": 187.52
  }
}
```

#### Error (404 Not Found)

```json
{
  "error": "NOT_FOUND",
  "message": "Category not found or does not exist"
}
```

---

## Response Fields

### Category Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Category ID (e.g., "cat_food") |
| `name` | string | Display name |
| `description` | string | Description of category |
| `color` | string | Hex color for UI |
| `icon` | string | Icon name |
| `is_predefined` | boolean | true for system categories |

### Stats Object

| Field | Type | Description |
|-------|------|-------------|
| `total_spent` | decimal | Sum of all transaction amounts |
| `transaction_count` | integer | Number of transactions |
| `average_transaction` | decimal | mean(amounts) |
| `min_transaction` | decimal | Smallest transaction |
| `max_transaction` | decimal | Largest transaction |
| `last_transaction_at` | ISO8601 | Most recent transaction timestamp |

### Transaction Object

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | string | Unique transaction ID |
| `date` | ISO8601 | Transaction date |
| `merchant_name` | string | Merchant name from statement |
| `amount` | decimal | Transaction amount |
| `categorization_method` | string | "rule_based", "fuzzy", "llm", "manual", "none" |
| `confidence` | float | 0.0-1.0; 1.0 for rule/manual, LLM score, 0.0 for uncategorized |
| `can_recategorize` | boolean | true if user can manually recategorize |

---

## Error Handling

### Invalid Period Format
- Status: 400 Bad Request
- `period` must be YYYY-MM (e.g., "2026-07")

### Invalid Date Range
- Status: 400 Bad Request
- `period_start` must be <= `period_end`

### Invalid Amount Filter
- Status: 400 Bad Request
- `min_amount` must be <= `max_amount`

### Category Not Found
- Status: 404 Not Found

### Limit Exceeded
- Status: 400 Bad Request
- `limit` cannot exceed 1000

## Performance Targets

- **List categories**: <100ms
- **Get category stats**: <500ms (for typical month with 100+ transactions)
- **Get transactions (paginated)**: <1000ms for 100 transactions

## Testing Contract

### Contract Tests

```bash
# Test 1: List all categories
GET /api/v1/categories
Expect: 200 OK, categories array with 10+ predefined categories

# Test 2: List categories with stats
GET /api/v1/categories?include_stats=true&period=2026-07
Expect: 200 OK, each category includes stats object

# Test 3: Get category stats
GET /api/v1/categories/cat_food/stats?period=2026-07
Expect: 200 OK, stats for Food category in July 2026

# Test 4: Get category transactions with filters
GET /api/v1/categories/cat_food/transactions?min_amount=100&max_amount=500&min_confidence=0.8
Expect: 200 OK, filtered transaction list

# Test 5: Invalid period format
GET /api/v1/categories?period=invalid
Expect: 400 Bad Request, error: "VALIDATION_ERROR"

# Test 6: Pagination
GET /api/v1/categories/cat_food/transactions?limit=10&offset=10
Expect: 200 OK, second page of transactions
```
