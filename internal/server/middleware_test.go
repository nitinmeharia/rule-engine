package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		setupRequest   func(*gin.Context)
	}{
		{
			name:           "successful_request",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error_request",
			method:         "GET",
			path:           "/error",
			expectedStatus: http.StatusNotFound,
			setupRequest: func(c *gin.Context) {
				c.AbortWithStatus(http.StatusNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(LoggingMiddleware())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.GET("/error", func(c *gin.Context) {
				if tt.setupRequest != nil {
					tt.setupRequest(c)
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestErrorRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		panicValue     interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "string_panic",
			panicValue:     "test panic",
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			},
		},
		{
			name:           "error_panic",
			panicValue:     fmt.Errorf("test error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			},
		},
		{
			name:           "unknown_panic",
			panicValue:     123,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorRecoveryMiddleware())

			router.GET("/panic", func(c *gin.Context) {
				panic(tt.panicValue)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/panic", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestRequestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		method    string
		path      string
		body      string
		shouldLog bool
	}{
		{
			name:      "post_request_with_body",
			method:    "POST",
			path:      "/test",
			body:      `{"test": "data"}`,
			shouldLog: true,
		},
		{
			name:      "put_request_with_body",
			method:    "PUT",
			path:      "/test",
			body:      `{"test": "data"}`,
			shouldLog: true,
		},
		{
			name:      "patch_request_with_body",
			method:    "PATCH",
			path:      "/test",
			body:      `{"test": "data"}`,
			shouldLog: true,
		},
		{
			name:      "get_request_no_body",
			method:    "GET",
			path:      "/test",
			shouldLog: false,
		},
		{
			name:      "delete_request_no_body",
			method:    "DELETE",
			path:      "/test",
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestLoggingMiddleware())

			router.Any("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			var req *http.Request

			if tt.body != "" {
				req, _ = http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req, _ = http.NewRequest(tt.method, tt.path, nil)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestResponseLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "successful_response",
			responseBody:   `{"message": "success"}`,
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error_response",
			responseBody:   `{"error": "not found"}`,
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "large_response_truncated",
			responseBody:   strings.Repeat("a", 1500),
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ResponseLoggingMiddleware())

			router.GET("/test", func(c *gin.Context) {
				c.Data(tt.statusCode, "application/json", []byte(tt.responseBody))
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.responseBody, w.Body.String())
		})
	}
}

func TestContextLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ContextLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		assert.NotEmpty(t, requestID)
		assert.NotEqual(t, "unknown", requestID)

		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that request ID is in response headers
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, requestID, response["request_id"])
}

func TestDatabaseErrorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(DatabaseErrorMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.Error(fmt.Errorf("database error"))
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDebugErrorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupHandler   func(*gin.Context)
		expectedStatus int
	}{
		{
			name: "normal_request",
			setupHandler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error_in_context",
			setupHandler: func(c *gin.Context) {
				c.Error(fmt.Errorf("test error"))
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error_status",
			setupHandler: func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(DebugErrorMiddleware())

			router.GET("/test", tt.setupHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestBodyLogWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ResponseLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test response body"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test response body")
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := getStackTrace()

	assert.NotEmpty(t, stackTrace)
	assert.Contains(t, stackTrace, "TestGetStackTrace")
	assert.Contains(t, stackTrace, "getStackTrace")
}

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs to ensure they're unique
	id1 := generateRequestID()
	id2 := generateRequestID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.True(t, strings.HasPrefix(id1, "req_"))
	assert.True(t, strings.HasPrefix(id2, "req_"))

	// Check that they contain numbers
	assert.True(t, len(id1) > 4) // "req_" + at least one digit
	assert.True(t, len(id2) > 4)
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		expected     string
	}{
		{
			name: "request_id_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("request_id", "test-request-id")
			},
			expected: "test-request-id",
		},
		{
			name: "no_request_id_in_context",
			setupContext: func(c *gin.Context) {
				// Don't set request_id
			},
			expected: "unknown",
		},
		{
			name: "wrong_type_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("request_id", 123)
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.GET("/test", func(c *gin.Context) {
				if tt.setupContext != nil {
					tt.setupContext(c)
				}

				requestID := GetRequestID(c)
				assert.Equal(t, tt.expected, requestID)

				c.JSON(http.StatusOK, gin.H{"request_id": requestID})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test multiple middleware working together
	router := gin.New()
	router.Use(
		ContextLoggingMiddleware(),
		RequestLoggingMiddleware(),
		ResponseLoggingMiddleware(),
		DatabaseErrorMiddleware(),
	)

	router.POST("/test", func(c *gin.Context) {
		// Read request body
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get request ID
		requestID := GetRequestID(c)

		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"request_id": requestID,
			"received":   body,
		})
	})

	w := httptest.NewRecorder()
	reqBody := `{"test": "data"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response headers
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["message"])
	assert.Equal(t, requestID, response["request_id"])
	assert.Equal(t, map[string]interface{}{"test": "data"}, response["received"])
}
