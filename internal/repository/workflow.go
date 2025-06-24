package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// WorkflowRepository implements domain.WorkflowRepository
type WorkflowRepository struct {
	queries *db.Queries
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(queries *db.Queries) *WorkflowRepository {
	return &WorkflowRepository{
		queries: queries,
	}
}

// Create creates a new workflow
func (r *WorkflowRepository) Create(ctx context.Context, workflow *domain.Workflow) error {
	params := db.CreateWorkflowParams{
		Namespace:  workflow.Namespace,
		WorkflowID: workflow.WorkflowID,
		Version:    workflow.Version,
		Status:     &workflow.Status,
		StartAt:    workflow.StartAt,
		Steps:      workflow.Steps,
		CreatedBy:  workflow.CreatedBy,
	}

	return r.queries.CreateWorkflow(ctx, params)
}

// GetByID retrieves a workflow by ID and version
func (r *WorkflowRepository) GetByID(ctx context.Context, namespace, workflowID string, version int32) (*domain.Workflow, error) {
	params := db.GetWorkflowParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
		Version:    version,
	}

	dbWorkflow, err := r.queries.GetWorkflow(ctx, params)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	return convertDBWorkflowToDomain(dbWorkflow), nil
}

// GetActiveVersion retrieves the active version of a workflow
func (r *WorkflowRepository) GetActiveVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	params := db.GetActiveWorkflowVersionParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	}

	dbWorkflow, err := r.queries.GetActiveWorkflowVersion(ctx, params)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	return convertDBWorkflowToDomain(dbWorkflow), nil
}

// GetDraftVersion retrieves the draft version of a workflow
func (r *WorkflowRepository) GetDraftVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	params := db.GetDraftWorkflowVersionParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	}

	dbWorkflow, err := r.queries.GetDraftWorkflowVersion(ctx, params)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	return convertDBWorkflowToDomain(dbWorkflow), nil
}

// List retrieves all workflows in a namespace
func (r *WorkflowRepository) List(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	dbWorkflows, err := r.queries.ListWorkflows(ctx, namespace)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	workflows := make([]*domain.Workflow, len(dbWorkflows))
	for i, dbWorkflow := range dbWorkflows {
		workflows[i] = convertDBWorkflowToDomain(dbWorkflow)
	}

	return workflows, nil
}

// ListActive retrieves all active workflows in a namespace
func (r *WorkflowRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	dbWorkflows, err := r.queries.ListActiveWorkflows(ctx, namespace)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	workflows := make([]*domain.Workflow, len(dbWorkflows))
	for i, dbWorkflow := range dbWorkflows {
		workflows[i] = convertDBWorkflowToDomain(dbWorkflow)
	}

	return workflows, nil
}

// ListVersions retrieves all versions of a workflow
func (r *WorkflowRepository) ListVersions(ctx context.Context, namespace, workflowID string) ([]*domain.Workflow, error) {
	params := db.ListWorkflowVersionsParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	}

	dbWorkflows, err := r.queries.ListWorkflowVersions(ctx, params)
	if err != nil {
		return nil, mapWorkflowError(err)
	}

	workflows := make([]*domain.Workflow, len(dbWorkflows))
	for i, dbWorkflow := range dbWorkflows {
		workflows[i] = convertDBWorkflowToDomain(dbWorkflow)
	}

	return workflows, nil
}

// Update updates a workflow
func (r *WorkflowRepository) Update(ctx context.Context, workflow *domain.Workflow) error {
	params := db.UpdateWorkflowParams{
		Namespace:  workflow.Namespace,
		WorkflowID: workflow.WorkflowID,
		Version:    workflow.Version,
		StartAt:    workflow.StartAt,
		Steps:      workflow.Steps,
		CreatedBy:  workflow.CreatedBy,
	}

	return r.queries.UpdateWorkflow(ctx, params)
}

// Publish publishes a workflow (draft → active)
func (r *WorkflowRepository) Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error {
	// First, deactivate any existing active version
	err := r.queries.DeactivateWorkflow(ctx, db.DeactivateWorkflowParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	})
	if err != nil {
		return mapWorkflowError(err)
	}

	// Then publish the new version
	params := db.PublishWorkflowParams{
		Namespace:   namespace,
		WorkflowID:  workflowID,
		Version:     version,
		PublishedBy: &publishedBy,
	}

	err = r.queries.PublishWorkflow(ctx, params)
	if err != nil {
		return mapWorkflowError(err)
	}

	// Refresh the namespace checksum to trigger cache refresh
	err = r.queries.RefreshNamespaceChecksum(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to refresh namespace checksum: %w", err)
	}

	return nil
}

// Deactivate deactivates a workflow (active → inactive)
func (r *WorkflowRepository) Deactivate(ctx context.Context, namespace, workflowID string) error {
	params := db.DeactivateWorkflowParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	}

	return r.queries.DeactivateWorkflow(ctx, params)
}

// Delete deletes a workflow version
func (r *WorkflowRepository) Delete(ctx context.Context, namespace, workflowID string, version int32) error {
	params := db.DeleteWorkflowParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
		Version:    version,
	}

	return r.queries.DeleteWorkflow(ctx, params)
}

// GetMaxVersion gets the maximum version number for a workflow
func (r *WorkflowRepository) GetMaxVersion(ctx context.Context, namespace, workflowID string) (int32, error) {
	params := db.GetMaxWorkflowVersionParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
	}

	result, err := r.queries.GetMaxWorkflowVersion(ctx, params)
	if err != nil {
		return 0, mapWorkflowError(err)
	}

	// Handle the interface{} return type from sqlc
	switch v := result.(type) {
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for max version: %T", result)
	}
}

// Exists checks if a workflow version exists
func (r *WorkflowRepository) Exists(ctx context.Context, namespace, workflowID string, version int32) (bool, error) {
	params := db.WorkflowExistsParams{
		Namespace:  namespace,
		WorkflowID: workflowID,
		Version:    version,
	}

	return r.queries.WorkflowExists(ctx, params)
}

// convertDBWorkflowToDomain converts a database workflow to domain workflow
func convertDBWorkflowToDomain(dbWorkflow *db.Workflow) *domain.Workflow {
	return &domain.Workflow{
		Namespace:   dbWorkflow.Namespace,
		WorkflowID:  dbWorkflow.WorkflowID,
		Version:     dbWorkflow.Version,
		Status:      derefString(dbWorkflow.Status),
		StartAt:     dbWorkflow.StartAt,
		Steps:       dbWorkflow.Steps,
		CreatedBy:   dbWorkflow.CreatedBy,
		PublishedBy: dbWorkflow.PublishedBy,
		CreatedAt:   dbWorkflow.CreatedAt.Time,
		PublishedAt: derefTime(&dbWorkflow.PublishedAt),
	}
}

// mapWorkflowError maps database errors to domain errors
func mapWorkflowError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific database errors and map them to domain errors
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrWorkflowNotFound
	}

	// Return the original error if no specific mapping is found
	return err
}
