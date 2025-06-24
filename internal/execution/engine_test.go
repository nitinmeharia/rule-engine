package execution

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---
type MockCacheRepo struct{ mock.Mock }

func (m *MockCacheRepo) GetActiveConfigChecksum(ctx context.Context, ns string) (*domain.ActiveConfigMeta, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(*domain.ActiveConfigMeta), args.Error(1)
}

func (m *MockCacheRepo) UpsertActiveConfigChecksum(ctx context.Context, ns, checksum string) error {
	return nil
}
func (m *MockCacheRepo) RefreshNamespaceChecksum(ctx context.Context, ns string) error { return nil }
func (m *MockCacheRepo) ListAllActiveConfigChecksums(ctx context.Context) ([]*domain.ActiveConfigMeta, error) {
	return nil, nil
}
func (m *MockCacheRepo) DeleteActiveConfigChecksum(ctx context.Context, ns string) error { return nil }

type MockRuleRepo struct{ mock.Mock }

func (m *MockRuleRepo) ListActive(ctx context.Context, ns string) ([]*domain.Rule, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).([]*domain.Rule), args.Error(1)
}

func (m *MockRuleRepo) Create(ctx context.Context, rule *domain.Rule) error { return nil }
func (m *MockRuleRepo) GetByID(ctx context.Context, ns, ruleID string, version int32) (*domain.Rule, error) {
	return nil, nil
}
func (m *MockRuleRepo) GetActiveVersion(ctx context.Context, ns, ruleID string) (*domain.Rule, error) {
	return nil, nil
}
func (m *MockRuleRepo) GetDraftVersion(ctx context.Context, ns, ruleID string) (*domain.Rule, error) {
	return nil, nil
}
func (m *MockRuleRepo) List(ctx context.Context, ns string) ([]*domain.Rule, error) { return nil, nil }
func (m *MockRuleRepo) ListVersions(ctx context.Context, ns, ruleID string) ([]*domain.Rule, error) {
	return nil, nil
}
func (m *MockRuleRepo) Update(ctx context.Context, rule *domain.Rule) error { return nil }
func (m *MockRuleRepo) Publish(ctx context.Context, ns, ruleID string, version int32, publishedBy string) error {
	return nil
}
func (m *MockRuleRepo) Deactivate(ctx context.Context, ns, ruleID string) error { return nil }
func (m *MockRuleRepo) Delete(ctx context.Context, ns, ruleID string, version int32) error {
	return nil
}
func (m *MockRuleRepo) GetMaxVersion(ctx context.Context, ns, ruleID string) (int32, error) {
	return 0, nil
}
func (m *MockRuleRepo) Exists(ctx context.Context, ns, ruleID string, version int32) (bool, error) {
	return false, nil
}

type MockWorkflowRepo struct{ mock.Mock }

func (m *MockWorkflowRepo) ListActive(ctx context.Context, ns string) ([]*domain.Workflow, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepo) Create(ctx context.Context, workflow *domain.Workflow) error { return nil }
func (m *MockWorkflowRepo) GetByID(ctx context.Context, ns, workflowID string, version int32) (*domain.Workflow, error) {
	return nil, nil
}
func (m *MockWorkflowRepo) GetActiveVersion(ctx context.Context, ns, workflowID string) (*domain.Workflow, error) {
	return nil, nil
}
func (m *MockWorkflowRepo) GetDraftVersion(ctx context.Context, ns, workflowID string) (*domain.Workflow, error) {
	return nil, nil
}
func (m *MockWorkflowRepo) List(ctx context.Context, ns string) ([]*domain.Workflow, error) {
	return nil, nil
}
func (m *MockWorkflowRepo) ListVersions(ctx context.Context, ns, workflowID string) ([]*domain.Workflow, error) {
	return nil, nil
}
func (m *MockWorkflowRepo) Update(ctx context.Context, workflow *domain.Workflow) error { return nil }
func (m *MockWorkflowRepo) Publish(ctx context.Context, ns, workflowID string, version int32, publishedBy string) error {
	return nil
}
func (m *MockWorkflowRepo) Deactivate(ctx context.Context, ns, workflowID string) error { return nil }
func (m *MockWorkflowRepo) Delete(ctx context.Context, ns, workflowID string, version int32) error {
	return nil
}
func (m *MockWorkflowRepo) GetMaxVersion(ctx context.Context, ns, workflowID string) (int32, error) {
	return 0, nil
}
func (m *MockWorkflowRepo) Exists(ctx context.Context, ns, workflowID string, version int32) (bool, error) {
	return false, nil
}

