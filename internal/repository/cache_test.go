package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rule-engine/internal/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCacheQueries struct {
	mock.Mock
}

// Implement all db.Querier methods
func (m *MockCacheQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.ActiveConfigMetum), args.Error(1)
}

func (m *MockCacheQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockCacheQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *MockCacheQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.ActiveConfigMetum), args.Error(1)
}

// Stubs for all other db.Querier methods
func (m *MockCacheQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockCacheQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockCacheQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}
func (m *MockCacheQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}
func (m *MockCacheQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockCacheQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error { return nil }
func (m *MockCacheQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}
func (m *MockCacheQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}
func (m *MockCacheQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}
func (m *MockCacheQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}
func (m *MockCacheQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}
func (m *MockCacheQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}
func (m *MockCacheQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}
func (m *MockCacheQueries) DeleteNamespace(ctx context.Context, id string) error          { return nil }
func (m *MockCacheQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error { return nil }
func (m *MockCacheQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}
func (m *MockCacheQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}
func (m *MockCacheQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}
func (m *MockCacheQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}
func (m *MockCacheQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}
func (m *MockCacheQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}
func (m *MockCacheQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}
func (m *MockCacheQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}
func (m *MockCacheQueries) RefreshNamespaceChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockCacheQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}
func (m *MockCacheQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}
func (m *MockCacheQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}
func (m *MockCacheQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}
func (m *MockCacheQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error { return nil }
func (m *MockCacheQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}
func (m *MockCacheQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}

func TestCacheRepository_GetActiveConfigChecksum(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	meta := &db.ActiveConfigMetum{
		Namespace: "ns",
		Checksum:  "abc123",
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("GetActiveConfigChecksum", ctx, "ns").Return(meta, nil)
	result, err := repo.GetActiveConfigChecksum(ctx, "ns")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", result.Checksum)
	assert.Equal(t, "ns", result.Namespace)
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_GetActiveConfigChecksum_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	mockQ.On("GetActiveConfigChecksum", ctx, "ns").Return(nil, errors.New("not found"))
	result, err := repo.GetActiveConfigChecksum(ctx, "ns")
	assert.Error(t, err)
	assert.Nil(t, result)
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_UpsertActiveConfigChecksum(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	params := db.UpsertActiveConfigChecksumParams{
		Namespace: "ns",
		Checksum:  "abc123",
	}

	mockQ.On("UpsertActiveConfigChecksum", ctx, params).Return(nil)
	err := repo.UpsertActiveConfigChecksum(ctx, "ns", "abc123")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_UpsertActiveConfigChecksum_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	params := db.UpsertActiveConfigChecksumParams{
		Namespace: "ns",
		Checksum:  "abc123",
	}

	mockQ.On("UpsertActiveConfigChecksum", ctx, params).Return(errors.New("db error"))
	err := repo.UpsertActiveConfigChecksum(ctx, "ns", "abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert active config checksum")
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_DeleteActiveConfigChecksum(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	mockQ.On("DeleteActiveConfigChecksum", ctx, "ns").Return(nil)
	err := repo.DeleteActiveConfigChecksum(ctx, "ns")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_DeleteActiveConfigChecksum_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	mockQ.On("DeleteActiveConfigChecksum", ctx, "ns").Return(errors.New("db error"))
	err := repo.DeleteActiveConfigChecksum(ctx, "ns")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete active config checksum")
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_ListAllActiveConfigChecksums(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	meta1 := &db.ActiveConfigMetum{
		Namespace: "ns1",
		Checksum:  "abc123",
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	meta2 := &db.ActiveConfigMetum{
		Namespace: "ns2",
		Checksum:  "def456",
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	mockQ.On("ListAllActiveConfigChecksums", ctx).Return([]*db.ActiveConfigMetum{meta1, meta2}, nil)
	results, err := repo.ListAllActiveConfigChecksums(ctx)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ns1", results[0].Namespace)
	assert.Equal(t, "abc123", results[0].Checksum)
	assert.Equal(t, "ns2", results[1].Namespace)
	assert.Equal(t, "def456", results[1].Checksum)
	mockQ.AssertExpectations(t)
}

func TestCacheRepository_ListAllActiveConfigChecksums_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockCacheQueries)
	repo := NewCacheRepository(mockQ)

	mockQ.On("ListAllActiveConfigChecksums", ctx).Return(nil, errors.New("db error"))
	results, err := repo.ListAllActiveConfigChecksums(ctx)
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to list all active config checksums")
	mockQ.AssertExpectations(t)
}
