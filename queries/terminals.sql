-- name: CreateTerminal :exec
INSERT INTO terminals (namespace, terminal_id, created_by)
VALUES ($1, $2, $3);

-- name: GetTerminal :one
SELECT namespace, terminal_id, created_at, created_by
FROM terminals
WHERE namespace = $1 AND terminal_id = $2;

-- name: ListTerminals :many
SELECT namespace, terminal_id, created_at, created_by
FROM terminals
WHERE namespace = $1
ORDER BY terminal_id ASC;

-- name: DeleteTerminal :exec
DELETE FROM terminals
WHERE namespace = $1 AND terminal_id = $2;

-- name: TerminalExists :one
SELECT EXISTS(
    SELECT 1 FROM terminals
    WHERE namespace = $1 AND terminal_id = $2
);

-- name: CountTerminalsByNamespace :one
SELECT COUNT(*)
FROM terminals
WHERE namespace = $1; 