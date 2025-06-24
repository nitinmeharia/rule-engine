package repository

import (
	"context"

	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// FieldRepository defines DB operations for fields
type FieldRepository struct {
	db *db.Queries
}

func NewFieldRepository(q *db.Queries) *FieldRepository {
	return &FieldRepository{db: q}
}

func (r *FieldRepository) Create(ctx context.Context, field *domain.Field) error {
	var description *string
	if field.Description != "" {
		description = &field.Description
	}

	err := r.db.CreateField(ctx, db.CreateFieldParams{
		Namespace:   field.Namespace,
		FieldID:     field.FieldID,
		Type:        &field.Type, // Type is required but nullable in DB
		Description: description,
		CreatedBy:   field.CreatedBy,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *FieldRepository) Get(ctx context.Context, namespace, fieldID string) (*domain.Field, error) {
	f, err := r.db.GetField(ctx, db.GetFieldParams{
		Namespace: namespace,
		FieldID:   fieldID,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Field{
		Namespace:   f.Namespace,
		FieldID:     f.FieldID,
		Type:        derefString(f.Type),
		Description: derefString(f.Description),
		CreatedAt:   f.CreatedAt.Time,
		CreatedBy:   f.CreatedBy,
	}, nil
}

func (r *FieldRepository) List(ctx context.Context, namespace string) ([]*domain.Field, error) {
	rows, err := r.db.ListFields(ctx, namespace)
	if err != nil {
		return nil, err
	}
	var fields []*domain.Field
	for _, f := range rows {
		fields = append(fields, &domain.Field{
			Namespace:   f.Namespace,
			FieldID:     f.FieldID,
			Type:        derefString(f.Type),
			Description: derefString(f.Description),
			CreatedAt:   f.CreatedAt.Time,
			CreatedBy:   f.CreatedBy,
		})
	}
	return fields, nil
}

func (r *FieldRepository) Update(ctx context.Context, field *domain.Field) error {
	return r.db.UpdateField(ctx, db.UpdateFieldParams{
		Namespace:   field.Namespace,
		FieldID:     field.FieldID,
		Type:        &field.Type,
		Description: &field.Description,
		CreatedBy:   field.CreatedBy,
	})
}

func (r *FieldRepository) Delete(ctx context.Context, namespace, fieldID string) error {
	return r.db.DeleteField(ctx, db.DeleteFieldParams{
		Namespace: namespace,
		FieldID:   fieldID,
	})
}

func (r *FieldRepository) Exists(ctx context.Context, namespace, fieldID string) (bool, error) {
	exists, err := r.db.FieldExists(ctx, db.FieldExistsParams{
		Namespace: namespace,
		FieldID:   fieldID,
	})
	return exists, err
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
