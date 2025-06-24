package service

import (
	"context"
	"strings"
	"time"

	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/repository"
)

var allowedFieldTypes = map[string]struct{}{
	"number": {},
	"string": {},
}

type FieldService struct {
	repo *repository.FieldRepository
}

func NewFieldService(repo *repository.FieldRepository) *FieldService {
	return &FieldService{repo: repo}
}

func (s *FieldService) CreateField(ctx context.Context, namespace string, field *domain.Field) error {
	if err := s.validateField(field); err != nil {
		return err
	}

	// Check if field already exists
	exists, err := s.repo.Exists(ctx, namespace, field.FieldID)
	if err != nil {
		return domain.ErrInternalError
	}
	if exists {
		return domain.ErrFieldAlreadyExists
	}

	// Set namespace and created timestamp
	field.Namespace = namespace
	field.CreatedAt = time.Now()

	err = s.repo.Create(ctx, field)
	if err != nil {
		return err
	}

	return nil
}

func (s *FieldService) GetField(ctx context.Context, namespace, fieldID string) (*domain.Field, error) {
	field, err := s.repo.Get(ctx, namespace, fieldID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrFieldNotFound
		}
		return nil, domain.ErrInternalError
	}

	return field, nil
}

func (s *FieldService) ListFields(ctx context.Context, namespace string) ([]*domain.Field, error) {
	fields, err := s.repo.List(ctx, namespace)
	if err != nil {
		return nil, domain.ErrListError
	}
	return fields, nil
}

func (s *FieldService) UpdateField(ctx context.Context, field *domain.Field) error {
	if err := s.validateField(field); err != nil {
		return err
	}
	exists, err := s.repo.Exists(ctx, field.Namespace, field.FieldID)
	if err != nil {
		return domain.ErrInternalError
	}
	if !exists {
		return domain.ErrFieldNotFound
	}
	return s.repo.Update(ctx, field)
}

func (s *FieldService) DeleteField(ctx context.Context, namespace, fieldID string) error {
	exists, err := s.repo.Exists(ctx, namespace, fieldID)
	if err != nil {
		return domain.ErrInternalError
	}
	if !exists {
		return domain.ErrFieldNotFound
	}
	return s.repo.Delete(ctx, namespace, fieldID)
}

func (s *FieldService) validateField(field *domain.Field) error {
	if field == nil {
		return domain.ErrValidationError
	}

	if strings.TrimSpace(field.FieldID) == "" {
		return domain.ErrInvalidFieldID
	}

	if _, valid := allowedFieldTypes[field.Type]; !valid {
		return domain.ErrInvalidFieldType
	}

	// Description can be nil/empty as per documentation
	return nil
}
