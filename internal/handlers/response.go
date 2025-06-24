package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
)

// ResponseHandler provides centralized HTTP response formatting
type ResponseHandler struct{}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// Success sends a success response
func (h *ResponseHandler) Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, domain.SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error response
func (h *ResponseHandler) Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, domain.ErrorResponse{
		Code:    code,
		Error:   code,
		Message: message,
	})
}

// BadRequest sends a 400 Bad Request response
func (h *ResponseHandler) BadRequest(c *gin.Context, message string) {
	h.Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 Unauthorized response
func (h *ResponseHandler) Unauthorized(c *gin.Context, message string) {
	h.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden response
func (h *ResponseHandler) Forbidden(c *gin.Context, message string) {
	h.Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found response
func (h *ResponseHandler) NotFound(c *gin.Context, message string) {
	h.Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict response
func (h *ResponseHandler) Conflict(c *gin.Context, message string) {
	h.Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalServerError sends a 500 Internal Server Error response
func (h *ResponseHandler) InternalServerError(c *gin.Context, message string) {
	h.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// Created sends a 201 Created response
func (h *ResponseHandler) Created(c *gin.Context, data interface{}) {
	h.Success(c, http.StatusCreated, data)
}

// OK sends a 200 OK response
func (h *ResponseHandler) OK(c *gin.Context, data interface{}) {
	h.Success(c, http.StatusOK, data)
}

// NoContent sends a 204 No Content response
func (h *ResponseHandler) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// MapDomainErrorToResponse maps domain errors to appropriate HTTP responses
// This centralizes error handling and reduces boilerplate in handlers
func (h *ResponseHandler) MapDomainErrorToResponse(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's a domain APIError
	if apiErr, ok := err.(*domain.APIError); ok {
		c.JSON(apiErr.HTTPStatus(), apiErr)
		return
	}

	// Recursively unwrap and check error messages for validation and dependency errors
	unwrapErr := err
	for unwrapErr != nil {
		errMsg := unwrapErr.Error()

		// Check for validation errors first (highest priority)
		if strings.Contains(errMsg, "Validation Error") ||
			strings.Contains(errMsg, "validation error") ||
			strings.Contains(errMsg, "invalid") ||
			strings.Contains(errMsg, "missing") ||
			strings.Contains(errMsg, "required") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "VALIDATION_ERROR",
				"error":   "BAD_REQUEST",
				"message": errMsg,
			})
			return
		}

		// Check for dependency validation errors
		if strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "is not active") ||
			strings.Contains(errMsg, "invalid dependency") ||
			strings.Contains(errMsg, "dependency") ||
			strings.Contains(errMsg, "no rows in result set") ||
			strings.Contains(errMsg, "cyclic dependency") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "VALIDATION_ERROR",
				"error":   "BAD_REQUEST",
				"message": errMsg,
			})
			return
		}
		unwrapErr = errors.Unwrap(unwrapErr)
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_SERVER_ERROR",
		"error":   "INTERNAL_SERVER_ERROR",
		"message": "An unexpected error occurred",
	})
}

// ConvertNamespaceToResponse converts a domain Namespace to NamespaceResponse
func (h *ResponseHandler) ConvertNamespaceToResponse(namespace *domain.Namespace) domain.NamespaceResponse {
	return domain.NamespaceResponse{
		ID:          namespace.ID,
		Description: namespace.Description,
		CreatedAt:   namespace.CreatedAt,
		CreatedBy:   namespace.CreatedBy,
	}
}

// ConvertFieldToResponse converts a domain Field to FieldResponse
func (h *ResponseHandler) ConvertFieldToResponse(field *domain.Field) domain.FieldResponse {
	return domain.FieldResponse{
		FieldID:     field.FieldID,
		Type:        field.Type,
		Description: field.Description,
		CreatedAt:   field.CreatedAt,
		CreatedBy:   field.CreatedBy,
	}
}

