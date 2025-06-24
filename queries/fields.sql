-- name: CreateField :exec
INSERT INTO fields (namespace, field_id, type, description, created_by)
VALUES ($1, $2, $3, $4, $5);

-- name: GetField :one
SELECT namespace, field_id, type, description, created_at, created_by
FROM fields
WHERE namespace = $1 AND field_id = $2;

-- name: ListFields :many
SELECT namespace, field_id, type, description, created_at, created_by
FROM fields
WHERE namespace = $1
ORDER BY field_id ASC;

-- name: UpdateField :exec
UPDATE fields
SET type = $3, description = $4, created_by = $5
WHERE namespace = $1 AND field_id = $2;

-- name: DeleteField :exec
DELETE FROM fields
WHERE namespace = $1 AND field_id = $2;

-- name: FieldExists :one
SELECT EXISTS(
    SELECT 1 FROM fields
    WHERE namespace = $1 AND field_id = $2
);

-- name: CountFieldsByNamespace :one
SELECT COUNT(*)
FROM fields
WHERE namespace = $1;