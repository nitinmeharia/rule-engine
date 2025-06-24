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

// MockTerminalService is a mock implementation of TerminalServiceInterface
type MockTerminalService struct {
	mock.Mock
}

// Ensure MockTerminalService implements TerminalServiceInterface
var _ service.TerminalServiceInterface = (*MockTerminalService)(nil)

func (m *MockTerminalService) CreateTerminal(ctx context.Context, namespace string, terminal *domain.Terminal) error {
	args := m.Called(ctx, namespace, terminal)
	return args.Error(0)
}

func (m *MockTerminalService) GetTerminal(ctx context.Context, namespace, terminalID string) (*domain.Terminal, error) {
	args := m.Called(ctx, namespace, terminalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Terminal), args.Error(1)
}

func (m *MockTerminalService) ListTerminals(ctx context.Context, namespace string) ([]*domain.Terminal, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Terminal), args.Error(1)
}

func (m *MockTerminalService) DeleteTerminal(ctx context.Context, namespace, terminalID string) error {
	args := m.Called(ctx, namespace, terminalID)
	return args.Error(0)
}

func setupTerminalTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	// Set default client_id for tests
	c.Set("client_id", "test-client")

	return c, w
}

func TestNewTerminalHandler(t *testing.T) {
	mockService := new(MockTerminalService)
	handler := NewTerminalHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.terminalService)
	assert.NotNil(t, handler.responseHandler)
}

