-- +goose Up
-- Create functions table
-- Following LLD.txt database schema specifications

CREATE TABLE functions (
    namespace     text,
    function_id   text,
    version       int,
    status        text CHECK (status IN ('draft','active','inactive')),
    type          text CHECK (type IN ('max','sum','avg','in')),
    args          text[],          -- for numeric ops
    values        text[],          -- for 'in' op
    return_type   text CHECK (return_type IN ('number','bool')),
    created_by    text NOT NULL,
    published_by  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz,
    PRIMARY KEY (namespace, function_id, version)
);

-- Foreign key constraint
ALTER TABLE functions ADD CONSTRAINT fk_functions_namespace 
    FOREIGN KEY (namespace) REFERENCES namespaces(id) ON DELETE CASCADE;

-- Ensure max 1 active + 1 draft per (namespace,function_id)
CREATE UNIQUE INDEX uniq_function_active ON functions(namespace, function_id)
    WHERE status='active';
CREATE UNIQUE INDEX uniq_function_draft ON functions(namespace, function_id)
    WHERE status='draft';

-- Performance indexes
CREATE INDEX idx_functions_namespace ON functions(namespace);
CREATE INDEX idx_functions_status ON functions(status);
CREATE INDEX idx_functions_type ON functions(type);
CREATE INDEX idx_functions_created_by ON functions(created_by);
CREATE INDEX idx_functions_published_by ON functions(published_by);
CREATE INDEX idx_functions_created_at ON functions(created_at);
CREATE INDEX idx_functions_published_at ON functions(published_at);

-- +goose Down
-- Drop functions table and related indexes

DROP INDEX IF EXISTS idx_functions_published_at;
DROP INDEX IF EXISTS idx_functions_created_at;
DROP INDEX IF EXISTS idx_functions_published_by;
DROP INDEX IF EXISTS idx_functions_created_by;
DROP INDEX IF EXISTS idx_functions_type;
DROP INDEX IF EXISTS idx_functions_status;
DROP INDEX IF EXISTS idx_functions_namespace;
DROP INDEX IF EXISTS uniq_function_draft;
DROP INDEX IF EXISTS uniq_function_active;
DROP TABLE IF EXISTS functions; 