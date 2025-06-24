package domain

import (
	"encoding/json"
	"time"
)

// Namespace represents a logical grouping of rules, workflows, and configurations
type Namespace struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

// Field represents a typed field definition used in rule conditions
type Field struct {
	Namespace   string    `json:"namespace"`
	FieldID     string    `json:"fieldId"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

// Function represents a built-in function available for rule evaluation
type Function struct {
	Namespace   string     `json:"namespace"`
	FunctionID  string     `json:"functionId"`
	Version     int32      `json:"version"`
	Status      string     `json:"status"`
	Type        string     `json:"type"`
	Args        []string   `json:"args"`
	Values      []string   `json:"values"`
	ReturnType  string     `json:"returnType"`
	CreatedBy   string     `json:"createdBy"`
	PublishedBy *string    `json:"publishedBy,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

// Rule represents a business rule with conditions and logic
type Rule struct {
	Namespace   string          `json:"namespace"`
	RuleID      string          `json:"ruleId"`
	Version     int32           `json:"version"`
	Status      string          `json:"status"`
	Logic       string          `json:"logic"`
	Conditions  json.RawMessage `json:"conditions"`
	CreatedBy   string          `json:"createdBy"`
	PublishedBy *string         `json:"publishedBy,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	PublishedAt *time.Time      `json:"publishedAt,omitempty"`
}

// Workflow represents a workflow definition with steps
type Workflow struct {
	Namespace   string          `json:"namespace"`
	WorkflowID  string          `json:"workflowId"`
	Version     int32           `json:"version"`
	Status      string          `json:"status"`
	StartAt     string          `json:"startAt"`
	Steps       json.RawMessage `json:"steps"`
	CreatedBy   string          `json:"createdBy"`
	PublishedBy *string         `json:"publishedBy,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	PublishedAt *time.Time      `json:"publishedAt,omitempty"`
}

// Terminal represents a terminal node in workflows
type Terminal struct {
	Namespace  string    `json:"namespace"`
	TerminalID string    `json:"terminalId"`
	CreatedAt  time.Time `json:"createdAt"`
	CreatedBy  string    `json:"createdBy"`
}

// ActiveConfigMeta represents cache checksum metadata
type ActiveConfigMeta struct {
	Namespace string    `json:"namespace"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ExecutionRequest represents an execution request payload
type ExecutionRequest struct {
	Namespace  string                 `json:"namespace"`
	RuleID     *string                `json:"ruleId,omitempty"`
	WorkflowID *string                `json:"workflowId,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Trace      bool                   `json:"trace,omitempty"`
}

// ExecutionResponse represents an execution result
type ExecutionResponse struct {
	Result    interface{}            `json:"result"`
	Trace     *ExecutionTrace        `json:"trace,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ExecutionTrace represents execution trace information
type ExecutionTrace struct {
	Steps    []TraceStep `json:"steps"`
	Duration string      `json:"duration"`
	Version  string      `json:"version"`
}

// TraceStep represents a single step in execution trace
type TraceStep struct {
	Type     string      `json:"type"`
	ID       string      `json:"id"`
	Input    interface{} `json:"input,omitempty"`
	Output   interface{} `json:"output,omitempty"`
	Duration string      `json:"duration"`
	Error    *string     `json:"error,omitempty"`
}

// Status constants
const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Logic constants
const (
	LogicAND = "AND"
	LogicOR  = "OR"
)

// Field type constants
const (
	FieldTypeString  = "string"
	FieldTypeNumber  = "number"
	FieldTypeBoolean = "boolean"
	FieldTypeDate    = "date"
)

// Function type constants
const (
	FunctionTypeMax = "max"
	FunctionTypeSum = "sum"
	FunctionTypeAvg = "avg"
	FunctionTypeIn  = "in"
)

// Function return type constants
const (
	FunctionReturnTypeNumber = "number"
	FunctionReturnTypeBool   = "bool"
)
