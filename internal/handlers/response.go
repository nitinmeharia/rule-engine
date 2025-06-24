package handlers

import (
	"net/http"

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
		Error:   http.StatusText(statusCode),
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
	var response []domain.FunctionResponse
	for _, function := range functions {
		response = append(response, h.ConvertFunctionToResponse(function))
	}
	return response
}
