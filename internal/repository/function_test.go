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

type MockFunctionQueries struct {
	mock.Mock
}

// Implement all db.Querier methods
func (m *MockFunctionQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockFunctionQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GetFunctionRow), args.Error(1)
}

func (m *MockFunctionQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GetActiveFunctionVersionRow), args.Error(1)
}

func (m *MockFunctionQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GetDraftFunctionVersionRow), args.Error(1)
}

func (m *MockFunctionQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ListFunctionsRow), args.Error(1)
}

func (m *MockFunctionQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ListActiveFunctionsRow), args.Error(1)
}

func (m *MockFunctionQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ListFunctionVersionsRow), args.Error(1)
}

func (m *MockFunctionQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockFunctionQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockFunctionQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockFunctionQueries) RefreshNamespaceChecksum(ctx context.Context, namespace string) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockFunctionQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockFunctionQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	args := m.Called(ctx, arg)
	return args.Get(0), args.Error(1)
}

func (m *MockFunctionQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}

// Stubs for all other db.Querier methods
func (m *MockFunctionQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockFunctionQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockFunctionQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}
func (m *MockFunctionQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockFunctionQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error {
	return nil
}
func (m *MockFunctionQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}
func (m *MockFunctionQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}
func (m *MockFunctionQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}
func (m *MockFunctionQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}
func (m *MockFunctionQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockFunctionQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}
func (m *MockFunctionQueries) DeleteNamespace(ctx context.Context, id string) error { return nil }
func (m *MockFunctionQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error {
	return nil
}
func (m *MockFunctionQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}
func (m *MockFunctionQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}
func (m *MockFunctionQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFunctionQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockFunctionQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}
func (m *MockFunctionQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}
func (m *MockFunctionQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFunctionQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}
func (m *MockFunctionQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}
func (m *MockFunctionQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error {
	return nil
}
func (m *MockFunctionQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}
func (m *MockFunctionQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}
func (m *MockFunctionQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}

func TestFunctionRepository_Create(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	function := &domain.Function{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Status:     "draft",
		Type:       "in",
		Args:       []string{"arg1"},
		Values:     []string{"val1"},
		ReturnType: "bool",
		CreatedBy:  "user",
	}

	params := db.CreateFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Status:     &function.Status,
		Type:       &function.Type,
		Args:       function.Args,
		Values:     function.Values,
		ReturnType: &function.ReturnType,
		CreatedBy:  "user",
	}

	mockQ.On("CreateFunction", ctx, params).Return(nil)
	err := repo.Create(ctx, function)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	function := &domain.Function{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Type:       "in",
		CreatedBy:  "user",
	}

	params := db.CreateFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Status:     &function.Status,
		Type:       &function.Type,
		Args:       function.Args,
		Values:     function.Values,
		ReturnType: &function.ReturnType,
		CreatedBy:  "user",
	}

	mockQ.On("CreateFunction", ctx, params).Return(errors.New("db error"))
	err := repo.Create(ctx, function)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create function")
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.GetFunctionRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("active"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.GetFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
	}

	mockQ.On("GetFunction", ctx, params).Return(row, nil)
	f, err := repo.GetByID(ctx, "ns", "f1", 1)
	assert.NoError(t, err)
	assert.Equal(t, "f1", f.FunctionID)
	assert.Equal(t, "active", f.Status)
	assert.Equal(t, "in", f.Type)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_GetByID_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	params := db.GetFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
	}

	mockQ.On("GetFunction", ctx, params).Return(nil, errors.New("not found"))
	f, err := repo.GetByID(ctx, "ns", "f1", 1)
	assert.Error(t, err)
	assert.Nil(t, f)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_GetActiveVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.GetActiveFunctionVersionRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("active"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.GetActiveFunctionVersionParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}

	mockQ.On("GetActiveFunctionVersion", ctx, params).Return(row, nil)
	f, err := repo.GetActiveVersion(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.Equal(t, "f1", f.FunctionID)
	assert.Equal(t, "active", f.Status)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_GetDraftVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.GetDraftFunctionVersionRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("draft"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: nil,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Valid: false},
	}

	params := db.GetDraftFunctionVersionParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}

	mockQ.On("GetDraftFunctionVersion", ctx, params).Return(row, nil)
	f, err := repo.GetDraftVersion(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.Equal(t, "f1", f.FunctionID)
	assert.Equal(t, "draft", f.Status)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_List(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.ListFunctionsRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("active"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("ListFunctions", ctx, "ns").Return([]*db.ListFunctionsRow{row}, nil)
	functions, err := repo.List(ctx, "ns")
	assert.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, "f1", functions[0].FunctionID)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_ListActive(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.ListActiveFunctionsRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("active"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("ListActiveFunctions", ctx, "ns").Return([]*db.ListActiveFunctionsRow{row}, nil)
	functions, err := repo.ListActive(ctx, "ns")
	assert.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, "f1", functions[0].FunctionID)
	assert.Equal(t, "active", functions[0].Status)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_ListVersions(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	row := &db.ListFunctionVersionsRow{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		Status:      strPtr("active"),
		Type:        strPtr("in"),
		Args:        []string{"arg1"},
		Values:      []string{"val1"},
		CreatedBy:   "user",
		PublishedBy: strPtr("user"),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	params := db.ListFunctionVersionsParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}

	mockQ.On("ListFunctionVersions", ctx, params).Return([]*db.ListFunctionVersionsRow{row}, nil)
	functions, err := repo.ListVersions(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, "f1", functions[0].FunctionID)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Update(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	function := &domain.Function{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Type:       "in",
		Args:       []string{"arg1"},
		Values:     []string{"val1"},
		ReturnType: "bool",
		CreatedBy:  "user",
	}

	params := db.UpdateFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
		Type:       &function.Type,
		Args:       function.Args,
		Values:     function.Values,
		ReturnType: &function.ReturnType,
		CreatedBy:  "user",
	}

	mockQ.On("UpdateFunction", ctx, params).Return(nil)
	err := repo.Update(ctx, function)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Publish(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	// Mock deactivate call
	deactivateParams := db.DeactivateFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}
	mockQ.On("DeactivateFunction", ctx, deactivateParams).Return(nil)

	// Mock publish call
	publishParams := db.PublishFunctionParams{
		Namespace:   "ns",
		FunctionID:  "f1",
		Version:     1,
		PublishedBy: strPtr("user"),
	}
	mockQ.On("PublishFunction", ctx, publishParams).Return(nil)

	// Mock refresh checksum call
	mockQ.On("RefreshNamespaceChecksum", ctx, "ns").Return(nil)

	err := repo.Publish(ctx, "ns", "f1", 1, "user")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Deactivate(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	params := db.DeactivateFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}

	mockQ.On("DeactivateFunction", ctx, params).Return(nil)
	err := repo.Deactivate(ctx, "ns", "f1")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Delete(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	params := db.DeleteFunctionParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
	}

	mockQ.On("DeleteFunction", ctx, params).Return(nil)
	err := repo.Delete(ctx, "ns", "f1", 1)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_GetMaxVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	params := db.GetMaxFunctionVersionParams{
		Namespace:  "ns",
		FunctionID: "f1",
	}

	mockQ.On("GetMaxFunctionVersion", ctx, params).Return(int64(5), nil)
	version, err := repo.GetMaxVersion(ctx, "ns", "f1")
	assert.NoError(t, err)
	assert.Equal(t, int32(5), version)
	mockQ.AssertExpectations(t)
}

func TestFunctionRepository_Exists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockFunctionQueries)
	repo := NewFunctionRepository(mockQ)

	params := db.FunctionExistsParams{
		Namespace:  "ns",
		FunctionID: "f1",
		Version:    1,
	}

	mockQ.On("FunctionExists", ctx, params).Return(true, nil)
	exists, err := repo.Exists(ctx, "ns", "f1", 1)
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)
}