func TestTerminalHandler_CreateTerminal(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		requestBody    map[string]interface{}
		setupMock      func(*MockTerminalService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful creation",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "approve",
			},
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("CreateTerminal", mock.Anything, "test-ns", mock.MatchedBy(func(t *domain.Terminal) bool {
					return t.TerminalID == "approve"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"terminalId": "approve",
				"createdBy":  "test-client",
				"createdAt":  "0001-01-01T00:00:00Z",
			},
		},
		{
			name:        "missing terminal ID",
			namespace:   "test-ns",
			requestBody: map[string]interface{}{
				// Empty body
			},
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
		{
			name:      "empty terminal ID",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "",
			},
			setupMock:      func(mockService *MockTerminalService) {},
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
				"terminalId": "approve",
			},
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
		{
			name:      "terminal already exists",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "approve",
			},
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("CreateTerminal", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Terminal")).Return(domain.ErrTerminalAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"code":    "TERMINAL_ALREADY_EXISTS",
				"error":   "CONFLICT",
				"message": "Terminal already exists",
			},
		},
		{
			name:      "namespace not found",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "approve",
			},
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("CreateTerminal", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Terminal")).Return(domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:      "service error",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "approve",
			},
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("CreateTerminal", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Terminal")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":    "INTERNAL_SERVER_ERROR",
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
		},
		{
			name:      "malformed JSON",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"terminalId": "approve",
				// Missing closing brace to simulate malformed JSON
			},
			setupMock:      func(mockService *MockTerminalService) {},
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
			mockService := new(MockTerminalService)
			handler := NewTerminalHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupTerminalTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request body
			var jsonBody []byte
			if tt.name == "malformed JSON" {
				// Send actually malformed JSON
				jsonBody = []byte(`{"terminalId": "approve"`)
			} else {
				jsonBody, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/terminals", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateTerminal(c)

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
			if data, exists := tt.expectedBody["terminalId"]; exists {
				assert.Equal(t, data, response["terminalId"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTerminalHandler_GetTerminal(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		terminalID     string
		setupMock      func(*MockTerminalService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful retrieval",
			namespace:  "test-ns",
			terminalID: "approve",
			setupMock: func(mockService *MockTerminalService) {
				terminal := &domain.Terminal{
					TerminalID: "approve",
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				mockService.On("GetTerminal", mock.Anything, "test-ns", "approve").Return(terminal, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"terminalId": "approve",
				"createdBy":  "test-client",
				"createdAt":  "2023-01-01T00:00:00Z",
			},
		},
		{
			name:       "terminal not found",
			namespace:  "test-ns",
			terminalID: "non-existent-terminal",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("GetTerminal", mock.Anything, "test-ns", "non-existent-terminal").Return(nil, domain.ErrTerminalNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "TERMINAL_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Terminal not found",
			},
		},
		{
			name:       "namespace not found",
			namespace:  "non-existent-ns",
			terminalID: "approve",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("GetTerminal", mock.Anything, "non-existent-ns", "approve").Return(nil, domain.ErrNamespaceNotFound)
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
			terminalID:     "approve",
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
		{
			name:           "empty terminal ID",
			namespace:      "test-ns",
			terminalID:     "",
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Terminal ID is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTerminalService)
			handler := NewTerminalHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupTerminalTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "terminalId", Value: tt.terminalID},
			}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/terminals/"+tt.terminalID, nil)
			c.Request = req

			handler.GetTerminal(c)

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
			if data, exists := tt.expectedBody["terminalId"]; exists {
				assert.Equal(t, data, response["terminalId"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTerminalHandler_ListTerminals(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockTerminalService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "successful list with terminals",
			namespace: "test-ns",
			setupMock: func(mockService *MockTerminalService) {
				terminals := []*domain.Terminal{
					{
						TerminalID: "approve",
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						TerminalID: "reject",
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}
				mockService.On("ListTerminals", mock.Anything, "test-ns").Return(terminals, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{
					"terminalId": "approve",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
				},
				{
					"terminalId": "reject",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-02T00:00:00Z",
				},
			},
		},
		{
			name:      "successful list with empty result",
			namespace: "test-ns",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("ListTerminals", mock.Anything, "test-ns").Return([]*domain.Terminal{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []map[string]interface{}{},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("ListTerminals", mock.Anything, "non-existent-ns").Return(nil, domain.ErrNamespaceNotFound)
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
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTerminalService)
			handler := NewTerminalHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupTerminalTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/terminals", nil)
			c.Request = req

			handler.ListTerminals(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			// Check if response is an array (success case) or object (error case)
			if tt.expectedStatus == http.StatusOK {
				assert.IsType(t, []interface{}{}, response)
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

func TestTerminalHandler_DeleteTerminal(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		terminalID     string
		setupMock      func(*MockTerminalService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		// {
		// 	name:       "successful deletion",
		// 	namespace:  "test-ns",
		// 	terminalID: "approve",
		// 	setupMock: func(mockService *MockTerminalService) {
		// 		mockService.On("DeleteTerminal", mock.Anything, "test-ns", "approve").Return(nil)
		// 	},
		// 	expectedStatus: http.StatusNoContent,
		// 	expectedBody:   map[string]interface{}{},
		// },
		{
			name:       "terminal not found",
			namespace:  "test-ns",
			terminalID: "non-existent-terminal",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("DeleteTerminal", mock.Anything, "test-ns", "non-existent-terminal").Return(domain.ErrTerminalNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "TERMINAL_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Terminal not found",
			},
		},
		{
			name:       "namespace not found",
			namespace:  "non-existent-ns",
			terminalID: "approve",
			setupMock: func(mockService *MockTerminalService) {
				mockService.On("DeleteTerminal", mock.Anything, "non-existent-ns", "approve").Return(domain.ErrNamespaceNotFound)
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
			terminalID:     "approve",
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace ID is required",
			},
		},
		{
			name:           "empty terminal ID",
			namespace:      "test-ns",
			terminalID:     "",
			setupMock:      func(mockService *MockTerminalService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Terminal ID is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTerminalService)
			handler := NewTerminalHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupTerminalTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "terminalId", Value: tt.terminalID},
			}

			// Create request
			req := httptest.NewRequest("DELETE", "/v1/namespaces/"+tt.namespace+"/terminals/"+tt.terminalID, nil)
			c.Request = req

			handler.DeleteTerminal(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// For successful deletion, there should be no body
			if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, w.Body.String())
			} else {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				// Check error fields
				if tt.expectedBody["error"] != nil {
					assert.Equal(t, tt.expectedBody["error"], response["error"])
				}
				if tt.expectedBody["code"] != nil {
					assert.Equal(t, tt.expectedBody["code"], response["code"])
				}
				if tt.expectedBody["message"] != nil {
					assert.Equal(t, tt.expectedBody["message"], response["message"])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTerminalHandler_EdgeCases(t *testing.T) {
	t.Run("missing client_id in context", func(t *testing.T) {
		mockService := new(MockTerminalService)
		handler := NewTerminalHandler(mockService)

		c, w := setupTerminalTestContext()
		// Remove client_id from context
		c.Set("client_id", nil)

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"terminalId": "approve",
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/terminals", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateTerminal(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response["error"])
		assert.Equal(t, "Client ID not found", response["message"])
	})

	t.Run("service error during list", func(t *testing.T) {
		mockService := new(MockTerminalService)
		handler := NewTerminalHandler(mockService)

		c, w := setupTerminalTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Setup mock to return error
		mockService.On("ListTerminals", mock.Anything, "test-ns").Return(nil, errors.New("database error"))

		// Create request
		req := httptest.NewRequest("GET", "/v1/namespaces/test-ns/terminals", nil)
		c.Request = req

		handler.ListTerminals(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", response["code"])
		assert.Equal(t, "INTERNAL_SERVER_ERROR", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid terminal ID format", func(t *testing.T) {
		mockService := new(MockTerminalService)
		handler := NewTerminalHandler(mockService)

		// Set up mock expectation for CreateTerminal call
		mockService.On("CreateTerminal", mock.Anything, "test-ns", mock.MatchedBy(func(t *domain.Terminal) bool {
			return t.TerminalID == "very-long-terminal-id-that-exceeds-maximum-length-allowed-by-the-system"
		})).Return(nil)

		c, w := setupTerminalTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body with invalid terminal ID (very long)
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"terminalId": "very-long-terminal-id-that-exceeds-maximum-length-allowed-by-the-system",
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/terminals", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateTerminal(c)

		// Should still be processed by the service layer for validation
		assert.Equal(t, http.StatusCreated, w.Code)

		mockService.AssertExpectations(t)
	})
}
