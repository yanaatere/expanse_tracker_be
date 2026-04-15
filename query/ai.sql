-- name: GetTransactionsForAI :many
SELECT t.type, t.amount, t.description, t.transaction_date,
       t.category_id, t.sub_category_id
FROM transactions t
WHERE t.user_id = $1
  AND t.transaction_date >= $2
ORDER BY t.transaction_date DESC
LIMIT 200;

-- name: GetMonthlyCategoryTotals :many
SELECT
    COALESCE(t.category_id::text, 'Uncategorized') AS category_name,
    t.type,
    SUM(t.amount) AS total,
    COUNT(*) AS count
FROM transactions t
WHERE t.user_id = $1
  AND DATE_TRUNC('month', t.transaction_date) = DATE_TRUNC('month', CURRENT_DATE)
GROUP BY t.category_id, t.type;

-- name: GetAvgMonthlyCategorySpend :many
SELECT
    COALESCE(category_id::text, 'Uncategorized') AS category_name,
    AVG(monthly_total) AS avg_monthly
FROM (
    SELECT t.category_id,
           DATE_TRUNC('month', t.transaction_date) AS month,
           SUM(t.amount) AS monthly_total
    FROM transactions t
    WHERE t.user_id = $1
      AND t.type = 'expense'
      AND t.transaction_date >= CURRENT_DATE - INTERVAL '3 months'
    GROUP BY t.category_id, DATE_TRUNC('month', t.transaction_date)
) sub
GROUP BY category_id;
