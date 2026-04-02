-- name: GetUserBalance :one
SELECT 
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::numeric AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::numeric AS total_expense
FROM transactions
WHERE user_id = $1;

-- name: GetBalance :one
SELECT id, user_id, total_balance, created_at, updated_at
FROM balances
WHERE user_id = $1;

-- name: UpsertBalance :one
INSERT INTO balances (user_id, total_balance)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET total_balance = $2, updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, total_balance, created_at, updated_at;

-- name: AdjustBalance :one
INSERT INTO balances (user_id, total_balance)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET total_balance = balances.total_balance + $2, updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, total_balance, created_at, updated_at;

-- name: RecalculateBalance :one
INSERT INTO balances (user_id, total_balance)
VALUES ($1, (
    SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)
    FROM transactions WHERE user_id = $1
))
ON CONFLICT (user_id)
DO UPDATE SET 
    total_balance = (
        SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)
        FROM transactions WHERE user_id = $1
    ),
    updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, total_balance, created_at, updated_at;

-- name: GetMonthlyBalance :many
SELECT 
    DATE_TRUNC('month', transaction_date::timestamp)::date AS month,
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::numeric AS income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::numeric AS expense,
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)::numeric AS net
FROM transactions
WHERE user_id = $1
GROUP BY DATE_TRUNC('month', transaction_date::timestamp)
ORDER BY month DESC;

-- name: GetBalanceByDateRange :one
SELECT 
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::numeric AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::numeric AS total_expense,
    COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)::numeric AS balance
FROM transactions
WHERE user_id = $1 
  AND transaction_date >= $2 
  AND transaction_date <= $3;

