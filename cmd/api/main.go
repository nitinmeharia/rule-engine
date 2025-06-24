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

// Version is set at build time using -ldflags
var Version = "dev"

const (
	// Application metadata
	appName = "rule-engine"
)

func main() {
	// Initialize logger with basic configuration
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	log.Info().
		Str("app", appName).
		Str("version", Version).
		Msg("Starting Generic Rule Engine")

	// Initialize application
	app, err := bootstrap.Init(Version)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to initialize application")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application components
	appErrChan := make(chan error, 1)
	go func() {
		if err := app.Start(ctx); err != nil {
			appErrChan <- fmt.Errorf("application failed to start: %w", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for either application error or shutdown signal
	select {
	case err := <-appErrChan:
		log.Error().
			Err(err).
			Msg("Application error occurred")
		os.Exit(1)

	case sig := <-signalChan:
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")

		// Initiate graceful shutdown
		gracefulShutdown(app, cancel)
	}
}

// gracefulShutdown handles cleanup and shutdown procedures
func gracefulShutdown(app *bootstrap.Application, cancel context.CancelFunc) {
	log.Info().Msg("Starting graceful shutdown")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Cancel the main context
	cancel()

	// Stop application components
	log.Info().Msg("Stopping application components")
	if err := app.Stop(shutdownCtx); err != nil {
		log.Error().
			Err(err).
			Msg("Error during application shutdown")
	}

	// Close database connection
	log.Info().Msg("Closing database connection")
	app.DB.Close()

	log.Info().Msg("Graceful shutdown completed")
}

// To set the version at build time, use:
// go build -ldflags "-X main.Version=$(git describe --tags --always --dirty)"
