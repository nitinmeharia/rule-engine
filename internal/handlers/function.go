package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// FunctionHandler handles HTTP requests for functions
type FunctionHandler struct {
	functionService *service.FunctionService
	responseHandler *ResponseHandler
}

// NewFunctionHandler creates a new function handler
func NewFunctionHandler(functionService *service.FunctionService) *FunctionHandler {
	return &FunctionHandler{
		functionService: functionService,
		responseHandler: NewResponseHandler(),
	}
}

// CreateFunction handles POST /v1/namespaces/{namespace}/functions
func (h *FunctionHandler) CreateFunction(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	var req domain.CreateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	function := &domain.Function{
		FunctionID: req.ID,
		Type:       req.Type,
		Args:       req.Args,
		Values:     req.Values,
		CreatedBy:  clientID.(string),
	}

	err := h.functionService.CreateFunction(c.Request.Context(), namespace, function)
	if err != nil {
		switch err {
		case domain.ErrFunctionAlreadyExists:
			h.responseHandler.Conflict(c, "Function already exists")
		case domain.ErrInvalidFunctionID:
			h.responseHandler.BadRequest(c, "Invalid function ID")
		case domain.ErrInvalidFunctionType:
			h.responseHandler.BadRequest(c, "Invalid function type")
		case domain.ErrInvalidFunctionArgs:
			h.responseHandler.BadRequest(c, "Invalid function arguments")
		default:
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	response := domain.CreateFunctionResponse{
		Status:   "draft",
		Function: h.responseHandler.ConvertFunctionToResponse(function),
	}
	h.responseHandler.Created(c, response)
}

// GetFunction handles GET /v1/namespaces/{namespace}/functions/{functionId}
func (h *FunctionHandler) GetFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		h.responseHandler.BadRequest(c, "Namespace and function ID are required")
		return
	}

	function, err := h.functionService.GetFunction(c.Request.Context(), namespace, functionID)
	if err != nil {
		if err == domain.ErrFunctionNotFound {
			h.responseHandler.NotFound(c, "Function not found")
		} else {
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	response := h.responseHandler.ConvertFunctionToResponse(function)
	h.responseHandler.OK(c, response)
}

// ListFunctions handles GET /v1/namespaces/{namespace}/functions
func (h *FunctionHandler) ListFunctions(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	functions, err := h.functionService.ListFunctions(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.InternalServerError(c, "Internal server error")
		return
	}

	response := h.responseHandler.ConvertFunctionsToResponse(functions)
	h.responseHandler.OK(c, response)
}

// UpdateFunction handles PUT /v1/namespaces/{namespace}/functions/{functionId}/versions/draft
func (h *FunctionHandler) UpdateFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		h.responseHandler.BadRequest(c, "Namespace and function ID are required")
		return
	}

	var req domain.UpdateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	function := &domain.Function{
		Type:      req.Type,
		Args:      req.Args,
		Values:    req.Values,
		CreatedBy: clientID.(string),
	}

	err := h.functionService.UpdateFunction(c.Request.Context(), namespace, functionID, function)
	if err != nil {
		switch err {
		case domain.ErrFunctionNotFound:
			h.responseHandler.NotFound(c, "Function not found")
		case domain.ErrInvalidFunctionType:
			h.responseHandler.BadRequest(c, "Invalid function type")
		case domain.ErrInvalidFunctionArgs:
			h.responseHandler.BadRequest(c, "Invalid function arguments")
		default:
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	response := domain.UpdateFunctionResponse{
		Function: h.responseHandler.ConvertFunctionToResponse(function),
	}
	h.responseHandler.OK(c, response)
}

// PublishFunction handles POST /v1/namespaces/{namespace}/functions/{functionId}/publish
func (h *FunctionHandler) PublishFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		h.responseHandler.BadRequest(c, "Namespace and function ID are required")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	err := h.functionService.PublishFunction(c.Request.Context(), namespace, functionID, clientID.(string))
	if err != nil {
		if err == domain.ErrFunctionNotFound {
			h.responseHandler.NotFound(c, "Function not found")
		} else {
			h.responseHandler.InternalServerError(c, "Internal server error")
		}
		return
	}

	// Get the published function to return in response
	function, err := h.functionService.GetFunction(c.Request.Context(), namespace, functionID)
	if err != nil {
		h.responseHandler.InternalServerError(c, "Internal server error")
		return
	}

	response := h.responseHandler.ConvertFunctionToResponse(function)
	h.responseHandler.OK(c, response)
}
