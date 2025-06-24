package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQueries is a mock implementation of db.Querier
type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) CreateNamespace(ctx context.Context, arg db.CreateNamespaceParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQueries) GetNamespace(ctx context.Context, id string) (*db.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Namespace), args.Error(1)
}

func (m *MockQueries) ListNamespaces(ctx context.Context) ([]*db.Namespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Namespace), args.Error(1)
}

func (m *MockQueries) DeleteNamespace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Add all required methods as stubs to implement db.Querier interface
func (m *MockQueries) CountFieldsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}

func (m *MockQueries) CountTerminalsByNamespace(ctx context.Context, namespace string) (int64, error) {
	return 0, nil
}

func (m *MockQueries) CreateField(ctx context.Context, arg db.CreateFieldParams) error {
	return nil
}

func (m *MockQueries) CreateFunction(ctx context.Context, arg db.CreateFunctionParams) error {
	return nil
}

func (m *MockQueries) CreateRule(ctx context.Context, arg db.CreateRuleParams) error {
	return nil
}

func (m *MockQueries) CreateTerminal(ctx context.Context, arg db.CreateTerminalParams) error {
	return nil
}

func (m *MockQueries) CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) error {
	return nil
}

func (m *MockQueries) DeactivateFunction(ctx context.Context, arg db.DeactivateFunctionParams) error {
	return nil
}

func (m *MockQueries) DeactivateRule(ctx context.Context, arg db.DeactivateRuleParams) error {
	return nil
}

func (m *MockQueries) DeactivateWorkflow(ctx context.Context, arg db.DeactivateWorkflowParams) error {
	return nil
}

func (m *MockQueries) DeleteActiveConfigChecksum(ctx context.Context, namespace string) error {
	return nil
}

func (m *MockQueries) DeleteField(ctx context.Context, arg db.DeleteFieldParams) error {
	return nil
}

func (m *MockQueries) DeleteFunction(ctx context.Context, arg db.DeleteFunctionParams) error {
	return nil
}

func (m *MockQueries) DeleteRule(ctx context.Context, arg db.DeleteRuleParams) error {
	return nil
}

func (m *MockQueries) DeleteTerminal(ctx context.Context, arg db.DeleteTerminalParams) error {
	return nil
}

func (m *MockQueries) DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error {
	return nil
}

func (m *MockQueries) FieldExists(ctx context.Context, arg db.FieldExistsParams) (bool, error) {
	return false, nil
}

func (m *MockQueries) FunctionExists(ctx context.Context, arg db.FunctionExistsParams) (bool, error) {
	return false, nil
}

func (m *MockQueries) GetActiveConfigChecksum(ctx context.Context, namespace string) (*db.ActiveConfigMetum, error) {
	return nil, nil
}

func (m *MockQueries) GetActiveFunctionVersion(ctx context.Context, arg db.GetActiveFunctionVersionParams) (*db.GetActiveFunctionVersionRow, error) {
	return nil, nil
}

func (m *MockQueries) GetActiveRuleVersion(ctx context.Context, arg db.GetActiveRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) GetActiveWorkflowVersion(ctx context.Context, arg db.GetActiveWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) GetDraftFunctionVersion(ctx context.Context, arg db.GetDraftFunctionVersionParams) (*db.GetDraftFunctionVersionRow, error) {
	return nil, nil
}

func (m *MockQueries) GetDraftRuleVersion(ctx context.Context, arg db.GetDraftRuleVersionParams) (*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) GetDraftWorkflowVersion(ctx context.Context, arg db.GetDraftWorkflowVersionParams) (*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) GetField(ctx context.Context, arg db.GetFieldParams) (*db.GetFieldRow, error) {
	return nil, nil
}

func (m *MockQueries) GetFunction(ctx context.Context, arg db.GetFunctionParams) (*db.GetFunctionRow, error) {
	return nil, nil
}

func (m *MockQueries) GetMaxFunctionVersion(ctx context.Context, arg db.GetMaxFunctionVersionParams) (interface{}, error) {
	return nil, nil
}

func (m *MockQueries) GetMaxRuleVersion(ctx context.Context, arg db.GetMaxRuleVersionParams) (interface{}, error) {
	return nil, nil
}

