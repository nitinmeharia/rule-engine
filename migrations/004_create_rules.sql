-- +goose Up
-- Create rules table
-- Following LLD.txt database schema specifications

CREATE TABLE rules (
    namespace     text,
    rule_id       text,
    version       int,
    status        text CHECK (status IN ('draft','active','inactive')),
    logic         text CHECK (logic IN ('AND','OR')),
    conditions    jsonb NOT NULL,
    created_by    text NOT NULL,
    published_by  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz,
    PRIMARY KEY (namespace, rule_id, version)
);

-- Foreign key constraint
ALTER TABLE rules ADD CONSTRAINT fk_rules_namespace 
    FOREIGN KEY (namespace) REFERENCES namespaces(id) ON DELETE CASCADE;

-- Ensure max 1 active + 1 draft per (namespace,rule_id)
CREATE UNIQUE INDEX uniq_rule_active ON rules(namespace, rule_id) 
    WHERE status='active';
CREATE UNIQUE INDEX uniq_rule_draft ON rules(namespace, rule_id) 
    WHERE status='draft';

-- Performance indexes
CREATE INDEX idx_rules_namespace ON rules(namespace);
CREATE INDEX idx_rules_status ON rules(status);
CREATE INDEX idx_rules_logic ON rules(logic);
CREATE INDEX idx_rules_created_by ON rules(created_by);
CREATE INDEX idx_rules_published_by ON rules(published_by);
CREATE INDEX idx_rules_created_at ON rules(created_at);
CREATE INDEX idx_rules_published_at ON rules(published_at);

-- JSONB indexes for conditions queries
CREATE INDEX idx_rules_conditions_gin ON rules USING gin(conditions);

-- +goose Down
-- Drop rules table and related indexes

DROP INDEX IF EXISTS idx_rules_conditions_gin;
DROP INDEX IF EXISTS idx_rules_published_at;
DROP INDEX IF EXISTS idx_rules_created_at;
DROP INDEX IF EXISTS idx_rules_published_by;
DROP INDEX IF EXISTS idx_rules_created_by;
DROP INDEX IF EXISTS idx_rules_logic;
DROP INDEX IF EXISTS idx_rules_status;
DROP INDEX IF EXISTS idx_rules_namespace;
DROP INDEX IF EXISTS uniq_rule_draft;
DROP INDEX IF EXISTS uniq_rule_active;
DROP TABLE IF EXISTS rules; 