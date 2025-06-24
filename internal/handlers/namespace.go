package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// NamespaceHandler handles namespace HTTP requests
type NamespaceHandler struct {
	namespaceService service.NamespaceServiceInterface
}

// NewNamespaceHandler creates a new namespace handler
func NewNamespaceHandler(namespaceService service.NamespaceServiceInterface) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: namespaceService,
	}
}

// CreateNamespaceRequest represents the request body for creating a namespace
type CreateNamespaceRequest struct {
	ID          string `json:"id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// CreateNamespaceResponse represents the response for creating a namespace
type CreateNamespaceResponse struct {
	Success   bool              `json:"success"`
	Namespace NamespaceResponse `json:"namespace"`
}

// NamespaceResponse represents a namespace in API responses
type NamespaceResponse struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CreateNamespace creates a new namespace
func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
	var req CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrValidationError)
		return
	}

	// Get createdBy from JWT context
	createdBy, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, domain.ErrInternalError)
		return
	}

	namespace := &domain.Namespace{
		ID:          req.ID,
		Description: req.Description,
		CreatedBy:   createdBy.(string),
	}

	err := h.namespaceService.CreateNamespace(c.Request.Context(), namespace)
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	// Get the created namespace to return in response
	createdNamespace, err := h.namespaceService.GetNamespace(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrInternalError)
		return
	}

	response := CreateNamespaceResponse{
		Success: true,
		Namespace: NamespaceResponse{
			ID:          createdNamespace.ID,
			Description: createdNamespace.Description,
			CreatedAt:   createdNamespace.CreatedAt,
			CreatedBy:   createdNamespace.CreatedBy,
		},
	}

	c.JSON(http.StatusCreated, response)
}

// GetNamespace retrieves a namespace by ID
func (h *NamespaceHandler) GetNamespace(c *gin.Context) {
	id := c.Param("id")

	namespace, err := h.namespaceService.GetNamespace(c.Request.Context(), id)
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.JSON(http.StatusOK, namespace)
}

// ListNamespaces retrieves all namespaces
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.namespaceService.ListNamespaces(c.Request.Context())
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	response := make([]NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		response[i] = h.toNamespaceResponse(ns)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteNamespace deletes a namespace
func (h *NamespaceHandler) DeleteNamespace(c *gin.Context) {
	id := c.Param("id")

	err := h.namespaceService.DeleteNamespace(c.Request.Context(), id)
	if err != nil {
		statusCode, errorResp := h.mapError(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

// toNamespaceResponse converts domain namespace to response
func (h *NamespaceHandler) toNamespaceResponse(namespace *domain.Namespace) NamespaceResponse {
	return NamespaceResponse{
		ID:          namespace.ID,
		Description: namespace.Description,
		CreatedAt:   namespace.CreatedAt,
		CreatedBy:   namespace.CreatedBy,
	}
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
