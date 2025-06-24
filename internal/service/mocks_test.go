package service

import (
	"context"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNamespaceRepository is a mock implementation of NamespaceRepository
type MockNamespaceRepository struct {
	mock.Mock
}

func (m *MockNamespaceRepository) Create(ctx context.Context, namespace *domain.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockNamespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceRepository) List(ctx context.Context) ([]*domain.Namespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Namespace), args.Error(1)
}

func (m *MockNamespaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockFieldRepository is a mock implementation of FieldRepository
type MockFieldRepository struct {
	mock.Mock
}

func (m *MockFieldRepository) Create(ctx context.Context, field *domain.Field) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockFieldRepository) GetByID(ctx context.Context, namespace, fieldID string) (*domain.Field, error) {
	args := m.Called(ctx, namespace, fieldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Field), args.Error(1)
}

func (m *MockFieldRepository) GetActiveVersion(ctx context.Context, namespace, fieldID string) (*domain.Field, error) {
	args := m.Called(ctx, namespace, fieldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Field), args.Error(1)
}

func (m *MockFieldRepository) List(ctx context.Context, namespace string) ([]*domain.Field, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Field), args.Error(1)
}

func (m *MockFieldRepository) Update(ctx context.Context, field *domain.Field) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockFieldRepository) Delete(ctx context.Context, namespace, fieldID string) error {
	args := m.Called(ctx, namespace, fieldID)
	return args.Error(0)
}

func (m *MockFieldRepository) Exists(ctx context.Context, namespace, fieldID string) (bool, error) {
	args := m.Called(ctx, namespace, fieldID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFieldRepository) CountByNamespace(ctx context.Context, namespace string) (int64, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFieldRepository) NamespaceExists(ctx context.Context, namespace string) (bool, error) {
	args := m.Called(ctx, namespace)
	return args.Bool(0), args.Error(1)
}

// MockFunctionRepository is a mock implementation of FunctionRepository
type MockFunctionRepository struct {
	mock.Mock
}

func (m *MockFunctionRepository) Create(ctx context.Context, function *domain.Function) error {
	args := m.Called(ctx, function)
	return args.Error(0)
}

func (m *MockFunctionRepository) GetByID(ctx context.Context, namespace, functionID string, version int32) (*domain.Function, error) {
	args := m.Called(ctx, namespace, functionID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) GetActiveVersion(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	args := m.Called(ctx, namespace, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) GetDraftVersion(ctx context.Context, namespace, functionID string) (*domain.Function, error) {
	args := m.Called(ctx, namespace, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) List(ctx context.Context, namespace string) ([]*domain.Function, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Function, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) ListVersions(ctx context.Context, namespace, functionID string) ([]*domain.Function, error) {
	args := m.Called(ctx, namespace, functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *MockFunctionRepository) Update(ctx context.Context, function *domain.Function) error {
	args := m.Called(ctx, function)
	return args.Error(0)
}

func (m *MockFunctionRepository) Publish(ctx context.Context, namespace, functionID string, version int32, publishedBy string) error {
	args := m.Called(ctx, namespace, functionID, version, publishedBy)
	return args.Error(0)
}

func (m *MockFunctionRepository) Deactivate(ctx context.Context, namespace, functionID string) error {
	args := m.Called(ctx, namespace, functionID)
	return args.Error(0)
}

func (m *MockFunctionRepository) Delete(ctx context.Context, namespace, functionID string, version int32) error {
	args := m.Called(ctx, namespace, functionID, version)
	return args.Error(0)
}

func (m *MockFunctionRepository) GetMaxVersion(ctx context.Context, namespace, functionID string) (int32, error) {
	args := m.Called(ctx, namespace, functionID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockFunctionRepository) Exists(ctx context.Context, namespace, functionID string, version int32) (bool, error) {
	args := m.Called(ctx, namespace, functionID, version)
	return args.Bool(0), args.Error(1)
}

// MockRuleRepository is a mock implementation of RuleRepository
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, namespace, ruleID string, version int32) (*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) GetActiveVersion(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) GetDraftVersion(ctx context.Context, namespace, ruleID string) (*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) List(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Rule, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) ListVersions(ctx context.Context, namespace, ruleID string) ([]*domain.Rule, error) {
	args := m.Called(ctx, namespace, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Publish(ctx context.Context, namespace, ruleID string, version int32, publishedBy string) error {
	args := m.Called(ctx, namespace, ruleID, version, publishedBy)
	return args.Error(0)
}

func (m *MockRuleRepository) Deactivate(ctx context.Context, namespace, ruleID string) error {
	args := m.Called(ctx, namespace, ruleID)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, namespace, ruleID string, version int32) error {
	args := m.Called(ctx, namespace, ruleID, version)
	return args.Error(0)
}

func (m *MockRuleRepository) GetMaxVersion(ctx context.Context, namespace, ruleID string) (int32, error) {
	args := m.Called(ctx, namespace, ruleID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockRuleRepository) Exists(ctx context.Context, namespace, ruleID string, version int32) (bool, error) {
	args := m.Called(ctx, namespace, ruleID, version)
	return args.Bool(0), args.Error(1)
}

// MockWorkflowRepository is a mock implementation of WorkflowRepository
type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) Create(ctx context.Context, workflow *domain.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) GetByID(ctx context.Context, namespace, workflowID string, version int32) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) GetActiveVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) GetDraftVersion(ctx context.Context, namespace, workflowID string) (*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) List(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) ListActive(ctx context.Context, namespace string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) ListVersions(ctx context.Context, namespace, workflowID string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, namespace, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) Update(ctx context.Context, workflow *domain.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error {
	args := m.Called(ctx, namespace, workflowID, version, publishedBy)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Deactivate(ctx context.Context, namespace, workflowID string) error {
	args := m.Called(ctx, namespace, workflowID)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Delete(ctx context.Context, namespace, workflowID string, version int32) error {
	args := m.Called(ctx, namespace, workflowID, version)
	return args.Error(0)
}

func (m *MockWorkflowRepository) GetMaxVersion(ctx context.Context, namespace, workflowID string) (int32, error) {
	args := m.Called(ctx, namespace, workflowID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockWorkflowRepository) Exists(ctx context.Context, namespace, workflowID string, version int32) (bool, error) {
	args := m.Called(ctx, namespace, workflowID, version)
	return args.Bool(0), args.Error(1)
}

// MockTerminalRepository is a mock implementation of TerminalRepository
type MockTerminalRepository struct {
	mock.Mock
}

func (m *MockTerminalRepository) Create(ctx context.Context, terminal *domain.Terminal) error {
	args := m.Called(ctx, terminal)
	return args.Error(0)
}

func (m *MockTerminalRepository) GetByID(ctx context.Context, namespace, terminalID string) (*domain.Terminal, error) {
	args := m.Called(ctx, namespace, terminalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Terminal), args.Error(1)
}

func (m *MockTerminalRepository) List(ctx context.Context, namespace string) ([]*domain.Terminal, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Terminal), args.Error(1)
}

func (m *MockTerminalRepository) Update(ctx context.Context, terminal *domain.Terminal) error {
	args := m.Called(ctx, terminal)
	return args.Error(0)
}

func (m *MockTerminalRepository) Delete(ctx context.Context, namespace, terminalID string) error {
	args := m.Called(ctx, namespace, terminalID)
	return args.Error(0)
}

func (m *MockTerminalRepository) Exists(ctx context.Context, namespace, terminalID string) (bool, error) {
	args := m.Called(ctx, namespace, terminalID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTerminalRepository) CountByNamespace(ctx context.Context, namespace string) (int64, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int64), args.Error(1)
}
