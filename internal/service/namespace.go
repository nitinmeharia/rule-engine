package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/rule-engine/internal/domain"
)

// NamespaceServiceInterface defines the interface for namespace service operations
type NamespaceServiceInterface interface {
	CreateNamespace(ctx context.Context, namespace *domain.Namespace) error
	GetNamespace(ctx context.Context, id string) (*domain.Namespace, error)
	ListNamespaces(ctx context.Context) ([]*domain.Namespace, error)
	DeleteNamespace(ctx context.Context, id string) error
}

// NamespaceService handles namespace business logic
type NamespaceService struct {
	namespaceRepo domain.NamespaceRepository
}

// Ensure NamespaceService implements NamespaceServiceInterface
var _ NamespaceServiceInterface = (*NamespaceService)(nil)

// NewNamespaceService creates a new namespace service
func NewNamespaceService(namespaceRepo domain.NamespaceRepository) *NamespaceService {
	return &NamespaceService{
		namespaceRepo: namespaceRepo,
	}
}

// CreateNamespace creates a new namespace with validation
func (s *NamespaceService) CreateNamespace(ctx context.Context, namespace *domain.Namespace) error {
	// Validate namespace
	if err := s.validateNamespace(namespace); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if namespace already exists
	existing, err := s.namespaceRepo.GetByID(ctx, namespace.ID)
	if err != nil && err != domain.ErrNotFound {
		return fmt.Errorf("failed to check existing namespace: %w", err)
	}
	if existing != nil {
		return domain.ErrAlreadyExists
	}

	// Create namespace
	return s.namespaceRepo.Create(ctx, namespace)
}

// GetNamespace retrieves a namespace by ID
func (s *NamespaceService) GetNamespace(ctx context.Context, id string) (*domain.Namespace, error) {
	if id == "" {
		return nil, domain.ErrInvalidInput
	}

	return s.namespaceRepo.GetByID(ctx, id)
}

// ListNamespaces retrieves all namespaces
func (s *NamespaceService) ListNamespaces(ctx context.Context) ([]*domain.Namespace, error) {
	return s.namespaceRepo.List(ctx)
}

// DeleteNamespace deletes a namespace
func (s *NamespaceService) DeleteNamespace(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidInput
	}

	// Check if namespace exists
	existing, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	return s.namespaceRepo.Delete(ctx, id)
}

// validateNamespace validates namespace input
func (s *NamespaceService) validateNamespace(namespace *domain.Namespace) error {
	if namespace == nil {
		return domain.ErrInvalidInput
	}

	if namespace.ID == "" {
		return fmt.Errorf("namespace ID is required: %w", domain.ErrValidation)
	}

	if len(namespace.ID) > 50 {
		return fmt.Errorf("namespace ID too long (max 50 characters): %w", domain.ErrValidation)
	}

	if !isValidNamespaceID(namespace.ID) {
		return fmt.Errorf("namespace ID must contain only alphanumeric characters, hyphens, and underscores: %w", domain.ErrValidation)
	}

	if namespace.CreatedBy == "" {
		return fmt.Errorf("createdBy is required: %w", domain.ErrValidation)
	}

	if len(namespace.Description) > 500 {
		return fmt.Errorf("description too long (max 500 characters): %w", domain.ErrValidation)
	}

	return nil
}

// isValidNamespaceID checks if namespace ID follows naming conventions
func isValidNamespaceID(id string) bool {
	if len(id) == 0 {
		return false
	}

	for _, char := range id {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	// Cannot start or end with hyphen or underscore
	return !strings.HasPrefix(id, "-") && !strings.HasPrefix(id, "_") &&
		!strings.HasSuffix(id, "-") && !strings.HasSuffix(id, "_")
}
