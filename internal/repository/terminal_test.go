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

type MockTerminalQueries struct {
	mock.Mock
}

// Terminal-specific methods (properly implemented)
func (m *MockTerminalQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTerminalQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GetTerminalRow), args.Error(1)
}

func (m *MockTerminalQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ListTerminalsRow), args.Error(1)
}

func (m *MockTerminalQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTerminalQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}

func (m *MockTerminalQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int64), args.Error(1)
}

// Stubs for all other db.Querier interface methods
func (m *MockTerminalQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockTerminalQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}
func (m *MockTerminalQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}
func (m *MockTerminalQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockTerminalQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error {
	return nil
}
func (m *MockTerminalQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}
func (m *MockTerminalQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}
func (m *MockTerminalQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}
func (m *MockTerminalQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}
func (m *MockTerminalQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockTerminalQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}
func (m *MockTerminalQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}
func (m *MockTerminalQueries) DeleteNamespace(ctx context.Context, id string) error { return nil }
func (m *MockTerminalQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error {
	return nil
}
func (m *MockTerminalQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}
func (m *MockTerminalQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}
func (m *MockTerminalQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}
func (m *MockTerminalQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockTerminalQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}
func (m *MockTerminalQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}
func (m *MockTerminalQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}
func (m *MockTerminalQueries) RefreshNamespaceChecksum(ctx context.Context, ns string) error {
	return nil
}
func (m *MockTerminalQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}
func (m *MockTerminalQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}
func (m *MockTerminalQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}
func (m *MockTerminalQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error {
	return nil
}
func (m *MockTerminalQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}
func (m *MockTerminalQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}
func (m *MockTerminalQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}

func TestTerminalRepository_Create(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	terminal := &domain.Terminal{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
		CreatedBy:  "test-user",
	}

	params := db.CreateTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
		CreatedBy:  "test-user",
	}

	mockQ.On("CreateTerminal", ctx, params).Return(nil)

	err := repo.Create(ctx, terminal)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	terminal := &domain.Terminal{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
		CreatedBy:  "test-user",
	}

	params := db.CreateTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
		CreatedBy:  "test-user",
	}

	expectedErr := errors.New("database error")
	mockQ.On("CreateTerminal", ctx, params).Return(expectedErr)

	err := repo.Create(ctx, terminal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create terminal")
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	now := time.Now()
	dbTerminal := &db.GetTerminalRow{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
		CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		CreatedBy:  "test-user",
	}

	params := db.GetTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	mockQ.On("GetTerminal", ctx, params).Return(dbTerminal, nil)

	result, err := repo.GetByID(ctx, "test-ns", "test-terminal")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-ns", result.Namespace)
	assert.Equal(t, "test-terminal", result.TerminalID)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, "test-user", result.CreatedBy)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_GetByID_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	params := db.GetTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	expectedErr := errors.New("not found")
	mockQ.On("GetTerminal", ctx, params).Return(nil, expectedErr)

	result, err := repo.GetByID(ctx, "test-ns", "test-terminal")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_List(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	now := time.Now()
	dbTerminals := []*db.ListTerminalsRow{
		{
			Namespace:  "test-ns",
			TerminalID: "terminal-1",
			CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			CreatedBy:  "user-1",
		},
		{
			Namespace:  "test-ns",
			TerminalID: "terminal-2",
			CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			CreatedBy:  "user-2",
		},
	}

	mockQ.On("ListTerminals", ctx, "test-ns").Return(dbTerminals, nil)

	result, err := repo.List(ctx, "test-ns")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "terminal-1", result[0].TerminalID)
	assert.Equal(t, "terminal-2", result[1].TerminalID)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_List_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	expectedErr := errors.New("database error")
	mockQ.On("ListTerminals", ctx, "test-ns").Return(nil, expectedErr)

	result, err := repo.List(ctx, "test-ns")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list terminals")
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_Delete(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	params := db.DeleteTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	mockQ.On("DeleteTerminal", ctx, params).Return(nil)

	err := repo.Delete(ctx, "test-ns", "test-terminal")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_Delete_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	params := db.DeleteTerminalParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	expectedErr := errors.New("database error")
	mockQ.On("DeleteTerminal", ctx, params).Return(expectedErr)

	err := repo.Delete(ctx, "test-ns", "test-terminal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete terminal")
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_Exists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	params := db.TerminalExistsParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	mockQ.On("TerminalExists", ctx, params).Return(true, nil)

	exists, err := repo.Exists(ctx, "test-ns", "test-terminal")
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_Exists_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	params := db.TerminalExistsParams{
		Namespace:  "test-ns",
		TerminalID: "test-terminal",
	}

	expectedErr := errors.New("database error")
	mockQ.On("TerminalExists", ctx, params).Return(false, expectedErr)

	exists, err := repo.Exists(ctx, "test-ns", "test-terminal")
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "failed to check terminal existence")
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_CountByNamespace(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	mockQ.On("CountTerminalsByNamespace", ctx, "test-ns").Return(int64(5), nil)

	count, err := repo.CountByNamespace(ctx, "test-ns")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	mockQ.AssertExpectations(t)
}

func TestTerminalRepository_CountByNamespace_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockTerminalQueries)
	repo := NewTerminalRepository(mockQ)

	expectedErr := errors.New("database error")
	mockQ.On("CountTerminalsByNamespace", ctx, "test-ns").Return(int64(0), expectedErr)

	count, err := repo.CountByNamespace(ctx, "test-ns")
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "failed to count terminals")
	mockQ.AssertExpectations(t)
}
