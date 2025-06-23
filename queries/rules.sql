-- name: CreateRule :exec
INSERT INTO rules (namespace, rule_id, version, status, logic, conditions, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetRule :one
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1 AND rule_id = $2 AND version = $3;

-- name: GetActiveRuleVersion :one
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1 AND rule_id = $2 AND status = 'active';

-- name: GetDraftRuleVersion :one
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1 AND rule_id = $2 AND status = 'draft';

-- name: ListRules :many
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1
ORDER BY rule_id ASC, version DESC;

-- name: ListActiveRules :many
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1 AND status = 'active'
ORDER BY rule_id ASC;

-- name: ListRuleVersions :many
SELECT namespace, rule_id, version, status, logic, conditions, created_by, published_by, created_at, published_at
FROM rules
WHERE namespace = $1 AND rule_id = $2
ORDER BY version DESC;

-- name: UpdateRule :exec
UPDATE rules
SET logic = $4, conditions = $5, created_by = $6
WHERE namespace = $1 AND rule_id = $2 AND version = $3;

-- name: PublishRule :exec
UPDATE rules
SET status = 'active', published_by = $4, published_at = now()
WHERE namespace = $1 AND rule_id = $2 AND version = $3;

-- name: DeactivateRule :exec
UPDATE rules
SET status = 'inactive'
WHERE namespace = $1 AND rule_id = $2 AND status = 'active';

-- name: DeleteRule :exec
DELETE FROM rules
WHERE namespace = $1 AND rule_id = $2 AND version = $3;

-- name: GetMaxRuleVersion :one
SELECT COALESCE(MAX(version), 0) as max_version
FROM rules
WHERE namespace = $1 AND rule_id = $2;

-- name: RuleExists :one
SELECT EXISTS(
    SELECT 1 FROM rules
    WHERE namespace = $1 AND rule_id = $2 AND version = $3
); 