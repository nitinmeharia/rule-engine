package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rule-engine/internal/domain"
)

// Engine represents the rule execution engine
type Engine struct {
	cache           *NamespaceCache
	cacheRepo       domain.CacheRepository
	ruleRepo        domain.RuleRepository
	workflowRepo    domain.WorkflowRepository
	fieldRepo       domain.FieldRepository
	functionRepo    domain.FunctionRepository
	terminalRepo    domain.TerminalRepository
	refreshMutex    sync.RWMutex
	lastRefresh     map[string]time.Time
	refreshInterval time.Duration
}

// NamespaceCache holds cached configuration for a namespace
type NamespaceCache struct {
	mutex     sync.RWMutex
	data      map[string]*NamespaceConfig
	checksums map[string]string
}

// NamespaceConfig holds all active configuration for a namespace
type NamespaceConfig struct {
	Fields    map[string]*domain.Field
	Functions map[string]*domain.Function
	Rules     map[string]*domain.Rule
	Workflows map[string]*domain.Workflow
	Terminals map[string]*domain.Terminal
	Checksum  string
	UpdatedAt time.Time
}

// ExecutionContext holds context for rule/workflow execution
type ExecutionContext struct {
	Namespace string
	Data      map[string]interface{}
	Trace     bool
	TraceData *domain.ExecutionTrace
	StartTime time.Time
}

// NewEngine creates a new execution engine
func NewEngine(
	cacheRepo domain.CacheRepository,
	ruleRepo domain.RuleRepository,
	workflowRepo domain.WorkflowRepository,
	fieldRepo domain.FieldRepository,
	functionRepo domain.FunctionRepository,
	terminalRepo domain.TerminalRepository,
	refreshInterval time.Duration,
) *Engine {
	return &Engine{
		cache: &NamespaceCache{
			data:      make(map[string]*NamespaceConfig),
			checksums: make(map[string]string),
		},
		cacheRepo:       cacheRepo,
		ruleRepo:        ruleRepo,
		workflowRepo:    workflowRepo,
		fieldRepo:       fieldRepo,
		functionRepo:    functionRepo,
		terminalRepo:    terminalRepo,
		lastRefresh:     make(map[string]time.Time),
		refreshInterval: refreshInterval,
	}
}

// ExecuteRule executes a specific rule
func (e *Engine) ExecuteRule(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	if req.RuleID == nil {
		return nil, domain.ErrInvalidInput
	}

	// Ensure cache is fresh
	if err := e.ensureFreshCache(ctx, req.Namespace); err != nil {
		return nil, fmt.Errorf("failed to refresh cache: %w", err)
	}

	// Get namespace config
	config, err := e.getNamespaceConfig(req.Namespace)
	if err != nil {
		return nil, err
	}

	// Find rule
	rule, exists := config.Rules[*req.RuleID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	// Create execution context
	execCtx := &ExecutionContext{
		Namespace: req.Namespace,
		Data:      req.Data,
		Trace:     req.Trace,
		StartTime: time.Now(),
	}

	if req.Trace {
		execCtx.TraceData = &domain.ExecutionTrace{
			Steps:   []domain.TraceStep{},
			Version: "1.0.0",
		}
	}

	// Execute rule
	result, err := e.evaluateRule(rule, config, execCtx)
	if err != nil {
		return nil, fmt.Errorf("rule execution failed: %w", err)
	}

	// Build response
	response := &domain.ExecutionResponse{
		Result:    result,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"ruleId":    *req.RuleID,
			"namespace": req.Namespace,
		},
	}

	if req.Trace && execCtx.TraceData != nil {
		execCtx.TraceData.Duration = time.Since(execCtx.StartTime).String()
		response.Trace = execCtx.TraceData
	}

	return response, nil
}

