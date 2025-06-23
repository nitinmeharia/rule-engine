-- name: CreateFunction :exec
INSERT INTO functions (namespace, function_id, version, status, type, args, values, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetFunction :one
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1 AND function_id = $2 AND version = $3;

-- name: GetActiveFunctionVersion :one
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1 AND function_id = $2 AND status = 'active';

-- name: GetDraftFunctionVersion :one
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1 AND function_id = $2 AND status = 'draft';

-- name: ListFunctions :many
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1
ORDER BY function_id ASC, version DESC;

-- name: ListActiveFunctions :many
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1 AND status = 'active'
ORDER BY function_id ASC;

-- name: ListFunctionVersions :many
SELECT namespace, function_id, version, status, type, args, values, created_by, published_by, created_at, published_at
FROM functions
WHERE namespace = $1 AND function_id = $2
ORDER BY version DESC;

-- name: UpdateFunction :exec
UPDATE functions
SET type = $4, args = $5, values = $6, created_by = $7
WHERE namespace = $1 AND function_id = $2 AND version = $3;

-- name: PublishFunction :exec
UPDATE functions
SET status = 'active', published_by = $4, published_at = now()
WHERE namespace = $1 AND function_id = $2 AND version = $3;

-- name: DeactivateFunction :exec
UPDATE functions
SET status = 'inactive'
WHERE namespace = $1 AND function_id = $2 AND status = 'active';

-- name: DeleteFunction :exec
DELETE FROM functions
WHERE namespace = $1 AND function_id = $2 AND version = $3;

-- name: GetMaxFunctionVersion :one
SELECT COALESCE(MAX(version), 0) as max_version
FROM functions
WHERE namespace = $1 AND function_id = $2;

-- name: FunctionExists :one
SELECT EXISTS(
    SELECT 1 FROM functions
    WHERE namespace = $1 AND function_id = $2 AND version = $3
); 