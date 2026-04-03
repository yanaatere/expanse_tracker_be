-- name: CreateRecurringTransaction :one
INSERT INTO recurring_transactions (user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, next_execution_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at;

-- name: ListRecurringTransactions :many
SELECT id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at
FROM recurring_transactions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetRecurringTransaction :one
SELECT id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at
FROM recurring_transactions
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: UpdateRecurringTransaction :one
UPDATE recurring_transactions
SET title = $3, type = $4, amount = $5, category_id = $6, sub_category_id = $7,
    wallet_id = $8, frequency = $9, start_date = $10, end_date = $11,
    next_execution_date = $12, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at;

-- name: DeleteRecurringTransaction :exec
DELETE FROM recurring_transactions
WHERE id = $1 AND user_id = $2;

-- name: ListDueRecurringTransactions :many
SELECT id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at
FROM recurring_transactions
WHERE is_active = TRUE AND next_execution_date <= $1
ORDER BY next_execution_date ASC;

-- name: UpdateNextExecutionDate :one
UPDATE recurring_transactions
SET next_execution_date = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, user_id, title, type, amount, category_id, sub_category_id, wallet_id, frequency, start_date, end_date, is_active, next_execution_date, created_at, updated_at;

-- name: DeactivateRecurringTransaction :exec
UPDATE recurring_transactions
SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
