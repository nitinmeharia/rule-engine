-- +goose Up
-- Create namespaces table
-- Following LLD.txt database schema specifications

CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- future LIKE optimisations

CREATE TABLE namespaces (
    id          text PRIMARY KEY,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    created_by  text NOT NULL
);

-- Add index for created_by queries
CREATE INDEX idx_namespaces_created_by ON namespaces(created_by);

-- Add index for created_at queries (for audit purposes)
CREATE INDEX idx_namespaces_created_at ON namespaces(created_at);

-- +goose Down
-- Drop namespaces table and related indexes

DROP INDEX IF EXISTS idx_namespaces_created_at;
DROP INDEX IF EXISTS idx_namespaces_created_by;
DROP TABLE IF EXISTS namespaces; 