// ConvertFunctionToResponse converts a domain Function to FunctionResponse
func (h *ResponseHandler) ConvertFunctionToResponse(function *domain.Function) domain.FunctionResponse {
	return domain.FunctionResponse{
		ID:          function.FunctionID,
		Version:     function.Version,
		Status:      function.Status,
		Type:        function.Type,
		Args:        function.Args,
		Values:      function.Values,
		ReturnType:  function.ReturnType,
		CreatedAt:   function.CreatedAt,
		CreatedBy:   function.CreatedBy,
		PublishedAt: function.PublishedAt,
		PublishedBy: function.PublishedBy,
	}
}

// ConvertFunctionsToResponse converts a slice of domain Functions to FunctionResponse slice
func (h *ResponseHandler) ConvertFunctionsToResponse(functions []*domain.Function) []domain.FunctionResponse {
	response := make([]domain.FunctionResponse, 0, len(functions))
	for _, function := range functions {
		response = append(response, h.ConvertFunctionToResponse(function))
	}
	return response
}

// ConvertRuleToResponse converts a domain Rule to RuleResponse
func (h *ResponseHandler) ConvertRuleToResponse(rule *domain.Rule) domain.RuleResponse {
	return domain.RuleResponse{
		ID:          rule.RuleID,
		Version:     rule.Version,
		Status:      rule.Status,
		Logic:       rule.Logic,
		Conditions:  rule.Conditions,
		CreatedAt:   rule.CreatedAt,
		CreatedBy:   rule.CreatedBy,
		PublishedAt: rule.PublishedAt,
		PublishedBy: rule.PublishedBy,
	}
}

// ConvertRulesToResponse converts a slice of domain Rules to RuleResponse slice
func (h *ResponseHandler) ConvertRulesToResponse(rules []*domain.Rule) []domain.RuleResponse {
	response := make([]domain.RuleResponse, 0, len(rules))
	for _, rule := range rules {
		response = append(response, h.ConvertRuleToResponse(rule))
	}
	return response
}

// ConvertRuleVersionsToResponse converts a slice of domain Rules to RuleResponse slice
func (h *ResponseHandler) ConvertRuleVersionsToResponse(rules []*domain.Rule) []domain.RuleResponse {
	response := make([]domain.RuleResponse, len(rules))
	for i, rule := range rules {
		response[i] = h.ConvertRuleToResponse(rule)
	}
	return response
}

// ConvertTerminalToResponse converts a domain Terminal to TerminalResponse
func (h *ResponseHandler) ConvertTerminalToResponse(terminal *domain.Terminal) domain.TerminalResponse {
	return domain.TerminalResponse{
		TerminalID: terminal.TerminalID,
		CreatedAt:  terminal.CreatedAt,
		CreatedBy:  terminal.CreatedBy,
	}
}

// ConvertWorkflowToResponse converts a domain Workflow to WorkflowResponse
func (h *ResponseHandler) ConvertWorkflowToResponse(workflow *domain.Workflow) domain.WorkflowResponse {
	return domain.WorkflowResponse{
		ID:          workflow.WorkflowID,
		Version:     workflow.Version,
		Status:      workflow.Status,
		StartAt:     workflow.StartAt,
		Steps:       workflow.Steps,
		CreatedAt:   workflow.CreatedAt,
		CreatedBy:   workflow.CreatedBy,
		PublishedAt: workflow.PublishedAt,
		PublishedBy: workflow.PublishedBy,
	}
}

// ConvertWorkflowsToResponse converts a slice of domain Workflows to WorkflowResponse slice
func (h *ResponseHandler) ConvertWorkflowsToResponse(workflows []*domain.Workflow) []domain.WorkflowResponse {
	response := make([]domain.WorkflowResponse, 0, len(workflows))
	for _, workflow := range workflows {
		response = append(response, h.ConvertWorkflowToResponse(workflow))
	}
	return response
}
