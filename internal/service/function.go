package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rule-engine/internal/domain"
)

// FunctionService handles business logic for functions
type FunctionService struct {
	repo domain.FunctionRepository
}

// NewFunctionService creates a new function service
func NewFunctionService(repo domain.FunctionRepository) *FunctionService {
	return &FunctionService{repo: repo}
}

// CreateFunction creates a new function
func (s *FunctionService) CreateFunction(ctx context.Context, namespace string, function *domain.Function) error {
	if err := s.validateFunction(function); err != nil {
		return err
	}

	// Check if namespace exists (assuming we have access to namespace repo)
	// This would need to be injected or checked differently in a real implementation

	// Check if function already exists (draft or active)
	draftExists, err := s.repo.GetDraftVersion(ctx, namespace, function.FunctionID)
	if err == nil && draftExists != nil {
		return domain.ErrFunctionAlreadyExists
	}

	activeExists, err := s.repo.GetActiveVersion(ctx, namespace, function.FunctionID)
	if err == nil && activeExists != nil {
		return domain.ErrFunctionAlreadyExists
	}

	// Get next version number
	maxVersion, err := s.repo.GetMaxVersion(ctx, namespace, function.FunctionID)
	if err != nil {
		return domain.ErrInternalError
	}

	// Set function properties
	function.Namespace = namespace
	function.Version = maxVersion + 1
	function.Status = domain.StatusDraft
	function.CreatedAt = time.Now()
	function.ReturnType = s.computeReturnType(function.Type)

	// Create the function
	return s.repo.Create(ctx, function)
}

// GetFunction retrieves a function by ID
func (s *FunctionService) GetFunction(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	// Try to get active version first
	function, err := s.repo.GetActiveVersion(ctx, namespace, functionID)
	if err == nil {
		return function, nil
	}

	// If no active version, try draft version
	function, err = s.repo.GetDraftVersion(ctx, namespace, functionID)
	if err != nil {
		return nil, domain.ErrFunctionNotFound
	}

	return function, nil
}

// GetFunctionVersion retrieves a specific version of a function
func (s *FunctionService) GetFunctionVersion(ctx context.Context, namespace, functionID string, version int32) (*domain.Function, error) {
	function, err := s.repo.GetByID(ctx, namespace, functionID, version)
	if err != nil {
		return nil, domain.ErrFunctionNotFound
	}
	return function, nil
}

// ListFunctions lists all functions in a namespace
func (s *FunctionService) ListFunctions(ctx context.Context, namespace string) ([]*domain.Function, error) {
	functions, err := s.repo.List(ctx, namespace)
	if err != nil {
		return nil, domain.ErrInternalError
	}
	return functions, nil
}

// ListActiveFunctions lists all active functions in a namespace
func (s *FunctionService) ListActiveFunctions(ctx context.Context, namespace string) ([]*domain.Function, error) {
	functions, err := s.repo.ListActive(ctx, namespace)
	if err != nil {
		return nil, domain.ErrInternalError
	}
	return functions, nil
}

// ListFunctionVersions lists all versions of a function
func (s *FunctionService) ListFunctionVersions(ctx context.Context, namespace, functionID string) ([]*domain.Function, error) {
	functions, err := s.repo.ListVersions(ctx, namespace, functionID)
	if err != nil {
		return nil, domain.ErrInternalError
	}
	return functions, nil
}

