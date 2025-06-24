-- +goose Up
-- Create refresh_checksum function
-- Following LLD.txt database schema specifications

CREATE OR REPLACE FUNCTION refresh_checksum(ns text) RETURNS void AS $$
DECLARE
    config_data text;
    new_checksum text;
BEGIN
    -- Simple implementation for now
    config_data := ns || '_config_data';
    new_checksum := encode(digest(config_data, 'sha256'), 'hex');
    
    -- Insert or update the checksum
    INSERT INTO active_config_meta (namespace, checksum, updated_at)
    VALUES (ns, new_checksum, now())
    ON CONFLICT (namespace) 
    DO UPDATE SET 
        checksum = EXCLUDED.checksum,
        updated_at = EXCLUDED.updated_at;
END;
$$ LANGUAGE plpgsql;

-- +goose Down
-- Drop refresh_checksum function

DROP FUNCTION IF EXISTS refresh_checksum(text); 