type MockFieldRepo struct{ mock.Mock }

func (m *MockFieldRepo) List(ctx context.Context, ns string) ([]*domain.Field, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).([]*domain.Field), args.Error(1)
}

func (m *MockFieldRepo) Create(ctx context.Context, field *domain.Field) error { return nil }
func (m *MockFieldRepo) GetByID(ctx context.Context, ns, fieldID string) (*domain.Field, error) {
	return nil, nil
}
func (m *MockFieldRepo) Update(ctx context.Context, field *domain.Field) error { return nil }
func (m *MockFieldRepo) Delete(ctx context.Context, ns, fieldID string) error  { return nil }
func (m *MockFieldRepo) Exists(ctx context.Context, ns, fieldID string) (bool, error) {
	return false, nil
}
func (m *MockFieldRepo) CountByNamespace(ctx context.Context, ns string) (int64, error) {
	return 0, nil
}

type MockFunctionRepo struct{ mock.Mock }

func (m *MockFunctionRepo) ListActive(ctx context.Context, ns string) ([]*domain.Function, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *MockFunctionRepo) Create(ctx context.Context, function *domain.Function) error { return nil }
func (m *MockFunctionRepo) GetByID(ctx context.Context, ns, functionID string, version int32) (*domain.Function, error) {
	return nil, nil
}
func (m *MockFunctionRepo) GetActiveVersion(ctx context.Context, ns, functionID string) (*domain.Function, error) {
	return nil, nil
}
func (m *MockFunctionRepo) GetDraftVersion(ctx context.Context, ns, functionID string) (*domain.Function, error) {
	return nil, nil
}
func (m *MockFunctionRepo) List(ctx context.Context, ns string) ([]*domain.Function, error) {
	return nil, nil
}
func (m *MockFunctionRepo) ListVersions(ctx context.Context, ns, functionID string) ([]*domain.Function, error) {
	return nil, nil
}
func (m *MockFunctionRepo) Update(ctx context.Context, function *domain.Function) error { return nil }
func (m *MockFunctionRepo) Publish(ctx context.Context, ns, functionID string, version int32, publishedBy string) error {
	return nil
}
func (m *MockFunctionRepo) Deactivate(ctx context.Context, ns, functionID string) error { return nil }
func (m *MockFunctionRepo) Delete(ctx context.Context, ns, functionID string, version int32) error {
	return nil
}
func (m *MockFunctionRepo) GetMaxVersion(ctx context.Context, ns, functionID string) (int32, error) {
	return 0, nil
}
func (m *MockFunctionRepo) Exists(ctx context.Context, ns, functionID string, version int32) (bool, error) {
	return false, nil
}

type MockTerminalRepo struct{ mock.Mock }

func (m *MockTerminalRepo) List(ctx context.Context, ns string) ([]*domain.Terminal, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).([]*domain.Terminal), args.Error(1)
}

func (m *MockTerminalRepo) Create(ctx context.Context, terminal *domain.Terminal) error { return nil }
func (m *MockTerminalRepo) GetByID(ctx context.Context, ns, terminalID string) (*domain.Terminal, error) {
	return nil, nil
}
func (m *MockTerminalRepo) Delete(ctx context.Context, ns, terminalID string) error { return nil }
func (m *MockTerminalRepo) Exists(ctx context.Context, ns, terminalID string) (bool, error) {
	return false, nil
}
func (m *MockTerminalRepo) CountByNamespace(ctx context.Context, ns string) (int64, error) {
	return 0, nil
}

func setupEngine(t *testing.T) *Engine {
	cacheRepo := new(MockCacheRepo)
	ruleRepo := new(MockRuleRepo)
	workflowRepo := new(MockWorkflowRepo)
	fieldRepo := new(MockFieldRepo)
	functionRepo := new(MockFunctionRepo)
	terminalRepo := new(MockTerminalRepo)
	return NewEngine(cacheRepo, ruleRepo, workflowRepo, fieldRepo, functionRepo, terminalRepo, 0)
}

