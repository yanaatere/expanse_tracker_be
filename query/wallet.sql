-- name: CreateWallet :one
INSERT INTO wallets (user_id, name, type, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
SET name = $1, type = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3 AND user_id = $4
RETURNING *;

-- name: DeleteWallet :exec
DELETE FROM wallets WHERE id = $1 AND user_id = $2;
