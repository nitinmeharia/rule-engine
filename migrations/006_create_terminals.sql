-- +goose Up
-- Create terminals table
-- Following LLD.txt database schema specifications

CREATE TABLE terminals (
    namespace   text REFERENCES namespaces(id) ON DELETE CASCADE,
    terminal_id text,
    created_by  text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, terminal_id)
);

-- Performance indexes
CREATE INDEX idx_terminals_namespace ON terminals(namespace);
CREATE INDEX idx_terminals_created_by ON terminals(created_by);
CREATE INDEX idx_terminals_created_at ON terminals(created_at);

-- +goose Down
-- Drop terminals table and related indexes

DROP INDEX IF EXISTS idx_terminals_created_at;
DROP INDEX IF EXISTS idx_terminals_created_by;
DROP INDEX IF EXISTS idx_terminals_namespace;
DROP TABLE IF EXISTS terminals; 