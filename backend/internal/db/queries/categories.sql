-- Categories queries for transaction categorization

-- name: GetAllCategories :many
SELECT id, name, description, color, icon, is_predefined, created_at, updated_at
FROM categories
ORDER BY name ASC;

-- name: GetCategoryByID :one
SELECT id, name, description, color, icon, is_predefined, created_at, updated_at
FROM categories
WHERE id = $1;

-- name: GetCategoryByName :one
SELECT id, name, description, color, icon, is_predefined, created_at, updated_at
FROM categories
WHERE name = $1;

-- name: CreateCategory :one
INSERT INTO categories (id, name, description, color, icon, is_predefined, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, name, description, color, icon, is_predefined, created_at, updated_at;

-- name: MerchantDictionaryInsert :exec
INSERT INTO merchant_dictionary (id, merchant_name, merchant_pattern, category_id, source, confidence, match_type, frequency, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 0, NOW(), NOW());

-- name: GetMerchantByName :one
SELECT id, merchant_name, merchant_pattern, category_id, source, confidence, match_type, frequency, created_at, updated_at
FROM merchant_dictionary
WHERE merchant_name = $1
LIMIT 1;

-- name: SearchMerchantsByPattern :many
SELECT id, merchant_name, merchant_pattern, category_id, source, confidence, match_type, frequency, created_at, updated_at
FROM merchant_dictionary
WHERE merchant_name ILIKE $1
ORDER BY frequency DESC
LIMIT $2;

-- name: TransactionCategoryInsert :exec
INSERT INTO transaction_categories (id, user_id, transaction_id, category_id, method, llm_provider, confidence, llm_explanation, assigned_by_user_id, assigned_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW());

-- name: GetTransactionCategory :one
SELECT id, user_id, transaction_id, category_id, method, llm_provider, confidence, llm_explanation, assigned_by_user_id, assigned_at, updated_at
FROM transaction_categories
WHERE transaction_id = $1;

-- name: GetTransactionsByCategory :many
SELECT id, user_id, transaction_id, category_id, method, llm_provider, confidence, llm_explanation, assigned_by_user_id, assigned_at, updated_at
FROM transaction_categories
WHERE user_id = $1 AND category_id = $2 AND assigned_at >= $3 AND assigned_at <= $4
ORDER BY assigned_at DESC
LIMIT $5;

-- name: UpdateTransactionCategory :exec
UPDATE transaction_categories
SET category_id = $1, method = $2, confidence = $3, assigned_by_user_id = $4, updated_at = NOW()
WHERE transaction_id = $5;

-- name: CategoryStatsUpsert :exec
INSERT INTO category_stats (id, user_id, category_id, period, total_spent, transaction_count, average_transaction, min_transaction, max_transaction, last_transaction_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), NOW())
ON CONFLICT (user_id, category_id, period) DO UPDATE SET
  total_spent = category_stats.total_spent + EXCLUDED.total_spent,
  transaction_count = category_stats.transaction_count + EXCLUDED.transaction_count,
  average_transaction = EXCLUDED.total_spent / EXCLUDED.transaction_count,
  max_transaction = GREATEST(category_stats.max_transaction, EXCLUDED.max_transaction),
  min_transaction = LEAST(category_stats.min_transaction, EXCLUDED.min_transaction),
  last_transaction_at = NOW(),
  updated_at = NOW();

-- name: GetCategoryStats :one
SELECT id, user_id, category_id, period, total_spent, transaction_count, average_transaction, min_transaction, max_transaction, last_transaction_at, created_at, updated_at
FROM category_stats
WHERE user_id = $1 AND category_id = $2 AND period = $3;

-- name: GetCategoryStatsByPeriod :many
SELECT id, user_id, category_id, period, total_spent, transaction_count, average_transaction, min_transaction, max_transaction, last_transaction_at, created_at, updated_at
FROM category_stats
WHERE user_id = $1 AND period = $2
ORDER BY total_spent DESC;
