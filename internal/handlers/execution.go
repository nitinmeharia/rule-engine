package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/execution"
)

// ExecutionHandler handles execution HTTP requests
type ExecutionHandler struct {
	engine *execution.Engine
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(engine *execution.Engine) *ExecutionHandler {
	return &ExecutionHandler{
		engine: engine,
	}
}

// ExecuteRule executes a specific rule
func (h *ExecutionHandler) ExecuteRule(c *gin.Context) {
	namespace := c.Param("namespace")
	ruleID := c.Param("ruleId")

	var req domain.ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
			Details: map[string]interface{}{
				"validation": err.Error(),
			},
		})
		return
	}

	// Set namespace and ruleID from URL
	req.Namespace = namespace
	req.RuleID = &ruleID

	// Check for trace parameter
	if c.Query("trace") == "full" {
		req.Trace = true
	}

	// Execute rule
	response, err := h.engine.ExecuteRule(c.Request.Context(), &req)
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExecuteWorkflow executes a workflow
func (h *ExecutionHandler) ExecuteWorkflow(c *gin.Context) {
	namespace := c.Param("namespace")
	workflowID := c.Param("workflowId")

	var req domain.ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
			Details: map[string]interface{}{
				"validation": err.Error(),
			},
		})
		return
	}

	// Set namespace and workflowID from URL
	req.Namespace = namespace
	req.WorkflowID = &workflowID

	// Check for trace parameter
	if c.Query("trace") == "full" {
		req.Trace = true
	}

	// Execute workflow
	response, err := h.engine.ExecuteWorkflow(c.Request.Context(), &req)
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.JSON(http.StatusOK, response)
}

// mapError maps domain errors to HTTP responses
func (h *ExecutionHandler) mapError(err error) (int, ErrorResponse) {
	switch err {
	case domain.ErrNotFound:
		return http.StatusNotFound, ErrorResponse{
			Error: "Resource not found",
			Code:  "NOT_FOUND",
		}
	case domain.ErrInvalidInput:
		return http.StatusBadRequest, ErrorResponse{
			Error: "Invalid input",
			Code:  "INVALID_INPUT",
		}
	case domain.ErrValidation:
		return http.StatusBadRequest, ErrorResponse{
			Error: "Validation failed",
			Code:  "VALIDATION_ERROR",
		}
	case domain.ErrUnauthorized:
		return http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		}
	case domain.ErrForbidden:
		return http.StatusForbidden, ErrorResponse{
			Error: "Forbidden",
			Code:  "FORBIDDEN",
		}
	default:
		// Check if it's an execution error
		if isExecutionError(err) {
			return http.StatusUnprocessableEntity, ErrorResponse{
				Error: err.Error(),
				Code:  "EXECUTION_ERROR",
			}
		}

		return http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		}
	}
}

// isExecutionError checks if error is an execution error
func isExecutionError(err error) bool {
	return err != nil && (contains(err.Error(), "execution failed") ||
		contains(err.Error(), "evaluation failed") ||
		contains(err.Error(), "step execution failed") ||
		contains(err.Error(), "failed to parse") ||
		contains(err.Error(), "unsupported"))
}
