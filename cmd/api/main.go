package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rule-engine/internal/bootstrap"
)

const (
	// Application metadata
	appName    = "rule-engine"
	appVersion = "1.0.0"
)

func main() {
	// Initialize logger with basic configuration
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	log.Info().
		Str("app", appName).
		Str("version", appVersion).
		Msg("Starting Generic Rule Engine")

	// Initialize application
	app, err := bootstrap.Init()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to initialize application")
	}

	// Start HTTP server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Info().
			Int("port", app.Config.Server.Port).
			Msg("Starting HTTP server")

		if err := app.Server.Start(); err != nil {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Start cache refresh loop in a goroutine
	cacheCtx, cancelCache := context.WithCancel(context.Background())
	go func() {
		log.Info().
			Int("interval_seconds", app.Config.Cache.RefreshIntervalSec).
			Msg("Starting cache refresh loop")

		app.ExecutionService.StartRefreshLoop(cacheCtx)
	}()

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrChan:
		log.Error().
			Err(err).
			Msg("Server error occurred")
		os.Exit(1)

	case sig := <-signalChan:
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")

		// Initiate graceful shutdown
		gracefulShutdown(app, cancelCache)
	}
}

// gracefulShutdown handles cleanup and shutdown procedures
func gracefulShutdown(app *bootstrap.Application, cancelCache context.CancelFunc) {
	log.Info().Msg("Starting graceful shutdown")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop cache refresh loop
	log.Info().Msg("Stopping cache refresh loop")
	cancelCache()

	// Stop HTTP server
	log.Info().Msg("Stopping HTTP server")
	if err := app.Server.Stop(shutdownCtx); err != nil {
		log.Error().
			Err(err).
			Msg("Error during server shutdown")
	}

	// Close database connection
	log.Info().Msg("Closing database connection")
	app.DB.Close()

	log.Info().Msg("Graceful shutdown completed")
}
