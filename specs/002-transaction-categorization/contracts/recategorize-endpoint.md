# Contract: Recategorize Transaction Endpoint

**Feature**: Transaction Categorization | **Phase**: Phase 1 (US3: Category Management)

## Endpoint

```
POST /api/v1/transactions/{transaction_id}/recategorize
```

## Purpose

Allow users to manually correct a transaction's category after import. Updates the category assignment and optionally learns the correction for future merchant dictionary updates.

## Request

### Path Parameters
- `transaction_id` (string, UUID): ID of the transaction to recategorize

### Headers
```
Content-Type: application/json
Authorization: Bearer {jwt_token}
```

### Body

```json
{
  "category_id": "cat_shopping",
  "learn_correction": true,
  "notes": "This was actually an Amazon purchase, not miscellaneous"
}
```

### Validation
- `category_id` required, must be a valid predefined category ID
- `learn_correction` optional, default: false
- `notes` optional, max 500 characters

## Response

### Success (200 OK)

```json
{
  "transaction_id": "txn_001",
  "old_category": {
    "category_id": "cat_miscellaneous",
    "category_name": "Miscellaneous"
  },
  "new_category": {
    "category_id": "cat_shopping",
    "category_name": "Shopping"
  },
  "updated_at": "2026-07-12T15:45:00Z",
  "learned": true,
  "message": "Transaction recategorized successfully"
}
```

### Error (404 Not Found)

```json
{
  "error": "NOT_FOUND",
  "message": "Transaction not found or does not belong to the authenticated user"
}
```

### Error (400 Bad Request)

```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid category_id or request body",
  "details": {
    "category_id": "Unknown category ID"
  }
}
```

### Error (401 Unauthorized)

```json
{
  "error": "UNAUTHORIZED",
  "message": "Invalid or missing authentication token"
}
```

## Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | string | ID of the transaction that was recategorized |
| `old_category` | object | Previous category assignment |
| `new_category` | object | New category assignment |
| `updated_at` | ISO8601 | Timestamp of the update |
| `learned` | boolean | Whether the correction was learned for future merchant dictionary updates |
| `message` | string | Success message |

## Side Effects

### Category Stats Update
When a transaction is recategorized, the `category_stats` table for the affected month is updated:
- Old category: `total_spent` decremented, `transaction_count` decremented
- New category: `total_spent` incremented, `transaction_count` incremented

### Merchant Dictionary Learning (if `learn_correction=true`)
If the transaction's merchant name is not in the merchant_dictionary:
1. Add new entry: `merchant_name` → `new_category_id` with `source: "user_correction"`
2. Set `confidence: 90` (high confidence, but not 100% since it's a single correction)

If merchant already in dictionary with different category:
1. Log the discrepancy (user correction conflicts with existing mapping)
2. Update the entry's `frequency` counter for analysis
3. Optionally flag for admin review if conflict pattern emerges

## Error Handling

### Transaction Not Found
- Status: 404 Not Found
- Ensure transaction belongs to authenticated user (privacy check)

### Invalid Category
- Status: 400 Bad Request
- Only predefined categories allowed (custom categories deferred to Phase 3+)

### Concurrent Updates
- If transaction is being recategorized concurrently from another request:
  - Latest update wins (last-write-wins strategy)
  - Both requests return success (optimistic concurrency)

## Performance Targets

- Response time: <500ms
- Category stats update: <1s propagation
- Merchant dictionary learning: async, can be batched

## Testing Contract

### Contract Tests

```bash
# Test 1: Successful recategorization
POST /api/v1/transactions/txn_001/recategorize
Body: { category_id: "cat_shopping" }
Expect: 200 OK, old_category: original, new_category: "Shopping"

# Test 2: Recategorization with learning
POST /api/v1/transactions/txn_001/recategorize
Body: { category_id: "cat_shopping", learn_correction: true }
Expect: 200 OK, learned: true

# Test 3: Invalid category ID
POST /api/v1/transactions/txn_001/recategorize
Body: { category_id: "cat_invalid" }
Expect: 400 Bad Request, error: "VALIDATION_ERROR"

# Test 4: Transaction not found
POST /api/v1/transactions/nonexistent/recategorize
Body: { category_id: "cat_shopping" }
Expect: 404 Not Found

# Test 5: Concurrent updates (last-write-wins)
POST /api/v1/transactions/txn_001/recategorize (Request A)
POST /api/v1/transactions/txn_001/recategorize (Request B)
Expect: Both 200 OK, final_category matches Request B's category
```
