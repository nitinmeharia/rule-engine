package repository

import (
	"context"
	"fmt"

	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
)

// RuleRepository implements domain.RuleRepository
type RuleRepository struct {
	db *db.Queries
}

func NewRuleRepository(q *db.Queries) *RuleRepository {
	return &RuleRepository{db: q}
}

// Create creates a new rule
func (r *RuleRepository) Create(ctx context.Context, rule *domain.Rule) error {
	err := r.db.CreateRule(ctx, db.CreateRuleParams{
		Namespace:  rule.Namespace,
		RuleID:     rule.RuleID,
		Version:    rule.Version,
		Status:     &rule.Status,
		Logic:      &rule.Logic,
		Conditions: rule.Conditions,
		CreatedBy:  rule.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}
	return nil
}

// GetByID retrieves a rule by ID and version
func (r *RuleRepository) GetByID(ctx context.Context, namespace, ruleID string, version int32) (*domain.Rule, error) {
	rule, err := r.db.GetRule(ctx, db.GetRuleParams{
		Namespace: namespace,
		RuleID:    ruleID,
		Version:   version,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Rule{
		Namespace:   rule.Namespace,
		RuleID:      rule.RuleID,
		Version:     rule.Version,
		Status:      derefString(rule.Status),
		Logic:       derefString(rule.Logic),
		Conditions:  rule.Conditions,
		CreatedBy:   rule.CreatedBy,
		PublishedBy: rule.PublishedBy,
		CreatedAt:   rule.CreatedAt.Time,
		PublishedAt: derefTime(&rule.PublishedAt),
	}, nil
}

// GetActiveVersion retrieves the active version of a rule
func (r *RuleRepository) GetActiveVersion(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	rule, err := r.db.GetActiveRuleVersion(ctx, db.GetActiveRuleVersionParams{
		Namespace: namespace,
		RuleID:    ruleID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Rule{
		Namespace:   rule.Namespace,
		RuleID:      rule.RuleID,
		Version:     rule.Version,
		Status:      derefString(rule.Status),
		Logic:       derefString(rule.Logic),
		Conditions:  rule.Conditions,
		CreatedBy:   rule.CreatedBy,
		PublishedBy: rule.PublishedBy,
		CreatedAt:   rule.CreatedAt.Time,
		PublishedAt: derefTime(&rule.PublishedAt),
	}, nil
}

// GetDraftVersion retrieves the draft version of a rule
func (r *RuleRepository) GetDraftVersion(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	rule, err := r.db.GetDraftRuleVersion(ctx, db.GetDraftRuleVersionParams{
		Namespace: namespace,
		RuleID:    ruleID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Rule{
		Namespace:   rule.Namespace,
		RuleID:      rule.RuleID,
		Version:     rule.Version,
		Status:      derefString(rule.Status),
		Logic:       derefString(rule.Logic),
		Conditions:  rule.Conditions,
		CreatedBy:   rule.CreatedBy,
		PublishedBy: rule.PublishedBy,
		CreatedAt:   rule.CreatedAt.Time,
		PublishedAt: derefTime(&rule.PublishedAt),
	}, nil
}

// List retrieves all rules in a namespace
func (r *RuleRepository) List(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	rules, err := r.db.ListRules(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	result := make([]*domain.Rule, len(rules))
	for i, rule := range rules {
		result[i] = &domain.Rule{
			Namespace:   rule.Namespace,
			RuleID:      rule.RuleID,
			Version:     rule.Version,
			Status:      derefString(rule.Status),
			Logic:       derefString(rule.Logic),
			Conditions:  rule.Conditions,
			CreatedBy:   rule.CreatedBy,
			PublishedBy: rule.PublishedBy,
			CreatedAt:   rule.CreatedAt.Time,
			PublishedAt: derefTime(&rule.PublishedAt),
		}
	}

	return result, nil
}

// ListActive retrieves all active rules in a namespace
func (r *RuleRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	rules, err := r.db.ListActiveRules(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list active rules: %w", err)
	}

	result := make([]*domain.Rule, len(rules))
	for i, rule := range rules {
		result[i] = &domain.Rule{
			Namespace:   rule.Namespace,
			RuleID:      rule.RuleID,
			Version:     rule.Version,
			Status:      derefString(rule.Status),
			Logic:       derefString(rule.Logic),
			Conditions:  rule.Conditions,
			CreatedBy:   rule.CreatedBy,
			PublishedBy: rule.PublishedBy,
			CreatedAt:   rule.CreatedAt.Time,
			PublishedAt: derefTime(&rule.PublishedAt),
		}
	}

	return result, nil
}

// ListVersions retrieves all versions of a specific rule
func (r *RuleRepository) ListVersions(ctx context.Context, namespace, ruleID string) ([]*domain.Rule, error) {
	rules, err := r.db.ListRuleVersions(ctx, db.ListRuleVersionsParams{
		Namespace: namespace,
		RuleID:    ruleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list rule versions: %w", err)
	}

	result := make([]*domain.Rule, len(rules))
	for i, rule := range rules {
		result[i] = &domain.Rule{
			Namespace:   rule.Namespace,
			RuleID:      rule.RuleID,
			Version:     rule.Version,
			Status:      derefString(rule.Status),
			Logic:       derefString(rule.Logic),
			Conditions:  rule.Conditions,
			CreatedBy:   rule.CreatedBy,
			PublishedBy: rule.PublishedBy,
			CreatedAt:   rule.CreatedAt.Time,
			PublishedAt: derefTime(&rule.PublishedAt),
		}
	}

	return result, nil
}

// Update updates a rule
func (r *RuleRepository) Update(ctx context.Context, rule *domain.Rule) error {
	err := r.db.UpdateRule(ctx, db.UpdateRuleParams{
		Namespace:  rule.Namespace,
		RuleID:     rule.RuleID,
		Version:    rule.Version,
		Logic:      &rule.Logic,
		Conditions: rule.Conditions,
		CreatedBy:  rule.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	return nil
}

// Publish publishes a rule (makes it active)
func (r *RuleRepository) Publish(ctx context.Context, namespace, ruleID string, version int32, publishedBy string) error {
	// First deactivate any existing active version
	err := r.Deactivate(ctx, namespace, ruleID)
	if err != nil {
		return fmt.Errorf("failed to deactivate existing rule: %w", err)
	}

	// Then publish the new version
	err = r.db.PublishRule(ctx, db.PublishRuleParams{
		Namespace:   namespace,
		RuleID:      ruleID,
		Version:     version,
		PublishedBy: &publishedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to publish rule: %w", err)
	}

	// Refresh the namespace checksum to trigger cache refresh
	err = r.db.RefreshNamespaceChecksum(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to refresh namespace checksum: %w", err)
	}

	return nil
}

// Deactivate deactivates the active version of a rule
func (r *RuleRepository) Deactivate(ctx context.Context, namespace, ruleID string) error {
	err := r.db.DeactivateRule(ctx, db.DeactivateRuleParams{
		Namespace: namespace,
		RuleID:    ruleID,
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate rule: %w", err)
	}
	return nil
}

// Delete deletes a specific version of a rule
func (r *RuleRepository) Delete(ctx context.Context, namespace, ruleID string, version int32) error {
	err := r.db.DeleteRule(ctx, db.DeleteRuleParams{
		Namespace: namespace,
		RuleID:    ruleID,
		Version:   version,
	})
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}

// GetMaxVersion gets the maximum version number for a rule
func (r *RuleRepository) GetMaxVersion(ctx context.Context, namespace, ruleID string) (int32, error) {
	result, err := r.db.GetMaxRuleVersion(ctx, db.GetMaxRuleVersionParams{
		Namespace: namespace,
		RuleID:    ruleID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get max rule version: %w", err)
	}

	// Handle the interface{} return type from sqlc
	switch v := result.(type) {
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case float64:
		return int32(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for max version: %T", result)
	}
}

// Exists checks if a rule exists
func (r *RuleRepository) Exists(ctx context.Context, namespace, ruleID string, version int32) (bool, error) {
	exists, err := r.db.RuleExists(ctx, db.RuleExistsParams{
		Namespace: namespace,
		RuleID:    ruleID,
		Version:   version,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check rule existence: %w", err)
	}
	return exists, nil
}
