package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
		requestBody    interface{}
		clientID       string
		setupMock      func(*MockNamespaceService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful creation",
			requestBody: domain.CreateNamespaceRequest{
				ID:          "test-ns",
				Description: "Test namespace",
			},
			clientID: "test-client",
			setupMock: func(mockService *MockNamespaceService) {
				createdNamespace := &domain.Namespace{
					ID:          "test-ns",
					Description: "Test namespace",
					CreatedBy:   "test-client",
					CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(namespace *domain.Namespace) bool {
					return namespace.ID == "test-ns" &&
						namespace.Description == "Test namespace" &&
						namespace.CreatedBy == "test-client"
				})).Return(nil)
				mockService.On("GetNamespace", mock.Anything, "test-ns").Return(createdNamespace, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
				"namespace": map[string]interface{}{
					"id":          "test-ns",
					"description": "Test namespace",
					"createdBy":   "test-client",
				},
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "json",
			},
			clientID:       "test-client",
			setupMock:      func(mockService *MockNamespaceService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "BAD_REQUEST",
				"code":  "VALIDATION_ERROR",
			},
		},
		{
			name: "missing client ID",
			requestBody: domain.CreateNamespaceRequest{
				ID:          "test-ns",
				Description: "Test namespace",
			},
			clientID:       "",
			setupMock:      func(mockService *MockNamespaceService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
			},
		},
		{
			name: "namespace already exists",
			requestBody: domain.CreateNamespaceRequest{
				ID:          "test-ns",
				Description: "Test namespace",
			},
			clientID: "test-client",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(namespace *domain.Namespace) bool {
					return namespace.ID == "test-ns" &&
						namespace.Description == "Test namespace" &&
						namespace.CreatedBy == "test-client"
				})).Return(domain.ErrNamespaceAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "CONFLICT",
				"code":  "NAMESPACE_ALREADY_EXISTS",
			},
		},
		{
			name: "service error during creation",
			requestBody: domain.CreateNamespaceRequest{
				ID:          "test-ns",
				Description: "Test namespace",
			},
			clientID: "test-client",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(namespace *domain.Namespace) bool {
					return namespace.ID == "test-ns" &&
						namespace.Description == "Test namespace" &&
						namespace.CreatedBy == "test-client"
				})).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
			},
		},
		{
			name: "service error during fetch",
			requestBody: domain.CreateNamespaceRequest{
				ID:          "test-ns",
				Description: "Test namespace",
			},
			clientID: "test-client",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("CreateNamespace", mock.Anything, mock.MatchedBy(func(namespace *domain.Namespace) bool {
					return namespace.ID == "test-ns" &&
						namespace.Description == "Test namespace" &&
						namespace.CreatedBy == "test-client"
				})).Return(nil)
				mockService.On("GetNamespace", mock.Anything, "test-ns").Return((*domain.Namespace)(nil), errors.New("fetch error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()

			// Only set client_id if it's not empty (to test missing client_id case)
			if tt.clientID != "" {
				c.Set("client_id", tt.clientID)
			}

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

			// Check success field if present
			if success, exists := tt.expectedBody["success"]; exists {
				assert.Equal(t, success, response["success"])
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
				"id":          "test-ns",
				"description": "Test namespace",
				"createdBy":   "test-client",
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
				"error": "NOT_FOUND",
				"code":  "NAMESPACE_NOT_FOUND",
			},
		},
		{
			name:        "invalid input",
			namespaceID: "",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("GetNamespace", mock.Anything, "").Return((*domain.Namespace)(nil), domain.ErrInvalidNamespaceID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "BAD_REQUEST",
				"code":  "INVALID_NAMESPACE_ID",
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
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
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

			// Check namespace fields if present
			if id, exists := tt.expectedBody["id"]; exists {
				assert.Equal(t, id, response["id"])
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
			expectedBody: []map[string]interface{}{
				{
					"id":          "ns1",
					"description": "Namespace 1",
					"createdBy":   "user1",
				},
				{
					"id":          "ns2",
					"description": "Namespace 2",
					"createdBy":   "user2",
				},
			},
		},
		{
			name: "empty list",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("ListNamespaces", mock.Anything).Return([]*domain.Namespace{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []interface{}{},
		},
		{
			name: "service error",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("ListNamespaces", mock.Anything).Return(([]*domain.Namespace)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
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

			if tt.expectedStatus == http.StatusOK {
				if _, ok := tt.expectedBody.([]map[string]interface{}); ok && len(tt.expectedBody.([]map[string]interface{})) == 0 {
					// Empty list case
					var response []interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					assert.Len(t, response, 0)
				} else if expected, ok := tt.expectedBody.([]map[string]interface{}); ok {
					var response []map[string]interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					assert.Len(t, response, len(expected))
				}
			} else {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				expectedBody := tt.expectedBody.(map[string]interface{})
				assert.Equal(t, expectedBody["error"], response["error"])
				assert.Equal(t, expectedBody["code"], response["code"])
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
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "namespace not found",
			namespaceID: "non-existent",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("DeleteNamespace", mock.Anything, "non-existent").Return(domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "NOT_FOUND",
				"code":  "NAMESPACE_NOT_FOUND",
			},
		},
		{
			name:        "invalid input",
			namespaceID: "",
			setupMock: func(mockService *MockNamespaceService) {
				mockService.On("DeleteNamespace", mock.Anything, "").Return(domain.ErrInvalidNamespaceID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "BAD_REQUEST",
				"code":  "INVALID_NAMESPACE_ID",
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
				"error": "INTERNAL_ERROR",
				"code":  "INTERNAL_ERROR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNamespaceService)
			handler := NewNamespaceHandler(mockService)
			tt.setupMock(mockService)

			c, w := createTestContext()
			c.Request.Method = http.MethodDelete
			c.Params = gin.Params{{Key: "id", Value: tt.namespaceID}}

			handler.DeleteNamespace(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody["error"], response["error"])
				assert.Equal(t, tt.expectedBody["code"], response["code"])
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

func TestNamespaceHandler_MapError(t *testing.T) {
	handler := NewNamespaceHandler(nil)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{"not found", domain.ErrNamespaceNotFound, http.StatusNotFound, "NAMESPACE_NOT_FOUND"},
		{"already exists", domain.ErrNamespaceAlreadyExists, http.StatusConflict, "NAMESPACE_ALREADY_EXISTS"},
		{"invalid input", domain.ErrInvalidNamespaceID, http.StatusBadRequest, "INVALID_NAMESPACE_ID"},
		{"validation", domain.ErrValidationError, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"unauthorized", domain.ErrMissingAuthHeader, http.StatusUnauthorized, "MISSING_AUTH_HEADER"},
		{"forbidden", domain.ErrInsufficientPermissions, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS"},
		{"unknown error", errors.New("unknown error"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, response := handler.mapError(tt.err)
			assert.Equal(t, tt.expectedStatus, status)
			if apiErr, ok := response.(*domain.APIError); ok {
				assert.Equal(t, tt.expectedCode, apiErr.Code)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"validation error", domain.ErrValidationError, true},
		{"validation failed", errors.New("validation failed: namespace ID is required"), true},
		{"required error", errors.New("namespace ID is required"), true},
		{"too long error", errors.New("description too long"), true},
		{"invalid error", errors.New("invalid namespace ID"), true},
		{"other error", errors.New("database error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"exact match", "test", "test", true},
		{"prefix match", "test string", "test", true},
		{"suffix match", "test string", "string", true},
		{"middle match", "test string", "st st", true},
		{"no match", "test string", "xyz", false},
		{"empty substr", "test", "", true},
		{"empty str", "", "test", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected int
	}{
		{"exact match", "test", "test", 0},
		{"prefix match", "test string", "test", 0},
		{"suffix match", "test string", "string", 5},
		{"middle match", "test string", "st st", 2},
		{"no match", "test string", "xyz", -1},
		{"empty substr", "test", "", 0},
		{"empty str", "", "test", -1},
		{"both empty", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.str, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
