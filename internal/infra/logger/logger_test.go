package logger

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rule-engine/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DefaultConfig(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:        "info",
		Format:       "json",
		EnableCaller: false,
		TimeFormat:   "",
		LogPath:      "",
	}

	logger, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
}

func TestNew_ConsoleFormat(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:        "debug",
		Format:       "console",
		EnableCaller: true,
		TimeFormat:   "2006-01-02 15:04:05",
		LogPath:      "",
	}

	logger, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNew_InvalidLevel(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:        "invalid_level",
		Format:       "json",
		EnableCaller: false,
		TimeFormat:   "",
		LogPath:      "",
	}

	logger, err := New(cfg)
	require.NoError(t, err) // Should not error, should default to InfoLevel
	assert.NotNil(t, logger)
}

func TestNew_FileLogging(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	cfg := config.LoggerConfig{
		Level:        "info",
		Format:       "json",
		EnableCaller: false,
		TimeFormat:   "",
		LogPath:      tmpFile.Name(),
	}

	logger, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// Test that we can write to the log file
	logger.Info().Msg("test message")

	// Read the file to verify content
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}

func TestNew_InvalidLogPath(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:        "info",
		Format:       "json",
		EnableCaller: false,
		TimeFormat:   "",
		LogPath:      "/invalid/path/that/does/not/exist/test.log",
	}

	logger, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, logger)
	assert.Contains(t, err.Error(), "failed to open log file")
}

func TestGetTimeFormat_Default(t *testing.T) {
	format := getTimeFormat("")
	assert.Equal(t, time.RFC3339, format)
}

func TestGetTimeFormat_Custom(t *testing.T) {
	customFormat := "2006-01-02 15:04:05"
	format := getTimeFormat(customFormat)
	assert.Equal(t, customFormat, format)
}

func TestLogErrorWithStackTrace(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("test error")
	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "test_action",
	}

	LogErrorWithStackTrace(err, "test error message", fields)

	output := buf.String()
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "user_id")
	assert.Contains(t, output, "action")
	assert.Contains(t, output, "stack_trace")
}

func TestLogErrorWithContext(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("test error")
	context := map[string]interface{}{
		"request_id": "req-123",
		"method":     "GET",
		"path":       "/api/test",
	}

	LogErrorWithContext(err, "test context error", context)

	output := buf.String()
	assert.Contains(t, output, "test context error")
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "request_id")
	assert.Contains(t, output, "method")
	assert.Contains(t, output, "path")
	assert.Contains(t, output, "stack_trace")
}

func TestLogDatabaseError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("database connection failed")
	LogDatabaseError(err, "SELECT", "users", "SELECT * FROM users WHERE id = 1")

	output := buf.String()
	assert.Contains(t, output, "Database operation failed")
	assert.Contains(t, output, "database connection failed")
	assert.Contains(t, output, "SELECT")
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "SELECT * FROM users WHERE id = 1")
	assert.Contains(t, output, "stack_trace")
}

func TestLogHTTPError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("connection timeout")
	LogHTTPError(err, "POST", "/api/users", 500, "192.168.1.1")

	output := buf.String()
	assert.Contains(t, output, "HTTP request failed")
	assert.Contains(t, output, "connection timeout")
	assert.Contains(t, output, "POST")
	assert.Contains(t, output, "/api/users")
	assert.Contains(t, output, "500")
	assert.Contains(t, output, "192.168.1.1")
	assert.Contains(t, output, "stack_trace")
}

func TestLogValidationError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("invalid email format")
	LogValidationError(err, "email", "invalid-email", "email_format")

	output := buf.String()
	assert.Contains(t, output, "Validation failed")
	assert.Contains(t, output, "invalid email format")
	assert.Contains(t, output, "email")
	assert.Contains(t, output, "invalid-email")
	assert.Contains(t, output, "email_format")
	assert.Contains(t, output, "stack_trace")
}

func TestLogBusinessLogicError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("insufficient permissions")
	LogBusinessLogicError(err, "create", "user", "user-123")

	output := buf.String()
	assert.Contains(t, output, "Business logic error")
	assert.Contains(t, output, "insufficient permissions")
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "user")
	assert.Contains(t, output, "user-123")
	assert.Contains(t, output, "stack_trace")
}

func TestLogSecurityEvent(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	LogSecurityEvent("failed_login", "user-456", "login", "/auth/login", "10.0.0.1")

	output := buf.String()
	assert.Contains(t, output, "Security event")
	assert.Contains(t, output, "failed_login")
	assert.Contains(t, output, "user-456")
	assert.Contains(t, output, "login")
	assert.Contains(t, output, "/auth/login")
	assert.Contains(t, output, "10.0.0.1")
	assert.Contains(t, output, "stack_trace")
}

func TestLogPerformanceEvent(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	details := map[string]interface{}{
		"rows_affected": 100,
		"cache_hit":     true,
		"query_type":    "SELECT",
	}

	LogPerformanceEvent("database_query", 150*time.Millisecond, "users_table", details)

	output := buf.String()
	assert.Contains(t, output, "Performance event")
	assert.Contains(t, output, "database_query")
	assert.Contains(t, output, "users_table")
	assert.Contains(t, output, "rows_affected")
	assert.Contains(t, output, "cache_hit")
	assert.Contains(t, output, "query_type")
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := getStackTrace()
	assert.NotEmpty(t, stackTrace)
	assert.Contains(t, stackTrace, "TestGetStackTrace")
	assert.Contains(t, stackTrace, "logger_test.go")
}

func TestFormatErrorWithStackTrace(t *testing.T) {
	err := errors.New("test formatting error")
	formatted := FormatErrorWithStackTrace(err)

	assert.Contains(t, formatted, "Error: test formatting error")
	assert.Contains(t, formatted, "Stack Trace:")
	assert.Contains(t, formatted, "TestFormatErrorWithStackTrace")
}

func TestLogErrorWithStackTrace_NoFields(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	err := errors.New("simple error")
	LogErrorWithStackTrace(err, "simple error message")

	output := buf.String()
	assert.Contains(t, output, "simple error message")
	assert.Contains(t, output, "simple error")
	assert.Contains(t, output, "stack_trace")
}

func TestLogPerformanceEvent_EmptyDetails(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)
	log.Logger = testLogger

	LogPerformanceEvent("simple_operation", 50*time.Millisecond, "test_resource", nil)

	output := buf.String()
	assert.Contains(t, output, "Performance event")
	assert.Contains(t, output, "simple_operation")
	assert.Contains(t, output, "test_resource")
}
