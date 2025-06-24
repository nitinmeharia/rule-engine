package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/execution"
	"github.com/rule-engine/internal/handlers"
	"github.com/rule-engine/internal/infra"
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
	terminalHandler  *handlers.TerminalHandler
	workflowHandler  *handlers.WorkflowHandler
	executionHandler *handlers.ExecutionHandler
	adminHandler     *handlers.AdminHandler
	address          string
}

// New creates a new HTTP server instance
func New(cfg *config.Config, database *pgxpool.Pool, log *logger.Logger, engine *execution.Engine) (*Server, error) {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Register Prometheus metrics
	infra.RegisterMetrics()

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
	terminalRepo := repository.NewTerminalRepository(queries)
	workflowRepo := repository.NewWorkflowRepository(queries)

	// Initialize services
	namespaceService := service.NewNamespaceService(namespaceRepo)
	fieldService := service.NewFieldService(fieldRepo)
	functionService := service.NewFunctionService(functionRepo, namespaceRepo)
	ruleService := service.NewRuleService(ruleRepo, functionRepo, fieldRepo, namespaceRepo)
	terminalService := service.NewTerminalService(terminalRepo, namespaceRepo)
	workflowService := service.NewWorkflowService(workflowRepo, ruleRepo, terminalRepo, namespaceRepo)

	// Initialize handlers
	namespaceHandler := handlers.NewNamespaceHandler(namespaceService)
	fieldHandler := handlers.NewFieldHandler(fieldService)
	functionHandler := handlers.NewFunctionHandler(functionService)
	ruleHandler := handlers.NewRuleHandler(ruleService)
	terminalHandler := handlers.NewTerminalHandler(terminalService)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)
	executionHandler := handlers.NewExecutionHandler(engine)
	adminHandler := handlers.NewAdminHandler(engine)

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
		terminalHandler:  terminalHandler,
		workflowHandler:  workflowHandler,
		executionHandler: executionHandler,
		adminHandler:     adminHandler,
		address:          cfg.Server.GetServerAddress(),
	}

	// Health check endpoint (no authentication required)
	router.GET("/health", server.healthCheck)

	// Metrics endpoint (no authentication required)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
			namespaces.DELETE("/:id/functions/:functionId/versions/:version", RequireRole("admin"), functionHandler.DeleteFunction)

			// Rule routes within namespace - nested under namespace ID
			namespaces.GET("/:id/rules", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.ListRules)
			namespaces.POST("/:id/rules", RequireRole("admin"), ruleHandler.CreateRule)
			namespaces.GET("/:id/rules/:ruleId", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.GetRule)
			namespaces.GET("/:id/rules/:ruleId/versions/draft", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.GetDraftRule)
			namespaces.PUT("/:id/rules/:ruleId/versions/draft", RequireRole("admin"), ruleHandler.UpdateRule)
			namespaces.POST("/:id/rules/:ruleId/publish", RequireRole("admin"), ruleHandler.PublishRule)
			namespaces.GET("/:id/rules/:ruleId/history", RequireAnyRole("admin", "viewer", "executor"), ruleHandler.ListRuleVersions)
			namespaces.DELETE("/:id/rules/:ruleId/versions/:version", RequireRole("admin"), ruleHandler.DeleteRule)

			// Terminal routes within namespace - nested under namespace ID
			namespaces.GET("/:id/terminals", RequireAnyRole("admin", "viewer", "executor"), terminalHandler.ListTerminals)
			namespaces.POST("/:id/terminals", RequireRole("admin"), terminalHandler.CreateTerminal)
			namespaces.GET("/:id/terminals/:terminalId", RequireAnyRole("admin", "viewer", "executor"), terminalHandler.GetTerminal)
			namespaces.DELETE("/:id/terminals/:terminalId", RequireRole("admin"), terminalHandler.DeleteTerminal)

			// Workflow routes within namespace - nested under namespace ID
			namespaces.GET("/:id/workflows", RequireAnyRole("admin", "viewer", "executor"), workflowHandler.ListWorkflows)
			namespaces.POST("/:id/workflows", RequireRole("admin"), workflowHandler.CreateWorkflow)
			namespaces.GET("/:id/workflows/:workflowId", RequireAnyRole("admin", "viewer", "executor"), workflowHandler.GetWorkflow)
			namespaces.GET("/:id/workflows/:workflowId/versions/:version", RequireAnyRole("admin", "viewer", "executor"), workflowHandler.GetWorkflowVersion)
			namespaces.PUT("/:id/workflows/:workflowId/versions/:version", RequireRole("admin"), workflowHandler.UpdateWorkflow)
			namespaces.POST("/:id/workflows/:workflowId/versions/:version/publish", RequireRole("admin"), workflowHandler.PublishWorkflow)
			namespaces.POST("/:id/workflows/:workflowId/deactivate", RequireRole("admin"), workflowHandler.DeactivateWorkflow)
			namespaces.DELETE("/:id/workflows/:workflowId/versions/:version", RequireRole("admin"), workflowHandler.DeleteWorkflow)
			namespaces.GET("/:id/workflows/active", RequireAnyRole("admin", "viewer", "executor"), workflowHandler.ListActiveWorkflows)
			namespaces.GET("/:id/workflows/:workflowId/versions", RequireAnyRole("admin", "viewer", "executor"), workflowHandler.ListWorkflowVersions)
		}

		// Execution API
		execute := v1.Group("/execute")
		{
			execute.POST("/namespaces/:namespace/workflows/:workflowId", RequireRole("executor"), executionHandler.ExecuteWorkflow)
		}
	}

	// Admin routes (require admin role)
	admin := router.Group("/admin")
	{
		admin.GET("/cache/stats/:namespace", RequireRole("admin"), adminHandler.GetCacheStats)
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
