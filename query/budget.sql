-- name: GetBudgetsByUser :many
SELECT id, user_id, category_id, category_name, budget_limit, period, title, notification_enabled, created_at, updated_at
FROM budgets
WHERE user_id = $1
ORDER BY period, category_name;

-- name: GetBudgetByID :one
SELECT id, user_id, category_id, category_name, budget_limit, period, title, notification_enabled, created_at, updated_at
FROM budgets
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, category_name, budget_limit, period, title, notification_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, category_name)
DO UPDATE SET
    budget_limit         = EXCLUDED.budget_limit,
    period               = EXCLUDED.period,
    title                = EXCLUDED.title,
    notification_enabled = EXCLUDED.notification_enabled,
    updated_at           = NOW()
RETURNING id, user_id, category_id, category_name, budget_limit, period, title, notification_enabled, created_at, updated_at;

-- name: UpdateBudget :one
UPDATE budgets
SET category_id          = $3,
    category_name        = $4,
    budget_limit         = $5,
    period               = $6,
    title                = $7,
    notification_enabled = $8,
    updated_at           = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, category_id, category_name, budget_limit, period, title, notification_enabled, created_at, updated_at;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1 AND user_id = $2;
