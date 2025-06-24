package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rule-engine/internal/config"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// New creates a new logger instance with the provided configuration
func New(cfg config.LoggerConfig) (*Logger, error) {
	// Set global log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output writer
	var writer io.Writer = os.Stdout

	// If LogPath is specified, use file-based logging
	if cfg.LogPath != "" {
		file, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", cfg.LogPath, err)
		}
		writer = file
	}

	// Configure output format
	if strings.ToLower(cfg.Format) == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: getTimeFormat(cfg.TimeFormat),
		}
	}

	// Create logger
	logger := zerolog.New(writer)

	// Add timestamp
	logger = logger.With().Timestamp().Logger()

	// Add caller information if enabled
	if cfg.EnableCaller {
		logger = logger.With().Caller().Logger()
	}

	// Set as global logger
	log.Logger = logger

	return &Logger{
		Logger: &logger,
	}, nil
}

// getTimeFormat returns the appropriate time format
func getTimeFormat(format string) string {
	if format == "" {
		return time.RFC3339
	}
	return format
}

// LogErrorWithStackTrace logs an error with stack trace
func LogErrorWithStackTrace(err error, message string, fields ...map[string]interface{}) {
	event := log.Error().
		Str("error", err.Error()).
		Str("stack_trace", getStackTrace())

	// Add additional fields if provided
	if len(fields) > 0 {
		for key, value := range fields[0] {
			switch v := value.(type) {
			case string:
				event = event.Str(key, v)
			case int:
				event = event.Int(key, v)
			case bool:
				event = event.Bool(key, v)
			default:
				event = event.Interface(key, v)
			}
		}
	}

	event.Msg(message)
}

// LogErrorWithContext logs an error with context information
func LogErrorWithContext(err error, message string, context map[string]interface{}) {
	event := log.Error().
		Str("error", err.Error()).
		Str("stack_trace", getStackTrace())

	// Add context fields
	for key, value := range context {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case bool:
			event = event.Bool(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	event.Msg(message)
}

// LogDatabaseError logs database-specific errors with additional context
func LogDatabaseError(err error, operation string, table string, query string) {
	log.Error().
		Str("error", err.Error()).
		Str("operation", operation).
		Str("table", table).
		Str("query", query).
		Str("stack_trace", getStackTrace()).
		Msg("Database operation failed")
}

// LogHTTPError logs HTTP errors with request context
func LogHTTPError(err error, method string, path string, statusCode int, clientIP string) {
	log.Error().
		Str("error", err.Error()).
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Str("client_ip", clientIP).
		Str("stack_trace", getStackTrace()).
		Msg("HTTP request failed")
}

// LogValidationError logs validation errors with field details
func LogValidationError(err error, field string, value interface{}, rule string) {
	log.Error().
		Str("error", err.Error()).
		Str("field", field).
		Interface("value", value).
		Str("rule", rule).
		Str("stack_trace", getStackTrace()).
		Msg("Validation failed")
}

// LogBusinessLogicError logs business logic errors with operation context
func LogBusinessLogicError(err error, operation string, entity string, entityID string) {
	log.Error().
		Str("error", err.Error()).
		Str("operation", operation).
		Str("entity", entity).
		Str("entity_id", entityID).
		Str("stack_trace", getStackTrace()).
		Msg("Business logic error")
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(event string, userID string, action string, resource string, clientIP string) {
	log.Warn().
		Str("event", event).
		Str("user_id", userID).
		Str("action", action).
		Str("resource", resource).
		Str("client_ip", clientIP).
		Str("stack_trace", getStackTrace()).
		Msg("Security event")
}

// LogPerformanceEvent logs performance-related events
func LogPerformanceEvent(operation string, duration time.Duration, resource string, details map[string]interface{}) {
	event := log.Info().
		Str("operation", operation).
		Dur("duration", duration).
		Str("resource", resource)

	// Add additional details
	for key, value := range details {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case bool:
			event = event.Bool(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	event.Msg("Performance event")
}

// getStackTrace returns the current stack trace as a string
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// FormatErrorWithStackTrace formats an error with stack trace for logging
func FormatErrorWithStackTrace(err error) string {
	return fmt.Sprintf("Error: %s\nStack Trace:\n%s", err.Error(), getStackTrace())
}