func TestEngine_ExecuteRule_Success(t *testing.T) {
	// Set up mocks first
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	ruleID := "r1"

	// Set up test data
	cond := map[string]interface{}{
		"a": map[string]interface{}{"field": "x", "operator": "eq", "value": 5},
	}
	condBytes, _ := json.Marshal(cond)
	rule := &domain.Rule{RuleID: ruleID, Logic: domain.LogicAND, Conditions: condBytes}
	field := &domain.Field{FieldID: "x", Type: "number"}

	// Set up mock expectations - return the same data that will be in the cache
	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "dummy"}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{field}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{rule}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{}, nil)

	// Create engine with mocks
	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)

	// Manually set cache to avoid refresh
	config := &NamespaceConfig{
		Rules:     map[string]*domain.Rule{ruleID: rule},
		Fields:    map[string]*domain.Field{"x": field},
		Functions: map[string]*domain.Function{},
		Terminals: map[string]*domain.Terminal{},
		Workflows: map[string]*domain.Workflow{},
	}
	e.cache.mutex.Lock()
	e.cache.data[ns] = config
	e.cache.checksums[ns] = "dummy"
	e.cache.mutex.Unlock()

	req := &domain.ExecutionRequest{Namespace: ns, RuleID: &ruleID, Data: map[string]interface{}{"x": 5}}
	resp, err := e.ExecuteRule(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, true, resp.Result)
}

func TestEngine_ExecuteRule_RuleNotFound(t *testing.T) {
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	ruleID := "r1"

	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "dummy"}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{}, nil)

	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)

	config := &NamespaceConfig{
		Rules:     map[string]*domain.Rule{},
		Fields:    map[string]*domain.Field{},
		Functions: map[string]*domain.Function{},
		Terminals: map[string]*domain.Terminal{},
		Workflows: map[string]*domain.Workflow{},
	}
	e.cache.mutex.Lock()
	e.cache.data[ns] = config
	e.cache.checksums[ns] = "dummy"
	e.cache.mutex.Unlock()

	req := &domain.ExecutionRequest{Namespace: ns, RuleID: &ruleID, Data: map[string]interface{}{"x": 5}}
	_, err := e.ExecuteRule(context.Background(), req)
	assert.Error(t, err)
}

func TestEngine_ExecuteWorkflow_Success(t *testing.T) {
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	wfID := "wf1"

	steps := map[string]interface{}{
		"s1": map[string]interface{}{"type": "terminal", "terminalId": "t1"},
	}
	stepsBytes, _ := json.Marshal(steps)
	wf := &domain.Workflow{WorkflowID: wfID, StartAt: "s1", Steps: stepsBytes}
	terminal := &domain.Terminal{TerminalID: "t1"}

	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "dummy"}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{terminal}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{wf}, nil)

	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)

	config := &NamespaceConfig{
		Workflows: map[string]*domain.Workflow{wfID: wf},
		Terminals: map[string]*domain.Terminal{"t1": terminal},
		Rules:     map[string]*domain.Rule{},
		Fields:    map[string]*domain.Field{},
		Functions: map[string]*domain.Function{},
	}
	e.cache.mutex.Lock()
	e.cache.data[ns] = config
	e.cache.checksums[ns] = "dummy"
	e.cache.mutex.Unlock()

	req := &domain.ExecutionRequest{Namespace: ns, WorkflowID: &wfID, Data: map[string]interface{}{}}
	resp, err := e.ExecuteWorkflow(context.Background(), req)
	assert.NoError(t, err)
	assert.Nil(t, resp.Result)
}

func TestEngine_ExecuteWorkflow_WorkflowNotFound(t *testing.T) {
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	wfID := "wf1"

	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "dummy"}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{}, nil)

	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)

	config := &NamespaceConfig{
		Workflows: map[string]*domain.Workflow{},
		Terminals: map[string]*domain.Terminal{},
		Rules:     map[string]*domain.Rule{},
		Fields:    map[string]*domain.Field{},
		Functions: map[string]*domain.Function{},
	}
	e.cache.mutex.Lock()
	e.cache.data[ns] = config
	e.cache.checksums[ns] = "dummy"
	e.cache.mutex.Unlock()

	req := &domain.ExecutionRequest{Namespace: ns, WorkflowID: &wfID, Data: map[string]interface{}{}}
	_, err := e.ExecuteWorkflow(context.Background(), req)
	assert.Error(t, err)
}

