-- Transactions queries

-- name: CreateTransaction :one
INSERT INTO transactions (
    transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at;

-- name: GetTransactionsByStatement :many
SELECT transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at
FROM transactions
WHERE statement_id = $1
ORDER BY transaction_date DESC;

-- name: GetTransactionsByUserAndDate :many
SELECT transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at
FROM transactions
WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date <= $3
ORDER BY transaction_date DESC;

-- name: GetTransactionsByUser :many
SELECT transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at
FROM transactions
WHERE user_id = $1
ORDER BY transaction_date DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByBank :many
SELECT transaction_id, user_id, statement_id, transaction_date, merchant,
    amount, type, balance, description, currency, imported_at,
    bank_code, account_number_hash, raw_data, created_at, updated_at
FROM transactions
WHERE user_id = $1 AND bank_code = $2
ORDER BY transaction_date DESC;

-- name: DeleteTransactionsByStatement :exec
DELETE FROM transactions
WHERE statement_id = $1;

-- name: CountTransactionsByStatement :one
SELECT COUNT(*) FROM transactions
WHERE statement_id = $1;
