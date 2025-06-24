package handlers

import (
	"net/http"
	"strings"

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
		switch err {
		case domain.ErrNamespaceAlreadyExists:
			h.responseHandler.Conflict(c, "Namespace already exists")
		case domain.ErrInvalidNamespaceID:
			h.responseHandler.BadRequest(c, "Invalid namespace ID")
		case domain.ErrInvalidDescription:
			h.responseHandler.BadRequest(c, "Invalid description")
		default:
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
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
		if err == domain.ErrNamespaceNotFound {
			h.responseHandler.NotFound(c, "Namespace not found")
		} else {
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	response := h.responseHandler.ConvertNamespaceToResponse(namespace)
	h.responseHandler.OK(c, response)
}

// ListNamespaces handles GET /v1/namespaces
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.namespaceService.ListNamespaces(c.Request.Context())
	if err != nil {
		h.responseHandler.InternalServerError(c, "Internal server error")
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
		if err == domain.ErrNamespaceNotFound {
			h.responseHandler.NotFound(c, "Namespace not found")
		} else {
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	h.responseHandler.NoContent(c)
}

// mapError maps domain errors to appropriate HTTP responses
func (h *NamespaceHandler) mapError(err error) (int, interface{}) {
	apiErr, ok := err.(*domain.APIError)
	if ok {
		return apiErr.HTTPStatus(), apiErr
	}

	// Fallback for non-APIError errors
	return http.StatusInternalServerError, domain.ErrInternalError
}

// isValidationError checks if an error is a validation error
func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an APIError with validation-related codes
	if apiErr, ok := err.(*domain.APIError); ok {
		return apiErr.Code == domain.ErrCodeValidationError ||
			apiErr.Code == domain.ErrCodeInvalidNamespaceID ||
			apiErr.Code == domain.ErrCodeInvalidFieldID ||
			apiErr.Code == domain.ErrCodeInvalidDescription
	}

	// Fallback to string checking for non-APIError errors
	errStr := err.Error()
	return strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "binding") ||
		strings.Contains(errStr, "required") ||
		strings.Contains(errStr, "too long") ||
		strings.Contains(errStr, "invalid")
}

// contains checks if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					indexOf(str, substr) != -1)))
}

// indexOf returns the index of substr in str, or -1 if not found
func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
