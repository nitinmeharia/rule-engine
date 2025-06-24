package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRuleQueries struct {
	mock.Mock
}

// Implement all db.Querier methods
func (m *MockRuleQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRuleQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Rule), args.Error(1)
}

func (m *MockRuleQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRuleQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRuleQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRuleQueries) RefreshNamespaceChecksum(ctx context.Context, namespace string) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockRuleQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRuleQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	args := m.Called(ctx, arg)
	return args.Get(0), args.Error(1)
}

func (m *MockRuleQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}

// Stubs for all other db.Querier methods
func (m *MockRuleQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockRuleQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockRuleQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}
func (m *MockRuleQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}
func (m *MockRuleQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockRuleQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}
func (m *MockRuleQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}
func (m *MockRuleQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}
func (m *MockRuleQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}
func (m *MockRuleQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockRuleQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}
func (m *MockRuleQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}
func (m *MockRuleQueries) DeleteNamespace(ctx context.Context, id string) error { return nil }
func (m *MockRuleQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}
func (m *MockRuleQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}
func (m *MockRuleQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}
func (m *MockRuleQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}
func (m *MockRuleQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockRuleQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}
func (m *MockRuleQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}
func (m *MockRuleQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}
func (m *MockRuleQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}
func (m *MockRuleQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}
func (m *MockRuleQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}
func (m *MockRuleQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}
func (m *MockRuleQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}

func TestRuleRepository_Create(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &domain.Rule{
		Namespace:  "ns",
		RuleID:     "r1",
		Version:    1,
		Status:     "draft",
		Logic:      "AND",
		Conditions: []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:  "user",
	}

	params := db.CreateRuleParams{
		Namespace:  "ns",
		RuleID:     "r1",
		Version:    1,
		Status:     &rule.Status,
		Logic:      &rule.Logic,
		Conditions: rule.Conditions,
		CreatedBy:  "user",
	}

	mockQ.On("CreateRule", ctx, params).Return(nil)
	err := repo.Create(ctx, rule)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &domain.Rule{
		Namespace: "ns",
		RuleID:    "r1",
		Version:   1,
		Status:    "draft",
		CreatedBy: "user",
	}

	params := db.CreateRuleParams{
		Namespace:  "ns",
		RuleID:     "r1",
		Version:    1,
		Status:     &rule.Status,
		Logic:      &rule.Logic,
		Conditions: rule.Conditions,
		CreatedBy:  "user",
	}

	mockQ.On("CreateRule", ctx, params).Return(errors.New("db error"))
	err := repo.Create(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create rule")
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("active"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.GetRuleParams{
		Namespace: "ns",
		RuleID:    "r1",
		Version:   1,
	}

	mockQ.On("GetRule", ctx, params).Return(rule, nil)
	r, err := repo.GetByID(ctx, "ns", "r1", 1)
	assert.NoError(t, err)
	assert.Equal(t, "r1", r.RuleID)
	assert.Equal(t, "active", r.Status)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_GetActiveVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("active"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.GetActiveRuleVersionParams{
		Namespace: "ns",
		RuleID:    "r1",
	}

	mockQ.On("GetActiveRuleVersion", ctx, params).Return(rule, nil)
	r, err := repo.GetActiveVersion(ctx, "ns", "r1")
	assert.NoError(t, err)
	assert.Equal(t, "r1", r.RuleID)
	assert.Equal(t, "active", r.Status)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_GetDraftVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("draft"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: nil,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Valid: false},
	}

	params := db.GetDraftRuleVersionParams{
		Namespace: "ns",
		RuleID:    "r1",
	}

	mockQ.On("GetDraftRuleVersion", ctx, params).Return(rule, nil)
	r, err := repo.GetDraftVersion(ctx, "ns", "r1")
	assert.NoError(t, err)
	assert.Equal(t, "r1", r.RuleID)
	assert.Equal(t, "draft", r.Status)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_List(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("active"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("ListRules", ctx, "ns").Return([]*db.Rule{rule}, nil)
	rules, err := repo.List(ctx, "ns")
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "r1", rules[0].RuleID)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_ListActive(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("active"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("ListActiveRules", ctx, "ns").Return([]*db.Rule{rule}, nil)
	rules, err := repo.ListActive(ctx, "ns")
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "r1", rules[0].RuleID)
	assert.Equal(t, "active", rules[0].Status)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_ListVersions(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &db.Rule{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		Status:      strPtr("active"),
		Logic:       strPtr("AND"),
		Conditions:  []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.ListRuleVersionsParams{
		Namespace: "ns",
		RuleID:    "r1",
	}

	mockQ.On("ListRuleVersions", ctx, params).Return([]*db.Rule{rule}, nil)
	rules, err := repo.ListVersions(ctx, "ns", "r1")
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, "r1", rules[0].RuleID)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Update(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	rule := &domain.Rule{
		Namespace:  "ns",
		RuleID:     "r1",
		Version:    1,
		Logic:      "AND",
		Conditions: []byte(`[{"fieldId":"f1","operator":"eq","value":"test"}]`),
		CreatedBy:  "user",
	}

	params := db.UpdateRuleParams{
		Namespace:  "ns",
		RuleID:     "r1",
		Version:    1,
		Logic:      &rule.Logic,
		Conditions: rule.Conditions,
		CreatedBy:  "user",
	}

	mockQ.On("UpdateRule", ctx, params).Return(nil)
	err := repo.Update(ctx, rule)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Publish(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	// Mock deactivate call
	deactivateParams := db.DeactivateRuleParams{
		Namespace: "ns",
		RuleID:    "r1",
	}
	mockQ.On("DeactivateRule", ctx, deactivateParams).Return(nil)

	// Mock publish call
	publishParams := db.PublishRuleParams{
		Namespace:   "ns",
		RuleID:      "r1",
		Version:     1,
		PublishedBy: strPtr("user"),
	}
	mockQ.On("PublishRule", ctx, publishParams).Return(nil)

	// Mock refresh checksum call
	mockQ.On("RefreshNamespaceChecksum", ctx, "ns").Return(nil)

	err := repo.Publish(ctx, "ns", "r1", 1, "user")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Deactivate(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	params := db.DeactivateRuleParams{
		Namespace: "ns",
		RuleID:    "r1",
	}

	mockQ.On("DeactivateRule", ctx, params).Return(nil)
	err := repo.Deactivate(ctx, "ns", "r1")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Delete(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	params := db.DeleteRuleParams{
		Namespace: "ns",
		RuleID:    "r1",
		Version:   1,
	}

	mockQ.On("DeleteRule", ctx, params).Return(nil)
	err := repo.Delete(ctx, "ns", "r1", 1)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_GetMaxVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	params := db.GetMaxRuleVersionParams{
		Namespace: "ns",
		RuleID:    "r1",
	}

	mockQ.On("GetMaxRuleVersion", ctx, params).Return(int64(5), nil)
	version, err := repo.GetMaxVersion(ctx, "ns", "r1")
	assert.NoError(t, err)
	assert.Equal(t, int32(5), version)
	mockQ.AssertExpectations(t)
}

func TestRuleRepository_Exists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockRuleQueries)
	repo := NewRuleRepository(mockQ)

	params := db.RuleExistsParams{
		Namespace: "ns",
		RuleID:    "r1",
		Version:   1,
	}

	mockQ.On("RuleExists", ctx, params).Return(true, nil)
	exists, err := repo.Exists(ctx, "ns", "r1", 1)
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)
}

