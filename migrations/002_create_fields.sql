-- +goose Up
-- Create fields table
-- Following LLD.txt database schema specifications

CREATE TABLE fields (
    namespace    text REFERENCES namespaces(id) ON DELETE CASCADE,
    field_id     text,
    type         text CHECK (type IN ('number','string')),
    description  text,
    created_by   text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, field_id)
);

-- Indexes for performance
CREATE INDEX idx_fields_namespace ON fields(namespace);
CREATE INDEX idx_fields_type ON fields(type);
CREATE INDEX idx_fields_created_by ON fields(created_by);
CREATE INDEX idx_fields_created_at ON fields(created_at);

-- +goose Down
-- Drop fields table and related indexes

DROP INDEX IF EXISTS idx_fields_created_at;
DROP INDEX IF EXISTS idx_fields_created_by;
DROP INDEX IF EXISTS idx_fields_type;
DROP INDEX IF EXISTS idx_fields_namespace;
DROP TABLE IF EXISTS fields; 