package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNamespaceService is a mock implementation of NamespaceServiceInterface
type MockNamespaceService struct {
	mock.Mock
}

// Ensure MockNamespaceService implements NamespaceServiceInterface
var _ service.NamespaceServiceInterface = (*MockNamespaceService)(nil)

func (m *MockNamespaceService) CreateNamespace(ctx context.Context, namespace *domain.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockNamespaceService) GetNamespace(ctx context.Context, id string) (*domain.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceService) ListNamespaces(ctx context.Context) ([]*domain.Namespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceService) DeleteNamespace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	// Set default client_id for tests
	c.Set("client_id", "test-client")

	return c, w
}

func TestNewNamespaceHandler(t *testing.T) {
	mockService := new(MockNamespaceService)
	handler := NewNamespaceHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.namespaceService)
}

func TestNamespaceHandler_CreateNamespace(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(*MockNamespaceService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"id":          "test-ns",
				"description": "Test namespace",
			},
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(n *domain.Namespace) bool {
					return n.ID == "test-ns" && n.Description == "Test namespace"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":          "test-ns",
					"description": "Test namespace",
					"createdBy":   "test-client",
					"createdAt":   "0001-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "service error during creation",
			requestBody: map[string]interface{}{
				"id":          "test-ns",
				"description": "Test namespace",
			},
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(n *domain.Namespace) bool {
					return n.ID == "test-ns" && n.Description == "Test namespace"
				})).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":    "INTERNAL_SERVER_ERROR",
				"error":   "Internal Server Error",
				"message": "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/namespaces", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateNamespace(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			// Check basic fields
			assert.Equal(t, tt.expectedBody["error"], response["error"])
			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["message"], response["message"])

			// Check success field if present
			if success, exists := tt.expectedBody["success"]; exists {
				assert.Equal(t, success, response["success"])
			}

			// Check data field if present
			if data, exists := tt.expectedBody["data"]; exists {
				assert.Equal(t, data, response["data"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNamespaceHandler_GetNamespace(t *testing.T) {
	tests := []struct {
		name           string
		namespaceID    string
		setupMock      func(*MockNamespaceService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful retrieval",
			namespaceID: "test-ns",
			setupMock: func(mockService *MockNamespaceService) {
				namespace := &domain.Namespace{
					ID:          "test-ns",
					Description: "Test namespace",
					CreatedBy:   "test-client",
					CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				mockService.On("GetNamespace", mock.Anything, "test-ns").Return(namespace, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":          "test-ns",
					"description": "Test namespace",
					"createdBy":   "test-client",
					"createdAt":   "2023-01-01T00:00:00Z",
				},
			},
		},
		{
			name:        "namespace not found",
			namespaceID: "non-existent",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("GetNamespace", mock.Anything, "non-existent").Return((*domain.Namespace)(nil), domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":   "NOT_FOUND",
				"code":    "NAMESPACE_NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:        "invalid input",
			namespaceID: "",
			setupMock: func(mockService *MockNamespaceService) {
				// No mock expectation needed - handler returns early for empty ID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Bad Request",
				"code":    "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
		{
			name:        "service error",
			namespaceID: "test-ns",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("GetNamespace", mock.Anything, "test-ns").Return((*domain.Namespace)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Internal Server Error",
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()
			c.Params = gin.Params{{Key: "id", Value: tt.namespaceID}}

			handler.GetNamespace(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			assert.Equal(t, tt.expectedBody["error"], response["error"])
			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["message"], response["message"])

			// Check success field if present
			if success, exists := tt.expectedBody["success"]; exists {
				assert.Equal(t, success, response["success"])
			}

			// Check data field if present
			if data, exists := tt.expectedBody["data"]; exists {
				assert.Equal(t, data, response["data"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNamespaceHandler_ListNamespaces(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockNamespaceService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful list",
			setupMock: func(mockService *MockNamespaceService) {
				namespaces := []*domain.Namespace{
					{
						ID:          "ns1",
						Description: "Namespace 1",
						CreatedBy:   "user1",
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "ns2",
						Description: "Namespace 2",
						CreatedBy:   "user2",
						CreatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}
				mockService.On("ListNamespaces", mock.Anything).Return(namespaces, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": []interface{}{
					map[string]interface{}{
						"id":          "ns1",
						"description": "Namespace 1",
						"createdBy":   "user1",
						"createdAt":   "2023-01-01T00:00:00Z",
					},
					map[string]interface{}{
						"id":          "ns2",
						"description": "Namespace 2",
						"createdBy":   "user2",
						"createdAt":   "2023-01-02T00:00:00Z",
					},
				},
			},
		},
		{
			name: "empty list",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("ListNamespaces", mock.Anything).Return([]*domain.Namespace{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data":    nil,
			},
		},
		{
			name: "service error",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("ListNamespaces", mock.Anything).Return(([]*domain.Namespace)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Internal Server Error",
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()

			handler.ListNamespaces(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if expectedBody, ok := tt.expectedBody.(map[string]interface{}); ok {
				if errorCode, exists := expectedBody["error"]; exists {
					assert.Equal(t, errorCode, response["error"])
				}
				if code, exists := expectedBody["code"]; exists {
					assert.Equal(t, code, response["code"])
				}
				if message, exists := expectedBody["message"]; exists {
					assert.Equal(t, message, response["message"])
				}
				if success, exists := expectedBody["success"]; exists {
					assert.Equal(t, success, response["success"])
				}
				if data, exists := expectedBody["data"]; exists {
					actualData := response["data"]
					if data == nil {
						assert.Nil(t, actualData)
					} else {
						assert.True(t, reflect.DeepEqual(data, actualData), "expected: %#v, got: %#v", data, actualData)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNamespaceHandler_DeleteNamespace(t *testing.T) {
	tests := []struct {
		name           string
		namespaceID    string
		setupMock      func(*MockNamespaceService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful deletion",
			namespaceID: "test-ns",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("DeleteNamespace", mock.Anything, "test-ns").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid input",
			namespaceID: "",
			setupMock: func(mockService *MockNamespaceService) {
				// No mock expectation needed - handler returns early for empty ID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Bad Request",
				"code":    "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
		{
			name:        "namespace not found",
			namespaceID: "non-existent",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("DeleteNamespace", mock.Anything, "non-existent").Return(domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":   "NOT_FOUND",
				"code":    "NAMESPACE_NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:        "service error",
			namespaceID: "test-ns",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("DeleteNamespace", mock.Anything, "test-ns").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Internal Server Error",
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()

			// Create a proper DELETE request
			req := httptest.NewRequest(http.MethodDelete, "/v1/namespaces/"+tt.namespaceID, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.namespaceID}}

			handler.DeleteNamespace(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK && tt.expectedBody != nil {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody["error"], response["error"])
				assert.Equal(t, tt.expectedBody["code"], response["code"])
				assert.Equal(t, tt.expectedBody["message"], response["message"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestNamespaceHandler_ToNamespaceResponse(t *testing.T) {
	handler := NewNamespaceHandler(nil)

	namespace := &domain.Namespace{
		ID:          "test-ns",
		Description: "Test namespace",
		CreatedBy:   "test-client",
		CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	response := handler.responseHandler.ConvertNamespaceToResponse(namespace)

	assert.Equal(t, "test-ns", response.ID)
	assert.Equal(t, "Test namespace", response.Description)
	assert.Equal(t, "test-client", response.CreatedBy)
	assert.Equal(t, namespace.CreatedAt, response.CreatedAt)
}
