package service

import (
	"context"
	"strings"

	"github.com/rule-engine/internal/domain"
)

// TerminalServiceInterface defines the interface for terminal service operations
type TerminalServiceInterface interface {
	CreateTerminal(ctx context.Context, namespace string, terminal *domain.Terminal) error
	GetTerminal(ctx context.Context, namespace, terminalID string) (*domain.Terminal, error)
	ListTerminals(ctx context.Context, namespace string) ([]*domain.Terminal, error)
	DeleteTerminal(ctx context.Context, namespace, terminalID string) error
}

// TerminalService handles terminal business logic
type TerminalService struct {
	terminalRepo  domain.TerminalRepository
	namespaceRepo domain.NamespaceRepository
}

// Ensure TerminalService implements TerminalServiceInterface
var _ TerminalServiceInterface = (*TerminalService)(nil)

// NewTerminalService creates a new terminal service
func NewTerminalService(terminalRepo domain.TerminalRepository, namespaceRepo domain.NamespaceRepository) *TerminalService {
	return &TerminalService{
		terminalRepo:  terminalRepo,
		namespaceRepo: namespaceRepo,
	}
}

// CreateTerminal creates a new terminal with validation
func (s *TerminalService) CreateTerminal(ctx context.Context, namespace string, terminal *domain.Terminal) error {
	if err := terminal.Validate(); err != nil {
		return err
	}

	// Set the namespace from the path parameter
	terminal.Namespace = namespace

	// Check if namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") || err == domain.ErrNamespaceNotFound {
			return domain.ErrNamespaceNotFound
		}
		return domain.ErrInternalError
	}

	// Check if terminal already exists
	exists, err := s.terminalRepo.Exists(ctx, namespace, terminal.TerminalID)
	if err != nil {
		return domain.ErrInternalError
	}
	if exists {
		return domain.ErrTerminalAlreadyExists
	}

	return s.terminalRepo.Create(ctx, terminal)
}

// GetTerminal retrieves a terminal by ID
func (s *TerminalService) GetTerminal(ctx context.Context, namespace, terminalID string) (*domain.Terminal, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, domain.ErrInvalidNamespaceID
	}

	if strings.TrimSpace(terminalID) == "" {
		return nil, domain.ErrInvalidTerminalID
	}

	// Check if namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") || err == domain.ErrNamespaceNotFound {
			return nil, domain.ErrNamespaceNotFound
		}
		return nil, domain.ErrInternalError
	}

	terminal, err := s.terminalRepo.GetByID(ctx, namespace, terminalID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrTerminalNotFound
		}
		return nil, domain.ErrInternalError
	}

	return terminal, nil
}

// ListTerminals retrieves all terminals for a namespace
func (s *TerminalService) ListTerminals(ctx context.Context, namespace string) ([]*domain.Terminal, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, domain.ErrInvalidNamespaceID
	}

	// Check if namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") || err == domain.ErrNamespaceNotFound {
			return nil, domain.ErrNamespaceNotFound
		}
		return nil, domain.ErrInternalError
	}

	terminals, err := s.terminalRepo.List(ctx, namespace)
	if err != nil {
		return nil, domain.ErrListError
	}
	return terminals, nil
}

// DeleteTerminal deletes a terminal
func (s *TerminalService) DeleteTerminal(ctx context.Context, namespace, terminalID string) error {
	if strings.TrimSpace(namespace) == "" {
		return domain.ErrInvalidNamespaceID
	}

	if strings.TrimSpace(terminalID) == "" {
		return domain.ErrInvalidTerminalID
	}

	// Check if namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") || err == domain.ErrNamespaceNotFound {
			return domain.ErrNamespaceNotFound
		}
		return domain.ErrInternalError
	}

	// Check if terminal exists
	exists, err := s.terminalRepo.Exists(ctx, namespace, terminalID)
	if err != nil {
		return domain.ErrInternalError
	}
	if !exists {
		return domain.ErrTerminalNotFound
	}

	return s.terminalRepo.Delete(ctx, namespace, terminalID)
}
