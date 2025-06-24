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

// MockWorkflowService is a mock implementation of WorkflowServiceInterface
type MockWorkflowService struct {
	mock.Mock
}

// Ensure MockWorkflowService implements WorkflowServiceInterface
var _ service.WorkflowServiceInterface = (*MockWorkflowService)(nil)

func (m *MockWorkflowService) Create(ctx context.Context, workflow *domain.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowService) GetByID(ctx context.Context, namespace, workflowID string, version int32) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) GetActiveVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) GetDraftVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) List(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) ListActive(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) ListVersions(ctx context.Context, namespace, workflowID string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowService) Update(ctx context.Context, workflow *domain.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowService) Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error {
	args := m.Called(ctx, namespace, workflowID, version, publishedBy)
	return args.Error(0)
}

func (m *MockWorkflowService) Deactivate(ctx context.Context, namespace, workflowID string) error {
	args := m.Called(ctx, namespace, workflowID)
	return args.Error(0)
}

func (m *MockWorkflowService) Delete(ctx context.Context, namespace, workflowID string, version int32) error {
	args := m.Called(ctx, namespace, workflowID, version)
	return args.Error(0)
}

func setupWorkflowTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a proper request with context
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	// Set default client_id for tests
	c.Set("client_id", "test-client")

	return c, w
}

func TestNewWorkflowHandler(t *testing.T) {
	mockService := new(MockWorkflowService)
	handler := NewWorkflowHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.workflowService)
	assert.NotNil(t, handler.responseHandler)
}

func TestWorkflowHandler_CreateWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		requestBody    map[string]interface{}
		setupMock      func(*MockWorkflowService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful creation",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":      "approval_workflow",
				"startAt": "start",
				"steps": map[string]interface{}{
					"start": map[string]interface{}{
						"type":    "rule",
						"ruleId":  "check_eligibility",
						"onTrue":  "approve",
						"onFalse": "reject",
					},
					"approve": map[string]interface{}{
						"type": "terminal",
					},
					"reject": map[string]interface{}{
						"type": "terminal",
					},
				},
			},
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(w *domain.Workflow) bool {
					return w.WorkflowID == "approval_workflow" && w.StartAt == "start"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"workflowId": "approval_workflow",
					"startAt":    "start",
					"createdBy":  "test-client",
					"status":     "draft",
					"version":    float64(1),
				},
			},
		},
		{
			name:      "missing workflow ID",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"startAt": "start",
				"steps":   map[string]interface{}{},
			},
			setupMock:      func(mockService *MockWorkflowService) {},
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
				"id":      "approval_workflow",
				"startAt": "start",
				"steps":   map[string]interface{}{},
			},
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace is required",
			},
		},
		{
			name:      "workflow already exists",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":      "approval_workflow",
				"startAt": "start",
				"steps":   map[string]interface{}{},
			},
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*domain.Workflow")).Return(domain.ErrWorkflowAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"code":    "WORKFLOW_ALREADY_EXISTS",
				"error":   "CONFLICT",
				"message": "Workflow already exists",
			},
		},
		{
			name:      "namespace not found",
			namespace: "test-ns",
			requestBody: map[string]interface{}{
				"id":      "approval_workflow",
				"startAt": "start",
				"steps":   map[string]interface{}{},
			},
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*domain.Workflow")).Return(domain.ErrNamespaceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "NAMESPACE_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Namespace not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkflowService)
			handler := NewWorkflowHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupWorkflowTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/workflows", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.CreateWorkflow(c)

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

			mockService.AssertExpectations(t)
		})
	}
}

