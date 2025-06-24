package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

// MockRuleService is a mock implementation of RuleService
type MockRuleService struct {
	mock.Mock
}

// Ensure MockRuleService implements RuleServiceInterface
var _ service.RuleServiceInterface = (*MockRuleService)(nil)

// Ensure MockRuleService implements the required methods
func (m *MockRuleService) CreateRule(ctx context.Context, namespace string, rule *domain.Rule) error {
	args := m.Called(ctx, namespace, rule)
	return args.Error(0)
}

func (m *MockRuleService) GetRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleService) GetDraftRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleService) ListRules(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleService) ListRuleVersions(ctx context.Context, namespace, ruleID string) ([]*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleService) UpdateRule(ctx context.Context, namespace, ruleID string, rule *domain.Rule) error {
	args := m.Called(ctx, namespace, ruleID, rule)
	return args.Error(0)
}

func (m *MockRuleService) PublishRule(ctx context.Context, namespace, ruleID, clientID string) error {
	args := m.Called(ctx, namespace, ruleID, clientID)
	return args.Error(0)
}

func (m *MockRuleService) DeleteRule(ctx context.Context, namespace, ruleID string, version int32) error {
	args := m.Called(ctx, namespace, ruleID, version)
	return args.Error(0)
}

func setupRuleTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	// Set default client_id for tests
	c.Set("client_id", "test-client")

	return c, w
}

func TestNewRuleHandler(t *testing.T) {
	mockService := new(MockRuleService)
	handler := NewRuleHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.ruleService)
	assert.NotNil(t, handler.responseHandler)
}

