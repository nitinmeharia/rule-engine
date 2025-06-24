-- +goose Up
-- Create a function to calculate and upsert a canonical checksum for a namespace's active configuration.

CREATE OR REPLACE FUNCTION refresh_namespace_checksum(ns_id text)
RETURNS void AS $$
DECLARE
    -- Variable to hold the canonical JSON representation of the active config
    active_config_jsonb jsonb;
    -- Variable to hold the calculated checksum
    new_checksum text;
BEGIN
    -- Aggregate all active entities for the given namespace into a single,
    -- canonical JSONB object. Using jsonb_agg ensures that the order of
    -- elements within the arrays does not affect the final object, and
    -- jsonb_build_object provides a consistent structure.
    SELECT jsonb_build_object(
        'fields', (SELECT jsonb_agg(f.*) FROM (SELECT * FROM fields WHERE namespace = ns_id ORDER BY field_id) as f),
        'terminals', (SELECT jsonb_agg(t.*) FROM (SELECT * FROM terminals WHERE namespace = ns_id ORDER BY terminal_id) as t),
        'functions', (SELECT jsonb_agg(fn.*) FROM (SELECT * FROM functions WHERE namespace = ns_id AND status = 'active' ORDER BY function_id, version) as fn),
        'rules', (SELECT jsonb_agg(r.*) FROM (SELECT * FROM rules WHERE namespace = ns_id AND status = 'active' ORDER BY rule_id, version) as r),
        'workflows', (SELECT jsonb_agg(w.*) FROM (SELECT * FROM workflows WHERE namespace = ns_id AND status = 'active' ORDER BY workflow_id, version) as w)
    ) INTO active_config_jsonb;

    -- Calculate the SHA256 checksum of the canonical JSONB object.
    -- The object is cast to text before hashing.
    new_checksum := encode(digest(active_config_jsonb::text, 'sha256'), 'hex');

    -- Insert the new checksum into the metadata table. If a checksum for the
    -- namespace already exists, update it. This is an UPSERT operation.
    INSERT INTO active_config_meta (namespace, checksum, updated_at)
    VALUES (ns_id, new_checksum, now())
    ON CONFLICT (namespace)
    DO UPDATE SET
        checksum = EXCLUDED.checksum,
        updated_at = EXCLUDED.updated_at;

END;
$$ LANGUAGE plpgsql;

-- +goose Down
-- Drop the function

DROP FUNCTION IF EXISTS refresh_namespace_checksum(text); 