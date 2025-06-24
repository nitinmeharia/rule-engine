package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupExecutionTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	return c, w
}

func TestNewExecutionHandler(t *testing.T) {
	mockEngine := new(MockEngine)
	handler := NewExecutionHandler(mockEngine)

	assert.NotNil(t, handler)
	assert.Equal(t, mockEngine, handler.engine)
	assert.NotNil(t, handler.responseHandler)
}

func TestExecutionHandler_ExecuteRule(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		ruleID         string
		requestBody    map[string]interface{}
		queryParams    map[string]string
		setupMock      func(*MockEngine)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful rule execution",
			namespace: "test-ns",
			ruleID:    "approval_rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
					"age":    30,
					"credit": "good",
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				response := &domain.ExecutionResponse{
					Result:    true,
					Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Metadata: map[string]interface{}{
						"ruleId":    "approval_rule",
						"namespace": "test-ns",
					},
				}
				mockEngine.On("ExecuteRule", mock.Anything, mock.MatchedBy(func(req *domain.ExecutionRequest) bool {
					return req.Namespace == "test-ns" && *req.RuleID == "approval_rule"
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"result":    true,
				"timestamp": "2023-01-01T00:00:00Z",
				"metadata": map[string]interface{}{
					"ruleId":    "approval_rule",
					"namespace": "test-ns",
				},
			},
		},
		{
			name:      "successful rule execution with trace",
			namespace: "test-ns",
			ruleID:    "approval_rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			queryParams: map[string]string{"trace": "full"},
			setupMock: func(mockEngine *MockEngine) {
				response := &domain.ExecutionResponse{
					Result:    false,
					Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Trace: &domain.ExecutionTrace{
						Steps: []domain.TraceStep{
							{
								Type:     "condition_check",
								ID:       "check_eligibility",
								Output:   false,
								Duration: "1ms",
							},
						},
						Version:  "1.0.0",
						Duration: "1ms",
					},
				}
				mockEngine.On("ExecuteRule", mock.Anything, mock.MatchedBy(func(req *domain.ExecutionRequest) bool {
					return req.Namespace == "test-ns" && *req.RuleID == "approval_rule" && req.Trace == true
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"result":    false,
				"timestamp": "2023-01-01T00:00:00Z",
				"trace": map[string]interface{}{
					"steps": []interface{}{
						map[string]interface{}{
							"type":     "condition_check",
							"id":       "check_eligibility",
							"output":   false,
							"duration": "1ms",
						},
					},
					"version":  "1.0.0",
					"duration": "1ms",
				},
			},
		},
		{
			name:      "rule not found",
			namespace: "test-ns",
			ruleID:    "non-existent-rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("ExecuteRule", mock.Anything, mock.AnythingOfType("*domain.ExecutionRequest")).Return(nil, domain.ErrRuleNotFound)
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
			ruleID:    "approval_rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("ExecuteRule", mock.Anything, mock.AnythingOfType("*domain.ExecutionRequest")).Return(nil, domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:        "invalid request body",
			namespace:   "test-ns",
			ruleID:      "approval_rule",
			requestBody: map[string]interface{}{
				// Missing data field
			},
			setupMock:      func(mockEngine *MockEngine) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body: INVALID_EXECUTION_DATA: Invalid execution data",
			},
		},
		{
			name:      "malformed JSON",
			namespace: "test-ns",
			ruleID:    "approval_rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock:      func(mockEngine *MockEngine) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body: unexpected EOF",
			},
		},
		{
			name:      "engine execution error",
			namespace: "test-ns",
			ruleID:    "approval_rule",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("ExecuteRule", mock.Anything, mock.AnythingOfType("*domain.ExecutionRequest")).Return(nil, errors.New("execution failed"))
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
			handler := NewExecutionHandler(mockEngine)
			tt.setupMock(mockEngine)

			c, w := setupExecutionTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "namespace", Value: tt.namespace},
				{Key: "ruleId", Value: tt.ruleID},
			}

			// Create request body
			var jsonBody []byte
			if tt.name == "malformed JSON" {
				jsonBody = []byte(`{"data": {"income": 50000}`) // Missing closing brace
			} else {
				jsonBody, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/rules/"+tt.ruleID+"/execute", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Add query parameters
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			c.Request = req

			handler.ExecuteRule(c)

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
			if result, exists := tt.expectedBody["result"]; exists {
				assert.Equal(t, result, response["result"])
			}
			if timestamp, exists := tt.expectedBody["timestamp"]; exists {
				assert.Equal(t, timestamp, response["timestamp"])
			}
			if metadata, exists := tt.expectedBody["metadata"]; exists {
				assert.Equal(t, metadata, response["metadata"])
			}
			if trace, exists := tt.expectedBody["trace"]; exists {
				assert.Equal(t, trace, response["trace"])
			}

			mockEngine.AssertExpectations(t)
		})
	}
}

