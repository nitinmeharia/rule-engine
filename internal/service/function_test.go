package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rule-engine/internal/domain"
)

func TestFunctionService_CreateFunction(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		function      *domain.Function
		setupMock     func(*MockFunctionRepository, *MockNamespaceRepository)
		expectedError bool
	}{
		{
			name:      "successful creation - max function",
			namespace: "test-namespace",
			function: &domain.Function{
				FunctionID: "test_function",
				Type:       "max",
				Args:       []string{"field1", "field2"},
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "test_function").Return(nil, assert.AnError)
				mockFuncRepo.On("GetActiveVersion", mock.Anything, "test-namespace", "test_function").Return(nil, assert.AnError)
				mockFuncRepo.On("GetMaxVersion", mock.Anything, "test-namespace", "test_function").Return(int32(0), nil)
				mockFuncRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Function")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "successful creation - in function",
			namespace: "test-namespace",
			function: &domain.Function{
				FunctionID: "test_function",
				Type:       "in",
				Values:     []string{"value1", "value2"},
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "test_function").Return(nil, assert.AnError)
				mockFuncRepo.On("GetActiveVersion", mock.Anything, "test-namespace", "test_function").Return(nil, assert.AnError)
				mockFuncRepo.On("GetMaxVersion", mock.Anything, "test-namespace", "test_function").Return(int32(0), nil)
				mockFuncRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Function")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "function already exists - draft",
			namespace: "test-namespace",
			function: &domain.Function{
				FunctionID: "existing_function",
				Type:       "max",
				Args:       []string{"field1"},
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				existingFunction := &domain.Function{
					Namespace:  "test-namespace",
					FunctionID: "existing_function",
					Version:    1,
					Type:       "max",
					Args:       []string{"field1"},
					Status:     "draft",
					CreatedBy:  "test-user",
				}
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "existing_function").Return(existingFunction, nil)
			},
			expectedError: true,
		},
		{
			name:      "invalid function type",
			namespace: "test-namespace",
			function: &domain.Function{
				FunctionID: "test_function",
				Type:       "invalid_type",
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				// No mock setup needed for validation error
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			mockNsRepo := &MockNamespaceRepository{}
			tt.setupMock(mockFuncRepo, mockNsRepo)

			service := NewFunctionService(mockFuncRepo, mockNsRepo)
			err := service.CreateFunction(context.Background(), tt.namespace, tt.function)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.namespace, tt.function.Namespace)
				assert.Equal(t, int32(1), tt.function.Version)
				assert.Equal(t, "draft", tt.function.Status)
				assert.NotEmpty(t, tt.function.ReturnType)
			}

			mockFuncRepo.AssertExpectations(t)
			mockNsRepo.AssertExpectations(t)
		})
	}
}

func TestFunctionService_GetFunction(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		functionID       string
		setupMock        func(*MockFunctionRepository)
		expectedFunction *domain.Function
		expectedError    bool
	}{
		{
			name:       "successful retrieval",
			namespace:  "test-namespace",
			functionID: "test_function",
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				function := &domain.Function{
					Namespace:  "test-namespace",
					FunctionID: "test_function",
					Type:       "max",
					Args:       []string{"field1", "field2"},
					Status:     "active",
					CreatedBy:  "test-user",
				}
				mockFuncRepo.On("GetActiveVersion", mock.Anything, "test-namespace", "test_function").Return(function, nil)
			},
			expectedFunction: &domain.Function{
				Namespace:  "test-namespace",
				FunctionID: "test_function",
				Type:       "max",
				Args:       []string{"field1", "field2"},
				Status:     "active",
				CreatedBy:  "test-user",
			},
			expectedError: false,
		},
		{
			name:       "function not found",
			namespace:  "test-namespace",
			functionID: "non-existent-function",
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("GetActiveVersion", mock.Anything, "test-namespace", "non-existent-function").Return(nil, assert.AnError)
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "non-existent-function").Return(nil, assert.AnError)
			},
			expectedFunction: nil,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			tt.setupMock(mockFuncRepo)

			service := NewFunctionService(mockFuncRepo, nil)
			function, err := service.GetFunction(context.Background(), tt.namespace, tt.functionID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, function)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFunction, function)
			}

			mockFuncRepo.AssertExpectations(t)
		})
	}
}

func TestFunctionService_ListFunctions(t *testing.T) {
	tests := []struct {
		name              string
		namespace         string
		setupMock         func(*MockFunctionRepository)
		expectedFunctions []*domain.Function
		expectedError     bool
	}{
		{
			name:      "successful listing",
			namespace: "test-namespace",
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				functions := []*domain.Function{
					{
						Namespace:  "test-namespace",
						FunctionID: "function1",
						Type:       "max",
						Args:       []string{"field1"},
						Status:     "active",
						CreatedBy:  "test-user",
					},
					{
						Namespace:  "test-namespace",
						FunctionID: "function2",
						Type:       "in",
						Values:     []string{"value1", "value2"},
						Status:     "draft",
						CreatedBy:  "test-user",
					},
				}
				mockFuncRepo.On("List", mock.Anything, "test-namespace").Return(functions, nil)
			},
			expectedFunctions: []*domain.Function{
				{
					Namespace:  "test-namespace",
					FunctionID: "function1",
					Type:       "max",
					Args:       []string{"field1"},
					Status:     "active",
					CreatedBy:  "test-user",
				},
				{
					Namespace:  "test-namespace",
					FunctionID: "function2",
					Type:       "in",
					Values:     []string{"value1", "value2"},
					Status:     "draft",
					CreatedBy:  "test-user",
				},
			},
			expectedError: false,
		},
		{
			name:      "empty list",
			namespace: "test-namespace",
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("List", mock.Anything, "test-namespace").Return([]*domain.Function{}, nil)
			},
			expectedFunctions: []*domain.Function{},
			expectedError:     false,
		},
		{
			name:      "list error",
			namespace: "test-namespace",
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("List", mock.Anything, "test-namespace").Return(nil, assert.AnError)
			},
			expectedFunctions: nil,
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			tt.setupMock(mockFuncRepo)

			service := NewFunctionService(mockFuncRepo, nil)
			functions, err := service.ListFunctions(context.Background(), tt.namespace)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, functions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFunctions, functions)
			}

			mockFuncRepo.AssertExpectations(t)
		})
	}
}

