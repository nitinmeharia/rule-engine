package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rule-engine/internal/domain"
)

// WorkflowServiceInterface defines the interface for workflow operations
type WorkflowServiceInterface interface {
	Create(ctx context.Context, workflow *domain.Workflow) error
	GetByID(ctx context.Context, namespace, workflowID string, version int32) (*domain.Workflow, error)
	GetActiveVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error)
	GetDraftVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error)
	List(ctx context.Context, namespace string) ([]*domain.Workflow, error)
	ListActive(ctx context.Context, namespace string) ([]*domain.Workflow, error)
	ListVersions(ctx context.Context, namespace, workflowID string) ([]*domain.Workflow, error)
	Update(ctx context.Context, workflow *domain.Workflow) error
	Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error
	Deactivate(ctx context.Context, namespace, workflowID string) error
	Delete(ctx context.Context, namespace, workflowID string, version int32) error
}

// WorkflowService implements workflow business logic
type WorkflowService struct {
	workflowRepo domain.WorkflowRepository
	ruleRepo     domain.RuleRepository
	terminalRepo domain.TerminalRepository
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(
	workflowRepo domain.WorkflowRepository,
	ruleRepo domain.RuleRepository,
	terminalRepo domain.TerminalRepository,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		ruleRepo:     ruleRepo,
		terminalRepo: terminalRepo,
	}
}

// Create creates a new workflow
func (s *WorkflowService) Create(ctx context.Context, workflow *domain.Workflow) error {
	// Validate the workflow
	if err := workflow.Validate(); err != nil {
		// Wrap validation errors in domain.APIError to ensure proper HTTP 400 responses
		if _, ok := err.(*domain.APIError); ok {
			return err // already wrapped
		}
		return domain.NewAPIError(domain.ErrCodeValidationError, err.Error())
	}

	// Check if workflow already exists (draft or active)
	draftExists, err := s.workflowRepo.GetDraftVersion(ctx, workflow.Namespace, workflow.WorkflowID)
	if err == nil && draftExists != nil {
		return domain.ErrWorkflowAlreadyExists
	}

	activeExists, err := s.workflowRepo.GetActiveVersion(ctx, workflow.Namespace, workflow.WorkflowID)
	if err == nil && activeExists != nil {
		return domain.ErrWorkflowAlreadyExists
	}

	// Get next version number
	maxVersion, err := s.workflowRepo.GetMaxVersion(ctx, workflow.Namespace, workflow.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get max version: %w", err)
	}

	// Set workflow properties
	workflow.Version = maxVersion + 1
	workflow.Status = domain.StatusDraft

	err = s.workflowRepo.Create(ctx, workflow)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a workflow by ID and version
func (s *WorkflowService) GetByID(ctx context.Context, namespace, workflowID string, version int32) (*domain.Workflow, error) {
	return s.workflowRepo.GetByID(ctx, namespace, workflowID, version)
}

// GetActiveVersion retrieves the active version of a workflow
func (s *WorkflowService) GetActiveVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	return s.workflowRepo.GetActiveVersion(ctx, namespace, workflowID)
}

// GetDraftVersion retrieves the draft version of a workflow
func (s *WorkflowService) GetDraftVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	return s.workflowRepo.GetDraftVersion(ctx, namespace, workflowID)
}

// List retrieves all workflows in a namespace
func (s *WorkflowService) List(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	return s.workflowRepo.List(ctx, namespace)
}

// ListActive retrieves all active workflows in a namespace
func (s *WorkflowService) ListActive(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	return s.workflowRepo.ListActive(ctx, namespace)
}

// ListVersions retrieves all versions of a workflow
func (s *WorkflowService) ListVersions(ctx context.Context, namespace, workflowID string) ([]*domain.Workflow, error) {
	return s.workflowRepo.ListVersions(ctx, namespace, workflowID)
}

// Update updates a workflow
func (s *WorkflowService) Update(ctx context.Context, workflow *domain.Workflow) error {
	// Validate the workflow
	if err := workflow.Validate(); err != nil {
		// Wrap validation errors in domain.APIError to ensure proper HTTP 400 responses
		if _, ok := err.(*domain.APIError); ok {
			return err // already wrapped
		}
		return domain.NewAPIError(domain.ErrCodeValidationError, err.Error())
	}

	// Check if workflow exists
	exists, err := s.workflowRepo.Exists(ctx, workflow.Namespace, workflow.WorkflowID, workflow.Version)
	if err != nil {
		return fmt.Errorf("failed to check workflow existence: %w", err)
	}
	if !exists {
		return domain.ErrWorkflowNotFound
	}

	// Ensure workflow is in draft status for updates
	currentWorkflow, err := s.workflowRepo.GetByID(ctx, workflow.Namespace, workflow.WorkflowID, workflow.Version)
	if err != nil {
		return err
	}
	if currentWorkflow.Status != domain.StatusDraft {
		return domain.ErrWorkflowNotDraft
	}

	return s.workflowRepo.Update(ctx, workflow)
}

