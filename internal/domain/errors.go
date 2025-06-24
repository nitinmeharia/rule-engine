package domain

import (
	"fmt"
	"net/http"
)

// Error codes for API responses
const (
	// Namespace errors
	ErrCodeNamespaceAlreadyExists = "NAMESPACE_ALREADY_EXISTS"
	ErrCodeNamespaceNotFound      = "NAMESPACE_NOT_FOUND"
	ErrCodeInvalidNamespaceID     = "INVALID_NAMESPACE_ID"
	ErrCodeInvalidDescription     = "INVALID_DESCRIPTION"

	// Field errors
	ErrCodeFieldAlreadyExists = "FIELD_ALREADY_EXISTS"
	ErrCodeFieldNotFound      = "FIELD_NOT_FOUND"
	ErrCodeInvalidFieldID     = "INVALID_FIELD_ID"
	ErrCodeInvalidFieldType   = "INVALID_FIELD_TYPE"

	// Function errors
	ErrCodeFunctionAlreadyExists = "FUNCTION_ALREADY_EXISTS"
	ErrCodeFunctionNotFound      = "FUNCTION_NOT_FOUND"
	ErrCodeInvalidFunctionID     = "INVALID_FUNCTION_ID"
	ErrCodeInvalidFunctionType   = "INVALID_FUNCTION_TYPE"
	ErrCodeInvalidFunctionArgs   = "INVALID_FUNCTION_ARGS"

	// Rule errors
	ErrCodeRuleAlreadyExists     = "RULE_ALREADY_EXISTS"
	ErrCodeRuleNotFound          = "RULE_NOT_FOUND"
	ErrCodeInvalidRuleID         = "INVALID_RULE_ID"
	ErrCodeInvalidRuleLogic      = "INVALID_RULE_LOGIC"
	ErrCodeInvalidRuleConditions = "INVALID_RULE_CONDITIONS"
	ErrCodeDraftExists           = "DRAFT_EXISTS"
	ErrCodeDependencyInactive    = "PUBLISH_DEPENDENCY_INACTIVE"

	// Workflow errors
	ErrCodeWorkflowAlreadyExists   = "WORKFLOW_ALREADY_EXISTS"
	ErrCodeWorkflowNotFound        = "WORKFLOW_NOT_FOUND"
	ErrCodeInvalidWorkflowID       = "INVALID_WORKFLOW_ID"
	ErrCodeInvalidWorkflowStartAt  = "INVALID_WORKFLOW_START_AT"
	ErrCodeWorkflowExecutionFailed = "WORKFLOW_EXECUTION_FAILED"

	// Terminal errors
	ErrCodeTerminalAlreadyExists = "TERMINAL_ALREADY_EXISTS"
	ErrCodeTerminalNotFound      = "TERMINAL_NOT_FOUND"
	ErrCodeInvalidTerminalID     = "INVALID_TERMINAL_ID"

	// Authentication & Authorization errors
	ErrCodeMissingAuthHeader       = "MISSING_AUTH_HEADER"
	ErrCodeInvalidJWTToken         = "INVALID_JWT_TOKEN"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// Validation errors
	ErrCodeValidationError = "VALIDATION_ERROR"

	// Precondition errors
	ErrCodePreconditionFailed = "PRECONDITION_FAILED"

	// Internal errors
	ErrCodeInternalError = "INTERNAL_ERROR"
	ErrCodeListError     = "LIST_ERROR"

	// Cache/Config errors
	ErrCodeInvalidChecksum = "INVALID_CHECKSUM"

	// Execution errors
	ErrCodeInvalidExecutionRequest = "INVALID_EXECUTION_REQUEST"
	ErrCodeInvalidExecutionData    = "INVALID_EXECUTION_DATA"

	// New errors
	ErrCodeFunctionNotActive = "FUNCTION_NOT_ACTIVE"
	ErrCodeRuleNotActive     = "RULE_NOT_ACTIVE"
)

