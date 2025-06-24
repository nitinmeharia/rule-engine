#!/bin/bash

# Setup test database with proper migrations
set -e

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="rule_engine_test"
DB_USER="postgres"
DB_PASSWORD="postgres"

export PGPASSWORD=$DB_PASSWORD

echo "Setting up test database..."

# Function to extract and run Up migration
run_migration() {
    local migration_file=$1
    echo "Running migration: $migration_file"
    awk '/-- \\+goose Up/{flag=1; next} /-- \\+goose Down/{flag=0} flag' "$migration_file" | \
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
}

# Run migrations in order
run_migration "migrations/001_create_namespaces.sql"
run_migration "migrations/002_create_fields.sql"
run_migration "migrations/003_create_functions.sql"
run_migration "migrations/004_create_rules.sql"
run_migration "migrations/005_create_workflows.sql"
run_migration "migrations/006_create_terminals.sql"
run_migration "migrations/007_create_active_config_meta.sql"
run_migration "migrations/008_create_checksum_function.sql"

echo "Test database setup complete!" 