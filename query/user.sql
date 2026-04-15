-- name: CreateUser :one
INSERT INTO users (username, email, password)
VALUES ($1, $2, $3)
RETURNING id, username, email, password, created_at, updated_at, is_premium;

-- name: GetUser :one
SELECT id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium
FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET username = $2, email = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium
FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium
FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdatePassword :one
UPDATE users
SET password = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium;

-- name: SetPasswordResetToken :one
UPDATE users
SET password_reset_token = $2, password_reset_expires = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium;

-- name: GetUserByResetToken :one
SELECT id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium
FROM users
WHERE password_reset_token = $1 AND password_reset_expires > CURRENT_TIMESTAMP LIMIT 1;

-- name: ClearPasswordResetToken :one
UPDATE users
SET password_reset_token = NULL, password_reset_expires = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium;

-- name: SetUserPremium :one
UPDATE users
SET is_premium = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, password, password_reset_token, password_reset_expires, created_at, updated_at, is_premium;
