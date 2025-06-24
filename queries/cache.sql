-- name: GetActiveConfigChecksum :one
SELECT namespace, checksum, updated_at
FROM active_config_meta
WHERE namespace = $1;

-- name: UpsertActiveConfigChecksum :exec
INSERT INTO active_config_meta (namespace, checksum, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (namespace)
DO UPDATE SET
    checksum = EXCLUDED.checksum,
    updated_at = EXCLUDED.updated_at;

-- name: RefreshNamespaceChecksum :exec
SELECT refresh_checksum($1);

-- name: ListAllActiveConfigChecksums :many
SELECT namespace, checksum, updated_at
FROM active_config_meta
ORDER BY namespace ASC;

-- name: DeleteActiveConfigChecksum :exec
DELETE FROM active_config_meta
WHERE namespace = $1; 