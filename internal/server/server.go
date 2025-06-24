package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/handlers"
	"github.com/rule-engine/internal/infra/logger"
	"github.com/rule-engine/internal/models/db"
	"github.com/rule-engine/internal/repository"
	"github.com/rule-engine/internal/service"
)

// Server represents the HTTP server
type Server struct {
	httpServer       *http.Server
	router           *gin.Engine
	config           *config.Config
	db               *pgxpool.Pool
	logger           *logger.Logger
	namespaceHandler *handlers.NamespaceHandler
	fieldHandler     *handlers.FieldHandler
	functionHandler  *handlers.FunctionHandler
	ruleHandler      *handlers.RuleHandler
	address          string
}

// New creates a new HTTP server instance
func New(cfg *config.Config, database *pgxpool.Pool, log *logger.Logger) (*Server, error) {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add comprehensive logging and error handling middleware
	router.Use(
		DebugErrorMiddleware(),      // Debug error catching
		ErrorRecoveryMiddleware(),   // Panic recovery with stack traces
		ContextLoggingMiddleware(),  // Request ID and context logging
		RequestLoggingMiddleware(),  // Request body logging
		ResponseLoggingMiddleware(), // Response logging
		DatabaseErrorMiddleware(),   // Database error logging
		JWTMiddleware(&cfg.JWT),     // JWT authentication
		LoggingMiddleware(),         // Standard HTTP logging
	)

	// Initialize database queries
	queries := db.New(database)

	// Initialize repositories
	namespaceRepo := repository.NewNamespaceRepository(queries)
	fieldRepo := repository.NewFieldRepository(queries)
	functionRepo := repository.NewFunctionRepository(queries)
	ruleRepo := repository.NewRuleRepository(queries)

	// Initialize services
	namespaceService := service.NewNamespaceService(namespaceRepo)
	fieldService := service.NewFieldService(fieldRepo)
	functionService := service.NewFunctionService(functionRepo)
	ruleService := service.NewRuleService(ruleRepo, functionRepo, fieldRepo)

	// Initialize handlers
	namespaceHandler := handlers.NewNamespaceHandler(namespaceService)
	fieldHandler := handlers.NewFieldHandler(fieldService)
	functionHandler := handlers.NewFunctionHandler(functionService)
	ruleHandler := handlers.NewRuleHandler(ruleService)

	// Create server instance first
	server := &Server{
		router:           router,
		config:           cfg,
		db:               database,
		logger:           log,
		namespaceHandler: namespaceHandler,
		fieldHandler:     fieldHandler,
		functionHandler:  functionHandler,
		ruleHandler:      ruleHandler,
		address:          cfg.Server.GetServerAddress(),
	}

	// Health check endpoint (no authentication required)
	router.GET("/health", server.healthCheck)

	// API v1 routes with authentication
	v1 := router.Group("/v1")
	{
		// Namespace routes - require admin role for write operations
		namespaces := v1.Group("/namespaces")
		{
			namespaces.GET("", RequireAnyRole("admin", "viewer", "executor"), namespaceHandler.ListNamespaces)
			namespaces.POST("", RequireRole("admin"), namespaceHandler.CreateNamespace)
			namespaces.GET("/:id", RequireAnyRole("admin", "viewer", "executor"), namespaceHandler.GetNamespace)
			namespaces.DELETE("/:id", RequireRole("admin"), namespaceHandler.DeleteNamespace)

			// Field routes within namespace - nested under namespace ID
			namespaces.GET("/:id/fields", RequireAnyRole("admin", "viewer", "executor"), fieldHandler.ListFields)
			namespaces.POST("/:id/fields", RequireRole("admin"), fieldHandler.CreateField)

			// Function routes within namespace - nested under namespace ID
			namespaces.GET("/:id/functions", RequireAnyRole("admin", "viewer", "executor"), functionHandler.ListFunctions)
			namespaces.POST("/:id/functions", RequireRole("admin"), functionHandler.CreateFunction)
			namespaces.GET("/:id/functions/:functionId", RequireAnyRole("admin", "viewer", "executor"), functionHandler.GetFunction)
			namespaces.PUT("/:id/functions/:functionId/versions/draft", RequireRole("admin"), functionHandler.UpdateFunction)
			namespaces.POST("/:id/functions/:functionId/publish", RequireRole("admin"), functionHandler.PublishFunction)

			// Rule routes within namespace - nested under namespace ID
			namespaces.GET("/:id/rules", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.ListRules)
			namespaces.POST("/:id/rules", RequireRole("admin"), ruleHandler.CreateRule)
			namespaces.GET("/:id/rules/:ruleId", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.GetRule)
			namespaces.GET("/:id/rules/:ruleId/versions/draft", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.GetDraftRule)
			namespaces.PUT("/:id/rules/:ruleId/versions/draft", RequireRole("admin"), ruleHandler.UpdateRule)
			namespaces.POST("/:id/rules/:ruleId/publish", RequireRole("admin"), ruleHandler.PublishRule)
			namespaces.GET("/:id/rules/:ruleId/history", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.ListRuleVersions)
			namespaces.DELETE("/:id/rules/:ruleId/versions/:version", RequireRole("admin"), ruleHandler.DeleteRule)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	server.httpServer = srv

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting HTTP server")

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Stopping HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/v1")
	{
		// Namespace routes
		namespaces := v1.Group("/namespaces")
		{
			namespaces.POST("", s.namespaceHandler.CreateNamespace)
			namespaces.GET("", s.namespaceHandler.ListNamespaces)
			namespaces.GET("/:id", s.namespaceHandler.GetNamespace)
			namespaces.DELETE("/:id", s.namespaceHandler.DeleteNamespace)
		}

		// TODO: Add other resource routes (fields, functions, rules, workflows, terminals)
	}
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0", // TODO: Get from build info
	})
}

// ServeHTTP implements http.Handler interface for testing
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