// ExecuteWorkflow executes a workflow
func (e *Engine) ExecuteWorkflow(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	if req.WorkflowID == nil {
		return nil, domain.ErrInvalidInput
	}

	// Ensure cache is fresh
	if err := e.ensureFreshCache(ctx, req.Namespace); err != nil {
		return nil, fmt.Errorf("failed to refresh cache: %w", err)
	}

	// Get namespace config
	config, err := e.getNamespaceConfig(req.Namespace)
	if err != nil {
		return nil, err
	}

	// Find workflow
	workflow, exists := config.Workflows[*req.WorkflowID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	// Create execution context
	execCtx := &ExecutionContext{
		Namespace: req.Namespace,
		Data:      req.Data,
		Trace:     req.Trace,
		StartTime: time.Now(),
	}

	if req.Trace {
		execCtx.TraceData = &domain.ExecutionTrace{
			Steps:   []domain.TraceStep{},
			Version: "1.0.0",
		}
	}

	// Execute workflow
	result, err := e.evaluateWorkflow(workflow, config, execCtx)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	// Build response
	response := &domain.ExecutionResponse{
		Result:    result,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"workflowId": *req.WorkflowID,
			"namespace":  req.Namespace,
		},
	}

	if req.Trace && execCtx.TraceData != nil {
		execCtx.TraceData.Duration = time.Since(execCtx.StartTime).String()
		response.Trace = execCtx.TraceData
	}

	return response, nil
}

// evaluateRule evaluates a rule using AST-based evaluation
func (e *Engine) evaluateRule(rule *domain.Rule, config *NamespaceConfig, ctx *ExecutionContext) (interface{}, error) {
	stepStart := time.Now()

	// Parse conditions
	var conditions map[string]interface{}
	if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
		return nil, fmt.Errorf("failed to parse rule conditions: %w", err)
	}

	// Evaluate conditions based on logic (AND/OR)
	result := false
	if rule.Logic == domain.LogicAND {
		result = e.evaluateANDConditions(conditions, config, ctx)
	} else if rule.Logic == domain.LogicOR {
		result = e.evaluateORConditions(conditions, config, ctx)
	} else {
		return nil, fmt.Errorf("unsupported logic type: %s", rule.Logic)
	}

	// Add trace step
	if ctx.Trace && ctx.TraceData != nil {
		ctx.TraceData.Steps = append(ctx.TraceData.Steps, domain.TraceStep{
			Type:     "rule",
			ID:       rule.RuleID,
			Input:    conditions,
			Output:   result,
			Duration: time.Since(stepStart).String(),
		})
	}

	return result, nil
}

// evaluateWorkflow evaluates a workflow by executing steps
func (e *Engine) evaluateWorkflow(workflow *domain.Workflow, config *NamespaceConfig, ctx *ExecutionContext) (interface{}, error) {
	stepStart := time.Now()

	// Parse steps
	var steps map[string]interface{}
	if err := json.Unmarshal(workflow.Steps, &steps); err != nil {
		return nil, fmt.Errorf("failed to parse workflow steps: %w", err)
	}

	// Start execution from startAt step
	currentStep := workflow.StartAt
	result := interface{}(nil)

	for {
		stepData, exists := steps[currentStep]
		if !exists {
			return nil, fmt.Errorf("step not found: %s", currentStep)
		}

		stepMap, ok := stepData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid step format: %s", currentStep)
		}

		stepType, _ := stepMap["type"].(string)

		// Execute step based on type
		stepResult, nextStep, err := e.executeWorkflowStep(stepType, stepMap, config, ctx)
		if err != nil {
			return nil, fmt.Errorf("step execution failed (%s): %w", currentStep, err)
		}

		result = stepResult

		// Check if we're done
		if nextStep == "" || stepType == "terminal" {
			break
		}

		currentStep = nextStep
	}

	// Add trace step
	if ctx.Trace && ctx.TraceData != nil {
		ctx.TraceData.Steps = append(ctx.TraceData.Steps, domain.TraceStep{
			Type:     "workflow",
			ID:       workflow.WorkflowID,
			Input:    steps,
			Output:   result,
			Duration: time.Since(stepStart).String(),
		})
	}

	return result, nil
}