func TestFunctionService_UpdateFunction(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		functionID    string
		function      *domain.Function
		setupMock     func(*MockFunctionRepository)
		expectedError bool
	}{
		{
			name:       "successful update",
			namespace:  "test-namespace",
			functionID: "test_function",
			function: &domain.Function{
				FunctionID: "test_function",
				Type:       "max",
				Args:       []string{"field1", "field2", "field3"},
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				existingFunction := &domain.Function{
					Namespace:  "test-namespace",
					FunctionID: "test_function",
					Version:    1,
					Type:       "max",
					Args:       []string{"field1"},
					Status:     "draft",
					CreatedBy:  "test-user",
				}
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "test_function").Return(existingFunction, nil)
				mockFuncRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Function")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:       "function not found",
			namespace:  "test-namespace",
			functionID: "non-existent-function",
			function: &domain.Function{
				FunctionID: "non-existent-function",
				Type:       "max",
				Args:       []string{"field1"},
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "non-existent-function").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name:       "invalid function type",
			namespace:  "test-namespace",
			functionID: "test_function",
			function: &domain.Function{
				FunctionID: "test_function",
				Type:       "invalid_type",
				CreatedBy:  "test-user",
			},
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				// No mock setup needed for validation error
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			tt.setupMock(mockFuncRepo)

			service := NewFunctionService(mockFuncRepo, nil)
			err := service.UpdateFunction(context.Background(), tt.namespace, tt.functionID, tt.function)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockFuncRepo.AssertExpectations(t)
		})
	}
}

func TestFunctionService_PublishFunction(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		functionID    string
		publishedBy   string
		setupMock     func(*MockFunctionRepository, *MockNamespaceRepository)
		expectedError bool
	}{
		{
			name:        "successful publish",
			namespace:   "test-namespace",
			functionID:  "test_function",
			publishedBy: "test-user",
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				draftFunction := &domain.Function{
					Namespace:  "test-namespace",
					FunctionID: "test_function",
					Version:    1,
					Type:       "max",
					Args:       []string{"field1"},
					Status:     "draft",
					CreatedBy:  "test-user",
				}
				namespace := &domain.Namespace{
					ID: "test-namespace",
				}
				mockNsRepo.On("GetByID", mock.Anything, "test-namespace").Return(namespace, nil)
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "test_function").Return(draftFunction, nil)
				mockFuncRepo.On("Publish", mock.Anything, "test-namespace", "test_function", int32(1), "test-user").Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "namespace not found",
			namespace:   "non-existent-namespace",
			functionID:  "test_function",
			publishedBy: "test-user",
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				mockNsRepo.On("GetByID", mock.Anything, "non-existent-namespace").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name:        "function not found",
			namespace:   "test-namespace",
			functionID:  "non-existent-function",
			publishedBy: "test-user",
			setupMock: func(mockFuncRepo *MockFunctionRepository, mockNsRepo *MockNamespaceRepository) {
				namespace := &domain.Namespace{
					ID: "test-namespace",
				}
				mockNsRepo.On("GetByID", mock.Anything, "test-namespace").Return(namespace, nil)
				mockFuncRepo.On("GetDraftVersion", mock.Anything, "test-namespace", "non-existent-function").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			mockNsRepo := &MockNamespaceRepository{}
			tt.setupMock(mockFuncRepo, mockNsRepo)

			service := NewFunctionService(mockFuncRepo, mockNsRepo)
			err := service.PublishFunction(context.Background(), tt.namespace, tt.functionID, tt.publishedBy)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockFuncRepo.AssertExpectations(t)
			mockNsRepo.AssertExpectations(t)
		})
	}
}

func TestFunctionService_DeleteFunction(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		functionID    string
		version       int32
		setupMock     func(*MockFunctionRepository)
		expectedError bool
	}{
		{
			name:       "successful deletion",
			namespace:  "test-namespace",
			functionID: "test_function",
			version:    1,
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("Exists", mock.Anything, "test-namespace", "test_function", int32(1)).Return(true, nil)
				mockFuncRepo.On("Delete", mock.Anything, "test-namespace", "test_function", int32(1)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:       "delete error",
			namespace:  "test-namespace",
			functionID: "test_function",
			version:    1,
			setupMock: func(mockFuncRepo *MockFunctionRepository) {
				mockFuncRepo.On("Exists", mock.Anything, "test-namespace", "test_function", int32(1)).Return(true, nil)
				mockFuncRepo.On("Delete", mock.Anything, "test-namespace", "test_function", int32(1)).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFuncRepo := &MockFunctionRepository{}
			tt.setupMock(mockFuncRepo)

			service := NewFunctionService(mockFuncRepo, nil)
			err := service.DeleteFunction(context.Background(), tt.namespace, tt.functionID, tt.version)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockFuncRepo.AssertExpectations(t)
		})
	}
}
