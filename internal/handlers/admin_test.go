package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEngine is a mock implementation of the execution engine
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) GetCacheInfo(namespace string) (*execution.CacheInfo, error) {
	args := m.Called(namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.CacheInfo), args.Error(1)
}

func (m *MockEngine) ExecuteRule(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ExecutionResponse), args.Error(1)
}

func (m *MockEngine) ExecuteWorkflow(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ExecutionResponse), args.Error(1)
}

func setupAdminTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	return c, w
}

func TestNewAdminHandler(t *testing.T) {
	mockEngine := new(MockEngine)
	handler := NewAdminHandler(mockEngine)

	assert.NotNil(t, handler)
	assert.Equal(t, mockEngine, handler.engine)
	assert.NotNil(t, handler.responseHandler)
}

func TestAdminHandler_GetCacheStats(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockEngine)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful cache stats - fresh cache",
			namespace: "test-ns",
			setupMock: func(mockEngine *MockEngine) {
				cacheInfo := &execution.CacheInfo{
					Checksum:    "abc123",
					LastRefresh: time.Now().Add(-time.Minute), // 1 minute ago
				}
				mockEngine.On("GetCacheInfo", "test-ns").Return(cacheInfo, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"namespace":        "test-ns",
					"checksum":         "abc123",
					"cacheStatus":      "fresh",
					"isStale":          false,
					"stalenessSeconds": mock.AnythingOfType("float64"),
				},
			},
		},
		{
			name:      "successful cache stats - stale cache",
			namespace: "test-ns",
			setupMock: func(mockEngine *MockEngine) {
				cacheInfo := &execution.CacheInfo{
					Checksum:    "abc123",
					LastRefresh: time.Now().Add(-10 * time.Minute), // 10 minutes ago
				}
				mockEngine.On("GetCacheInfo", "test-ns").Return(cacheInfo, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"namespace":        "test-ns",
					"checksum":         "abc123",
					"cacheStatus":      "stale",
					"isStale":          true,
					"stalenessSeconds": mock.AnythingOfType("float64"),
				},
			},
		},
		{
			name:      "successful cache stats - not initialized",
			namespace: "test-ns",
			setupMock: func(mockEngine *MockEngine) {
				cacheInfo := &execution.CacheInfo{
					Checksum:    "abc123",
					LastRefresh: time.Time{}, // Zero time
				}
				mockEngine.On("GetCacheInfo", "test-ns").Return(cacheInfo, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"namespace":        "test-ns",
					"checksum":         "abc123",
					"cacheStatus":      "not_initialized",
					"isStale":          true,
					"stalenessSeconds": mock.AnythingOfType("float64"),
				},
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			setupMock:      func(mockEngine *MockEngine) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace is required",
			},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("GetCacheInfo", "non-existent-ns").Return(nil, domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:      "engine error",
			namespace: "test-ns",
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("GetCacheInfo", "test-ns").Return(nil, errors.New("engine error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":    "INTERNAL_SERVER_ERROR",
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockEngine)
			handler := NewAdminHandler(mockEngine)
			tt.setupMock(mockEngine)

			c, w := setupAdminTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "namespace", Value: tt.namespace}}

			// Create request
			req := httptest.NewRequest("GET", "/admin/cache/"+tt.namespace+"/stats", nil)
			c.Request = req

			handler.GetCacheStats(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			// Check basic fields
			if tt.expectedBody["error"] != nil {
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}
			if tt.expectedBody["code"] != nil {
				assert.Equal(t, tt.expectedBody["code"], response["code"])
			}
			if tt.expectedBody["message"] != nil {
				assert.Equal(t, tt.expectedBody["message"], response["message"])
			}

			// Check success fields if present
			if success, exists := tt.expectedBody["success"]; exists {
				assert.Equal(t, success, response["success"])
			}

			// Check data fields if present
			if data, exists := tt.expectedBody["data"]; exists {
				assert.IsType(t, map[string]interface{}{}, response["data"])
				responseData := response["data"].(map[string]interface{})
				expectedData := data.(map[string]interface{})

				// Check specific fields
				if namespace, exists := expectedData["namespace"]; exists {
					assert.Equal(t, namespace, responseData["namespace"])
				}
				if checksum, exists := expectedData["checksum"]; exists {
					assert.Equal(t, checksum, responseData["checksum"])
				}
				if cacheStatus, exists := expectedData["cacheStatus"]; exists {
					assert.Equal(t, cacheStatus, responseData["cacheStatus"])
				}
				if isStale, exists := expectedData["isStale"]; exists {
					assert.Equal(t, isStale, responseData["isStale"])
				}
				if _, exists := expectedData["stalenessSeconds"]; exists {
					assert.IsType(t, float64(0), responseData["stalenessSeconds"])
				}
			}

			mockEngine.AssertExpectations(t)
		})
	}
}

func TestAdminHandler_GetCacheStatus(t *testing.T) {
	tests := []struct {
		name           string
		lastRefresh    time.Time
		isStale        bool
		expectedStatus string
	}{
		{
			name:           "not initialized",
			lastRefresh:    time.Time{},
			isStale:        true,
			expectedStatus: "not_initialized",
		},
		{
			name:           "stale cache",
			lastRefresh:    time.Now().Add(-10 * time.Minute),
			isStale:        true,
			expectedStatus: "stale",
		},
		{
			name:           "fresh cache",
			lastRefresh:    time.Now().Add(-time.Minute),
			isStale:        false,
			expectedStatus: "fresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockEngine)
			handler := NewAdminHandler(mockEngine)

			cacheInfo := &execution.CacheInfo{
				Checksum:    "test-checksum",
				LastRefresh: tt.lastRefresh,
			}

			status := handler.getCacheStatus(cacheInfo, tt.isStale)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestAdminHandler_EdgeCases(t *testing.T) {
	t.Run("cache info with zero time", func(t *testing.T) {
		mockEngine := new(MockEngine)
		handler := NewAdminHandler(mockEngine)

		c, w := setupAdminTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "namespace", Value: "test-ns"}}

		// Setup mock with zero time
		cacheInfo := &execution.CacheInfo{
			Checksum:    "abc123",
			LastRefresh: time.Time{}, // Zero time
		}
		mockEngine.On("GetCacheInfo", "test-ns").Return(cacheInfo, nil)

		// Create request
		req := httptest.NewRequest("GET", "/admin/cache/test-ns/stats", nil)
		c.Request = req

		handler.GetCacheStats(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["success"])
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "not_initialized", data["cacheStatus"])
		assert.Equal(t, true, data["isStale"])

		mockEngine.AssertExpectations(t)
	})

	t.Run("very old cache", func(t *testing.T) {
		mockEngine := new(MockEngine)
		handler := NewAdminHandler(mockEngine)

		c, w := setupAdminTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "namespace", Value: "test-ns"}}

		// Setup mock with very old time
		cacheInfo := &execution.CacheInfo{
			Checksum:    "abc123",
			LastRefresh: time.Now().Add(-24 * time.Hour), // 24 hours ago
		}
		mockEngine.On("GetCacheInfo", "test-ns").Return(cacheInfo, nil)

		// Create request
		req := httptest.NewRequest("GET", "/admin/cache/test-ns/stats", nil)
		c.Request = req

		handler.GetCacheStats(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["success"])
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "stale", data["cacheStatus"])
		assert.Equal(t, true, data["isStale"])
		assert.Greater(t, data["stalenessSeconds"].(float64), float64(300)) // Greater than 5 minutes

		mockEngine.AssertExpectations(t)
	})
}