func TestExecutionHandler_ExecuteWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		workflowID     string
		requestBody    map[string]interface{}
		queryParams    map[string]string
		setupMock      func(*MockEngine)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful workflow execution",
			namespace:  "test-ns",
			workflowID: "approval_workflow",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
					"age":    30,
					"credit": "good",
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				response := &domain.ExecutionResponse{
					Result:    "approved",
					Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Metadata: map[string]interface{}{
						"workflowId": "approval_workflow",
						"namespace":  "test-ns",
					},
				}
				mockEngine.On("ExecuteWorkflow", mock.Anything, mock.MatchedBy(func(req *domain.ExecutionRequest) bool {
					return req.Namespace == "test-ns" && *req.WorkflowID == "approval_workflow"
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"result":    "approved",
				"timestamp": "2023-01-01T00:00:00Z",
				"metadata": map[string]interface{}{
					"workflowId": "approval_workflow",
					"namespace":  "test-ns",
				},
			},
		},
		{
			name:       "workflow not found",
			namespace:  "test-ns",
			workflowID: "non-existent-workflow",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("ExecuteWorkflow", mock.Anything, mock.AnythingOfType("*domain.ExecutionRequest")).Return(nil, domain.ErrWorkflowNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "WORKFLOW_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Workflow not found",
			},
		},
		{
			name:       "namespace not found",
			namespace:  "non-existent-ns",
			workflowID: "approval_workflow",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"income": 50000,
				},
			},
			setupMock: func(mockEngine *MockEngine) {
				mockEngine.On("ExecuteWorkflow", mock.Anything, mock.AnythingOfType("*domain.ExecutionRequest")).Return(nil, domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
		{
			name:        "invalid request body",
			namespace:   "test-ns",
			workflowID:  "approval_workflow",
			requestBody: map[string]interface{}{
				// Missing data field
			},
			setupMock:      func(mockEngine *MockEngine) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid request body: INVALID_EXECUTION_DATA: Invalid execution data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockEngine)
			handler := NewExecutionHandler(mockEngine)
			tt.setupMock(mockEngine)

			c, w := setupExecutionTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "namespace", Value: tt.namespace},
				{Key: "workflowId", Value: tt.workflowID},
			}

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/workflows/"+tt.workflowID+"/execute", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Add query parameters
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			c.Request = req

			handler.ExecuteWorkflow(c)

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
			if result, exists := tt.expectedBody["result"]; exists {
				assert.Equal(t, result, response["result"])
			}
			if timestamp, exists := tt.expectedBody["timestamp"]; exists {
				assert.Equal(t, timestamp, response["timestamp"])
			}
			if metadata, exists := tt.expectedBody["metadata"]; exists {
				assert.Equal(t, metadata, response["metadata"])
			}

			mockEngine.AssertExpectations(t)
		})
	}
}

func TestExecutionHandler_EdgeCases(t *testing.T) {
	t.Run("empty request body", func(t *testing.T) {
		mockEngine := new(MockEngine)
		handler := NewExecutionHandler(mockEngine)

		c, w := setupExecutionTestContext()

		// Set parameters
		c.Params = gin.Params{
			{Key: "namespace", Value: "test-ns"},
			{Key: "ruleId", Value: "approval_rule"},
		}

		// Create empty request body
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/rules/approval_rule/execute", bytes.NewBufferString(""))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.ExecuteRule(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "BAD_REQUEST", response["error"])
		assert.Contains(t, response["message"].(string), "Invalid request body")
	})

	t.Run("trace parameter with different values", func(t *testing.T) {
		mockEngine := new(MockEngine)
		handler := NewExecutionHandler(mockEngine)

		c, w := setupExecutionTestContext()

		// Set parameters
		c.Params = gin.Params{
			{Key: "namespace", Value: "test-ns"},
			{Key: "ruleId", Value: "approval_rule"},
		}

		// Create request with trace=partial (should not enable trace)
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{
				"income": 50000,
			},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/rules/approval_rule/execute?trace=partial", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Setup mock to expect trace=false
		response := &domain.ExecutionResponse{
			Result:    true,
			Timestamp: time.Now(),
		}
		mockEngine.On("ExecuteRule", mock.Anything, mock.MatchedBy(func(req *domain.ExecutionRequest) bool {
			return req.Trace == false
		})).Return(response, nil)

		handler.ExecuteRule(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockEngine.AssertExpectations(t)
	})
}
