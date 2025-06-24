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

// MockFunctionService is a mock implementation of FunctionService
type MockFunctionService struct {
	mock.Mock
}

// Ensure MockFunctionService implements FunctionServiceInterface
var _ service.FunctionServiceInterface = (*MockFunctionService)(nil)

// Ensure MockFunctionService implements the required methods
func (m *MockFunctionService) CreateFunction(ctx context.Context, namespace string, function *domain.Function) error {
	args := m.Called(ctx, namespace, function)
	return args.Error(0)
}

func (m *MockFunctionService) GetFunction(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	args := m.Called(ctx, namespace, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *MockFunctionService) ListFunctions(ctx context.Context, namespace string) ([]*domain.Function, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *MockFunctionService) UpdateFunction(ctx context.Context, namespace, functionID string, function *domain.Function) error {
	args := m.Called(ctx, namespace, functionID, function)
	return args.Error(0)
}

func (m *MockFunctionService) PublishFunction(ctx context.Context, namespace, functionID, clientID string) error {
	args := m.Called(ctx, namespace, functionID, clientID)
	return args.Error(0)
}

func (m *MockFunctionService) DeleteFunction(ctx context.Context, namespace, functionID string, version int32) error {
	args := m.Called(ctx, namespace, functionID, version)
	return args.Error(0)
}

func setupFunctionTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	// Set default client_id for tests
	c.Set("client_id", "test-client")

	return c, w
}

func TestNewFunctionHandler(t *testing.T) {
	mockService := new(MockFunctionService)
	handler := NewFunctionHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.functionService)
	assert.NotNil(t, handler.responseHandler)
}

