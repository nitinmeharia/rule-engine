package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflow_ValidateAllPathsLeadToTerminals(t *testing.T) {
	tests := []struct {
		name        string
		steps       map[string]interface{}
		startAt     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid workflow with all paths leading to terminals",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
				"step3": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: false,
		},
		{
			name: "valid workflow with nested rule paths",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule2",
					"onTrue":  "step4",
					"onFalse": "step5",
				},
				"step3": map[string]interface{}{
					"type": "terminal",
				},
				"step4": map[string]interface{}{
					"type": "terminal",
				},
				"step5": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: false,
		},
		{
			name: "invalid workflow - rule step missing onTrue path",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onFalse": "step2",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: The 'onTrue' path for step 'step1' does not lead to a terminal.",
		},
		{
			name: "invalid workflow - rule step missing onFalse path",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":   "rule",
					"ruleId": "rule1",
					"onTrue": "step2",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: The 'onFalse' path for step 'step1' does not lead to a terminal.",
		},
		{
			name: "invalid workflow - onTrue path leads to non-terminal",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule2",
					"onTrue":  "step4",
					"onFalse": "step5",
				},
				"step3": map[string]interface{}{
					"type": "terminal",
				},
				"step4": map[string]interface{}{
					"type": "terminal",
				},
				// step5 is missing - this should cause validation to fail
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: The 'onTrue' path for step 'step1' does not lead to a terminal.",
		},
		{
			name: "invalid workflow - onFalse path leads to non-terminal",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
				"step3": map[string]interface{}{
					"type":   "rule",
					"ruleId": "rule2",
					"onTrue": "step4",
					// missing onFalse
				},
				"step4": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: The 'onFalse' path for step 'step1' does not lead to a terminal.",
		},
		{
			name: "invalid workflow - step leads to non-existent step",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
				// step3 is missing but referenced
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: The 'onFalse' path for step 'step1' does not lead to a terminal.",
		},
		{
			name: "valid workflow - single terminal step",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: false,
		},
		{
			name: "invalid workflow - unknown step type",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type": "unknown",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: Step 'step1' is invalid or missing and does not lead to a terminal.",
		},
		{
			name: "invalid workflow - malformed step data",
			steps: map[string]interface{}{
				"step1": "not a map",
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "Validation Error: Step 'step1' is invalid or missing and does not lead to a terminal.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepsJSON, err := json.Marshal(tt.steps)
			if err != nil {
				t.Fatalf("Failed to marshal steps: %v", err)
			}

			workflow := &Workflow{
				WorkflowID: "test-workflow",
				StartAt:    tt.startAt,
				Steps:      stepsJSON,
			}

			err = workflow.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestWorkflow_ValidateWorkflowSteps_CycleDetection(t *testing.T) {
	tests := []struct {
		name        string
		steps       map[string]interface{}
		startAt     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid workflow - no cycles",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
				"step3": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: false,
		},
		{
			name: "invalid workflow - direct cycle",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step1", // Direct cycle
				},
				"step2": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "cyclic dependency detected in workflow steps",
		},
		{
			name: "invalid workflow - indirect cycle",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule2",
					"onTrue":  "step3",
					"onFalse": "step4",
				},
				"step3": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule3",
					"onTrue":  "step1", // Creates cycle: step1 -> step2 -> step3 -> step1
					"onFalse": "step5",
				},
				"step4": map[string]interface{}{
					"type": "terminal",
				},
				"step5": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			errorMsg:    "cyclic dependency detected in workflow steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepsJSON, err := json.Marshal(tt.steps)
			if err != nil {
				t.Fatalf("Failed to marshal steps: %v", err)
			}

			workflow := &Workflow{
				WorkflowID: "test-workflow",
				StartAt:    tt.startAt,
				Steps:      stepsJSON,
			}

			err = workflow.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestWorkflow_ValidateWorkflowSteps_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name        string
		steps       map[string]interface{}
		startAt     string
		expectError bool
		description string
	}{
		{
			name: "complex valid workflow with multiple branches",
			steps: map[string]interface{}{
				"start": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "initial_check",
					"onTrue":  "branch_a",
					"onFalse": "branch_b",
				},
				"branch_a": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "branch_a_check",
					"onTrue":  "terminal_approve",
					"onFalse": "terminal_reject",
				},
				"branch_b": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "branch_b_check",
					"onTrue":  "terminal_approve",
					"onFalse": "terminal_reject",
				},
				"terminal_approve": map[string]interface{}{
					"type": "terminal",
				},
				"terminal_reject": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "start",
			expectError: false,
			description: "Complex workflow with multiple branches all leading to terminals",
		},
		{
			name: "workflow with shared terminal",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule2",
					"onTrue":  "terminal_shared",
					"onFalse": "terminal_shared",
				},
				"step3": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule3",
					"onTrue":  "terminal_shared",
					"onFalse": "terminal_shared",
				},
				"terminal_shared": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: false,
			description: "Workflow where multiple paths lead to the same terminal",
		},
		{
			name: "workflow with missing terminal reference",
			steps: map[string]interface{}{
				"step1": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule1",
					"onTrue":  "step2",
					"onFalse": "step3",
				},
				"step2": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule2",
					"onTrue":  "terminal_approve",
					"onFalse": "terminal_reject",
				},
				"step3": map[string]interface{}{
					"type":    "rule",
					"ruleId":  "rule3",
					"onTrue":  "terminal_approve",
					"onFalse": "missing_terminal", // This terminal doesn't exist
				},
				"terminal_approve": map[string]interface{}{
					"type": "terminal",
				},
				"terminal_reject": map[string]interface{}{
					"type": "terminal",
				},
			},
			startAt:     "step1",
			expectError: true,
			description: "Workflow with reference to non-existent terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepsJSON, err := json.Marshal(tt.steps)
			if err != nil {
				t.Fatalf("Failed to marshal steps: %v", err)
			}

			workflow := &Workflow{
				WorkflowID: "test-workflow",
				StartAt:    tt.startAt,
				Steps:      stepsJSON,
			}

			err = workflow.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v for: %s", err, tt.description)
				}
			}
		})
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
