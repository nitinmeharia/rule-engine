-- name: CreateNamespace :exec
INSERT INTO namespaces (id, description, created_by)
VALUES ($1, $2, $3);

-- name: GetNamespace :one
SELECT id, description, created_at, created_by
FROM namespaces
WHERE id = $1;

-- name: ListNamespaces :many
SELECT id, description, created_at, created_by
FROM namespaces
ORDER BY created_at ASC;

-- name: DeleteNamespace :exec
DELETE FROM namespaces WHERE id = $1; 