package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestTerminalService_CreateTerminal(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		terminal := &domain.Terminal{
			TerminalID: "test-terminal",
			CreatedBy:  "user1",
		}

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("Exists", ctx, "test-ns", "test-terminal").Return(false, nil)
		mockTerminalRepo.On("Create", ctx, terminal).Return(nil)

		err := service.CreateTerminal(ctx, "test-ns", terminal)

		assert.NoError(t, err)
		assert.Equal(t, "test-ns", terminal.Namespace)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("validation error - empty terminal ID", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		terminal := &domain.Terminal{
			TerminalID: "",
			CreatedBy:  "user1",
		}

		err := service.CreateTerminal(ctx, "test-ns", terminal)

		assert.Error(t, err)
		mockTerminalRepo.AssertNotCalled(t, "Create")
		mockNamespaceRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		terminal := &domain.Terminal{
			TerminalID: "test-terminal",
			CreatedBy:  "user1",
		}

		mockNamespaceRepo.On("GetByID", ctx, "non-existent").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		err := service.CreateTerminal(ctx, "non-existent", terminal)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNamespaceNotFound, err)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("terminal already exists", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		terminal := &domain.Terminal{
			TerminalID: "existing-terminal",
			CreatedBy:  "user1",
		}

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("Exists", ctx, "test-ns", "existing-terminal").Return(true, nil)

		err := service.CreateTerminal(ctx, "test-ns", terminal)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTerminalAlreadyExists, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})
}

func TestTerminalService_GetTerminal(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		expectedTerminal := &domain.Terminal{
			TerminalID: "test-terminal",
		}

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("GetByID", ctx, "test-ns", "test-terminal").Return(expectedTerminal, nil)

		terminal, err := service.GetTerminal(ctx, "test-ns", "test-terminal")

		assert.NoError(t, err)
		assert.Equal(t, expectedTerminal, terminal)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("terminal not found", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("GetByID", ctx, "test-ns", "non-existent").Return((*domain.Terminal)(nil), sql.ErrNoRows)

		terminal, err := service.GetTerminal(ctx, "test-ns", "non-existent")

		assert.Error(t, err)
		assert.Nil(t, terminal)
		assert.Equal(t, domain.ErrTerminalNotFound, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		mockNamespaceRepo.On("GetByID", ctx, "non-existent").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		terminal, err := service.GetTerminal(ctx, "non-existent", "test-terminal")

		assert.Error(t, err)
		assert.Nil(t, terminal)
		assert.Equal(t, domain.ErrNamespaceNotFound, err)
		mockNamespaceRepo.AssertExpectations(t)
	})
}

func TestTerminalService_ListTerminals(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		expectedTerminals := []*domain.Terminal{
			{TerminalID: "terminal1"},
			{TerminalID: "terminal2"},
		}

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("List", ctx, "test-ns").Return(expectedTerminals, nil)

		terminals, err := service.ListTerminals(ctx, "test-ns")

		assert.NoError(t, err)
		assert.Equal(t, expectedTerminals, terminals)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("List", ctx, "test-ns").Return(([]*domain.Terminal)(nil), assert.AnError)

		terminals, err := service.ListTerminals(ctx, "test-ns")

		assert.Error(t, err)
		assert.Nil(t, terminals)
		assert.Equal(t, domain.ErrListError, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})
}

func TestTerminalService_DeleteTerminal(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("Exists", ctx, "test-ns", "test-terminal").Return(true, nil)
		mockTerminalRepo.On("Delete", ctx, "test-ns", "test-terminal").Return(nil)

		err := service.DeleteTerminal(ctx, "test-ns", "test-terminal")

		assert.NoError(t, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("terminal not found", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("Exists", ctx, "test-ns", "test-terminal").Return(false, nil)

		err := service.DeleteTerminal(ctx, "test-ns", "test-terminal")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTerminalNotFound, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})

	t.Run("deletion error", func(t *testing.T) {
		mockTerminalRepo := new(MockTerminalRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)
		service := NewTerminalService(mockTerminalRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockTerminalRepo.On("Exists", ctx, "test-ns", "test-terminal").Return(true, nil)
		mockTerminalRepo.On("Delete", ctx, "test-ns", "test-terminal").Return(assert.AnError)

		err := service.DeleteTerminal(ctx, "test-ns", "test-terminal")

		assert.Error(t, err)
		mockTerminalRepo.AssertExpectations(t)
		mockNamespaceRepo.AssertExpectations(t)
	})
}