func TestWorkflowHandler_GetWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		workflowID     string
		setupMock      func(*MockWorkflowService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful retrieval - active version",
			namespace:  "test-ns",
			workflowID: "approval_workflow",
			setupMock: func(mockService *MockWorkflowService) {
				workflow := &domain.Workflow{
					WorkflowID: "approval_workflow",
					StartAt:    "start",
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "active",
					Version:    1,
				}
				mockService.On("GetActiveVersion", mock.Anything, "test-ns", "approval_workflow").Return(workflow, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"workflowId": "approval_workflow",
					"startAt":    "start",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "active",
					"version":    float64(1),
				},
			},
		},
		{
			name:       "successful retrieval - draft version",
			namespace:  "test-ns",
			workflowID: "approval_workflow",
			setupMock: func(mockService *MockWorkflowService) {
				workflow := &domain.Workflow{
					WorkflowID: "approval_workflow",
					StartAt:    "start",
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "draft",
					Version:    1,
				}
				mockService.On("GetActiveVersion", mock.Anything, "test-ns", "approval_workflow").Return(nil, domain.ErrWorkflowNotFound)
				mockService.On("GetDraftVersion", mock.Anything, "test-ns", "approval_workflow").Return(workflow, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"workflowId": "approval_workflow",
					"startAt":    "start",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "draft",
					"version":    float64(1),
				},
			},
		},
		{
			name:       "workflow not found",
			namespace:  "test-ns",
			workflowID: "non-existent-workflow",
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("GetActiveVersion", mock.Anything, "test-ns", "non-existent-workflow").Return(nil, domain.ErrWorkflowNotFound)
				mockService.On("GetDraftVersion", mock.Anything, "test-ns", "non-existent-workflow").Return(nil, domain.ErrWorkflowNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "WORKFLOW_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Workflow not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			workflowID:     "approval_workflow",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and workflow ID are required",
			},
		},
		{
			name:           "empty workflow ID",
			namespace:      "test-ns",
			workflowID:     "",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace and workflow ID are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkflowService)
			handler := NewWorkflowHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupWorkflowTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "workflowId", Value: tt.workflowID},
			}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/workflows/"+tt.workflowID, nil)
			c.Request = req

			handler.GetWorkflow(c)

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

			mockService.AssertExpectations(t)
		})
	}
}

func TestWorkflowHandler_GetWorkflowVersion(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		workflowID     string
		version        string
		setupMock      func(*MockWorkflowService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful retrieval",
			namespace:  "test-ns",
			workflowID: "approval_workflow",
			version:    "2",
			setupMock: func(mockService *MockWorkflowService) {
				workflow := &domain.Workflow{
					WorkflowID: "approval_workflow",
					StartAt:    "start",
					CreatedBy:  "test-client",
					CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:     "active",
					Version:    2,
				}
				mockService.On("GetByID", mock.Anything, "test-ns", "approval_workflow", int32(2)).Return(workflow, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"workflowId": "approval_workflow",
					"startAt":    "start",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "active",
					"version":    float64(2),
				},
			},
		},
		{
			name:           "invalid version number",
			namespace:      "test-ns",
			workflowID:     "approval_workflow",
			version:        "invalid",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Invalid version number",
			},
		},
		{
			name:       "workflow not found",
			namespace:  "test-ns",
			workflowID: "non-existent-workflow",
			version:    "1",
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("GetByID", mock.Anything, "test-ns", "non-existent-workflow", int32(1)).Return(nil, domain.ErrWorkflowNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "WORKFLOW_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Workflow not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			workflowID:     "approval_workflow",
			version:        "1",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace, workflow ID, and version are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkflowService)
			handler := NewWorkflowHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupWorkflowTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "workflowId", Value: tt.workflowID},
				{Key: "version", Value: tt.version},
			}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/workflows/"+tt.workflowID+"/versions/"+tt.version, nil)
			c.Request = req

			handler.GetWorkflowVersion(c)

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

			mockService.AssertExpectations(t)
		})
	}
}

