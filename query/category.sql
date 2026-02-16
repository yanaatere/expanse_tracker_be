-- name: CreateCategory :one
INSERT INTO categories (name, description, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, description, created_at, updated_at;

-- name: GetCategory :one
SELECT id, name, description, created_at, updated_at
FROM categories
WHERE id = $1 LIMIT 1;

-- name: ListCategories :many
SELECT id, name, description, created_at, updated_at
FROM categories
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;