// Publish publishes a workflow (draft → active)
func (s *WorkflowService) Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error {
	// Get the workflow to validate
	workflow, err := s.workflowRepo.GetByID(ctx, namespace, workflowID, version)
	if err != nil {
		return err
	}

	// Ensure workflow is in draft status
	if workflow.Status != domain.StatusDraft {
		return domain.ErrWorkflowNotDraft
	}

	// Validate dependencies before publishing
	if err := s.validateDependencies(ctx, workflow); err != nil {
		return err
	}

	return s.workflowRepo.Publish(ctx, namespace, workflowID, version, publishedBy)
}

// Deactivate deactivates a workflow (active → inactive)
func (s *WorkflowService) Deactivate(ctx context.Context, namespace, workflowID string) error {
	// Get the active workflow to validate
	workflow, err := s.workflowRepo.GetActiveVersion(ctx, namespace, workflowID)
	if err != nil {
		return err
	}

	// Ensure workflow is active
	if workflow.Status != domain.StatusActive {
		return domain.ErrWorkflowNotActive
	}

	return s.workflowRepo.Deactivate(ctx, namespace, workflowID)
}

// Delete deletes a workflow version
func (s *WorkflowService) Delete(ctx context.Context, namespace, workflowID string, version int32) error {
	// Get the workflow to validate
	workflow, err := s.workflowRepo.GetByID(ctx, namespace, workflowID, version)
	if err != nil {
		return err
	}

	// Prevent deletion of active workflows
	if workflow.Status == domain.StatusActive {
		return domain.ErrWorkflowActive
	}

	return s.workflowRepo.Delete(ctx, namespace, workflowID, version)
}

// validateDependencies validates that all dependencies (rules and terminals) exist and are active
func (s *WorkflowService) validateDependencies(ctx context.Context, workflow *domain.Workflow) error {
	// Parse workflow steps from JSON
	var steps map[string]interface{}
	if err := json.Unmarshal(workflow.Steps, &steps); err != nil {
		// Wrap JSON unmarshaling errors in domain.APIError to ensure proper HTTP 400 responses
		return domain.NewAPIError(domain.ErrCodeValidationError, fmt.Sprintf("invalid workflow steps format: %v", err))
	}

	// Extract all rule and terminal references from workflow steps
	ruleRefs := make(map[string]bool)
	terminalRefs := make(map[string]bool)

	for _, stepData := range steps {
		stepMap, ok := stepData.(map[string]interface{})
		if !ok {
			continue
		}

		stepType, _ := stepMap["type"].(string)

		// Check rule references
		if stepType == "rule" {
			if ruleID, exists := stepMap["ruleId"].(string); exists && ruleID != "" {
				ruleRefs[ruleID] = true
			}
		}

		// Check terminal references
		if stepType == "terminal" {
			if terminalID, exists := stepMap["terminalId"].(string); exists && terminalID != "" {
				terminalRefs[terminalID] = true
			}
		}
	}

	// Validate rule dependencies
	for ruleID := range ruleRefs {
		rule, err := s.ruleRepo.GetActiveVersion(ctx, workflow.Namespace, ruleID)
		if err != nil {
			if err == domain.ErrRuleNotFound {
				return domain.NewAPIError(domain.ErrCodeDependencyInactive, fmt.Sprintf("rule '%s' not found", ruleID))
			}
			return fmt.Errorf("failed to validate rule '%s': %w", ruleID, err)
		}

		if rule.Status != domain.StatusActive {
			return domain.NewAPIError(domain.ErrCodeDependencyInactive, fmt.Sprintf("rule '%s' is not active (status: %s)", ruleID, rule.Status))
		}
	}

	// Validate terminal dependencies
	for terminalID := range terminalRefs {
		_, err := s.terminalRepo.GetByID(ctx, workflow.Namespace, terminalID)
		if err != nil {
			if err == domain.ErrTerminalNotFound {
				return domain.NewAPIError(domain.ErrCodeDependencyInactive, fmt.Sprintf("terminal '%s' not found", terminalID))
			}
			return fmt.Errorf("failed to validate terminal '%s': %w", terminalID, err)
		}
	}

	return nil
}
