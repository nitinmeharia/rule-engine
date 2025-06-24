package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/rule-engine/internal/domain"
)

// RuleServiceInterface defines the interface for rule service operations
type RuleServiceInterface interface {
	CreateRule(ctx context.Context, namespace string, rule *domain.Rule) error
	GetRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error)
	GetDraftRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error)
	ListRules(ctx context.Context, namespace string) ([]*domain.Rule, error)
	ListRuleVersions(ctx context.Context, namespace, ruleID string) ([]*domain.Rule, error)
	UpdateRule(ctx context.Context, namespace, ruleID string, rule *domain.Rule) error
	PublishRule(ctx context.Context, namespace, ruleID, publishedBy string) error
	DeleteRule(ctx context.Context, namespace, ruleID string, version int32) error
}

// RuleService handles business logic for rules
type RuleService struct {
	repo          domain.RuleRepository
	functionRepo  domain.FunctionRepository
	fieldRepo     domain.FieldRepository
	namespaceRepo domain.NamespaceRepository
}

// Ensure RuleService implements RuleServiceInterface
var _ RuleServiceInterface = (*RuleService)(nil)

// NewRuleService creates a new rule service
func NewRuleService(repo domain.RuleRepository, functionRepo domain.FunctionRepository, fieldRepo domain.FieldRepository, namespaceRepo domain.NamespaceRepository) *RuleService {
	return &RuleService{
		repo:          repo,
		functionRepo:  functionRepo,
		fieldRepo:     fieldRepo,
		namespaceRepo: namespaceRepo,
	}
}

// CreateRule creates a new rule
func (s *RuleService) CreateRule(ctx context.Context, namespace string, rule *domain.Rule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	// Check if rule already exists (draft or active)
	draftExists, err := s.repo.GetDraftVersion(ctx, namespace, rule.RuleID)
	if err == nil && draftExists != nil {
		return domain.ErrRuleAlreadyExists
	}

	activeExists, err := s.repo.GetActiveVersion(ctx, namespace, rule.RuleID)
	if err == nil && activeExists != nil {
		return domain.ErrRuleAlreadyExists
	}

	// Get next version number
	maxVersion, err := s.repo.GetMaxVersion(ctx, namespace, rule.RuleID)
	if err != nil {
		return domain.ErrInternalError
	}

	// Set rule properties
	rule.Namespace = namespace
	rule.Version = maxVersion + 1
	rule.Status = domain.StatusDraft
	rule.CreatedAt = time.Now()

	// Validate conditions structure
	if err := s.validateConditions(rule.Conditions); err != nil {
		return err
	}

	// Create the rule
	return s.repo.Create(ctx, rule)
}

// GetRule retrieves a rule by ID (returns active version by default)
func (s *RuleService) GetRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	rule, err := s.repo.GetActiveVersion(ctx, namespace, ruleID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrRuleNotFound
		}
		return nil, domain.ErrInternalError
	}

	return rule, nil
}

// GetRuleVersion retrieves a specific version of a rule
func (s *RuleService) GetRuleVersion(ctx context.Context, namespace, ruleID string, version int32) (*domain.Rule, error) {
	rule, err := s.repo.GetByID(ctx, namespace, ruleID, version)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrRuleNotFound
		}
		return nil, domain.ErrInternalError
	}

	return rule, nil
}

// GetDraftRule retrieves the draft version of a rule
func (s *RuleService) GetDraftRule(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	rule, err := s.repo.GetDraftVersion(ctx, namespace, ruleID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, domain.ErrRuleNotFound
		}
		return nil, domain.ErrInternalError
	}

	return rule, nil
}

// ListRules retrieves all rules in a namespace
func (s *RuleService) ListRules(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	rules, err := s.repo.List(ctx, namespace)
	if err != nil {
		return nil, domain.ErrListError
	}
	return rules, nil
}

// ListActiveRules retrieves all active rules in a namespace
func (s *RuleService) ListActiveRules(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	rules, err := s.repo.ListActive(ctx, namespace)
	if err != nil {
		return nil, domain.ErrListError
	}
	return rules, nil
}