func TestFunctionHandler_CreateFunction(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		requestBody    map[string]interface{}
		setupMock      func(*MockFunctionService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful creation with max function",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":   "max_income",
				"type": "max",
				"args": []string{"income", "bonus"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("CreateFunction", mock.Anything, "test-ns", mock.MatchedBy(func(f *domain.Function) bool {
					return f.FunctionID == "max_income" && f.Type == "max" && len(f.Args) == 2
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status": "draft",
				"function": map[string]interface{}{
					"id":         "max_income",
					"type":       "max",
					"args":       []interface{}{"income", "bonus"},
					"values":     nil,
					"createdBy":  "test-client",
					"createdAt":  "0001-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
					"returnType": "",
				},
			},
		},
		{
			name:      "successful creation with in function",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":     "valid_occupation",
				"type":   "in",
				"values": []string{"salaried", "self_employed"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("CreateFunction", mock.Anything, "test-ns", mock.MatchedBy(func(f *domain.Function) bool {
					return f.FunctionID == "valid_occupation" && f.Type == "in" && len(f.Values) == 2
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status": "draft",
				"function": map[string]interface{}{
					"id":         "valid_occupation",
					"type":       "in",
					"args":       nil,
					"values":     []interface{}{"salaried", "self_employed"},
					"createdBy":  "test-client",
					"createdAt":  "0001-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
					"returnType": "",
				},
			},
		},
		{
			name:      "missing function ID",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"type": "max",
				"args": []string{"income"},
			},
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
		{
			name:      "missing function type",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":   "max_income",
				"args": []string{"income"},
			},
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
		{
			name:      "empty namespace",
			namespace: "",
			requestBody: map[string]interface{}{
				"id":   "max_income",
				"type": "max",
				"args": []string{"income"},
			},
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace is required",
			},
		},
		{
			name:      "function already exists",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":   "max_income",
				"type": "max",
				"args": []string{"income"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("CreateFunction", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Function")).Return(domain.ErrFunctionAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"code":    "FUNCTION_ALREADY_EXISTS",
				"error":   "CONFLICT",
				"message": "Function already exists",
			},
		},
		{
			name:      "namespace not found",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":   "max_income",
				"type": "max",
				"args": []string{"income"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("CreateFunction", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Function")).Return(domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:      "malformed JSON",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id": "max_income",
				// Missing closing brace to simulate malformed JSON
			},
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFunctionService)
			handler := NewFunctionHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupFunctionTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/functions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateFunction(c)

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
			if status, exists := tt.expectedBody["status"]; exists {
				assert.Equal(t, status, response["status"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestFunctionHandler_GetFunction(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		functionID     string
		setupMock      func(*MockFunctionService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful retrieval",
			namespace:  "test-ns",
			functionID: "max_income",
			setupMock: func(mockService *MockFunctionService) {
				function := &domain.Function{
					FunctionID: "max_income",
					Type:       "max",
					Args:       []string{"income", "bonus"},
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "active",
					Version:    1,
				}
				mockService.On("GetFunction", mock.Anything, "test-ns", "max_income").Return(function, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         "max_income",
				"type":       "max",
				"args":       []interface{}{"income", "bonus"},
				"values":     nil,
				"createdBy":  "test-client",
				"createdAt":  "2023-01-01T00:00:00Z",
				"status":     "active",
				"version":    float64(1),
				"returnType": "",
			},
		},
		{
			name:       "function not found",
			namespace:  "test-ns",
			functionID: "non-existent-function",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("GetFunction", mock.Anything, "test-ns", "non-existent-function").Return(nil, domain.ErrFunctionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "FUNCTION_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Function not found",
			},
		},
		{
			name:       "namespace not found",
			namespace:  "non-existent-ns",
			functionID: "max_income",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("GetFunction", mock.Anything, "non-existent-ns", "max_income").Return(nil, domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			functionID:     "max_income",
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and function ID are required",
			},
		},
		{
			name:           "empty function ID",
			namespace:      "test-ns",
			functionID:     "",
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and function ID are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFunctionService)
			handler := NewFunctionHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupFunctionTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "functionId", Value: tt.functionID},
			}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/functions/"+tt.functionID, nil)
			c.Request = req

			handler.GetFunction(c)

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

			// Check data field if present (for successful responses)
			if data, exists := tt.expectedBody["id"]; exists {
				// For successful responses, data is wrapped in envelope
				if response["success"] == true {
					dataField := response["data"].(map[string]interface{})
					assert.Equal(t, data, dataField["id"])
				} else {
					// For error responses, check directly
					assert.Equal(t, data, response["id"])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestFunctionHandler_ListFunctions(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockFunctionService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "successful list with functions",
			namespace: "test-ns",
			setupMock: func(mockService *MockFunctionService) {
				functions := []*domain.Function{
					{
						FunctionID: "max_income",
						Type:       "max",
						Args:       []string{"income", "bonus"},
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						Status:     "active",
						Version:    1,
					},
					{
						FunctionID: "valid_occupation",
						Type:       "in",
						Values:     []string{"salaried", "self_employed"},
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
						Status:     "draft",
						Version:    1,
					},
				}
				mockService.On("ListFunctions", mock.Anything, "test-ns").Return(functions, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{
					"id":         "max_income",
					"type":       "max",
					"args":       []interface{}{"income", "bonus"},
					"values":     nil,
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "active",
					"version":    float64(1),
					"returnType": "",
				},
				{
					"id":         "valid_occupation",
					"type":       "in",
					"args":       nil,
					"values":     []interface{}{"salaried", "self_employed"},
					"createdBy":  "test-client",
					"createdAt":  "2023-01-02T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
					"returnType": "",
				},
			},
		},
		{
			name:      "successful list with empty result",
			namespace: "test-ns",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("ListFunctions", mock.Anything, "test-ns").Return([]*domain.Function{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []map[string]interface{}{},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("ListFunctions", mock.Anything, "non-existent-ns").Return(nil, domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFunctionService)
			handler := NewFunctionHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupFunctionTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/functions", nil)
			c.Request = req

			handler.ListFunctions(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			// Check if response is an array (success case) or object (error case)
			if tt.expectedStatus == http.StatusOK {
				// For successful responses, data is wrapped in envelope
				assert.IsType(t, map[string]interface{}{}, response)
				responseMap := response.(map[string]interface{})
				assert.Equal(t, true, responseMap["success"])
				assert.IsType(t, []interface{}{}, responseMap["data"])
			} else {
				assert.IsType(t, map[string]interface{}{}, response)
				errorResponse := response.(map[string]interface{})
				expectedError := tt.expectedBody.(map[string]interface{})
				assert.Equal(t, expectedError["code"], errorResponse["code"])
				assert.Equal(t, expectedError["error"], errorResponse["error"])
				assert.Equal(t, expectedError["message"], errorResponse["message"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestFunctionHandler_UpdateFunction(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		functionID     string
		requestBody    map[string]interface{}
		setupMock      func(*MockFunctionService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful update",
			namespace:  "test-ns",
			functionID: "max_income",
			requestBody: map[string]interface{}{
				"type": "max",
				"args": []string{"income", "bonus", "commission"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("UpdateFunction", mock.Anything, "test-ns", "max_income", mock.MatchedBy(func(f *domain.Function) bool {
					return f.Type == "max" && len(f.Args) == 3
				})).Return(nil)

				updatedFunction := &domain.Function{
					FunctionID: "max_income",
					Type:       "max",
					Args:       []string{"income", "bonus", "commission"},
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "draft",
					Version:    2,
				}
				mockService.On("GetFunction", mock.Anything, "test-ns", "max_income").Return(updatedFunction, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"function": map[string]interface{}{
					"id":         "max_income",
					"type":       "max",
					"args":       []interface{}{"income", "bonus", "commission"},
					"values":     nil,
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(2),
					"returnType": "",
				},
			},
		},
		{
			name:       "function not found",
			namespace:  "test-ns",
			functionID: "non-existent-function",
			requestBody: map[string]interface{}{
				"type": "max",
				"args": []string{"income"},
			},
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("UpdateFunction", mock.Anything, "test-ns", "non-existent-function", mock.AnythingOfType("*domain.Function")).Return(domain.ErrFunctionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "FUNCTION_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Function not found",
			},
		},
		{
			name:        "missing request body",
			namespace:   "test-ns",
			functionID:  "max_income",
			requestBody: map[string]interface{}{
				// Empty body
			},
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFunctionService)
			handler := NewFunctionHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupFunctionTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "functionId", Value: tt.functionID},
			}

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/v1/namespaces/"+tt.namespace+"/functions/"+tt.functionID+"/versions/draft", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.UpdateFunction(c)

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

			mockService.AssertExpectations(t)
		})
	}
}

func TestFunctionHandler_PublishFunction(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		functionID     string
		setupMock      func(*MockFunctionService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful publish",
			namespace:  "test-ns",
			functionID: "max_income",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("PublishFunction", mock.Anything, "test-ns", "max_income", "test-client").Return(nil)

				publishedFunction := &domain.Function{
					FunctionID: "max_income",
					Type:       "max",
					Args:       []string{"income", "bonus"},
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "active",
					Version:    1,
				}
				mockService.On("GetFunction", mock.Anything, "test-ns", "max_income").Return(publishedFunction, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         "max_income",
				"type":       "max",
				"args":       []interface{}{"income", "bonus"},
				"values":     nil,
				"createdBy":  "test-client",
				"createdAt":  "2023-01-01T00:00:00Z",
				"status":     "active",
				"version":    float64(1),
				"returnType": "",
			},
		},
		{
			name:       "function not found",
			namespace:  "test-ns",
			functionID: "non-existent-function",
			setupMock: func(mockService *MockFunctionService) {
				mockService.On("PublishFunction", mock.Anything, "test-ns", "non-existent-function", "test-client").Return(domain.ErrFunctionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "FUNCTION_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Function not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			functionID:     "max_income",
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and function ID are required",
			},
		},
		{
			name:           "empty function ID",
			namespace:      "test-ns",
			functionID:     "",
			setupMock:      func(mockService *MockFunctionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and function ID are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFunctionService)
			handler := NewFunctionHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupFunctionTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "functionId", Value: tt.functionID},
			}

			// Create request
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/functions/"+tt.functionID+"/publish", nil)
			c.Request = req

			handler.PublishFunction(c)

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

			// Check data field if present
			if data, exists := tt.expectedBody["id"]; exists {
				// For successful responses, data is wrapped in envelope
				if response["success"] == true {
					dataField := response["data"].(map[string]interface{})
					assert.Equal(t, data, dataField["id"])
				} else {
					// For error responses, check directly
					assert.Equal(t, data, response["id"])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestFunctionHandler_EdgeCases(t *testing.T) {
	t.Run("missing client_id in context", func(t *testing.T) {
		mockService := new(MockFunctionService)
		handler := NewFunctionHandler(mockService)

		c, w := setupFunctionTestContext()
		// Remove client_id from context
		c.Set("client_id", nil)

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"id":   "max_income",
			"type": "max",
			"args": []string{"income"},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/functions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateFunction(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response["error"])
		assert.Equal(t, "Client ID not found", response["message"])
	})

	t.Run("get function error after publish", func(t *testing.T) {
		mockService := new(MockFunctionService)
		handler := NewFunctionHandler(mockService)

		c, w := setupFunctionTestContext()

		// Set parameters
		c.Params = gin.Params{
			{Key: "id", Value: "test-ns"},
			{Key: "functionId", Value: "max_income"},
		}

		// Setup mocks
		mockService.On("PublishFunction", mock.Anything, "test-ns", "max_income", "test-client").Return(nil)
		mockService.On("GetFunction", mock.Anything, "test-ns", "max_income").Return(nil, errors.New("database error"))

		// Create request
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/functions/max_income/publish", nil)
		c.Request = req

		handler.PublishFunction(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", response["error"])
		assert.Equal(t, "An unexpected error occurred", response["message"])

		mockService.AssertExpectations(t)
	})
}
