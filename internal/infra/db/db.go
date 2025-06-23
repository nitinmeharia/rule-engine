package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rule-engine/internal/config"
)

// New creates a new database connection pool using pgx
func New(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Build connection URL
	connURL := cfg.GetDatabaseURL()

	// Parse config
	poolConfig, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
