-- +goose Up
-- Create active_config_meta table
-- Following LLD.txt database schema specifications

CREATE TABLE active_config_meta (
    namespace   text PRIMARY KEY REFERENCES namespaces(id) ON DELETE CASCADE,
    checksum    text NOT NULL,
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Performance indexes
CREATE INDEX idx_active_config_meta_updated_at ON active_config_meta(updated_at);
CREATE INDEX idx_active_config_meta_checksum ON active_config_meta(checksum);

-- +goose Down
-- Drop active_config_meta table and related indexes

DROP INDEX IF EXISTS idx_active_config_meta_checksum;
DROP INDEX IF EXISTS idx_active_config_meta_updated_at;
DROP TABLE IF EXISTS active_config_meta; 