// ListRuleVersions retrieves all versions of a specific rule
func (s *RuleService) ListRuleVersions(ctx context.Context, namespace, ruleID string) ([]*domain.Rule, error) {
	rules, err := s.repo.ListVersions(ctx, namespace, ruleID)
	if err != nil {
		return nil, domain.ErrListError
	}
	return rules, nil
}

// UpdateRule updates a draft rule
func (s *RuleService) UpdateRule(ctx context.Context, namespace, ruleID string, rule *domain.Rule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	// Get the draft version
	draftRule, err := s.repo.GetDraftVersion(ctx, namespace, ruleID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.ErrRuleNotFound
		}
		return domain.ErrInternalError
	}

	// Update the draft rule
	draftRule.Logic = rule.Logic
	draftRule.Conditions = rule.Conditions
	draftRule.CreatedBy = rule.CreatedBy

	// Validate conditions structure
	if err := s.validateConditions(draftRule.Conditions); err != nil {
		return err
	}

	return s.repo.Update(ctx, draftRule)
}

// PublishRule publishes a draft rule (makes it active)
func (s *RuleService) PublishRule(ctx context.Context, namespace, ruleID, publishedBy string) error {
	// Defensive: Check namespace existence
	ns, err := s.namespaceRepo.GetByID(ctx, namespace)
	if err != nil || ns == nil {
		return domain.ErrNamespaceNotFound
	}

	// Get the draft version
	draftRule, err := s.repo.GetDraftVersion(ctx, namespace, ruleID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.ErrRuleNotFound
		}
		return domain.ErrInternalError
	}

	// Validate conditions structure
	if err := s.validateConditions(draftRule.Conditions); err != nil {
		return err
	}

	// CRITICAL: Validate dependencies before publishing
	if err := s.validateDependencies(ctx, namespace, draftRule.Conditions); err != nil {
		return err
	}

	// Publish the rule
	return s.repo.Publish(ctx, namespace, ruleID, draftRule.Version, publishedBy)
}

// DeleteRule deletes a specific version of a rule
func (s *RuleService) DeleteRule(ctx context.Context, namespace, ruleID string, version int32) error {
	// Check if rule exists
	exists, err := s.repo.Exists(ctx, namespace, ruleID, version)
	if err != nil {
		return domain.ErrInternalError
	}
	if !exists {
		return domain.ErrRuleNotFound
	}

	return s.repo.Delete(ctx, namespace, ruleID, version)
}

// validateConditions validates the conditions structure
func (s *RuleService) validateConditions(conditions json.RawMessage) error {
	if conditions == nil {
		return domain.ErrInvalidRuleConditions
	}

	// Parse conditions to validate structure
	var conditionsArray []map[string]interface{}
	if err := json.Unmarshal(conditions, &conditionsArray); err != nil {
		return domain.ErrInvalidRuleConditions
	}

	if len(conditionsArray) == 0 {
		return domain.ErrInvalidRuleConditions
	}

	// Validate each condition
	for _, condition := range conditionsArray {
		if err := s.validateCondition(condition); err != nil {
			return domain.ErrInvalidRuleConditions
		}
	}

	return nil
}

// validateCondition validates a single condition
func (s *RuleService) validateCondition(condition map[string]interface{}) error {
	// Check required fields
	conditionType, ok := condition["type"].(string)
	if !ok {
		return domain.ErrInvalidRuleConditions
	}

	switch conditionType {
	case "field":
		return s.validateFieldCondition(condition)
	case "function":
		return s.validateFunctionCondition(condition)
	case "rule":
		return s.validateRuleCondition(condition)
	default:
		return domain.ErrInvalidRuleConditions
	}
}

// validateFieldCondition validates a field condition
func (s *RuleService) validateFieldCondition(condition map[string]interface{}) error {
	// Required fields: fieldId, operator, value
	if _, ok := condition["fieldId"].(string); !ok {
		return domain.ErrInvalidRuleConditions
	}
	if _, ok := condition["operator"].(string); !ok {
		return domain.ErrInvalidRuleConditions
	}
	if _, ok := condition["value"]; !ok {
		return domain.ErrInvalidRuleConditions
	}

	// Validate operator
	operator := condition["operator"].(string)
	validOperators := []string{"==", "!=", ">", "<", ">=", "<="}
	isValid := false
	for _, op := range validOperators {
		if operator == op {
			isValid = true
			break
		}
	}
	if !isValid {
		return domain.ErrInvalidRuleConditions
	}

	return nil
}

