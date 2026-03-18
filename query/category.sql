-- name: CreateCategory :one
INSERT INTO categories (name, description, parent_id, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, description, parent_id, created_at, updated_at;

-- name: GetCategory :one
SELECT id, name, description, parent_id, created_at, updated_at
FROM categories
WHERE id = $1 LIMIT 1;

-- name: ListCategories :many
SELECT id, name, description, parent_id, created_at, updated_at
FROM categories
WHERE parent_id IS NULL
ORDER BY name;

-- name: ListSubCategories :many
SELECT id, name, description, parent_id, created_at, updated_at
FROM categories
WHERE parent_id = $1
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, parent_id, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;
