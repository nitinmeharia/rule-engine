package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFieldQueries struct {
	mock.Mock
}

func (m *MockFieldQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockFieldQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GetFieldRow), args.Error(1)
}
func (m *MockFieldQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ListFieldsRow), args.Error(1)
}
func (m *MockFieldQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockFieldQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockFieldQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}
func (m *MockFieldQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockFieldQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockFieldQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Namespace), args.Error(1)
}

// Add all missing db.Querier methods
func (m *MockFieldQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}
func (m *MockFieldQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockFieldQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error { return nil }
func (m *MockFieldQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}
func (m *MockFieldQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}
func (m *MockFieldQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}
func (m *MockFieldQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}
func (m *MockFieldQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}
func (m *MockFieldQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockFieldQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}
func (m *MockFieldQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error { return nil }
func (m *MockFieldQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}
func (m *MockFieldQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}
func (m *MockFieldQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFieldQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFieldQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}
func (m *MockFieldQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}
func (m *MockFieldQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}
func (m *MockFieldQueries) RefreshNamespaceChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockFieldQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFieldQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFieldQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}
func (m *MockFieldQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error { return nil }
func (m *MockFieldQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}
func (m *MockFieldQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}
func (m *MockFieldQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFieldQueries) DeleteNamespace(ctx context.Context, id string) error { return nil }
func (m *MockFieldQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}

func TestFieldRepository_Create(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	field := &domain.Field{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        "number",
		Description: "desc",
		CreatedBy:   "user",
	}
	params := db.CreateFieldParams{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        &field.Type,
		Description: &field.Description,
		CreatedBy:   "user",
	}
	mockQ.On("CreateField", ctx, params).Return(nil)
	err := repo.Create(ctx, field)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	row := &db.GetFieldRow{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        strPtr("string"),
		Description: strPtr("test field"),
		CreatedBy:   "user",
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	params := db.GetFieldParams{Namespace: "ns", FieldID: "f1"}
	mockQ.On("GetField", ctx, params).Return(row, nil)
	f, err := repo.GetByID(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.Equal(t, "f1", f.FieldID)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_List(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	row := &db.ListFieldsRow{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        strPtr("number"),
		Description: strPtr("desc"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		CreatedBy:   "user",
	}
	mockQ.On("ListFields", ctx, "ns").Return([]*db.ListFieldsRow{row}, nil)
	fields, err := repo.List(ctx, "ns")
	assert.NoError(t, err)
	assert.Len(t, fields, 1)
	assert.Equal(t, "f1", fields[0].FieldID)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_Update(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	field := &domain.Field{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        "number",
		Description: "desc",
		CreatedBy:   "user",
	}
	params := db.UpdateFieldParams{
		Namespace:   "ns",
		FieldID:     "f1",
		Type:        &field.Type,
		Description: &field.Description,
		CreatedBy:   "user",
	}
	mockQ.On("UpdateField", ctx, params).Return(nil)
	err := repo.Update(ctx, field)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_Delete(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	params := db.DeleteFieldParams{Namespace: "ns", FieldID: "f1"}
	mockQ.On("DeleteField", ctx, params).Return(nil)
	err := repo.Delete(ctx, "ns", "f1")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_Exists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	params := db.FieldExistsParams{Namespace: "ns", FieldID: "f1"}
	mockQ.On("FieldExists", ctx, params).Return(true, nil)
	exists, err := repo.Exists(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_CountByNamespace(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	mockQ.On("CountFieldsByNamespace", ctx, "ns").Return(int64(5), nil)
	count, err := repo.CountByNamespace(ctx, "ns")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	mockQ.AssertExpectations(t)
}

func TestFieldRepository_NamespaceExists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFieldQueries)
	repo := NewFieldRepository(mockQ)
	ns := &db.Namespace{ID: "ns"}
	mockQ.On("GetNamespace", ctx, "ns").Return(ns, nil)
	exists, err := repo.NamespaceExists(ctx, "ns")
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)

	mockQ2 := new(MockFieldQueries)
	repo2 := NewFieldRepository(mockQ2)
	mockQ2.On("GetNamespace", ctx, "missing").Return(nil, pgx.ErrNoRows)
	exists, err = repo2.NamespaceExists(ctx, "missing")
	assert.NoError(t, err)
	assert.False(t, exists)
	mockQ2.AssertExpectations(t)
}

func strPtr(s string) *string { return &s }
