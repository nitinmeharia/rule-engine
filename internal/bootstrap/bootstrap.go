package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/execution"
	"github.com/rule-engine/internal/infra/db"
	"github.com/rule-engine/internal/infra/logger"
	modelsdb "github.com/rule-engine/internal/models/db"
	"github.com/rule-engine/internal/repository"
	"github.com/rule-engine/internal/server"
)

// Application holds all the core application dependencies
type Application struct {
	Config       *config.Config
	DB           *pgxpool.Pool
	Server       *server.Server
	Logger       *logger.Logger
	Engine       *execution.Engine
	CacheRefresh *execution.CacheRefreshService
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

	// Initialize database queries
	queries := modelsdb.New(database)

	// Initialize repositories
	fieldRepo := repository.NewFieldRepository(queries)
	functionRepo := repository.NewFunctionRepository(queries)
	ruleRepo := repository.NewRuleRepository(queries)
	terminalRepo := repository.NewTerminalRepository(queries)
	workflowRepo := repository.NewWorkflowRepository(queries)
	cacheRepo := repository.NewCacheRepository(queries)

	// Initialize execution engine
	engine := execution.NewEngine(
		cacheRepo,
		ruleRepo,
		workflowRepo,
		fieldRepo,
		functionRepo,
		terminalRepo,
		time.Duration(cfg.Cache.RefreshIntervalSec)*time.Second,
	)

	// Convert pgxpool to sql.DB for cache refresh service
	sqlDB, err := sql.Open("postgres", database.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to create sql.DB: %w", err)
	}

	// Initialize cache refresh service
	cacheRefresh := execution.NewCacheRefreshService(
		engine,
		sqlDB,
		cfg,
		log,
	)

	// Initialize HTTP server
	srv, err := server.New(cfg, database, log, engine)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	app := &Application{
		Config:       cfg,
		DB:           database,
		Server:       srv,
		Logger:       log,
		Engine:       engine,
		CacheRefresh: cacheRefresh,
	}

	return app, nil
}

// Start starts all application components
func (app *Application) Start(ctx context.Context) error {
	// Start cache refresh service
	if err := app.CacheRefresh.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cache refresh service: %w", err)
	}

	// Start HTTP server
	return app.Server.Start()
}

// Stop gracefully stops all application components
func (app *Application) Stop(ctx context.Context) error {
	// Stop cache refresh service
	if err := app.CacheRefresh.Stop(); err != nil {
		app.Logger.Error().Err(err).Msg("Failed to stop cache refresh service")
	}

	// Stop HTTP server
	return app.Server.Stop(ctx)
}
