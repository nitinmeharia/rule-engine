package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// WorkflowHandler handles HTTP requests for workflows
type WorkflowHandler struct {
	workflowService service.WorkflowServiceInterface
	responseHandler *ResponseHandler
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(workflowService service.WorkflowServiceInterface) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		responseHandler: NewResponseHandler(),
	}
}

// CreateWorkflow handles POST /v1/namespaces/{namespace}/workflows
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	var req domain.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists || clientID == nil {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	workflow := &domain.Workflow{
		Namespace:  namespace,
		WorkflowID: req.ID,
		StartAt:    req.StartAt,
		Steps:      req.Steps,
		CreatedBy:  clientID.(string),
	}

	err := h.workflowService.Create(c.Request.Context(), workflow)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowToResponse(workflow)
	h.responseHandler.Created(c, response)
}

// GetWorkflow handles GET /v1/namespaces/{namespace}/workflows/{workflowId}
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")

	if namespace == "" || workflowID == "" {
		h.responseHandler.BadRequest(c, "Namespace and workflow ID are required")
		return
	}

	// Try to get active version first
	workflow, err := h.workflowService.GetActiveVersion(c.Request.Context(), namespace, workflowID)
	if err != nil {
		// If active version not found, try to get draft version
		if err == domain.ErrWorkflowNotFound {
			workflow, err = h.workflowService.GetDraftVersion(c.Request.Context(), namespace, workflowID)
			if err != nil {
				h.responseHandler.MapDomainErrorToResponse(c, err)
				return
			}
		} else {
			h.responseHandler.MapDomainErrorToResponse(c, err)
			return
		}
	}

	response := h.responseHandler.ConvertWorkflowToResponse(workflow)
	h.responseHandler.OK(c, response)
}

// GetWorkflowVersion handles GET /v1/namespaces/{namespace}/workflows/{workflowId}/versions/{version}
func (h *WorkflowHandler) GetWorkflowVersion(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")
	versionStr := c.Param("version")

	if namespace == "" || workflowID == "" || versionStr == "" {
		h.responseHandler.BadRequest(c, "Namespace, workflow ID, and version are required")
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		h.responseHandler.BadRequest(c, "Invalid version number")
		return
	}

	workflow, err := h.workflowService.GetByID(c.Request.Context(), namespace, workflowID, int32(version))
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowToResponse(workflow)
	h.responseHandler.OK(c, response)
}

// ListWorkflows handles GET /v1/namespaces/{namespace}/workflows
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	workflows, err := h.workflowService.List(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowsToResponse(workflows)
	h.responseHandler.OK(c, response)
}

// ListActiveWorkflows handles GET /v1/namespaces/{namespace}/workflows/active
func (h *WorkflowHandler) ListActiveWorkflows(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	workflows, err := h.workflowService.ListActive(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowsToResponse(workflows)
	h.responseHandler.OK(c, response)
}

// ListWorkflowVersions handles GET /v1/namespaces/{namespace}/workflows/{workflowId}/versions
func (h *WorkflowHandler) ListWorkflowVersions(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")

	if namespace == "" || workflowID == "" {
		h.responseHandler.BadRequest(c, "Namespace and workflow ID are required")
		return
	}

	workflows, err := h.workflowService.ListVersions(c.Request.Context(), namespace, workflowID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowsToResponse(workflows)
	h.responseHandler.OK(c, response)
}

// UpdateWorkflow handles PUT /v1/namespaces/{namespace}/workflows/{workflowId}/versions/{version}
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")
	versionStr := c.Param("version")

	if namespace == "" || workflowID == "" || versionStr == "" {
		h.responseHandler.BadRequest(c, "Namespace, workflow ID, and version are required")
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		h.responseHandler.BadRequest(c, "Invalid version number")
		return
	}

	var req domain.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists || clientID == nil {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	workflow := &domain.Workflow{
		Namespace:  namespace,
		WorkflowID: workflowID,
		Version:    int32(version),
		StartAt:    req.StartAt,
		Steps:      req.Steps,
		CreatedBy:  clientID.(string),
	}

	err = h.workflowService.Update(c.Request.Context(), workflow)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertWorkflowToResponse(workflow)
	h.responseHandler.OK(c, response)
}

// PublishWorkflow handles POST /v1/namespaces/{namespace}/workflows/{workflowId}/versions/{version}/publish
func (h *WorkflowHandler) PublishWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")
	versionStr := c.Param("version")

	if namespace == "" || workflowID == "" || versionStr == "" {
		h.responseHandler.BadRequest(c, "Namespace, workflow ID, and version are required")
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		h.responseHandler.BadRequest(c, "Invalid version number")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists || clientID == nil {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	err = h.workflowService.Publish(c.Request.Context(), namespace, workflowID, int32(version), clientID.(string))
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := domain.PublishWorkflowResponse{
		Status: "active",
	}
	c.JSON(http.StatusOK, response)
}

// DeactivateWorkflow handles POST /v1/namespaces/{namespace}/workflows/{workflowId}/deactivate
func (h *WorkflowHandler) DeactivateWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")

	if namespace == "" || workflowID == "" {
		h.responseHandler.BadRequest(c, "Namespace and workflow ID are required")
		return
	}

	err := h.workflowService.Deactivate(c.Request.Context(), namespace, workflowID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteWorkflow handles DELETE /v1/namespaces/{namespace}/workflows/{workflowId}/versions/{version}
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	namespace := c.Param("id")
	workflowID := c.Param("workflowId")
	versionStr := c.Param("version")

	if namespace == "" || workflowID == "" || versionStr == "" {
		h.responseHandler.BadRequest(c, "Namespace, workflow ID, and version are required")
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		h.responseHandler.BadRequest(c, "Invalid version number")
		return
	}

	err = h.workflowService.Delete(c.Request.Context(), namespace, workflowID, int32(version))
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
