package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNamespaceRepository is a mock implementation of NamespaceRepository
type MockNamespaceRepository struct {
	mock.Mock
}

func (m *MockNamespaceRepository) Create(ctx context.Context, namespace *domain.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockNamespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceRepository) List(ctx context.Context) ([]*domain.Namespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNamespaceService_CreateNamespace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		// Mock that namespace doesn't exist (using sql.ErrNoRows which contains "no rows in result set")
		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), sql.ErrNoRows)
		mockRepo.On("Create", ctx, namespace).Return(nil)

		err := service.CreateNamespace(ctx, namespace)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - empty ID", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - ID too long", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "this-is-a-very-long-namespace-id-that-exceeds-the-maximum-allowed-length",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - invalid characters", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test@namespace",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - starts with hyphen", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "-test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - empty createdBy", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "VALIDATION_ERROR", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - description too long", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		longDescription := ""
		for i := 0; i < 501; i++ {
			longDescription += "a"
		}

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: longDescription,
			CreatedBy:   "user1",
		}

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_DESCRIPTION", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("namespace already exists", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "existing-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		existingNamespace := &domain.Namespace{
			ID:          "existing-ns",
			Description: "Existing namespace",
			CreatedBy:   "user2",
		}

		mockRepo.On("GetByID", ctx, "existing-ns").Return(existingNamespace, nil)

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "NAMESPACE_ALREADY_EXISTS", apiErr.Code)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error during existence check", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), assert.AnError)

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INTERNAL_ERROR", apiErr.Code)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error during creation", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), sql.ErrNoRows)
		mockRepo.On("Create", ctx, namespace).Return(assert.AnError)

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil namespace", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		err := service.CreateNamespace(ctx, nil)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "VALIDATION_ERROR", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestNamespaceService_GetNamespace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		expectedNamespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return(expectedNamespace, nil)

		result, err := service.GetNamespace(ctx, "test-ns")

		assert.NoError(t, err)
		assert.Equal(t, expectedNamespace, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty ID", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		result, err := service.GetNamespace(ctx, "")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		assert.Nil(t, result)
		mockRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		result, err := service.GetNamespace(ctx, "non-existent")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "NAMESPACE_NOT_FOUND", apiErr.Code)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNamespaceService_ListNamespaces(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		expectedNamespaces := []*domain.Namespace{
			{
				ID:          "ns1",
				Description: "Namespace 1",
				CreatedBy:   "user1",
			},
			{
				ID:          "ns2",
				Description: "Namespace 2",
				CreatedBy:   "user2",
			},
		}

		mockRepo.On("List", ctx).Return(expectedNamespaces, nil)

		result, err := service.ListNamespaces(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedNamespaces, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("List", ctx).Return([]*domain.Namespace{}, nil)

		result, err := service.ListNamespaces(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("List", ctx).Return(([]*domain.Namespace)(nil), assert.AnError)

		result, err := service.ListNamespaces(ctx)

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "LIST_ERROR", apiErr.Code)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNamespaceService_DeleteNamespace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		existingNamespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return(existingNamespace, nil)
		mockRepo.On("Delete", ctx, "test-ns").Return(nil)

		err := service.DeleteNamespace(ctx, "test-ns")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty ID", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		err := service.DeleteNamespace(ctx, "")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INVALID_NAMESPACE_ID", apiErr.Code)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "Delete")
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		err := service.DeleteNamespace(ctx, "non-existent")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "NAMESPACE_NOT_FOUND", apiErr.Code)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Delete")
	})

	t.Run("repository error during existence check", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), assert.AnError)

		err := service.DeleteNamespace(ctx, "test-ns")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "INTERNAL_ERROR", apiErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error during deletion", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		existingNamespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return(existingNamespace, nil)
		mockRepo.On("Delete", ctx, "test-ns").Return(assert.AnError)

		err := service.DeleteNamespace(ctx, "test-ns")

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("namespace not found (nil return)", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent").Return((*domain.Namespace)(nil), nil)

		err := service.DeleteNamespace(ctx, "non-existent")

		assert.Error(t, err)
		apiErr, ok := err.(*domain.APIError)
		assert.True(t, ok)
		assert.Equal(t, "NAMESPACE_NOT_FOUND", apiErr.Code)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Delete")
	})
}

func TestIsValidNamespaceID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid alphanumeric", "test123", true},
		{"valid with hyphens", "test-namespace", true},
		{"valid with underscores", "test_namespace", true},
		{"valid mixed", "test-namespace_123", true},
		{"empty string", "", false},
		{"starts with hyphen", "-test", false},
		{"starts with underscore", "_test", false},
		{"ends with hyphen", "test-", false},
		{"ends with underscore", "test_", false},
		{"contains special chars", "test@namespace", false},
		{"contains spaces", "test namespace", false},
		{"contains dots", "test.namespace", false},
		{"uppercase letters", "TEST", true},
		{"single character", "a", true},
		{"single digit", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNamespaceID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}
