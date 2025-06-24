package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LoggingMiddleware provides comprehensive request/response logging with error handling
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log request details
		log.Info().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Str("client_ip", param.ClientIP).
			Str("user_agent", param.Request.UserAgent()).
			Time("timestamp", param.TimeStamp).
			Msg("HTTP Request")

		// Log error details if status code indicates error
		if param.StatusCode >= 400 {
			log.Error().
				Str("method", param.Method).
				Str("path", param.Path).
				Int("status", param.StatusCode).
				Dur("latency", param.Latency).
				Str("client_ip", param.ClientIP).
				Str("error", param.ErrorMessage).
				Msg("HTTP Request Error")
		}

		return ""
	})
}

// ErrorRecoveryMiddleware recovers from panics and logs stack traces
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Error().
				Str("error", err).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Str("stack_trace", getStackTrace()).
				Msg("Panic recovered - string error")
		} else if err, ok := recovered.(error); ok {
			log.Error().
				Str("error", err.Error()).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Str("stack_trace", getStackTrace()).
				Msg("Panic recovered - error type")
		} else {
			log.Error().
				Str("error", fmt.Sprintf("%v", recovered)).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Str("stack_trace", getStackTrace()).
				Msg("Panic recovered - unknown type")
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}

// RequestLoggingMiddleware logs request body for debugging (only in development)
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log request body for POST/PUT/PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// Read the request body
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Error().
					Str("error", err.Error()).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("stack_trace", getStackTrace()).
					Msg("Failed to read request body")
			} else {
				// Log request body (truncate if too long)
				bodyStr := string(body)
				if len(bodyStr) > 1000 {
					bodyStr = bodyStr[:1000] + "... (truncated)"
				}

				log.Info().
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("body", bodyStr).
					Msg("Request body")

				// Restore the request body
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		c.Next()
	}
}

// ResponseLoggingMiddleware logs response details
func ResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Log response details
		statusCode := c.Writer.Status()
		responseBody := blw.body.String()

		// Truncate response body if too long
		if len(responseBody) > 1000 {
			responseBody = responseBody[:1000] + "... (truncated)"
		}

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", statusCode).
			Str("response_body", responseBody).
			Msg("HTTP Response")

		// Log error responses with more detail
		if statusCode >= 400 {
			log.Error().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", statusCode).
				Str("response_body", responseBody).
				Str("client_ip", c.ClientIP()).
				Msg("HTTP Response Error")
		}
	}
}

// ContextLoggingMiddleware adds request ID and logging context
func ContextLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := generateRequestID()

		// Add request ID to context
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)

		// Log request start
		log.Info().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("Request started")

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Log request completion
		log.Info().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Msg("Request completed")
	}
}

// DatabaseErrorMiddleware logs database errors with stack traces
func DatabaseErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Error().
					Str("error", err.Error()).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("client_ip", c.ClientIP()).
					Str("stack_trace", getStackTrace()).
					Msg("Request error")
			}
		}
	}
}

// DebugErrorMiddleware logs detailed error information for debugging
func DebugErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("stack_trace", getStackTrace()).
					Msg("DEBUG: Panic caught in debug middleware")
				panic(r) // Re-panic to let the recovery middleware handle it
			}
		}()

		c.Next()

		// Check for errors after processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Error().
					Str("error", err.Error()).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("DEBUG: Error found in context")
			}
		}

		// Log response status if it's an error
		if c.Writer.Status() >= 400 {
			log.Error().
				Int("status", c.Writer.Status()).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Msg("DEBUG: Error response status")
		}
	}
}

// Helper functions

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}