// executeWorkflowStep executes a single workflow step
func (e *Engine) executeWorkflowStep(stepType string, stepData map[string]interface{}, config *NamespaceConfig, ctx *ExecutionContext) (interface{}, string, error) {
	switch stepType {
	case "rule":
		ruleID, _ := stepData["ruleId"].(string)
		if ruleID == "" {
			return nil, "", fmt.Errorf("missing ruleId in rule step")
		}

		rule, exists := config.Rules[ruleID]
		if !exists {
			return nil, "", fmt.Errorf("rule not found: %s", ruleID)
		}

		result, err := e.evaluateRule(rule, config, ctx)
		if err != nil {
			return nil, "", err
		}

		// Determine next step based on result
		nextStep := ""
		if result == true {
			nextStep, _ = stepData["onTrue"].(string)
		} else {
			nextStep, _ = stepData["onFalse"].(string)
		}

		return result, nextStep, nil

	case "terminal":
		terminalID, _ := stepData["terminalId"].(string)
		if terminalID == "" {
			return nil, "", fmt.Errorf("missing terminalId in terminal step")
		}

		_, exists := config.Terminals[terminalID]
		if !exists {
			return nil, "", fmt.Errorf("terminal not found: %s", terminalID)
		}

		result, _ := stepData["result"]
		return result, "", nil

	default:
		return nil, "", fmt.Errorf("unsupported step type: %s", stepType)
	}
}

// evaluateANDConditions evaluates conditions with AND logic
func (e *Engine) evaluateANDConditions(conditions map[string]interface{}, config *NamespaceConfig, ctx *ExecutionContext) bool {
	for _, condition := range conditions {
		if !e.evaluateCondition(condition, config, ctx) {
			return false
		}
	}
	return true
}

// evaluateORConditions evaluates conditions with OR logic
func (e *Engine) evaluateORConditions(conditions map[string]interface{}, config *NamespaceConfig, ctx *ExecutionContext) bool {
	for _, condition := range conditions {
		if e.evaluateCondition(condition, config, ctx) {
			return true
		}
	}
	return false
}

// evaluateCondition evaluates a single condition
func (e *Engine) evaluateCondition(condition interface{}, config *NamespaceConfig, ctx *ExecutionContext) bool {
	conditionMap, ok := condition.(map[string]interface{})
	if !ok {
		return false
	}

	// Extract condition components
	field, _ := conditionMap["field"].(string)
	operator, _ := conditionMap["operator"].(string)
	value := conditionMap["value"]

	// Get field value from data
	dataValue, exists := ctx.Data[field]
	if !exists {
		return false
	}

	// Evaluate based on operator
	return e.evaluateOperator(dataValue, operator, value, config, ctx)
}

// evaluateOperator evaluates a comparison operator
func (e *Engine) evaluateOperator(dataValue interface{}, operator string, expectedValue interface{}, config *NamespaceConfig, ctx *ExecutionContext) bool {
	switch operator {
	case "eq", "equals":
		return dataValue == expectedValue
	case "ne", "not_equals":
		return dataValue != expectedValue
	case "gt", "greater_than":
		return compareNumbers(dataValue, expectedValue) > 0
	case "gte", "greater_than_or_equal":
		return compareNumbers(dataValue, expectedValue) >= 0
	case "lt", "less_than":
		return compareNumbers(dataValue, expectedValue) < 0
	case "lte", "less_than_or_equal":
		return compareNumbers(dataValue, expectedValue) <= 0
	case "in":
		return e.evaluateInOperator(dataValue, expectedValue)
	case "not_in":
		return !e.evaluateInOperator(dataValue, expectedValue)
	default:
		return false
	}
}