func TestEngine_ConditionOperators(t *testing.T) {
	e := setupEngine(t)
	ctx := &ExecutionContext{Data: map[string]interface{}{"x": 5}}
	config := &NamespaceConfig{}
	assert.True(t, e.evaluateOperator(5, "eq", 5, config, ctx))
	assert.False(t, e.evaluateOperator(5, "ne", 5, config, ctx))
	assert.True(t, e.evaluateOperator(6, "gt", 5, config, ctx))
	assert.True(t, e.evaluateOperator(5, "gte", 5, config, ctx))
	assert.True(t, e.evaluateOperator(4, "lt", 5, config, ctx))
	assert.True(t, e.evaluateOperator(5, "lte", 5, config, ctx))
	assert.True(t, e.evaluateOperator("a", "in", []interface{}{"a", "b"}, config, ctx))
	assert.True(t, e.evaluateOperator("a", "not_in", []interface{}{"b", "c"}, config, ctx))
}

func TestEngine_Tracing(t *testing.T) {
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	ruleID := "r1"

	cond := map[string]interface{}{"a": map[string]interface{}{"field": "x", "operator": "eq", "value": 5}}
	condBytes, _ := json.Marshal(cond)
	rule := &domain.Rule{RuleID: ruleID, Logic: domain.LogicAND, Conditions: condBytes}

	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "dummy"}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{{FieldID: "x", Type: "number"}}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{rule}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{}, nil)

	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)

	config := &NamespaceConfig{
		Rules:     map[string]*domain.Rule{ruleID: rule},
		Fields:    map[string]*domain.Field{"x": {FieldID: "x", Type: "number"}},
		Functions: map[string]*domain.Function{},
		Terminals: map[string]*domain.Terminal{},
		Workflows: map[string]*domain.Workflow{},
	}
	e.cache.mutex.Lock()
	e.cache.data[ns] = config
	e.cache.checksums[ns] = "dummy"
	e.cache.mutex.Unlock()

	req := &domain.ExecutionRequest{Namespace: ns, RuleID: &ruleID, Data: map[string]interface{}{"x": 5}, Trace: true}
	resp, err := e.ExecuteRule(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Trace)
	assert.GreaterOrEqual(t, len(resp.Trace.Steps), 1)
}

func TestEngine_RefreshCache(t *testing.T) {
	mockCacheRepo := new(MockCacheRepo)
	mockFieldRepo := new(MockFieldRepo)
	mockFunctionRepo := new(MockFunctionRepo)
	mockRuleRepo := new(MockRuleRepo)
	mockTerminalRepo := new(MockTerminalRepo)
	mockWorkflowRepo := new(MockWorkflowRepo)

	ns := "test-ns"
	mockCacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{Namespace: ns, Checksum: "abc", UpdatedAt: time.Now()}, nil)
	mockFieldRepo.On("List", mock.Anything, ns).Return([]*domain.Field{}, nil)
	mockFunctionRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Function{}, nil)
	mockRuleRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Rule{}, nil)
	mockTerminalRepo.On("List", mock.Anything, ns).Return([]*domain.Terminal{}, nil)
	mockWorkflowRepo.On("ListActive", mock.Anything, ns).Return([]*domain.Workflow{}, nil)

	e := NewEngine(mockCacheRepo, mockRuleRepo, mockWorkflowRepo, mockFieldRepo, mockFunctionRepo, mockTerminalRepo, 0)
	e.cache.data[ns] = &NamespaceConfig{}
	e.cache.checksums[ns] = "old"
	e.lastRefresh[ns] = time.Now().Add(-time.Hour)
	err := e.ensureFreshCache(context.Background(), ns)
	assert.NoError(t, err)
}

func TestEngine_ErrorHandling(t *testing.T) {
	e := setupEngine(t)
	ns := "test-ns"
	cacheRepo := e.cacheRepo.(*MockCacheRepo)
	cacheRepo.On("GetActiveConfigChecksum", mock.Anything, ns).Return(&domain.ActiveConfigMeta{}, errors.New("fail")).Once()
	err := e.ensureFreshCache(context.Background(), ns)
	assert.Error(t, err)
}