func TestWorkflowHandler_ListWorkflows(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockWorkflowService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "successful list with workflows",
			namespace: "test-ns",
			setupMock: func(mockService *MockWorkflowService) {
				workflows := []*domain.Workflow{
					{
						WorkflowID: "approval_workflow",
						StartAt:    "start",
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						Status:     "active",
						Version:    1,
					},
					{
						WorkflowID: "rejection_workflow",
						StartAt:    "start",
						CreatedBy:  "test-client",
						CreatedAt:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
						Status:     "draft",
						Version:    1,
					},
				}
				mockService.On("List", mock.Anything, "test-ns").Return(workflows, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{
					"workflowId": "approval_workflow",
					"startAt":    "start",
					"createdBy":  "test-client",
					"createdAt":  "2023-01-01T00:00:00Z",
					"status":     "active",
					"version":    float64(1),
				},
				{
					"workflowId": "rejection_workflow",
					"startAt":    "start",
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
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("List", mock.Anything, "test-ns").Return([]*domain.Workflow{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []map[string]interface{}{},
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-ns",
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("List", mock.Anything, "non-existent-ns").Return(nil, domain.ErrNamespaceNotFound)
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
			setupMock:      func(mockService *MockWorkflowService) {},
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
			mockService := new(MockWorkflowService)
			handler := NewWorkflowHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupWorkflowTestContext()

			// Set namespace parameter
			c.Params = gin.Params{{Key: "id", Value: tt.namespace}}

			// Create request
			req := httptest.NewRequest("GET", "/v1/namespaces/"+tt.namespace+"/workflows", nil)
			c.Request = req

			handler.ListWorkflows(c)

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

func TestWorkflowHandler_PublishWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		workflowID     string
		version        string
		setupMock      func(*MockWorkflowService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful publish",
			namespace:  "test-ns",
			workflowID: "approval_workflow",
			version:    "1",
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("Publish", mock.Anything, "test-ns", "approval_workflow", int32(1), "test-client").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "active",
			},
		},
		{
			name:       "workflow not found",
			namespace:  "test-ns",
			workflowID: "non-existent-workflow",
			version:    "1",
			setupMock: func(mockService *MockWorkflowService) {
				mockService.On("Publish", mock.Anything, "test-ns", "non-existent-workflow", int32(1), "test-client").Return(domain.ErrWorkflowNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    "WORKFLOW_NOT_FOUND",
				"error":   "NOT_FOUND",
				"message": "Workflow not found",
			},
		},
		{
			name:           "empty namespace",
			namespace:      "",
			workflowID:     "approval_workflow",
			version:        "1",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace, workflow ID, and version are required",
			},
		},
		{
			name:           "empty workflow ID",
			namespace:      "test-ns",
			workflowID:     "",
			version:        "1",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace, workflow ID, and version are required",
			},
		},
		{
			name:           "empty version",
			namespace:      "test-ns",
			workflowID:     "approval_workflow",
			version:        "",
			setupMock:      func(mockService *MockWorkflowService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    "BAD_REQUEST",
				"error":   "BAD_REQUEST",
				"message": "Namespace, workflow ID, and version are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkflowService)
			handler := NewWorkflowHandler(mockService)
			tt.setupMock(mockService)

			c, w := setupWorkflowTestContext()

			// Set parameters
			c.Params = gin.Params{
				{Key: "id", Value: tt.namespace},
				{Key: "workflowId", Value: tt.workflowID},
				{Key: "version", Value: tt.version},
			}

			// Create request
			req := httptest.NewRequest("POST", "/v1/namespaces/"+tt.namespace+"/workflows/"+tt.workflowID+"/versions/"+tt.version+"/publish", nil)
			c.Request = req

			handler.PublishWorkflow(c)

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

func TestWorkflowHandler_EdgeCases(t *testing.T) {
	t.Run("missing client_id in context", func(t *testing.T) {
		mockService := new(MockWorkflowService)
		handler := NewWorkflowHandler(mockService)

		c, w := setupWorkflowTestContext()
		// Remove client_id from context
		c.Set("client_id", nil)

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create request body
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"id":      "approval_workflow",
			"startAt": "start",
			"steps":   map[string]interface{}{},
		})
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/workflows", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateWorkflow(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response["error"])
		assert.Equal(t, "Client ID not found", response["message"])
	})

	t.Run("malformed JSON in create request", func(t *testing.T) {
		mockService := new(MockWorkflowService)
		handler := NewWorkflowHandler(mockService)

		c, w := setupWorkflowTestContext()

		// Set namespace parameter
		c.Params = gin.Params{{Key: "id", Value: "test-ns"}}

		// Create malformed JSON request body
		req := httptest.NewRequest("POST", "/v1/namespaces/test-ns/workflows", bytes.NewBufferString(`{"id": "approval_workflow"`))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateWorkflow(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "BAD_REQUEST", response["error"])
		assert.Equal(t, "Invalid request body", response["message"])
	})
}
