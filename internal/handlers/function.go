package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// FunctionHandler handles HTTP requests for functions
type FunctionHandler struct {
	functionService *service.FunctionService
}

// NewFunctionHandler creates a new function handler
func NewFunctionHandler(functionService *service.FunctionService) *FunctionHandler {
	return &FunctionHandler{
		functionService: functionService,
	}
}

// CreateFunction handles POST /v1/namespaces/{namespace}/functions
func (h *FunctionHandler) CreateFunction(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace is required"})
		return
	}

	var req struct {
		ID     string   `json:"id" binding:"required"`
		Type   string   `json:"type" binding:"required"`
		Args   []string `json:"args,omitempty"`
		Values []string `json:"values,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Client ID not found"})
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
			c.JSON(http.StatusConflict, gin.H{"error": "Function already exists"})
		case domain.ErrInvalidFunctionID:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		case domain.ErrInvalidFunctionType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function type"})
		case domain.ErrInvalidFunctionArgs:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function arguments"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "draft",
		"function": gin.H{
			"id":         function.FunctionID,
			"version":    function.Version,
			"status":     function.Status,
			"type":       function.Type,
			"args":       function.Args,
			"values":     function.Values,
			"returnType": function.ReturnType,
			"createdAt":  function.CreatedAt,
			"createdBy":  function.CreatedBy,
		},
	})
}

// GetFunction handles GET /v1/namespaces/{namespace}/functions/{functionId}
func (h *FunctionHandler) GetFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and function ID are required"})
		return
	}

	function, err := h.functionService.GetFunction(c.Request.Context(), namespace, functionID)
	if err != nil {
		if err == domain.ErrFunctionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          function.FunctionID,
		"version":     function.Version,
		"status":      function.Status,
		"type":        function.Type,
		"args":        function.Args,
		"values":      function.Values,
		"returnType":  function.ReturnType,
		"createdAt":   function.CreatedAt,
		"createdBy":   function.CreatedBy,
		"publishedAt": function.PublishedAt,
		"publishedBy": function.PublishedBy,
	})
}

// ListFunctions handles GET /v1/namespaces/{namespace}/functions
func (h *FunctionHandler) ListFunctions(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace is required"})
		return
	}

	functions, err := h.functionService.ListFunctions(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var response []gin.H
	for _, function := range functions {
		response = append(response, gin.H{
			"id":          function.FunctionID,
			"version":     function.Version,
			"status":      function.Status,
			"type":        function.Type,
			"args":        function.Args,
			"values":      function.Values,
			"returnType":  function.ReturnType,
			"createdAt":   function.CreatedAt,
			"createdBy":   function.CreatedBy,
			"publishedAt": function.PublishedAt,
			"publishedBy": function.PublishedBy,
		})
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFunction handles PUT /v1/namespaces/{namespace}/functions/{functionId}/versions/draft
func (h *FunctionHandler) UpdateFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and function ID are required"})
		return
	}

	var req struct {
		Type   string   `json:"type" binding:"required"`
		Args   []string `json:"args,omitempty"`
		Values []string `json:"values,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Client ID not found"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		case domain.ErrInvalidFunctionType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function type"})
		case domain.ErrInvalidFunctionArgs:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function arguments"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"function": gin.H{
			"id":         function.FunctionID,
			"version":    function.Version,
			"status":     function.Status,
			"type":       function.Type,
			"args":       function.Args,
			"values":     function.Values,
			"returnType": function.ReturnType,
			"createdAt":  function.CreatedAt,
			"createdBy":  function.CreatedBy,
		},
	})
}

// PublishFunction handles POST /v1/namespaces/{namespace}/functions/{functionId}/publish
func (h *FunctionHandler) PublishFunction(c *gin.Context) {
	namespace := c.Param("id")
	functionID := c.Param("functionId")

	if namespace == "" || functionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and function ID are required"})
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Client ID not found"})
		return
	}

	err := h.functionService.PublishFunction(c.Request.Context(), namespace, functionID, clientID.(string))
	if err != nil {
		switch err {
		case domain.ErrFunctionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Get the published function
	function, err := h.functionService.GetFunction(c.Request.Context(), namespace, functionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "active",
		"function": gin.H{
			"id":          function.FunctionID,
			"version":     function.Version,
			"status":      function.Status,
			"type":        function.Type,
			"args":        function.Args,
			"values":      function.Values,
			"returnType":  function.ReturnType,
			"createdAt":   function.CreatedAt,
			"createdBy":   function.CreatedBy,
			"publishedAt": function.PublishedAt,
			"publishedBy": function.PublishedBy,
		},
	})
}
