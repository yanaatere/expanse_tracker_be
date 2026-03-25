-- name: CreateWallet :one
INSERT INTO wallets (user_id, name, type, currency, balance, goals, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetWallet :one
SELECT * FROM wallets
WHERE id = $1 AND user_id = $2;

-- name: ListWallets :many
SELECT * FROM wallets
WHERE user_id = $1
ORDER BY name;

-- name: UpdateWallet :one
UPDATE wallets
SET name = $1, type = $2, currency = $3, balance = $4, goals = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $6 AND user_id = $7
RETURNING *;

-- name: DeleteWallet :exec
DELETE FROM wallets WHERE id = $1 AND user_id = $2;

-- name: AdjustWalletBalance :one
UPDATE wallets
SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2 AND user_id = $3
RETURNING *;
