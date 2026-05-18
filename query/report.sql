-- name: GetTransactionsByDateRange :many
SELECT t.id, t.user_id, t.type, t.amount, t.description,
       t.category_id, t.sub_category_id,
       t.wallet_id, w.name as wallet_name,
       t.receipt_image_url,
       t.transaction_date, t.created_at, t.updated_at
FROM transactions t
LEFT JOIN wallets w ON t.wallet_id = w.id
WHERE t.user_id = $1
  AND t.transaction_date >= $2
  AND t.transaction_date <= $3
ORDER BY t.transaction_date DESC, t.created_at DESC;
