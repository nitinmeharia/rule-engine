package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rule-engine/internal/domain"
)

func TestFieldService_CreateField(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		field         *domain.Field
		setupMock     func(*MockFieldRepository)
		expectedError bool
	}{
		{
			name:      "successful creation",
			namespace: "test-namespace",
			field: &domain.Field{
				FieldID:     "test_field",
				Type:        "string",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("NamespaceExists", mock.Anything, "test-namespace").Return(true, nil)
				mockRepo.On("Exists", mock.Anything, "test-namespace", "test_field").Return(false, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Field")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "field already exists",
			namespace: "test-namespace",
			field: &domain.Field{
				FieldID:     "existing_field",
				Type:        "string",
				Description: "Existing field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("NamespaceExists", mock.Anything, "test-namespace").Return(true, nil)
				mockRepo.On("Exists", mock.Anything, "test-namespace", "existing_field").Return(true, nil)
			},
			expectedError: true,
		},
		{
			name:      "namespace not found",
			namespace: "non-existent-namespace",
			field: &domain.Field{
				FieldID:     "test_field",
				Type:        "string",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("NamespaceExists", mock.Anything, "non-existent-namespace").Return(false, nil)
			},
			expectedError: true,
		},
		{
			name:      "invalid field type",
			namespace: "test-namespace",
			field: &domain.Field{
				FieldID:     "test_field",
				Type:        "invalid_type",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				// No mock setup needed for validation error
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFieldRepository{}
			tt.setupMock(mockRepo)

			service := NewFieldService(mockRepo)
			err := service.CreateField(context.Background(), tt.namespace, tt.field)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.namespace, tt.field.Namespace)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFieldService_GetField(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		fieldID       string
		setupMock     func(*MockFieldRepository)
		expectedField *domain.Field
		expectedError bool
	}{
		{
			name:      "successful retrieval",
			namespace: "test-namespace",
			fieldID:   "test_field",
			setupMock: func(mockRepo *MockFieldRepository) {
				field := &domain.Field{
					Namespace:   "test-namespace",
					FieldID:     "test_field",
					Type:        "string",
					Description: "Test field",
					CreatedBy:   "test-user",
				}
				mockRepo.On("GetByID", mock.Anything, "test-namespace", "test_field").Return(field, nil)
			},
			expectedField: &domain.Field{
				Namespace:   "test-namespace",
				FieldID:     "test_field",
				Type:        "string",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			expectedError: false,
		},
		{
			name:      "field not found",
			namespace: "test-namespace",
			fieldID:   "non-existent-field",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("GetByID", mock.Anything, "test-namespace", "non-existent-field").Return(nil, assert.AnError)
			},
			expectedField: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFieldRepository{}
			tt.setupMock(mockRepo)

			service := NewFieldService(mockRepo)
			field, err := service.GetField(context.Background(), tt.namespace, tt.fieldID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, field)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedField, field)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFieldService_ListFields(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		setupMock      func(*MockFieldRepository)
		expectedFields []*domain.Field
		expectedError  bool
	}{
		{
			name:      "successful listing",
			namespace: "test-namespace",
			setupMock: func(mockRepo *MockFieldRepository) {
				fields := []*domain.Field{
					{
						Namespace:   "test-namespace",
						FieldID:     "field1",
						Type:        "string",
						Description: "Field 1",
						CreatedBy:   "test-user",
					},
					{
						Namespace:   "test-namespace",
						FieldID:     "field2",
						Type:        "number",
						Description: "Field 2",
						CreatedBy:   "test-user",
					},
				}
				mockRepo.On("List", mock.Anything, "test-namespace").Return(fields, nil)
			},
			expectedFields: []*domain.Field{
				{
					Namespace:   "test-namespace",
					FieldID:     "field1",
					Type:        "string",
					Description: "Field 1",
					CreatedBy:   "test-user",
				},
				{
					Namespace:   "test-namespace",
					FieldID:     "field2",
					Type:        "number",
					Description: "Field 2",
					CreatedBy:   "test-user",
				},
			},
			expectedError: false,
		},
		{
			name:      "empty list",
			namespace: "test-namespace",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("List", mock.Anything, "test-namespace").Return([]*domain.Field{}, nil)
			},
			expectedFields: []*domain.Field{},
			expectedError:  false,
		},
		{
			name:      "list error",
			namespace: "test-namespace",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("List", mock.Anything, "test-namespace").Return(nil, assert.AnError)
			},
			expectedFields: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFieldRepository{}
			tt.setupMock(mockRepo)

			service := NewFieldService(mockRepo)
			fields, err := service.ListFields(context.Background(), tt.namespace)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, fields)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFields, fields)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFieldService_UpdateField(t *testing.T) {
	tests := []struct {
		name          string
		field         *domain.Field
		setupMock     func(*MockFieldRepository)
		expectedError bool
	}{
		{
			name: "successful update",
			field: &domain.Field{
				Namespace:   "test-namespace",
				FieldID:     "test_field",
				Type:        "number",
				Description: "Updated field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("Exists", mock.Anything, "test-namespace", "test_field").Return(true, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Field")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "field not found",
			field: &domain.Field{
				Namespace:   "test-namespace",
				FieldID:     "non-existent-field",
				Type:        "string",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("Exists", mock.Anything, "test-namespace", "non-existent-field").Return(false, nil)
			},
			expectedError: true,
		},
		{
			name: "invalid field type",
			field: &domain.Field{
				Namespace:   "test-namespace",
				FieldID:     "test_field",
				Type:        "invalid_type",
				Description: "Test field",
				CreatedBy:   "test-user",
			},
			setupMock: func(mockRepo *MockFieldRepository) {
				// No mock setup needed for validation error
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFieldRepository{}
			tt.setupMock(mockRepo)

			service := NewFieldService(mockRepo)
			err := service.UpdateField(context.Background(), tt.field)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFieldService_DeleteField(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		fieldID       string
		setupMock     func(*MockFieldRepository)
		expectedError bool
	}{
		{
			name:      "successful deletion",
			namespace: "test-namespace",
			fieldID:   "test_field",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("Exists", mock.Anything, "test-namespace", "test_field").Return(true, nil)
				mockRepo.On("Delete", mock.Anything, "test-namespace", "test_field").Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "field not found",
			namespace: "test-namespace",
			fieldID:   "non-existent-field",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("Exists", mock.Anything, "test-namespace", "non-existent-field").Return(false, nil)
			},
			expectedError: true,
		},
		{
			name:      "delete error",
			namespace: "test-namespace",
			fieldID:   "test_field",
			setupMock: func(mockRepo *MockFieldRepository) {
				mockRepo.On("Exists", mock.Anything, "test-namespace", "test_field").Return(true, nil)
				mockRepo.On("Delete", mock.Anything, "test-namespace", "test_field").Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFieldRepository{}
			tt.setupMock(mockRepo)

			service := NewFieldService(mockRepo)
			err := service.DeleteField(context.Background(), tt.namespace, tt.fieldID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