// evaluateInOperator evaluates the 'in' operator
func (e *Engine) evaluateInOperator(dataValue, expectedValue interface{}) bool {
	expectedSlice, ok := expectedValue.([]interface{})
	if !ok {
		return false
	}

	for _, item := range expectedSlice {
		if dataValue == item {
			return true
		}
	}
	return false
}

// compareNumbers compares two numeric values
func compareNumbers(a, b interface{}) int {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return 0
	}

	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

// toFloat64 converts interface{} to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// ensureFreshCache ensures the cache is fresh for the given namespace
func (e *Engine) ensureFreshCache(ctx context.Context, namespace string) error {
	e.refreshMutex.Lock()
	defer e.refreshMutex.Unlock()

	// Check if we need to refresh
	lastRefresh, exists := e.lastRefresh[namespace]
	if !exists || time.Since(lastRefresh) > e.refreshInterval {
		if err := e.refreshNamespaceCache(ctx, namespace); err != nil {
			return err
		}
		e.lastRefresh[namespace] = time.Now()
	}

	return nil
}

// refreshNamespaceCache refreshes the cache for a specific namespace
func (e *Engine) refreshNamespaceCache(ctx context.Context, namespace string) error {
	// Get current checksum
	currentChecksum, err := e.cacheRepo.GetActiveConfigChecksum(ctx, namespace)
	if err != nil && err != domain.ErrNotFound {
		return fmt.Errorf("failed to get current checksum: %w", err)
	}

	// Check if cache is already up to date
	e.cache.mutex.RLock()
	cachedChecksum, exists := e.cache.checksums[namespace]
	e.cache.mutex.RUnlock()

	if exists && currentChecksum != nil && cachedChecksum == currentChecksum.Checksum {
		return nil // Cache is up to date
	}

	// Load fresh configuration
	config, err := e.loadNamespaceConfig(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to load namespace config: %w", err)
	}

	// Update cache atomically
	e.cache.mutex.Lock()
	e.cache.data[namespace] = config
	if currentChecksum != nil {
		e.cache.checksums[namespace] = currentChecksum.Checksum
	}
	e.cache.mutex.Unlock()

	return nil
}

// loadNamespaceConfig loads all active configuration for a namespace
func (e *Engine) loadNamespaceConfig(ctx context.Context, namespace string) (*NamespaceConfig, error) {
	config := &NamespaceConfig{
		Fields:    make(map[string]*domain.Field),
		Functions: make(map[string]*domain.Function),
		Rules:     make(map[string]*domain.Rule),
		Workflows: make(map[string]*domain.Workflow),
		Terminals: make(map[string]*domain.Terminal),
		UpdatedAt: time.Now(),
	}

	// Load fields
	fields, err := e.fieldRepo.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}
	for _, field := range fields {
		config.Fields[field.FieldID] = field
	}

	// Load active functions
	functions, err := e.functionRepo.ListActive(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load functions: %w", err)
	}
	for _, function := range functions {
		config.Functions[function.FunctionID] = function
	}

	// Load active rules
	rules, err := e.ruleRepo.ListActive(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}
	for _, rule := range rules {
		config.Rules[rule.RuleID] = rule
	}

	// Load active workflows
	workflows, err := e.workflowRepo.ListActive(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}
	for _, workflow := range workflows {
		config.Workflows[workflow.WorkflowID] = workflow
	}

	// Load terminals
	terminals, err := e.terminalRepo.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load terminals: %w", err)
	}
	for _, terminal := range terminals {
		config.Terminals[terminal.TerminalID] = terminal
	}

	return config, nil
}

// getNamespaceConfig gets cached namespace configuration
func (e *Engine) getNamespaceConfig(namespace string) (*NamespaceConfig, error) {
	e.cache.mutex.RLock()
	defer e.cache.mutex.RUnlock()

	config, exists := e.cache.data[namespace]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return config, nil
}
