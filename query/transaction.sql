-- name: CreateTransaction :one
INSERT INTO transactions (user_id, type, amount, description, category_id, transaction_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, amount, description, category_id, transaction_date, created_at, updated_at;

-- name: GetTransaction :one
SELECT t.id, t.user_id, t.type, t.amount, t.description, t.category_id, c.name as category_name, t.transaction_date, t.created_at, t.updated_at
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.id = $1 AND t.user_id = $2 LIMIT 1;

-- name: ListTransactions :many
SELECT t.id, t.user_id, t.type, t.amount, t.description, t.category_id, c.name as category_name, t.transaction_date, t.created_at, t.updated_at
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1
ORDER BY t.transaction_date DESC, t.created_at DESC;

-- name: UpdateTransaction :one
UPDATE transactions
SET type = $3, amount = $4, description = $5, category_id = $6, transaction_date = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, type, amount, description, category_id, transaction_date, created_at, updated_at;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1 AND user_id = $2;

-- name: GetDashboardStats :one
SELECT 
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::numeric AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::numeric AS total_expense
FROM transactions
WHERE user_id = $1;
