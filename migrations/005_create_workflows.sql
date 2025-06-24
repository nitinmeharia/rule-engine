-- +goose Up
-- Create workflows table
-- Following LLD.txt database schema specifications

CREATE TABLE workflows (
    namespace     text,
    workflow_id   text,
    version       int,
    status        text CHECK (status IN ('draft','active','inactive')),
    start_at      text NOT NULL,
    steps         jsonb NOT NULL,
    created_by    text NOT NULL,
    published_by  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz,
    PRIMARY KEY (namespace, workflow_id, version)
);

-- Foreign key constraint
ALTER TABLE workflows ADD CONSTRAINT fk_workflows_namespace 
    FOREIGN KEY (namespace) REFERENCES namespaces(id) ON DELETE CASCADE;

-- Ensure max 1 active + 1 draft per (namespace,workflow_id)
CREATE UNIQUE INDEX uniq_workflow_active ON workflows(namespace, workflow_id) 
    WHERE status='active';
CREATE UNIQUE INDEX uniq_workflow_draft ON workflows(namespace, workflow_id) 
    WHERE status='draft';

-- Performance indexes
CREATE INDEX idx_workflows_namespace ON workflows(namespace);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_created_by ON workflows(created_by);
CREATE INDEX idx_workflows_published_by ON workflows(published_by);
CREATE INDEX idx_workflows_created_at ON workflows(created_at);
CREATE INDEX idx_workflows_published_at ON workflows(published_at);

-- JSONB indexes for steps queries
CREATE INDEX idx_workflows_steps_gin ON workflows USING gin(steps);

-- +goose Down
-- Drop workflows table and related indexes

DROP INDEX IF EXISTS idx_workflows_steps_gin;
DROP INDEX IF EXISTS idx_workflows_published_at;
DROP INDEX IF EXISTS idx_workflows_created_at;
DROP INDEX IF EXISTS idx_workflows_published_by;
DROP INDEX IF EXISTS idx_workflows_created_by;
DROP INDEX IF EXISTS idx_workflows_status;
DROP INDEX IF EXISTS idx_workflows_namespace;
DROP INDEX IF EXISTS uniq_workflow_draft;
DROP INDEX IF EXISTS uniq_workflow_active;
DROP TABLE IF EXISTS workflows; 