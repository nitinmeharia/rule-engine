package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// NamespaceRepository implements domain.NamespaceRepository
type NamespaceRepository struct {
	queries db.Querier
}

// NewNamespaceRepository creates a new namespace repository
func NewNamespaceRepository(queries db.Querier) domain.NamespaceRepository {
	return &NamespaceRepository{
		queries: queries,
	}
}

// Create creates a new namespace
func (r *NamespaceRepository) Create(ctx context.Context, namespace *domain.Namespace) error {
	description := &namespace.Description
	if namespace.Description == "" {
		description = nil
	}

	err := r.queries.CreateNamespace(ctx, db.CreateNamespaceParams{
		ID:          namespace.ID,
		Description: description,
		CreatedBy:   namespace.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	return nil
}

// GetByID retrieves a namespace by ID
func (r *NamespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	ns, err := r.queries.GetNamespace(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	description := ""
	if ns.Description != nil {
		description = *ns.Description
	}

	return &domain.Namespace{
		ID:          ns.ID,
		Description: description,
		CreatedAt:   ns.CreatedAt.Time,
		CreatedBy:   ns.CreatedBy,
	}, nil
}

// List retrieves all namespaces
func (r *NamespaceRepository) List(ctx context.Context) ([]*domain.Namespace, error) {
	namespaces, err := r.queries.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	result := make([]*domain.Namespace, len(namespaces))
	for i, ns := range namespaces {
		description := ""
		if ns.Description != nil {
			description = *ns.Description
		}

		result[i] = &domain.Namespace{
			ID:          ns.ID,
			Description: description,
			CreatedAt:   ns.CreatedAt.Time,
			CreatedBy:   ns.CreatedBy,
		}
	}

	return result, nil
}

// Delete deletes a namespace
func (r *NamespaceRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteNamespace(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	return nil
}