// UpdateFunction updates a function draft
func (s *FunctionService) UpdateFunction(ctx context.Context, namespace, functionID string, function *domain.Function) error {
	// Set the FunctionID from the URL parameter since the handler doesn't set it
	function.FunctionID = functionID

	if err := s.validateFunction(function); err != nil {
		return err
	}

	// Try to get the draft version first
	draft, err := s.repo.GetDraftVersion(ctx, namespace, functionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Draft doesn't exist, check if there's an active version
			active, err := s.repo.GetActiveVersion(ctx, namespace, functionID)
			if err != nil {
				if err == pgx.ErrNoRows {
					return domain.ErrFunctionNotFound
				}
				return err
			}
			if active == nil {
				return domain.ErrFunctionNotFound
			}
			// Create new draft from active version
			function.Namespace = namespace
			function.FunctionID = functionID
			function.Version = active.Version + 1
			function.Status = "draft"
			function.CreatedBy = function.CreatedBy
			return s.repo.Create(ctx, function)
		}
		return err
	}

	// Update existing draft
	function.Namespace = namespace
	function.FunctionID = functionID
	function.Version = draft.Version
	function.Status = "draft"
	function.CreatedBy = function.CreatedBy
	return s.repo.Update(ctx, function)
}

// PublishFunction publishes a function (draft → active)
func (s *FunctionService) PublishFunction(ctx context.Context, namespace, functionID, publishedBy string) error {
	// Get the draft version
	draft, err := s.repo.GetDraftVersion(ctx, namespace, functionID)
	if err != nil {
		return domain.ErrFunctionNotFound
	}

	// Validate dependencies (in a real implementation, you'd check if referenced fields exist)
	if err := s.validateFunctionDependencies(ctx, namespace, draft); err != nil {
		return err
	}

	// Publish the function
	return s.repo.Publish(ctx, namespace, functionID, draft.Version, publishedBy)
}

// DeleteFunction deletes a function
func (s *FunctionService) DeleteFunction(ctx context.Context, namespace, functionID string, version int32) error {
	// Check if function exists
	exists, err := s.repo.Exists(ctx, namespace, functionID, version)
	if err != nil {
		return domain.ErrInternalError
	}
	if !exists {
		return domain.ErrFunctionNotFound
	}

	// Delete the function
	return s.repo.Delete(ctx, namespace, functionID, version)
}

// validateFunction validates function data
func (s *FunctionService) validateFunction(function *domain.Function) error {
	if function.FunctionID == "" {
		return domain.ErrInvalidFunctionID
	}

	if function.Type == "" {
		return domain.ErrInvalidFunctionType
	}

	// Validate function type
	validTypes := []string{domain.FunctionTypeMax, domain.FunctionTypeSum, domain.FunctionTypeAvg, domain.FunctionTypeIn}
	isValidType := false
	for _, t := range validTypes {
		if function.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return domain.ErrInvalidFunctionType
	}

	// Validate based on function type
	switch function.Type {
	case domain.FunctionTypeMax, domain.FunctionTypeSum, domain.FunctionTypeAvg:
		if len(function.Args) == 0 {
			return domain.ErrInvalidFunctionArgs
		}
		if len(function.Values) > 0 {
			return domain.ErrInvalidFunctionArgs
		}
	case domain.FunctionTypeIn:
		if len(function.Values) == 0 {
			return domain.ErrInvalidFunctionArgs
		}
		if len(function.Args) > 0 {
			return domain.ErrInvalidFunctionArgs
		}
	}

	return nil
}

// validateFunctionDependencies validates that function dependencies exist
func (s *FunctionService) validateFunctionDependencies(ctx context.Context, namespace string, function *domain.Function) error {
	// For numeric functions, check if referenced fields exist
	if function.Type == domain.FunctionTypeMax || function.Type == domain.FunctionTypeSum || function.Type == domain.FunctionTypeAvg {
		// In a real implementation, you'd check if the fields in function.Args exist
		// For now, we'll just validate that args are not empty
		if len(function.Args) == 0 {
			return domain.ErrInvalidFunctionArgs
		}
	}

	return nil
}

// computeReturnType determines the return type based on function type
func (s *FunctionService) computeReturnType(functionType string) string {
	switch functionType {
	case domain.FunctionTypeMax, domain.FunctionTypeSum, domain.FunctionTypeAvg:
		return domain.FunctionReturnTypeNumber
	case domain.FunctionTypeIn:
		return domain.FunctionReturnTypeBool
	default:
		return ""
	}
}
