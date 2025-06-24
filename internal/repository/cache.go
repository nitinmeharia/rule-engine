package repository

import (
	"context"

	"github.com/rule-engine/internal/domain"
	db "github.com/rule-engine/internal/models/db"
)

type CacheRepository struct {
	db *db.Queries
}

func NewCacheRepository(q *db.Queries) *CacheRepository {
	return &CacheRepository{db: q}
}

func (r *CacheRepository) GetActiveConfigChecksum(ctx context.Context, namespace string) (*domain.ActiveConfigMeta, error) {
	meta, err := r.db.GetActiveConfigChecksum(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return mapActiveConfigMeta(meta), nil
}

func (r *CacheRepository) UpsertActiveConfigChecksum(ctx context.Context, namespace, checksum string) error {
	return r.db.UpsertActiveConfigChecksum(ctx, db.UpsertActiveConfigChecksumParams{
		Namespace: namespace,
		Checksum:  checksum,
	})
}

func (r *CacheRepository) RefreshNamespaceChecksum(ctx context.Context, namespace string) error {
	return r.db.RefreshNamespaceChecksum(ctx, namespace)
}

func (r *CacheRepository) ListAllActiveConfigChecksums(ctx context.Context) ([]*domain.ActiveConfigMeta, error) {
	items, err := r.db.ListAllActiveConfigChecksums(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.ActiveConfigMeta, 0, len(items))
	for _, item := range items {
		result = append(result, mapActiveConfigMeta(item))
	}
	return result, nil
}

func (r *CacheRepository) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return r.db.DeleteActiveConfigChecksum(ctx, namespace)
}

func mapActiveConfigMeta(dbMeta *db.ActiveConfigMetum) *domain.ActiveConfigMeta {
	if dbMeta == nil {
		return nil
	}
	return &domain.ActiveConfigMeta{
		Namespace: dbMeta.Namespace,
		Checksum:  dbMeta.Checksum,
		UpdatedAt: dbMeta.UpdatedAt.Time,
	}
}