// APIError represents a standardized API error response
type APIError struct {
	Code      string `json:"code"`
	ErrorType string `json:"error"`
	Message   string `json:"message"`
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *APIError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeNamespaceAlreadyExists, ErrCodeFieldAlreadyExists, ErrCodeFunctionAlreadyExists,
		ErrCodeRuleAlreadyExists, ErrCodeWorkflowAlreadyExists, ErrCodeTerminalAlreadyExists,
		ErrCodeDraftExists, ErrCodeDependencyInactive:
		return http.StatusConflict
	case ErrCodeNamespaceNotFound, ErrCodeFieldNotFound, ErrCodeFunctionNotFound,
		ErrCodeRuleNotFound, ErrCodeWorkflowNotFound, ErrCodeTerminalNotFound:
		return http.StatusNotFound
	case ErrCodeInvalidNamespaceID, ErrCodeInvalidFieldID, ErrCodeInvalidFunctionID,
		ErrCodeInvalidRuleID, ErrCodeInvalidWorkflowID, ErrCodeInvalidTerminalID,
		ErrCodeInvalidFieldType, ErrCodeInvalidFunctionType, ErrCodeInvalidRuleLogic,
		ErrCodeInvalidRuleConditions, ErrCodeInvalidDescription, ErrCodeInvalidFunctionArgs, ErrCodeValidationError,
		ErrCodeInvalidWorkflowStartAt, ErrCodeInvalidChecksum, ErrCodeInvalidExecutionRequest,
		ErrCodeInvalidExecutionData:
		return http.StatusBadRequest
	case ErrCodeMissingAuthHeader, ErrCodeInvalidJWTToken:
		return http.StatusUnauthorized
	case ErrCodeInsufficientPermissions:
		return http.StatusForbidden
	case ErrCodePreconditionFailed:
		return http.StatusPreconditionFailed
	case ErrCodeWorkflowExecutionFailed:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new APIError with the given code and message
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:      code,
		ErrorType: getErrorType(code),
		Message:   message,
	}
}

// getErrorType returns the error type string based on the error code
func getErrorType(code string) string {
	switch code {
	case ErrCodeMissingAuthHeader, ErrCodeInvalidJWTToken:
		return "UNAUTHORIZED"
	case ErrCodeInsufficientPermissions:
		return "FORBIDDEN"
	case ErrCodeNamespaceNotFound, ErrCodeFieldNotFound, ErrCodeFunctionNotFound,
		ErrCodeRuleNotFound, ErrCodeWorkflowNotFound, ErrCodeTerminalNotFound:
		return "NOT_FOUND"
	case ErrCodeNamespaceAlreadyExists, ErrCodeFieldAlreadyExists, ErrCodeFunctionAlreadyExists,
		ErrCodeRuleAlreadyExists, ErrCodeWorkflowAlreadyExists, ErrCodeTerminalAlreadyExists,
		ErrCodeDraftExists, ErrCodeDependencyInactive:
		return "CONFLICT"
	case ErrCodePreconditionFailed:
		return "PRECONDITION_FAILED"
	case ErrCodeWorkflowExecutionFailed:
		return "UNPROCESSABLE_ENTITY"
	case ErrCodeInvalidNamespaceID, ErrCodeInvalidFieldID, ErrCodeInvalidFunctionID,
		ErrCodeInvalidRuleID, ErrCodeInvalidWorkflowID, ErrCodeInvalidTerminalID,
		ErrCodeInvalidFieldType, ErrCodeInvalidFunctionType, ErrCodeInvalidRuleLogic,
		ErrCodeInvalidRuleConditions, ErrCodeInvalidDescription, ErrCodeInvalidFunctionArgs, ErrCodeValidationError,
		ErrCodeInvalidWorkflowStartAt, ErrCodeInvalidChecksum, ErrCodeInvalidExecutionRequest,
		ErrCodeInvalidExecutionData:
		return "BAD_REQUEST"
	default:
		return "INTERNAL_ERROR"
	}
}