func (m *MockQueries) GetMaxWorkflowVersion(ctx context.Context, arg db.GetMaxWorkflowVersionParams) (interface{}, error) {
	return nil, nil
}

func (m *MockQueries) GetRule(ctx context.Context, arg db.GetRuleParams) (*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) GetTerminal(ctx context.Context, arg db.GetTerminalParams) (*db.GetTerminalRow, error) {
	return nil, nil
}

func (m *MockQueries) GetWorkflow(ctx context.Context, arg db.GetWorkflowParams) (*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) ListActiveFunctions(ctx context.Context, namespace string) ([]*db.ListActiveFunctionsRow, error) {
	return nil, nil
}

func (m *MockQueries) ListActiveRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) ListActiveWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) ListAllActiveConfigChecksums(ctx context.Context) ([]*db.ActiveConfigMetum, error) {
	return nil, nil
}

func (m *MockQueries) ListFields(ctx context.Context, namespace string) ([]*db.ListFieldsRow, error) {
	return nil, nil
}

func (m *MockQueries) ListFunctionVersions(ctx context.Context, arg db.ListFunctionVersionsParams) ([]*db.ListFunctionVersionsRow, error) {
	return nil, nil
}

func (m *MockQueries) ListFunctions(ctx context.Context, namespace string) ([]*db.ListFunctionsRow, error) {
	return nil, nil
}

func (m *MockQueries) ListRuleVersions(ctx context.Context, arg db.ListRuleVersionsParams) ([]*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) ListRules(ctx context.Context, namespace string) ([]*db.Rule, error) {
	return nil, nil
}

func (m *MockQueries) ListTerminals(ctx context.Context, namespace string) ([]*db.ListTerminalsRow, error) {
	return nil, nil
}

func (m *MockQueries) ListWorkflowVersions(ctx context.Context, arg db.ListWorkflowVersionsParams) ([]*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) ListWorkflows(ctx context.Context, namespace string) ([]*db.Workflow, error) {
	return nil, nil
}

func (m *MockQueries) PublishFunction(ctx context.Context, arg db.PublishFunctionParams) error {
	return nil
}

func (m *MockQueries) PublishRule(ctx context.Context, arg db.PublishRuleParams) error {
	return nil
}

func (m *MockQueries) PublishWorkflow(ctx context.Context, arg db.PublishWorkflowParams) error {
	return nil
}

func (m *MockQueries) RefreshNamespaceChecksum(ctx context.Context, ns string) error {
	return nil
}

func (m *MockQueries) RuleExists(ctx context.Context, arg db.RuleExistsParams) (bool, error) {
	return false, nil
}

func (m *MockQueries) TerminalExists(ctx context.Context, arg db.TerminalExistsParams) (bool, error) {
	return false, nil
}

func (m *MockQueries) UpdateField(ctx context.Context, arg db.UpdateFieldParams) error {
	return nil
}

func (m *MockQueries) UpdateFunction(ctx context.Context, arg db.UpdateFunctionParams) error {
	return nil
}

func (m *MockQueries) UpdateRule(ctx context.Context, arg db.UpdateRuleParams) error {
	return nil
}

func (m *MockQueries) UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) error {
	return nil
}

func (m *MockQueries) UpsertActiveConfigChecksum(ctx context.Context, arg db.UpsertActiveConfigChecksumParams) error {
	return nil
}

func (m *MockQueries) WorkflowExists(ctx context.Context, arg db.WorkflowExistsParams) (bool, error) {
	return false, nil
}

func TestNamespaceRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		namespace := &domain.Namespace{
			ID:          "test-ns",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		expectedParams := db.CreateNamespaceParams{
			ID:          "test-ns",
			Description: &namespace.Description,
			CreatedBy:   "user1",
		}

		mockQueries.On("CreateNamespace", ctx, expectedParams).Return(nil)

		err := repo.Create(ctx, namespace)

		assert.NoError(t, err)
		mockQueries.AssertExpectations(t)
	})

	t.Run("creation with empty description", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		namespace := &domain.Namespace{
			ID:          "test-ns-2",
			Description: "",
			CreatedBy:   "user1",
		}

		expectedParams := db.CreateNamespaceParams{
			ID:          "test-ns-2",
			Description: nil,
			CreatedBy:   "user1",
		}

		mockQueries.On("CreateNamespace", ctx, expectedParams).Return(nil)

		err := repo.Create(ctx, namespace)

		assert.NoError(t, err)
		mockQueries.AssertExpectations(t)
	})

	t.Run("creation failure", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		namespace := &domain.Namespace{
			ID:          "test-ns-3",
			Description: "Test namespace",
			CreatedBy:   "user1",
		}

		expectedParams := db.CreateNamespaceParams{
			ID:          "test-ns-3",
			Description: &namespace.Description,
			CreatedBy:   "user1",
		}

		mockQueries.On("CreateNamespace", ctx, expectedParams).Return(sql.ErrConnDone)

		err := repo.Create(ctx, namespace)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create namespace")
		mockQueries.AssertExpectations(t)
	})
}

func TestNamespaceRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		description := "Test namespace"
		createdAt := time.Now()

		dbNamespace := &db.Namespace{
			ID:          "test-ns",
			Description: &description,
			CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
			CreatedBy:   "user1",
		}

		mockQueries.On("GetNamespace", ctx, "test-ns").Return(dbNamespace, nil)

		result, err := repo.GetByID(ctx, "test-ns")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-ns", result.ID)
		assert.Equal(t, "Test namespace", result.Description)
		assert.Equal(t, createdAt, result.CreatedAt)
		assert.Equal(t, "user1", result.CreatedBy)
		mockQueries.AssertExpectations(t)
	})

	t.Run("namespace not found", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("GetNamespace", ctx, "non-existent").Return((*db.Namespace)(nil), sql.ErrNoRows)

		result, err := repo.GetByID(ctx, "non-existent")

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, result)
		mockQueries.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("GetNamespace", ctx, "error-ns").Return((*db.Namespace)(nil), sql.ErrConnDone)

		result, err := repo.GetByID(ctx, "error-ns")

		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.Nil(t, result)
		mockQueries.AssertExpectations(t)
	})

	t.Run("namespace with nil description", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		createdAt := time.Now()

		dbNamespace := &db.Namespace{
			ID:          "test-ns-nil",
			Description: nil,
			CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
			CreatedBy:   "user1",
		}

		mockQueries.On("GetNamespace", ctx, "test-ns-nil").Return(dbNamespace, nil)

		result, err := repo.GetByID(ctx, "test-ns-nil")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-ns-nil", result.ID)
		assert.Equal(t, "", result.Description)
		assert.Equal(t, createdAt, result.CreatedAt)
		assert.Equal(t, "user1", result.CreatedBy)
		mockQueries.AssertExpectations(t)
	})
}

func TestNamespaceRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		description1 := "Namespace 1"
		description2 := "Namespace 2"
		createdAt := time.Now()

		dbNamespaces := []*db.Namespace{
			{
				ID:          "ns1",
				Description: &description1,
				CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
				CreatedBy:   "user1",
			},
			{
				ID:          "ns2",
				Description: &description2,
				CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
				CreatedBy:   "user2",
			},
		}

		mockQueries.On("ListNamespaces", ctx).Return(dbNamespaces, nil)

		result, err := repo.List(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ns1", result[0].ID)
		assert.Equal(t, "Namespace 1", result[0].Description)
		assert.Equal(t, "ns2", result[1].ID)
		assert.Equal(t, "Namespace 2", result[1].Description)
		mockQueries.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("ListNamespaces", ctx).Return([]*db.Namespace{}, nil)

		result, err := repo.List(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockQueries.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("ListNamespaces", ctx).Return([]*db.Namespace{}, sql.ErrConnDone)

		result, err := repo.List(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list namespaces")
		assert.Nil(t, result)
		mockQueries.AssertExpectations(t)
	})
}

func TestNamespaceRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("DeleteNamespace", ctx, "test-ns").Return(nil)

		err := repo.Delete(ctx, "test-ns")

		assert.NoError(t, err)
		mockQueries.AssertExpectations(t)
	})

	t.Run("deletion failure", func(t *testing.T) {
		mockQueries := new(MockQueries)
		repo := NewNamespaceRepository(mockQueries)

		mockQueries.On("DeleteNamespace", ctx, "error-ns").Return(sql.ErrConnDone)

		err := repo.Delete(ctx, "error-ns")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete namespace")
		mockQueries.AssertExpectations(t)
	})
}
