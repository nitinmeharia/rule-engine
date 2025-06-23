package domain

import (
	"context"
)

// NamespaceRepository defines operations for namespace management
type NamespaceRepository interface {
	Create(ctx context.Context, namespace *Namespace) error
	GetByID(ctx context.Context, id string) (*Namespace, error)
	List(ctx context.Context) ([]*Namespace, error)
	Delete(ctx context.Context, id string) error
}

// FieldRepository defines operations for field management
type FieldRepository interface {
	Create(ctx context.Context, field *Field) error
	GetByID(ctx context.Context, namespace, fieldID string) (*Field, error)
	List(ctx context.Context, namespace string) ([]*Field, error)
	Update(ctx context.Context, field *Field) error
	Delete(ctx context.Context, namespace, fieldID string) error
	Exists(ctx context.Context, namespace, fieldID string) (bool, error)
	CountByNamespace(ctx context.Context, namespace string) (int64, error)
}

// FunctionRepository defines operations for function management
type FunctionRepository interface {
	Create(ctx context.Context, function *Function) error
	GetByID(ctx context.Context, namespace, functionID string, version int32) (*Function, error)
	GetActiveVersion(ctx context.Context, namespace, functionID string) (*Function, error)
	GetDraftVersion(ctx context.Context, namespace, functionID string) (*Function, error)
	List(ctx context.Context, namespace string) ([]*Function, error)
	ListActive(ctx context.Context, namespace string) ([]*Function, error)
	ListVersions(ctx context.Context, namespace, functionID string) ([]*Function, error)
	Update(ctx context.Context, function *Function) error
	Publish(ctx context.Context, namespace, functionID string, version int32, publishedBy string) error
	Deactivate(ctx context.Context, namespace, functionID string) error
	Delete(ctx context.Context, namespace, functionID string, version int32) error
	GetMaxVersion(ctx context.Context, namespace, functionID string) (int32, error)
	Exists(ctx context.Context, namespace, functionID string, version int32) (bool, error)
}

// RuleRepository defines operations for rule management
type RuleRepository interface {
	Create(ctx context.Context, rule *Rule) error
	GetByID(ctx context.Context, namespace, ruleID string, version int32) (*Rule, error)
	GetActiveVersion(ctx context.Context, namespace, ruleID string) (*Rule, error)
	GetDraftVersion(ctx context.Context, namespace, ruleID string) (*Rule, error)
	List(ctx context.Context, namespace string) ([]*Rule, error)
	ListActive(ctx context.Context, namespace string) ([]*Rule, error)
	ListVersions(ctx context.Context, namespace, ruleID string) ([]*Rule, error)
	Update(ctx context.Context, rule *Rule) error
	Publish(ctx context.Context, namespace, ruleID string, version int32, publishedBy string) error
	Deactivate(ctx context.Context, namespace, ruleID string) error
	Delete(ctx context.Context, namespace, ruleID string, version int32) error
	GetMaxVersion(ctx context.Context, namespace, ruleID string) (int32, error)
	Exists(ctx context.Context, namespace, ruleID string, version int32) (bool, error)
}

// WorkflowRepository defines operations for workflow management
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *Workflow) error
	GetByID(ctx context.Context, namespace, workflowID string, version int32) (*Workflow, error)
	GetActiveVersion(ctx context.Context, namespace, workflowID string) (*Workflow, error)
	GetDraftVersion(ctx context.Context, namespace, workflowID string) (*Workflow, error)
	List(ctx context.Context, namespace string) ([]*Workflow, error)
	ListActive(ctx context.Context, namespace string) ([]*Workflow, error)
	ListVersions(ctx context.Context, namespace, workflowID string) ([]*Workflow, error)
	Update(ctx context.Context, workflow *Workflow) error
	Publish(ctx context.Context, namespace, workflowID string, version int32, publishedBy string) error
	Deactivate(ctx context.Context, namespace, workflowID string) error
	Delete(ctx context.Context, namespace, workflowID string, version int32) error
	GetMaxVersion(ctx context.Context, namespace, workflowID string) (int32, error)
	Exists(ctx context.Context, namespace, workflowID string, version int32) (bool, error)
}

// TerminalRepository defines operations for terminal management
type TerminalRepository interface {
	Create(ctx context.Context, terminal *Terminal) error
	GetByID(ctx context.Context, namespace, terminalID string) (*Terminal, error)
	List(ctx context.Context, namespace string) ([]*Terminal, error)
	Delete(ctx context.Context, namespace, terminalID string) error
	Exists(ctx context.Context, namespace, terminalID string) (bool, error)
	CountByNamespace(ctx context.Context, namespace string) (int64, error)
}

// CacheRepository defines operations for cache management
type CacheRepository interface {
	GetActiveConfigChecksum(ctx context.Context, namespace string) (*ActiveConfigMeta, error)
	UpsertActiveConfigChecksum(ctx context.Context, namespace, checksum string) error
	RefreshNamespaceChecksum(ctx context.Context, namespace string) error
	ListAllActiveConfigChecksums(ctx context.Context) ([]*ActiveConfigMeta, error)
	DeleteActiveConfigChecksum(ctx context.Context, namespace string) error
}
