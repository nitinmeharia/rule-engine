package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// NamespaceHandler handles HTTP requests for namespaces
type NamespaceHandler struct {
	namespaceService service.NamespaceServiceInterface
	responseHandler  *ResponseHandler
}

// NewNamespaceHandler creates a new namespace handler
func NewNamespaceHandler(namespaceService service.NamespaceServiceInterface) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: namespaceService,
		responseHandler:  NewResponseHandler(),
	}
}

// CreateNamespace handles POST /v1/namespaces
func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
	var req domain.CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	createdBy, exists := c.Get("client_id")
	if !exists {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	namespace := &domain.Namespace{
		ID:          req.ID,
		Description: req.Description,
		CreatedBy:   createdBy.(string),
	}

	err := h.namespaceService.CreateNamespace(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertNamespaceToResponse(namespace)
	h.responseHandler.Created(c, response)
}

// GetNamespace handles GET /v1/namespaces/{id}
func (h *NamespaceHandler) GetNamespace(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	namespace, err := h.namespaceService.GetNamespace(c.Request.Context(), id)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertNamespaceToResponse(namespace)
	h.responseHandler.OK(c, response)
}

// ListNamespaces handles GET /v1/namespaces
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.namespaceService.ListNamespaces(c.Request.Context())
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	var response []domain.NamespaceResponse
	for _, namespace := range namespaces {
		response = append(response, h.responseHandler.ConvertNamespaceToResponse(namespace))
	}
	h.responseHandler.OK(c, response)
}

// DeleteNamespace handles DELETE /v1/namespaces/{id}
func (h *NamespaceHandler) DeleteNamespace(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	err := h.namespaceService.DeleteNamespace(c.Request.Context(), id)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	h.responseHandler.NoContent(c)
}