// Common error instances
var (
	ErrNamespaceAlreadyExists = NewAPIError(ErrCodeNamespaceAlreadyExists, "Namespace already exists")
	ErrNamespaceNotFound      = NewAPIError(ErrCodeNamespaceNotFound, "Namespace not found")
	ErrInvalidNamespaceID     = NewAPIError(ErrCodeInvalidNamespaceID, "Invalid namespace ID")
	ErrInvalidDescription     = NewAPIError(ErrCodeInvalidDescription, "Description is required")

	ErrFieldAlreadyExists = NewAPIError(ErrCodeFieldAlreadyExists, "Field already exists")
	ErrFieldNotFound      = NewAPIError(ErrCodeFieldNotFound, "Field not found")
	ErrInvalidFieldID     = NewAPIError(ErrCodeInvalidFieldID, "Field ID is required")
	ErrInvalidFieldType   = NewAPIError(ErrCodeInvalidFieldType, "Field type must be 'number' or 'string'")

	ErrFunctionAlreadyExists = NewAPIError(ErrCodeFunctionAlreadyExists, "Function already exists")
	ErrFunctionNotFound      = NewAPIError(ErrCodeFunctionNotFound, "Function not found")
	ErrInvalidFunctionID     = NewAPIError(ErrCodeInvalidFunctionID, "Function ID is required")
	ErrInvalidFunctionType   = NewAPIError(ErrCodeInvalidFunctionType, "Invalid function type")
	ErrInvalidFunctionArgs   = NewAPIError(ErrCodeInvalidFunctionArgs, "Invalid function arguments")

	ErrRuleAlreadyExists     = NewAPIError(ErrCodeRuleAlreadyExists, "Rule already exists")
	ErrRuleNotFound          = NewAPIError(ErrCodeRuleNotFound, "Rule not found")
	ErrInvalidRuleID         = NewAPIError(ErrCodeInvalidRuleID, "Rule ID is required")
	ErrInvalidRuleLogic      = NewAPIError(ErrCodeInvalidRuleLogic, "Invalid rule logic")
	ErrInvalidRuleConditions = NewAPIError(ErrCodeInvalidRuleConditions, "Invalid rule conditions")
	ErrDraftExists           = NewAPIError(ErrCodeDraftExists, "Draft already exists")
	ErrDependencyInactive    = NewAPIError(ErrCodeDependencyInactive, "Dependent rule/function not active")

	ErrWorkflowAlreadyExists   = NewAPIError(ErrCodeWorkflowAlreadyExists, "Workflow already exists")
	ErrWorkflowNotFound        = NewAPIError(ErrCodeWorkflowNotFound, "Workflow not found")
	ErrInvalidWorkflowID       = NewAPIError(ErrCodeInvalidWorkflowID, "Workflow ID is required")
	ErrInvalidWorkflowStartAt  = NewAPIError(ErrCodeInvalidWorkflowStartAt, "Invalid workflow start at")
	ErrWorkflowExecutionFailed = NewAPIError(ErrCodeWorkflowExecutionFailed, "Workflow execution failed")

	ErrTerminalAlreadyExists = NewAPIError(ErrCodeTerminalAlreadyExists, "Terminal already exists")
	ErrTerminalNotFound      = NewAPIError(ErrCodeTerminalNotFound, "Terminal not found")
	ErrInvalidTerminalID     = NewAPIError(ErrCodeInvalidTerminalID, "Terminal ID is required")

	ErrMissingAuthHeader       = NewAPIError(ErrCodeMissingAuthHeader, "Missing Authorization header")
	ErrInvalidJWTToken         = NewAPIError(ErrCodeInvalidJWTToken, "Invalid JWT token")
	ErrInsufficientPermissions = NewAPIError(ErrCodeInsufficientPermissions, "Insufficient permissions")

	ErrValidationError    = NewAPIError(ErrCodeValidationError, "Validation error")
	ErrPreconditionFailed = NewAPIError(ErrCodePreconditionFailed, "Precondition failed")
	ErrInternalError      = NewAPIError(ErrCodeInternalError, "Internal server error")
	ErrListError          = NewAPIError(ErrCodeListError, "Failed to list resources")

	ErrInvalidChecksum = NewAPIError(ErrCodeInvalidChecksum, "Invalid checksum")

	ErrInvalidExecutionRequest = NewAPIError(ErrCodeInvalidExecutionRequest, "Invalid execution request")
	ErrInvalidExecutionData    = NewAPIError(ErrCodeInvalidExecutionData, "Invalid execution data")

	ErrFunctionNotActive = NewAPIError(ErrCodeFunctionNotActive, "Function not active")
	ErrRuleNotActive     = NewAPIError(ErrCodeRuleNotActive, "Rule not active")
)
