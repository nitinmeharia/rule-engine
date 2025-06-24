package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// FunctionRepository implements domain.FunctionRepository
type FunctionRepository struct {
	db *db.Queries
}

func NewFunctionRepository(q *db.Queries) *FunctionRepository {
	return &FunctionRepository{db: q}
}

// Create creates a new function
func (r *FunctionRepository) Create(ctx context.Context, function *domain.Function) error {
	err := r.db.CreateFunction(ctx, db.CreateFunctionParams{
		Namespace:  function.Namespace,
		FunctionID: function.FunctionID,
		Version:    function.Version,
		Status:     &function.Status,
		Type:       &function.Type,
		Args:       function.Args,
		Values:     function.Values,
		CreatedBy:  function.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}
	return nil
}

// GetByID retrieves a function by ID and version
func (r *FunctionRepository) GetByID(ctx context.Context, namespace, functionID string, version int32) (*domain.Function, error) {
	f, err := r.db.GetFunction(ctx, db.GetFunctionParams{
		Namespace:  namespace,
		FunctionID: functionID,
		Version:    version,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Function{
		Namespace:   f.Namespace,
		FunctionID:  f.FunctionID,
		Version:     f.Version,
		Status:      derefString(f.Status),
		Type:        derefString(f.Type),
		Args:        f.Args,
		Values:      f.Values,
		ReturnType:  computeReturnType(derefString(f.Type)),
		CreatedBy:   f.CreatedBy,
		PublishedBy: f.PublishedBy,
		CreatedAt:   f.CreatedAt.Time,
		PublishedAt: derefTime(&f.PublishedAt),
	}, nil
}

// GetActiveVersion retrieves the active version of a function
func (r *FunctionRepository) GetActiveVersion(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	f, err := r.db.GetActiveFunctionVersion(ctx, db.GetActiveFunctionVersionParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Function{
		Namespace:   f.Namespace,
		FunctionID:  f.FunctionID,
		Version:     f.Version,
		Status:      derefString(f.Status),
		Type:        derefString(f.Type),
		Args:        f.Args,
		Values:      f.Values,
		ReturnType:  computeReturnType(derefString(f.Type)),
		CreatedBy:   f.CreatedBy,
		PublishedBy: f.PublishedBy,
		CreatedAt:   f.CreatedAt.Time,
		PublishedAt: derefTime(&f.PublishedAt),
	}, nil
}

// GetDraftVersion retrieves the draft version of a function
func (r *FunctionRepository) GetDraftVersion(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	f, err := r.db.GetDraftFunctionVersion(ctx, db.GetDraftFunctionVersionParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Function{
		Namespace:   f.Namespace,
		FunctionID:  f.FunctionID,
		Version:     f.Version,
		Status:      derefString(f.Status),
		Type:        derefString(f.Type),
		Args:        f.Args,
		Values:      f.Values,
		ReturnType:  computeReturnType(derefString(f.Type)),
		CreatedBy:   f.CreatedBy,
		PublishedBy: f.PublishedBy,
		CreatedAt:   f.CreatedAt.Time,
		PublishedAt: derefTime(&f.PublishedAt),
	}, nil
}

// List retrieves all functions in a namespace
func (r *FunctionRepository) List(ctx context.Context, namespace string) ([]*domain.Function, error) {
	rows, err := r.db.ListFunctions(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	var functions []*domain.Function
	for _, f := range rows {
		functions = append(functions, &domain.Function{
			Namespace:   f.Namespace,
			FunctionID:  f.FunctionID,
			Version:     f.Version,
			Status:      derefString(f.Status),
			Type:        derefString(f.Type),
			Args:        f.Args,
			Values:      f.Values,
			ReturnType:  computeReturnType(derefString(f.Type)),
			CreatedBy:   f.CreatedBy,
			PublishedBy: f.PublishedBy,
			CreatedAt:   f.CreatedAt.Time,
			PublishedAt: derefTime(&f.PublishedAt),
		})
	}

	return functions, nil
}

// ListActive retrieves all active functions in a namespace
func (r *FunctionRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Function, error) {
	rows, err := r.db.ListActiveFunctions(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list active functions: %w", err)
	}

	var functions []*domain.Function
	for _, f := range rows {
		functions = append(functions, &domain.Function{
			Namespace:   f.Namespace,
			FunctionID:  f.FunctionID,
			Version:     f.Version,
			Status:      derefString(f.Status),
			Type:        derefString(f.Type),
			Args:        f.Args,
			Values:      f.Values,
			ReturnType:  computeReturnType(derefString(f.Type)),
			CreatedBy:   f.CreatedBy,
			PublishedBy: f.PublishedBy,
			CreatedAt:   f.CreatedAt.Time,
			PublishedAt: derefTime(&f.PublishedAt),
		})
	}

	return functions, nil
}

// ListVersions retrieves all versions of a function
func (r *FunctionRepository) ListVersions(ctx context.Context, namespace, functionID string) ([]*domain.Function, error) {
	rows, err := r.db.ListFunctionVersions(ctx, db.ListFunctionVersionsParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list function versions: %w", err)
	}

	var functions []*domain.Function
	for _, f := range rows {
		functions = append(functions, &domain.Function{
			Namespace:   f.Namespace,
			FunctionID:  f.FunctionID,
			Version:     f.Version,
			Status:      derefString(f.Status),
			Type:        derefString(f.Type),
			Args:        f.Args,
			Values:      f.Values,
			ReturnType:  computeReturnType(derefString(f.Type)),
			CreatedBy:   f.CreatedBy,
			PublishedBy: f.PublishedBy,
			CreatedAt:   f.CreatedAt.Time,
			PublishedAt: derefTime(&f.PublishedAt),
		})
	}

	return functions, nil
}

// Update updates a function
func (r *FunctionRepository) Update(ctx context.Context, function *domain.Function) error {
	err := r.db.UpdateFunction(ctx, db.UpdateFunctionParams{
		Namespace:  function.Namespace,
		FunctionID: function.FunctionID,
		Version:    function.Version,
		Type:       &function.Type,
		Args:       function.Args,
		Values:     function.Values,
		CreatedBy:  function.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

// Publish publishes a function (draft → active)
func (r *FunctionRepository) Publish(ctx context.Context, namespace, functionID string, version int32, publishedBy string) error {
	// First deactivate any existing active version
	err := r.db.DeactivateFunction(ctx, db.DeactivateFunctionParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate existing function: %w", err)
	}

	// Then publish the new version
	err = r.db.PublishFunction(ctx, db.PublishFunctionParams{
		Namespace:   namespace,
		FunctionID:  functionID,
		Version:     version,
		PublishedBy: &publishedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to publish function: %w", err)
	}
	return nil
}

// Deactivate deactivates a function
func (r *FunctionRepository) Deactivate(ctx context.Context, namespace, functionID string) error {
	err := r.db.DeactivateFunction(ctx, db.DeactivateFunctionParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate function: %w", err)
	}
	return nil
}

// Delete deletes a function
func (r *FunctionRepository) Delete(ctx context.Context, namespace, functionID string, version int32) error {
	err := r.db.DeleteFunction(ctx, db.DeleteFunctionParams{
		Namespace:  namespace,
		FunctionID: functionID,
		Version:    version,
	})
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	return nil
}

// GetMaxVersion gets the maximum version number for a function
func (r *FunctionRepository) GetMaxVersion(ctx context.Context, namespace, functionID string) (int32, error) {
	result, err := r.db.GetMaxFunctionVersion(ctx, db.GetMaxFunctionVersionParams{
		Namespace:  namespace,
		FunctionID: functionID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get max function version: %w", err)
	}

	// The result is interface{}, so we need to convert it
	switch v := result.(type) {
	case int64:
		return int32(v), nil
	case int32:
		return v, nil
	case int:
		return int32(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for max version: %T", result)
	}
}

// Exists checks if a function exists
func (r *FunctionRepository) Exists(ctx context.Context, namespace, functionID string, version int32) (bool, error) {
	exists, err := r.db.FunctionExists(ctx, db.FunctionExistsParams{
		Namespace:  namespace,
		FunctionID: functionID,
		Version:    version,
	})
	return exists, err
}

// Helper functions
func derefTime(t *pgtype.Timestamptz) *time.Time {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.Time
}

// computeReturnType determines the return type based on function type
func computeReturnType(functionType string) string {
	switch functionType {
	case domain.FunctionTypeMax, domain.FunctionTypeSum, domain.FunctionTypeAvg:
		return domain.FunctionReturnTypeNumber
	case domain.FunctionTypeIn:
		return domain.FunctionReturnTypeBool
	default:
		return ""
	}
}
