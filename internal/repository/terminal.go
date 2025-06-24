package repository

import (
	"context"
	"fmt"

	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// TerminalRepository implements domain.TerminalRepository
type TerminalRepository struct {
	queries db.Querier
}

// NewTerminalRepository creates a new terminal repository
func NewTerminalRepository(queries db.Querier) *TerminalRepository {
	return &TerminalRepository{queries: queries}
}

// Create creates a new terminal
func (r *TerminalRepository) Create(ctx context.Context, terminal *domain.Terminal) error {
	err := r.queries.CreateTerminal(ctx, db.CreateTerminalParams{
		Namespace:  terminal.Namespace,
		TerminalID: terminal.TerminalID,
		CreatedBy:  terminal.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create terminal: %w", err)
	}
	return nil
}

// GetByID retrieves a terminal by ID
func (r *TerminalRepository) GetByID(ctx context.Context, namespace, terminalID string) (*domain.Terminal, error) {
	terminal, err := r.queries.GetTerminal(ctx, db.GetTerminalParams{
		Namespace:  namespace,
		TerminalID: terminalID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Terminal{
		Namespace:  terminal.Namespace,
		TerminalID: terminal.TerminalID,
		CreatedAt:  terminal.CreatedAt.Time,
		CreatedBy:  terminal.CreatedBy,
	}, nil
}

// List retrieves all terminals for a namespace
func (r *TerminalRepository) List(ctx context.Context, namespace string) ([]*domain.Terminal, error) {
	terminals, err := r.queries.ListTerminals(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminals: %w", err)
	}

	result := make([]*domain.Terminal, len(terminals))
	for i, terminal := range terminals {
		result[i] = &domain.Terminal{
			Namespace:  terminal.Namespace,
			TerminalID: terminal.TerminalID,
			CreatedAt:  terminal.CreatedAt.Time,
			CreatedBy:  terminal.CreatedBy,
		}
	}

	return result, nil
}

// Delete deletes a terminal
func (r *TerminalRepository) Delete(ctx context.Context, namespace, terminalID string) error {
	err := r.queries.DeleteTerminal(ctx, db.DeleteTerminalParams{
		Namespace:  namespace,
		TerminalID: terminalID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete terminal: %w", err)
	}
	return nil
}

// Exists checks if a terminal exists
func (r *TerminalRepository) Exists(ctx context.Context, namespace, terminalID string) (bool, error) {
	exists, err := r.queries.TerminalExists(ctx, db.TerminalExistsParams{
		Namespace:  namespace,
		TerminalID: terminalID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check terminal existence: %w", err)
	}
	return exists, nil
}

// CountByNamespace counts terminals in a namespace
func (r *TerminalRepository) CountByNamespace(ctx context.Context, namespace string) (int64, error) {
	count, err := r.queries.CountTerminalsByNamespace(ctx, namespace)
	if err != nil {
		return 0, fmt.Errorf("failed to count terminals: %w", err)
	}
	return count, nil
}
