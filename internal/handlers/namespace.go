package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/auth"
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
	Description string `json:"description"`
}

// NamespaceResponse represents a namespace response
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
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
			Details: map[string]interface{}{
				"validation": err.Error(),
			},
		})
		return
	}

	// Get clientId from JWT context
	createdBy := auth.GetClientID(c)
	if createdBy == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Missing client ID",
			Code:  "MISSING_CLIENT_ID",
		})
		return
	}

	namespace := &domain.Namespace{
		ID:          req.ID,
		Description: req.Description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}

	err := h.namespaceService.CreateNamespace(c.Request.Context(), namespace)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "Namespace already exists",
				Code:  "ALREADY_EXISTS",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	// Fetch the created namespace to get the actual database values
	createdNamespace, err := h.namespaceService.GetNamespace(c.Request.Context(), namespace.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"namespace": createdNamespace,
	})
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

	c.JSON(http.StatusOK, h.toNamespaceResponse(namespace))
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

// mapError maps domain errors to HTTP responses
func (h *NamespaceHandler) mapError(err error) (int, ErrorResponse) {
	switch err {
	case domain.ErrNotFound:
		return http.StatusNotFound, ErrorResponse{
			Error: "Namespace not found",
			Code:  "NAMESPACE_NOT_FOUND",
		}
	case domain.ErrAlreadyExists:
		return http.StatusConflict, ErrorResponse{
			Error: "Namespace already exists",
			Code:  "NAMESPACE_ALREADY_EXISTS",
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
		// Check if it's a wrapped validation error
		if isValidationError(err) {
			return http.StatusBadRequest, ErrorResponse{
				Error: err.Error(),
				Code:  "VALIDATION_ERROR",
			}
		}

		return http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		}
	}
}

// isValidationError checks if error is a validation error
func isValidationError(err error) bool {
	return err != nil && (err == domain.ErrValidation ||
		(err.Error() != "" && (contains(err.Error(), "validation failed") ||
			contains(err.Error(), "required") ||
			contains(err.Error(), "too long") ||
			contains(err.Error(), "invalid"))))
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
