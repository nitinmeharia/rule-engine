package domain

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Namespace represents a logical grouping of rules, workflows, and configurations
type Namespace struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

// Validate validates the namespace data
func (n *Namespace) Validate() error {
	if n == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(n.ID) == "" {
		return ErrInvalidNamespaceID
	}

	if len(n.ID) > 50 {
		return ErrInvalidNamespaceID
	}

	// Check if ID contains only alphanumeric characters, hyphens, and underscores
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(n.ID) {
		return ErrInvalidNamespaceID
	}

	// Cannot start or end with hyphen or underscore
	if strings.HasPrefix(n.ID, "-") || strings.HasPrefix(n.ID, "_") ||
		strings.HasSuffix(n.ID, "-") || strings.HasSuffix(n.ID, "_") {
		return ErrInvalidNamespaceID
	}

	if strings.TrimSpace(n.CreatedBy) == "" {
		return ErrValidationError
	}

	if len(n.Description) > 500 {
		return ErrInvalidDescription
	}

	return nil
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

// Validate validates the field data
func (f *Field) Validate() error {
	if f == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(f.FieldID) == "" {
		return ErrInvalidFieldID
	}

	// Validate field type
	validTypes := map[string]struct{}{
		FieldTypeString:  {},
		FieldTypeNumber:  {},
		FieldTypeBoolean: {},
		FieldTypeDate:    {},
	}

	if _, valid := validTypes[f.Type]; !valid {
		return ErrInvalidFieldType
	}

	// Description can be nil/empty as per documentation
	return nil
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

// Validate validates the function data
func (f *Function) Validate() error {
	if f == nil {
		return ErrValidationError
	}

	if f.FunctionID == "" {
		return ErrInvalidFunctionID
	}

	if f.Type == "" {
		return ErrInvalidFunctionType
	}

	// Validate function type
	validTypes := []string{FunctionTypeMax, FunctionTypeSum, FunctionTypeAvg, FunctionTypeIn}
	isValidType := false
	for _, t := range validTypes {
		if f.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return ErrInvalidFunctionType
	}

	// Validate based on function type
	switch f.Type {
	case FunctionTypeMax, FunctionTypeSum, FunctionTypeAvg:
		if len(f.Args) == 0 {
			return ErrInvalidFunctionArgs
		}
		if len(f.Values) > 0 {
			return ErrInvalidFunctionArgs
		}
	case FunctionTypeIn:
		if len(f.Values) == 0 {
			return ErrInvalidFunctionArgs
		}
		if len(f.Args) > 0 {
			return ErrInvalidFunctionArgs
		}
	}

	return nil
}

// ComputeReturnType determines the return type based on function type
func (f *Function) ComputeReturnType() string {
	switch f.Type {
	case FunctionTypeMax, FunctionTypeSum, FunctionTypeAvg:
		return FunctionReturnTypeNumber
	case FunctionTypeIn:
		return FunctionReturnTypeBool
	default:
		return ""
	}
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

// Validate validates the rule data
func (r *Rule) Validate() error {
	if r == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(r.RuleID) == "" {
		return ErrInvalidRuleID
	}

	if strings.TrimSpace(r.Logic) == "" {
		return ErrInvalidRuleLogic
	}

	// Validate logic type
	validLogic := []string{LogicAND, LogicOR}
	isValidLogic := false
	for _, l := range validLogic {
		if r.Logic == l {
			isValidLogic = true
			break
		}
	}
	if !isValidLogic {
		return ErrInvalidRuleLogic
	}

	return nil
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

// Validate validates the workflow data
func (w *Workflow) Validate() error {
	if w == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(w.WorkflowID) == "" {
		return ErrInvalidWorkflowID
	}

	if strings.TrimSpace(w.StartAt) == "" {
		return ErrInvalidWorkflowStartAt
	}

	return nil
}

// Terminal represents a terminal node in workflows
type Terminal struct {
	Namespace  string    `json:"namespace"`
	TerminalID string    `json:"terminalId"`
	CreatedAt  time.Time `json:"createdAt"`
	CreatedBy  string    `json:"createdBy"`
}

// Validate validates the terminal data
func (t *Terminal) Validate() error {
	if t == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(t.TerminalID) == "" {
		return ErrInvalidTerminalID
	}

	return nil
}

// ActiveConfigMeta represents cache checksum metadata
type ActiveConfigMeta struct {
	Namespace string    `json:"namespace"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Validate validates the active config meta data
func (a *ActiveConfigMeta) Validate() error {
	if a == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(a.Namespace) == "" {
		return ErrInvalidNamespaceID
	}

	if strings.TrimSpace(a.Checksum) == "" {
		return ErrInvalidChecksum
	}

	return nil
}

// ExecutionRequest represents an execution request payload
type ExecutionRequest struct {
	Namespace  string                 `json:"namespace"`
	RuleID     *string                `json:"ruleId,omitempty"`
	WorkflowID *string                `json:"workflowId,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Trace      bool                   `json:"trace,omitempty"`
}

// Validate validates the execution request data
func (e *ExecutionRequest) Validate() error {
	if e == nil {
		return ErrValidationError
	}

	if strings.TrimSpace(e.Namespace) == "" {
		return ErrInvalidNamespaceID
	}

	if e.RuleID == nil && e.WorkflowID == nil {
		return ErrInvalidExecutionRequest
	}

	if e.RuleID != nil && e.WorkflowID != nil {
		return ErrInvalidExecutionRequest
	}

	if e.Data == nil {
		return ErrInvalidExecutionData
	}

	return nil
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
