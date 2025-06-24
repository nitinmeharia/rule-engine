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

type MockWorkflowQueries struct {
	mock.Mock
}

// Workflow-specific methods (properly implemented)
func (m *MockWorkflowQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockWorkflowQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Workflow), args.Error(1)
}

func (m *MockWorkflowQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockWorkflowQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockWorkflowQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockWorkflowQueries) RefreshNamespaceChecksum(ctx context.Context, ns string) error {
	args := m.Called(ctx, ns)
	return args.Error(0)
}

func (m *MockWorkflowQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockWorkflowQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	args := m.Called(ctx, arg)
	return args.Get(0), args.Error(1)
}

func (m *MockWorkflowQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}

// Stubs for all other db.Querier interface methods
func (m *MockWorkflowQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockWorkflowQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}
func (m *MockWorkflowQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}
func (m *MockWorkflowQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}
func (m *MockWorkflowQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	return nil
}
func (m *MockWorkflowQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error {
	return nil
}
func (m *MockWorkflowQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}
func (m *MockWorkflowQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeleteNamespace(ctx context.Context, id string) error { return nil }
func (m *MockWorkflowQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error {
	return nil
}
func (m *MockWorkflowQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}
func (m *MockWorkflowQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}
func (m *MockWorkflowQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}
func (m *MockWorkflowQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}
func (m *MockWorkflowQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}
func (m *MockWorkflowQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}
func (m *MockWorkflowQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}
func (m *MockWorkflowQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}
func (m *MockWorkflowQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}
func (m *MockWorkflowQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}
func (m *MockWorkflowQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error {
	return nil
}
func (m *MockWorkflowQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}

func TestWorkflowRepository_Create(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	workflow := &domain.Workflow{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		Status:     "draft",
		StartAt:    "2023-01-01T00:00:00Z",
		Steps:      []byte(`[]`),
		CreatedBy:  "test-user",
	}

	params := db.CreateWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		Status:     &workflow.Status,
		StartAt:    workflow.StartAt,
		Steps:      workflow.Steps,
		CreatedBy:  "test-user",
	}

	mockQ.On("CreateWorkflow", ctx, params).Return(nil)

	err := repo.Create(ctx, workflow)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	workflow := &domain.Workflow{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		Status:     "draft",
		StartAt:    "2023-01-01T00:00:00Z",
		Steps:      []byte(`[]`),
		CreatedBy:  "test-user",
	}

	params := db.CreateWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		Status:     &workflow.Status,
		StartAt:    workflow.StartAt,
		Steps:      workflow.Steps,
		CreatedBy:  "test-user",
	}

	expectedErr := errors.New("database error")
	mockQ.On("CreateWorkflow", ctx, params).Return(expectedErr)

	err := repo.Create(ctx, workflow)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflow := &db.Workflow{
		Namespace:   "test-ns",
		WorkflowID:  "test-workflow",
		Version:     1,
		Status:      &[]string{"active"}[0],
		StartAt:     "2023-01-01T00:00:00Z",
		Steps:       []byte(`[]`),
		CreatedBy:   "test-user",
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	params := db.GetWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
	}

	mockQ.On("GetWorkflow", ctx, params).Return(dbWorkflow, nil)

	result, err := repo.GetByID(ctx, "test-ns", "test-workflow", 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-ns", result.Namespace)
	assert.Equal(t, "test-workflow", result.WorkflowID)
	assert.Equal(t, int32(1), result.Version)
	assert.Equal(t, "active", result.Status)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_GetByID_Error(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	params := db.GetWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
	}

	expectedErr := errors.New("not found")
	mockQ.On("GetWorkflow", ctx, params).Return(nil, expectedErr)

	result, err := repo.GetByID(ctx, "test-ns", "test-workflow", 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_GetActiveVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflow := &db.Workflow{
		Namespace:   "test-ns",
		WorkflowID:  "test-workflow",
		Version:     2,
		Status:      &[]string{"active"}[0],
		StartAt:     "2023-01-01T00:00:00Z",
		Steps:       []byte(`[]`),
		CreatedBy:   "test-user",
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	params := db.GetActiveWorkflowVersionParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}

	mockQ.On("GetActiveWorkflowVersion", ctx, params).Return(dbWorkflow, nil)

	result, err := repo.GetActiveVersion(ctx, "test-ns", "test-workflow")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-ns", result.Namespace)
	assert.Equal(t, "test-workflow", result.WorkflowID)
	assert.Equal(t, int32(2), result.Version)
	assert.Equal(t, "active", result.Status)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_GetDraftVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflow := &db.Workflow{
		Namespace:   "test-ns",
		WorkflowID:  "test-workflow",
		Version:     3,
		Status:      &[]string{"draft"}[0],
		StartAt:     "2023-01-01T00:00:00Z",
		Steps:       []byte(`[]`),
		CreatedBy:   "test-user",
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	params := db.GetDraftWorkflowVersionParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}

	mockQ.On("GetDraftWorkflowVersion", ctx, params).Return(dbWorkflow, nil)

	result, err := repo.GetDraftVersion(ctx, "test-ns", "test-workflow")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-ns", result.Namespace)
	assert.Equal(t, "test-workflow", result.WorkflowID)
	assert.Equal(t, int32(3), result.Version)
	assert.Equal(t, "draft", result.Status)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_List(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflows := []*db.Workflow{
		{
			Namespace:   "test-ns",
			WorkflowID:  "workflow-1",
			Version:     1,
			Status:      &[]string{"active"}[0],
			StartAt:     "2023-01-01T00:00:00Z",
			Steps:       []byte(`[]`),
			CreatedBy:   "user-1",
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			Namespace:   "test-ns",
			WorkflowID:  "workflow-2",
			Version:     1,
			Status:      &[]string{"draft"}[0],
			StartAt:     "2023-01-01T00:00:00Z",
			Steps:       []byte(`[]`),
			CreatedBy:   "user-2",
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockQ.On("ListWorkflows", ctx, "test-ns").Return(dbWorkflows, nil)

	result, err := repo.List(ctx, "test-ns")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "workflow-1", result[0].WorkflowID)
	assert.Equal(t, "workflow-2", result[1].WorkflowID)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_ListActive(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflows := []*db.Workflow{
		{
			Namespace:   "test-ns",
			WorkflowID:  "workflow-1",
			Version:     1,
			Status:      &[]string{"active"}[0],
			StartAt:     "2023-01-01T00:00:00Z",
			Steps:       []byte(`[]`),
			CreatedBy:   "user-1",
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockQ.On("ListActiveWorkflows", ctx, "test-ns").Return(dbWorkflows, nil)

	result, err := repo.ListActive(ctx, "test-ns")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "workflow-1", result[0].WorkflowID)
	assert.Equal(t, "active", result[0].Status)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_ListVersions(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	now := time.Now()
	dbWorkflows := []*db.Workflow{
		{
			Namespace:   "test-ns",
			WorkflowID:  "test-workflow",
			Version:     1,
			Status:      &[]string{"active"}[0],
			StartAt:     "2023-01-01T00:00:00Z",
			Steps:       []byte(`[]`),
			CreatedBy:   "user-1",
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			Namespace:   "test-ns",
			WorkflowID:  "test-workflow",
			Version:     2,
			Status:      &[]string{"draft"}[0],
			StartAt:     "2023-01-01T00:00:00Z",
			Steps:       []byte(`[]`),
			CreatedBy:   "user-2",
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	params := db.ListWorkflowVersionsParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}

	mockQ.On("ListWorkflowVersions", ctx, params).Return(dbWorkflows, nil)

	result, err := repo.ListVersions(ctx, "test-ns", "test-workflow")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int32(1), result[0].Version)
	assert.Equal(t, int32(2), result[1].Version)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Update(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	workflow := &domain.Workflow{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		Status:     "draft",
		StartAt:    "2023-01-01T00:00:00Z",
		Steps:      []byte(`[]`),
		CreatedBy:  "test-user",
	}

	params := db.UpdateWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
		StartAt:    workflow.StartAt,
		Steps:      workflow.Steps,
		CreatedBy:  "test-user",
	}

	mockQ.On("UpdateWorkflow", ctx, params).Return(nil)

	err := repo.Update(ctx, workflow)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Publish(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	// Mock DeactivateWorkflow call
	deactivateParams := db.DeactivateWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}
	mockQ.On("DeactivateWorkflow", ctx, deactivateParams).Return(nil)

	// Mock PublishWorkflow call
	publishParams := db.PublishWorkflowParams{
		Namespace:   "test-ns",
		WorkflowID:  "test-workflow",
		Version:     2,
		PublishedBy: &[]string{"test-user"}[0],
	}
	mockQ.On("PublishWorkflow", ctx, publishParams).Return(nil)

	// Mock RefreshNamespaceChecksum call
	mockQ.On("RefreshNamespaceChecksum", ctx, "test-ns").Return(nil)

	err := repo.Publish(ctx, "test-ns", "test-workflow", 2, "test-user")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Deactivate(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	params := db.DeactivateWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}

	mockQ.On("DeactivateWorkflow", ctx, params).Return(nil)

	err := repo.Deactivate(ctx, "test-ns", "test-workflow")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Delete(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	params := db.DeleteWorkflowParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
	}

	mockQ.On("DeleteWorkflow", ctx, params).Return(nil)

	err := repo.Delete(ctx, "test-ns", "test-workflow", 1)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_GetMaxVersion(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	params := db.GetMaxWorkflowVersionParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
	}

	mockQ.On("GetMaxWorkflowVersion", ctx, params).Return(int32(5), nil)

	result, err := repo.GetMaxVersion(ctx, "test-ns", "test-workflow")
	assert.NoError(t, err)
	assert.Equal(t, int32(5), result)
	mockQ.AssertExpectations(t)
}

func TestWorkflowRepository_Exists(t *testing.T) {
	ctx := context.Background()
	mockQ := new(MockWorkflowQueries)
	repo := NewWorkflowRepository(mockQ)

	params := db.WorkflowExistsParams{
		Namespace:  "test-ns",
		WorkflowID: "test-workflow",
		Version:    1,
	}

	mockQ.On("WorkflowExists", ctx, params).Return(true, nil)

	exists, err := repo.Exists(ctx, "test-ns", "test-workflow", 1)
	assert.NoError(t, err)
	assert.True(t, exists)
	mockQ.AssertExpectations(t)
}
