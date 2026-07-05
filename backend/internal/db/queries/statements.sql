-- Statements queries

-- name: CreateStatement :one
INSERT INTO statements (
    statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at;

-- name: GetStatementByID :one
SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at
FROM statements
WHERE statement_id = $1;

-- name: GetStatementByFileHash :one
SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at
FROM statements
WHERE file_hash = $1;

-- name: GetStatementsByUser :many
SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at
FROM statements
WHERE user_id = $1
ORDER BY uploaded_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOverlappingStatements :many
SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at
FROM statements
WHERE user_id = $1
  AND bank_code = $2
  AND account_number_hash = $3
  AND (
    (statement_period_start <= $4 AND statement_period_end >= $5)
    OR (statement_period_start <= $5 AND statement_period_end >= $4)
  );

-- name: UpdateStatementStatus :one
UPDATE statements
SET status = $2, imported_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE statement_id = $1
RETURNING statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at;

-- name: UpdateStatementError :one
UPDATE statements
SET status = 'FAILED', error_log = $2, updated_at = CURRENT_TIMESTAMP
WHERE statement_id = $1
RETURNING statement_id, user_id, file_name, file_format, file_size_bytes,
    file_hash, bank_code, account_number_hash, statement_period_start,
    statement_period_end, transaction_count, status, error_log,
    uploaded_at, imported_at, created_at, updated_at;