func TestRuleHandler_CreateRule(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		requestBody    map[string]interface{}
		setupMock      func(*MockRuleService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful creation with AND logic",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":    "basic_approval",
				"logic": "AND",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock: func(mockService *MockRuleService) {
				mockService.On("CreateRule", mock.Anything, "test-ns", mock.MatchedBy(func(r *domain.Rule) bool {
					return r.RuleID == "basic_approval" && r.Logic == "AND"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status": "draft",
				"rule": map[string]interface{}{
					"id":         "basic_approval",
					"logic":      "AND",
					"conditions": []interface{}{},
					"createdBy":  "test-client",
					"createdAt":  "0001-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
				},
			},
		},
		{
			name:      "successful creation with OR logic",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":    "flexible_approval",
				"logic": "OR",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      60000,
					},
					{
						"type":       "function",
						"functionId": "good_credit",
						"operator":   ">=",
						"value":      750,
					},
				},
			},
			setupMock: func(mockService *MockRuleService) {
				mockService.On("CreateRule", mock.Anything, "test-ns", mock.MatchedBy(func(r *domain.Rule) bool {
					return r.RuleID == "flexible_approval" && r.Logic == "OR"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status": "draft",
				"rule": map[string]interface{}{
					"id":         "flexible_approval",
					"logic":      "OR",
					"conditions": []interface{}{},
					"createdBy":  "test-client",
					"createdAt":  "0001-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
				},
			},
		},
		{
			name:      "missing rule ID",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"logic": "AND",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
		{
			name:      "missing logic",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id": "basic_approval",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body",
			},
		},
		{
			name:      "missing conditions",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":    "basic_approval",
				"logic": "AND",
			},
			setupMock:      func(mockService *MockRuleService) {},
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
				"id":    "basic_approval",
				"logic": "AND",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace is required",
			},
		},
		{
			name:      "rule already exists",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":    "basic_approval",
				"logic": "AND",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock: func(mockService *MockRuleService) {
				mockService.On("CreateRule", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Rule")).Return(domain.ErrRuleAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"code":    "RULE_ALREADY_EXISTS",
				"error":   "CONFLICT",
				"message": "Rule already exists",
			},
		},
		{
			name:      "namespace not found",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":    "basic_approval",
				"logic": "AND",
				"conditions": []map[string]interface{}{
					{
						"type":       "function",
						"functionId": "high_income",
						"operator":   ">=",
						"value":      50000,
					},
				},
			},
			setupMock: func(mockService *MockRuleService) {
				mockService.On("CreateRule", mock.Anything, "test-ns", mock.AnythingOfType("*domain.Rule")).Return(domain.ErrNamespaceNotFound)
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
				"id": "basic_approval",
				// Missing closing brace to simulate malformed JSON
			},
			setupMock:      func(mockService *MockRuleService) {},
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
			mockService := new(MockRuleService)
			handler := NewRuleHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupRuleTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/rules", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateRule(c)

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

func TestRuleHandler_GetRule(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		ruleID         string
		setupMock      func(*MockRuleService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful retrieval",
			namespace: "test-ns",
			ruleID:    "basic_approval",
			setupMock: func(mockService *MockRuleService) {
				rule := &domain.Rule{
					RuleID:     "basic_approval",
					Logic:      "AND",
					Conditions: json.RawMessage(`[{"type":"function","functionId":"high_income","operator":">=","value":50000}]`),
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "active",
					Version:    1,
				}
				mockService.On("GetRule", mock.Anything, "test-ns", "basic_approval").Return(rule, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         "basic_approval",
				"logic":      "AND",
				"conditions": []interface{}{},
				"createdBy":  "test-client",
				"createdAt":  "2023-01-01T00:00:00Z",
				"status":     "active",
				"version":    float64(1),
			},
		},
		{
			name:      "rule not found",
			namespace: "test-ns",
			ruleID:    "non-existent-rule",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("GetRule", mock.Anything, "test-ns", "non-existent-rule").Return(nil, domain.ErrRuleNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "RULE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Rule not found",
			},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			ruleID:    "basic_approval",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("GetRule", mock.Anything, "non-existent-ns", "basic_approval").Return(nil, domain.ErrNamespaceNotFound)
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
			ruleID:         "basic_approval",
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and rule ID are required",
			},
		},
		{
			name:           "empty rule ID",
			namespace:      "test-ns",
			ruleID:         "",
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and rule ID are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRuleService)
			handler := NewRuleHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupRuleTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "ruleId", Value: tt.ruleID},
			}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/rules/"+tt.ruleID, nil)
			c.Request = req

			handler.GetRule(c)

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
				assert.Equal(t, data, response["id"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestRuleHandler_ListRules(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockRuleService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "successful list with rules",
			namespace: "test-ns",
			setupMock: func(mockService *MockRuleService) {
				rules := []*domain.Rule{
					{
						RuleID:     "basic_approval",
						Logic:      "AND",
						Conditions: json.RawMessage(`[{"type":"function","functionId":"high_income","operator":">=","value":50000}]`),
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						Status:     "active",
						Version:    1,
					},
					{
						RuleID:     "flexible_approval",
						Logic:      "OR",
						Conditions: json.RawMessage(`[{"type":"function","functionId":"high_income","operator":">=","value":60000}]`),
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
						Status:     "draft",
						Version:    1,
					},
				}
				mockService.On("ListRules", mock.Anything, "test-ns").Return(rules, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{
					"id":         "basic_approval",
					"logic":      "AND",
					"conditions": []interface{}{},
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "active",
					"version":    float64(1),
				},
				{
					"id":         "flexible_approval",
					"logic":      "OR",
					"conditions": []interface{}{},
					"createdBy":  "test-client",
					"createdAt":  "2023-01-02T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
				},
			},
		},
		{
			name:      "successful list with empty result",
			namespace: "test-ns",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("ListRules", mock.Anything, "test-ns").Return([]*domain.Rule{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []map[string]interface{}{},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("ListRules", mock.Anything, "non-existent-ns").Return(nil, domain.ErrNamespaceNotFound)
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
			setupMock:      func(mockService *MockRuleService) {},
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
			mockService := new(MockRuleService)
			handler := NewRuleHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupRuleTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/rules", nil)
			c.Request = req

			handler.ListRules(c)

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

func TestRuleHandler_PublishRule(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		ruleID         string
		setupMock      func(*MockRuleService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful publish",
			namespace: "test-ns",
			ruleID:    "basic_approval",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("PublishRule", mock.Anything, "test-ns", "basic_approval", "test-client").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "active",
			},
		},
		{
			name:      "rule not found",
			namespace: "test-ns",
			ruleID:    "non-existent-rule",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("PublishRule", mock.Anything, "test-ns", "non-existent-rule", "test-client").Return(domain.ErrRuleNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "RULE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Rule not found",
			},
		},
		{
			name:      "function dependency not found",
			namespace: "test-ns",
			ruleID:    "basic_approval",
			setupMock: func(mockService *MockRuleService) {
				mockService.On("PublishRule", mock.Anything, "test-ns", "basic_approval", "test-client").Return(domain.ErrFunctionNotFound)
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
			ruleID:         "basic_approval",
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and rule ID are required",
			},
		},
		{
			name:           "empty rule ID",
			namespace:      "test-ns",
			ruleID:         "",
			setupMock:      func(mockService *MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and rule ID are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRuleService)
			handler := NewRuleHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupRuleTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "ruleId", Value: tt.ruleID},
			}

			// Create request
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/rules/"+tt.ruleID+"/publish", nil)
			c.Request = req

			handler.PublishRule(c)

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

func TestRuleHandler_EdgeCases(t *testing.T) {
	t.Run("missing client_id in context", func(t *testing.T) {
		mockService := new(MockRuleService)
		handler := NewRuleHandler(mockService)

		c, w := setupRuleTestContext()
		// Remove client_id from context
		c.Set("client_id", nil)

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"id":    "basic_approval",
			"logic": "AND",
			"conditions": []map[string]interface{}{
				{
					"type":       "function",
					"functionId": "high_income",
					"operator":   ">=",
					"value":      50000,
				},
			},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/rules", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateRule(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response["error"])
		assert.Equal(t, "Client ID not found", response["message"])
	})

	t.Run("invalid logic type", func(t *testing.T) {
		mockService := new(MockRuleService)
		handler := NewRuleHandler(mockService)

		// Set up mock expectation for CreateRule call
		mockService.On("CreateRule", mock.Anything, "test-ns", mock.MatchedBy(func(r *domain.Rule) bool {
			return r.RuleID == "basic_approval" && r.Logic == "XOR"
		})).Return(domain.ErrInvalidRuleLogic)

		c, w := setupRuleTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body with invalid logic
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"id":    "basic_approval",
			"logic": "XOR", // Invalid logic type
			"conditions": []map[string]interface{}{
				{
					"type":       "function",
					"functionId": "high_income",
					"operator":   ">=",
					"value":      50000,
				},
			},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/rules", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateRule(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_RULE_LOGIC", response["code"])
		assert.Equal(t, "BAD_REQUEST", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("empty conditions array", func(t *testing.T) {
		mockService := new(MockRuleService)
		handler := NewRuleHandler(mockService)

		// Set up mock expectation for CreateRule call
		mockService.On("CreateRule", mock.Anything, "test-ns", mock.MatchedBy(func(r *domain.Rule) bool {
			return r.RuleID == "basic_approval" && r.Logic == "AND"
		})).Return(domain.ErrInvalidRuleConditions)

		c, w := setupRuleTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body with empty conditions
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"id":         "basic_approval",
			"logic":      "AND",
			"conditions": []map[string]interface{}{},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/rules", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateRule(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_RULE_CONDITIONS", response["code"])
		assert.Equal(t, "BAD_REQUEST", response["error"])

		mockService.AssertExpectations(t)
	})
}
