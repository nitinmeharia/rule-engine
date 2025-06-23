package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/infra/db"
	"github.com/rule-engine/internal/infra/logger"
	"github.com/rule-engine/internal/server"
)

// Application holds all the core application dependencies
type Application struct {
	Config           *config.Config
	DB               *pgxpool.Pool
	Server           *server.Server
	Logger           *logger.Logger
	ExecutionService ExecutionService // Interface to be defined
}

// ExecutionService interface for cache refresh loop
type ExecutionService interface {
	StartRefreshLoop(ctx context.Context)
}

// Init initializes the application with all dependencies
func Init() (*Application, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database
	database, err := db.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize HTTP server
	srv, err := server.New(cfg, database, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	// TODO: Initialize execution service with cache
	// For now, create a placeholder
	execService := &PlaceholderExecutionService{}

	app := &Application{
		Config:           cfg,
		DB:               database,
		Server:           srv,
		Logger:           log,
		ExecutionService: execService,
	}

	return app, nil
}

// PlaceholderExecutionService is a temporary implementation
type PlaceholderExecutionService struct{}

func (p *PlaceholderExecutionService) StartRefreshLoop(ctx context.Context) {
	// TODO: Implement cache refresh loop
	select {
	case <-ctx.Done():
		return
	}
}
