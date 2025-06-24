package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRuleService_CreateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		rule := &domain.Rule{
			RuleID:     "test-rule",
			Logic:      "AND",
			CreatedBy:  "user1",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		// Mock that rule doesn't exist
		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)
		mockRuleRepo.On("GetActiveVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)
		mockRuleRepo.On("GetMaxVersion", ctx, "test-ns", "test-rule").Return(int32(0), nil)
		mockRuleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Rule")).Return(nil)
		mockFieldRepo.On("Exists", mock.Anything, "test-ns", "age").Return(true, nil)

		err := service.CreateRule(ctx, "test-ns", rule)

		assert.NoError(t, err)
		assert.Equal(t, "test-ns", rule.Namespace)
		assert.Equal(t, int32(1), rule.Version)
		assert.Equal(t, domain.StatusDraft, rule.Status)
		assert.NotNil(t, rule.CreatedAt)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("validation error - empty rule ID", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		rule := &domain.Rule{
			RuleID:    "",
			Logic:     "AND",
			CreatedBy: "user1",
		}

		err := service.CreateRule(ctx, "test-ns", rule)

		assert.Error(t, err)
		mockRuleRepo.AssertNotCalled(t, "GetDraftVersion")
		mockRuleRepo.AssertNotCalled(t, "Create")
	})

	t.Run("rule already exists - draft", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		rule := &domain.Rule{
			RuleID:    "test-rule",
			Logic:     "AND",
			CreatedBy: "user1",
		}

		existingRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusDraft,
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return(existingRule, nil)

		err := service.CreateRule(ctx, "test-ns", rule)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrRuleAlreadyExists, err)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("rule already exists - active", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		rule := &domain.Rule{
			RuleID:    "test-rule",
			Logic:     "AND",
			CreatedBy: "user1",
		}

		existingRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusActive,
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)
		mockRuleRepo.On("GetActiveVersion", ctx, "test-ns", "test-rule").Return(existingRule, nil)

		err := service.CreateRule(ctx, "test-ns", rule)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrRuleAlreadyExists, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_GetRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		expectedRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusActive,
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetActiveVersion", ctx, "test-ns", "test-rule").Return(expectedRule, nil)

		rule, err := service.GetRule(ctx, "test-ns", "test-rule")

		assert.NoError(t, err)
		assert.Equal(t, expectedRule, rule)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("rule not found", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("GetActiveVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)

		rule, err := service.GetRule(ctx, "test-ns", "test-rule")

		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.Equal(t, domain.ErrRuleNotFound, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_GetDraftRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		expectedRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusDraft,
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return(expectedRule, nil)

		rule, err := service.GetDraftRule(ctx, "test-ns", "test-rule")

		assert.NoError(t, err)
		assert.Equal(t, expectedRule, rule)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("rule not found", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)

		rule, err := service.GetDraftRule(ctx, "test-ns", "test-rule")

		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.Equal(t, domain.ErrRuleNotFound, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_ListRules(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		expectedRules := []*domain.Rule{
			{
				RuleID:     "rule1",
				Namespace:  "test-ns",
				Version:    1,
				Status:     domain.StatusActive,
				Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
			},
			{
				RuleID:     "rule2",
				Namespace:  "test-ns",
				Version:    1,
				Status:     domain.StatusDraft,
				Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
			},
		}

		mockRuleRepo.On("List", ctx, "test-ns").Return(expectedRules, nil)

		rules, err := service.ListRules(ctx, "test-ns")

		assert.NoError(t, err)
		assert.Equal(t, expectedRules, rules)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("List", ctx, "test-ns").Return(([]*domain.Rule)(nil), assert.AnError)

		rules, err := service.ListRules(ctx, "test-ns")

		assert.Error(t, err)
		assert.Nil(t, rules)
		assert.Equal(t, domain.ErrListError, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_ListRuleVersions(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		expectedRules := []*domain.Rule{
			{
				RuleID:     "test-rule",
				Namespace:  "test-ns",
				Version:    1,
				Status:     domain.StatusActive,
				Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
			},
			{
				RuleID:     "test-rule",
				Namespace:  "test-ns",
				Version:    2,
				Status:     domain.StatusDraft,
				Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
			},
		}

		mockRuleRepo.On("ListVersions", ctx, "test-ns", "test-rule").Return(expectedRules, nil)

		rules, err := service.ListRuleVersions(ctx, "test-ns", "test-rule")

		assert.NoError(t, err)
		assert.Equal(t, expectedRules, rules)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("ListVersions", ctx, "test-ns", "test-rule").Return(([]*domain.Rule)(nil), assert.AnError)

		rules, err := service.ListRuleVersions(ctx, "test-ns", "test-rule")

		assert.Error(t, err)
		assert.Nil(t, rules)
		assert.Equal(t, domain.ErrListError, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_UpdateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		existingRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusDraft,
			Logic:      "AND",
			CreatedBy:  "user1",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		updateRule := &domain.Rule{
			RuleID:     "test-rule",
			Logic:      "OR",
			CreatedBy:  "user2",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return(existingRule, nil)
		mockRuleRepo.On("Update", ctx, mock.AnythingOfType("*domain.Rule")).Return(nil)
		mockFieldRepo.On("Exists", mock.Anything, "test-ns", "age").Return(true, nil)

		err := service.UpdateRule(ctx, "test-ns", "test-rule", updateRule)

		assert.NoError(t, err)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		updateRule := &domain.Rule{
			RuleID:     "",
			Logic:      "OR",
			CreatedBy:  "user2",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		err := service.UpdateRule(ctx, "test-ns", "test-rule", updateRule)

		assert.Error(t, err)
		mockRuleRepo.AssertNotCalled(t, "GetDraftVersion")
		mockRuleRepo.AssertNotCalled(t, "Update")
	})

	t.Run("rule not found", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		updateRule := &domain.Rule{
			RuleID:     "test-rule",
			Logic:      "OR",
			CreatedBy:  "user2",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)

		err := service.UpdateRule(ctx, "test-ns", "test-rule", updateRule)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrRuleNotFound, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_PublishRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful publish", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		draftRule := &domain.Rule{
			RuleID:     "test-rule",
			Namespace:  "test-ns",
			Version:    1,
			Status:     domain.StatusDraft,
			Logic:      "AND",
			CreatedBy:  "user1",
			Conditions: json.RawMessage(`[{"type":"field","fieldId":"age","operator":">","value":18}]`),
		}

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return(draftRule, nil)
		mockRuleRepo.On("Publish", ctx, "test-ns", "test-rule", int32(1), "user2").Return(nil)
		mockFieldRepo.On("Exists", mock.Anything, "test-ns", "age").Return(true, nil)

		err := service.PublishRule(ctx, "test-ns", "test-rule", "user2")

		assert.NoError(t, err)
		mockNamespaceRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return((*domain.Namespace)(nil), sql.ErrNoRows)

		err := service.PublishRule(ctx, "test-ns", "test-rule", "user2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNamespaceNotFound, err)
		mockNamespaceRepo.AssertExpectations(t)
		mockRuleRepo.AssertNotCalled(t, "GetDraftVersion")
	})

	t.Run("rule not found", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		namespace := &domain.Namespace{
			ID: "test-ns",
		}

		mockNamespaceRepo.On("GetByID", ctx, "test-ns").Return(namespace, nil)
		mockRuleRepo.On("GetDraftVersion", ctx, "test-ns", "test-rule").Return((*domain.Rule)(nil), sql.ErrNoRows)

		err := service.PublishRule(ctx, "test-ns", "test-rule", "user2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrRuleNotFound, err)
		mockNamespaceRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_DeleteRule(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("Exists", ctx, "test-ns", "test-rule", int32(1)).Return(true, nil)
		mockRuleRepo.On("Delete", ctx, "test-ns", "test-rule", int32(1)).Return(nil)

		err := service.DeleteRule(ctx, "test-ns", "test-rule", 1)

		assert.NoError(t, err)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockRuleRepo := new(MockRuleRepository)
		mockFunctionRepo := new(MockFunctionRepository)
		mockFieldRepo := new(MockFieldRepository)
		mockNamespaceRepo := new(MockNamespaceRepository)

		service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

		mockRuleRepo.On("Exists", ctx, "test-ns", "test-rule", int32(1)).Return(true, nil)
		mockRuleRepo.On("Delete", ctx, "test-ns", "test-rule", int32(1)).Return(assert.AnError)

		err := service.DeleteRule(ctx, "test-ns", "test-rule", 1)

		assert.Error(t, err)
		mockRuleRepo.AssertExpectations(t)
	})
}

func TestRuleService_validateConditions(t *testing.T) {
	mockRuleRepo := new(MockRuleRepository)
	mockFunctionRepo := new(MockFunctionRepository)
	mockFieldRepo := new(MockFieldRepository)
	mockNamespaceRepo := new(MockNamespaceRepository)

	service := NewRuleService(mockRuleRepo, mockFunctionRepo, mockFieldRepo, mockNamespaceRepo)

	t.Run("valid conditions", func(t *testing.T) {
		conditions := json.RawMessage(`[
			{
				"type": "field",
				"fieldId": "age",
				"operator": ">",
				"value": 18
			}
		]`)
		mockFieldRepo.On("Exists", mock.Anything, "test-ns", "age").Return(true, nil)
		// Patch: set the namespace in the service for this test
		serviceWithNamespace := *service
		serviceWithNamespace.fieldRepo = mockFieldRepo
		// Call validateConditions as if the rule is in namespace "test-ns"
		err := serviceWithNamespace.validateConditions(conditions)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		conditions := json.RawMessage(`invalid json`)
		err := service.validateConditions(conditions)
		assert.Error(t, err)
	})

	t.Run("missing type", func(t *testing.T) {
		conditions := json.RawMessage(`[
			{
				"fieldId": "age",
				"operator": ">",
				"value": 18
			}
		]`)
		err := service.validateConditions(conditions)
		assert.Error(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		conditions := json.RawMessage(`[
			{
				"type": "invalid",
				"fieldId": "age",
				"operator": ">",
				"value": 18
			}
		]`)
		err := service.validateConditions(conditions)
		assert.Error(t, err)
	})
}
