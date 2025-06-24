package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/execution"
)

// ExecutionHandler handles execution HTTP requests
type ExecutionHandler struct {
	engine          *execution.Engine
	responseHandler *ResponseHandler
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(engine *execution.Engine) *ExecutionHandler {
	return &ExecutionHandler{
		engine:          engine,
		responseHandler: NewResponseHandler(),
	}
}

// ExecuteRule executes a specific rule
func (h *ExecutionHandler) ExecuteRule(c *gin.Context) {
	namespace := c.Param("namespace")
	ruleID := c.Param("ruleId")

	var req domain.ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body: "+err.Error())
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
		h.responseHandler.MapDomainErrorToResponse(c, err)
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
		h.responseHandler.BadRequest(c, "Invalid request body: "+err.Error())
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
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// isExecutionError checks if error is an execution error
func isExecutionError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "execution failed") ||
		strings.Contains(err.Error(), "evaluation failed") ||
		strings.Contains(err.Error(), "step execution failed") ||
		strings.Contains(err.Error(), "failed to parse") ||
		strings.Contains(err.Error(), "unsupported"))
}
