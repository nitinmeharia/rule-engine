-- name: CreateWorkflow :exec
INSERT INTO workflows (namespace, workflow_id, version, status, start_at, steps, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetWorkflow :one
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1 AND workflow_id = $2 AND version = $3;

-- name: GetActiveWorkflowVersion :one
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1 AND workflow_id = $2 AND status = 'active';

-- name: GetDraftWorkflowVersion :one
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1 AND workflow_id = $2 AND status = 'draft';

-- name: ListWorkflows :many
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1
ORDER BY workflow_id ASC, version DESC;

-- name: ListActiveWorkflows :many
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1 AND status = 'active'
ORDER BY workflow_id ASC;

-- name: ListWorkflowVersions :many
SELECT namespace, workflow_id, version, status, start_at, steps, created_by, published_by, created_at, published_at
FROM workflows
WHERE namespace = $1 AND workflow_id = $2
ORDER BY version DESC;

-- name: UpdateWorkflow :exec
UPDATE workflows
SET start_at = $4, steps = $5, created_by = $6
WHERE namespace = $1 AND workflow_id = $2 AND version = $3;

-- name: PublishWorkflow :exec
UPDATE workflows
SET status = 'active', published_by = $4, published_at = now()
WHERE namespace = $1 AND workflow_id = $2 AND version = $3;

-- name: DeactivateWorkflow :exec
UPDATE workflows
SET status = 'inactive'
WHERE namespace = $1 AND workflow_id = $2 AND status = 'active';

-- name: DeleteWorkflow :exec
DELETE FROM workflows
WHERE namespace = $1 AND workflow_id = $2 AND version = $3;

-- name: GetMaxWorkflowVersion :one
SELECT COALESCE(MAX(version), 0) as max_version
FROM workflows
WHERE namespace = $1 AND workflow_id = $2;

-- name: WorkflowExists :one
SELECT EXISTS(
    SELECT 1 FROM workflows
    WHERE namespace = $1 AND workflow_id = $2 AND version = $3
); 