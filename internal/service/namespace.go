package service

import (
	"context"
	"regexp"
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
	if err := s.validateNamespace(namespace); err != nil {
		return err
	}

	// Check if namespace already exists
	existing, err := s.namespaceRepo.GetByID(ctx, namespace.ID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// Namespace doesn't exist, proceed with creation
		} else {
			return domain.ErrInternalError
		}
	} else if existing != nil {
		return domain.ErrNamespaceAlreadyExists
	}

	return s.namespaceRepo.Create(ctx, namespace)
}

// GetNamespace retrieves a namespace by ID
func (s *NamespaceService) GetNamespace(ctx context.Context, id string) (*domain.Namespace, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.ErrInvalidNamespaceID
	}

	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrNamespaceNotFound
		}
		return nil, domain.ErrInternalError
	}

	return namespace, nil
}

// ListNamespaces retrieves all namespaces
func (s *NamespaceService) ListNamespaces(ctx context.Context) ([]*domain.Namespace, error) {
	namespaces, err := s.namespaceRepo.List(ctx)
	if err != nil {
		return nil, domain.ErrListError
	}
	return namespaces, nil
}

// DeleteNamespace deletes a namespace
func (s *NamespaceService) DeleteNamespace(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidNamespaceID
	}

	// Check if namespace exists
	existing, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.ErrNamespaceNotFound
		}
		return domain.ErrInternalError
	}

	if existing == nil {
		return domain.ErrNamespaceNotFound
	}

	return s.namespaceRepo.Delete(ctx, id)
}

// validateNamespace validates namespace input
func (s *NamespaceService) validateNamespace(namespace *domain.Namespace) error {
	if namespace == nil {
		return domain.ErrValidationError
	}

	if strings.TrimSpace(namespace.ID) == "" {
		return domain.ErrInvalidNamespaceID
	}

	if len(namespace.ID) > 50 {
		return domain.ErrInvalidNamespaceID
	}

	// Check if ID contains only alphanumeric characters, hyphens, and underscores
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(namespace.ID) {
		return domain.ErrInvalidNamespaceID
	}

	if strings.TrimSpace(namespace.CreatedBy) == "" {
		return domain.ErrValidationError
	}

	if len(namespace.Description) > 500 {
		return domain.ErrInvalidDescription
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