// validateFunctionCondition validates a function condition
func (s *RuleService) validateFunctionCondition(condition map[string]interface{}) error {
	// Required fields: functionId, operator, value
	if _, ok := condition["functionId"].(string); !ok {
		return domain.ErrInvalidRuleConditions
	}
	if _, ok := condition["operator"].(string); !ok {
		return domain.ErrInvalidRuleConditions
	}
	if _, ok := condition["value"]; !ok {
		return domain.ErrInvalidRuleConditions
	}

	// Validate operator
	operator := condition["operator"].(string)
	validOperators := []string{"==", "!=", ">", "<", ">=", "<="}
	isValid := false
	for _, op := range validOperators {
		if operator == op {
			isValid = true
			break
		}
	}
	if !isValid {
		return domain.ErrInvalidRuleConditions
	}

	return nil
}

// validateRuleCondition validates a rule condition
func (s *RuleService) validateRuleCondition(condition map[string]interface{}) error {
	// Required fields: ruleId
	if _, ok := condition["ruleId"].(string); !ok {
		return domain.ErrInvalidRuleConditions
	}

	return nil
}

// validateDependencies validates the dependencies of a rule
func (s *RuleService) validateDependencies(ctx context.Context, namespace string, conditions json.RawMessage) error {
	// Parse conditions to extract dependencies
	var conditionsArray []map[string]interface{}
	if err := json.Unmarshal(conditions, &conditionsArray); err != nil {
		return domain.ErrInvalidRuleConditions
	}

	// Validate each condition's dependencies
	for _, condition := range conditionsArray {
		conditionType, ok := condition["type"].(string)
		if !ok {
			return domain.ErrInvalidRuleConditions
		}

		switch conditionType {
		case "function":
			if err := s.validateFunctionDependency(ctx, namespace, condition); err != nil {
				return err
			}
		case "field":
			if err := s.validateFieldDependency(ctx, namespace, condition); err != nil {
				return err
			}
		case "rule":
			if err := s.validateRuleDependency(ctx, namespace, condition); err != nil {
				return err
			}
		default:
			return domain.ErrInvalidRuleConditions
		}
	}

	return nil
}

// validateFunctionDependency validates that a referenced function exists and is active
func (s *RuleService) validateFunctionDependency(ctx context.Context, namespace string, condition map[string]interface{}) error {
	functionID, ok := condition["functionId"].(string)
	if !ok {
		return domain.ErrInvalidRuleConditions
	}

	// Check if function exists and is active
	function, err := s.functionRepo.GetActiveVersion(ctx, namespace, functionID)
	if err != nil {
		// Treat any error as not found for dependency validation
		return domain.ErrFunctionNotFound
	}

	// If function is found but not active, return error
	if function.Status != domain.StatusActive {
		return domain.ErrFunctionNotActive
	}

	return nil
}

// validateFieldDependency validates that a referenced field exists
func (s *RuleService) validateFieldDependency(ctx context.Context, namespace string, condition map[string]interface{}) error {
	fieldID, ok := condition["fieldId"].(string)
	if !ok {
		return domain.ErrInvalidRuleConditions
	}

	// Check if field exists
	exists, err := s.fieldRepo.Exists(ctx, namespace, fieldID)
	if err != nil {
		return domain.ErrInternalError
	}

	if !exists {
		return domain.ErrFieldNotFound
	}

	return nil
}

// validateRuleDependency validates that a referenced rule exists and is active
func (s *RuleService) validateRuleDependency(ctx context.Context, namespace string, condition map[string]interface{}) error {
	ruleID, ok := condition["ruleId"].(string)
	if !ok {
		return domain.ErrInvalidRuleConditions
	}

	// Check if rule exists and is active
	rule, err := s.repo.GetActiveVersion(ctx, namespace, ruleID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.ErrRuleNotFound
		}
		return domain.ErrInternalError
	}

	if rule == nil {
		return domain.ErrRuleNotFound
	}

	if rule.Status != domain.StatusActive {
		return domain.ErrRuleNotActive
	}

	return nil
}
