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
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Error:   "Bad Request",
			Message: "Invalid request body: " + err.Error(),
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
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Error:   "Bad Request",
			Message: "Invalid request body: " + err.Error(),
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
func (h *ExecutionHandler) mapError(err error) (int, interface{}) {
	apiErr, ok := err.(*domain.APIError)
	if ok {
		return apiErr.HTTPStatus(), apiErr
	}

	// Fallback for non-APIError errors
	return http.StatusInternalServerError, domain.ErrInternalError
}

// isExecutionError checks if error is an execution error
func isExecutionError(err error) bool {
	return err != nil && (contains(err.Error(), "execution failed") ||
		contains(err.Error(), "evaluation failed") ||
		contains(err.Error(), "step execution failed") ||
		contains(err.Error(), "failed to parse") ||
		contains(err.Error(), "unsupported"))
}
