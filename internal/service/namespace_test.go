package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
)

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

	t.Run("namespace already exists", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		existingNamespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Existing namespace",
			CreatedBy:   "user2",
		}

		mockRepo.On("GetByID", ctx, "test-ns").Return(existingNamespace, nil)

		err := service.CreateNamespace(ctx, namespace)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNamespaceAlreadyExists, err)
		mockRepo.AssertExpectations(t)
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

		namespace, err := service.GetNamespace(ctx, "test-ns")

		assert.NoError(t, err)
		assert.Equal(t, expectedNamespace, namespace)
		mockRepo.AssertExpectations(t)
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		namespace, err := service.GetNamespace(ctx, "test-ns")

		assert.Error(t, err)
		assert.Nil(t, namespace)
		assert.Equal(t, domain.ErrNamespaceNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), assert.AnError)

		namespace, err := service.GetNamespace(ctx, "test-ns")

		assert.Error(t, err)
		assert.Nil(t, namespace)
		assert.Equal(t, domain.ErrInternalError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestNamespaceService_ListNamespaces(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
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

		namespaces, err := service.ListNamespaces(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedNamespaces, namespaces)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		mockRepo.On("List", ctx).Return(([]*domain.Namespace)(nil), assert.AnError)

		namespaces, err := service.ListNamespaces(ctx)

		assert.Error(t, err)
		assert.Nil(t, namespaces)
		assert.Equal(t, domain.ErrListError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestNamespaceService_DeleteNamespace(t *testing.T) {
	ctx := context.Background()

	t.Run("successful_deletion", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		// Mock GetByID to return a valid namespace
		mockRepo.On("GetByID", ctx, "test-ns").Return(&domain.Namespace{ID: "test-ns"}, nil)
		mockRepo.On("Delete", ctx, "test-ns").Return(nil)

		err := service.DeleteNamespace(ctx, "test-ns")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockRepo := new(MockNamespaceRepository)
		service := NewNamespaceService(mockRepo)

		// Mock GetByID to return a valid namespace
		mockRepo.On("GetByID", ctx, "test-ns").Return(&domain.Namespace{ID: "test-ns"}, nil)
		mockRepo.On("Delete", ctx, "test-ns").Return(assert.AnError)

		err := service.DeleteNamespace(ctx, "test-ns")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestIsValidNamespaceID(t *testing.T) {
	t.Run("valid namespace IDs", func(t *testing.T) {
		validIDs := []string{
			"test",
			"test-namespace",
			"test123",
			"test_namespace",
			"test-namespace-123",
			"a",
			"test-namespace-with-very-long-name-that-is-still-valid",
		}

		for _, id := range validIDs {
			t.Run(id, func(t *testing.T) {
				assert.True(t, isValidNamespaceID(id), "Expected %s to be valid", id)
			})
		}
	})

	t.Run("invalid namespace IDs", func(t *testing.T) {
		invalidIDs := []string{
			"",
			"-test",
			"test-",
			"test@namespace",
			"test#namespace",
			"test$namespace",
			"test%namespace",
			"test^namespace",
			"test&namespace",
			"test*namespace",
			"test(namespace",
			"test)namespace",
			"test+namespace",
			"test=namespace",
			"test[namespace",
			"test]namespace",
			"test{namespace",
			"test}namespace",
			"test|namespace",
			"test\\namespace",
			"test:namespace",
			"test;namespace",
			"test\"namespace",
			"test'namespace",
			"test<namespace",
			"test>namespace",
			"test,namespace",
			"test.namespace",
			"test/namespace",
			"test?namespace",
			"test~namespace",
			"test`namespace",
			"test!namespace",
		}

		for _, id := range invalidIDs {
			t.Run(id, func(t *testing.T) {
				assert.False(t, isValidNamespaceID(id), "Expected %s to be invalid", id)
			})
		}
	})
}
