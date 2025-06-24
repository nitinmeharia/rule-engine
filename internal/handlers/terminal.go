package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// TerminalHandler handles HTTP requests for terminals
type TerminalHandler struct {
	terminalService service.TerminalServiceInterface
	responseHandler *ResponseHandler
}

// NewTerminalHandler creates a new terminal handler
func NewTerminalHandler(terminalService service.TerminalServiceInterface) *TerminalHandler {
	return &TerminalHandler{
		terminalService: terminalService,
		responseHandler: NewResponseHandler(),
	}
}

// CreateTerminal handles POST /v1/namespaces/{id}/terminals
func (h *TerminalHandler) CreateTerminal(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	var req domain.CreateTerminalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHandler.BadRequest(c, "Invalid request body")
		return
	}

	// Get client ID from context
	createdBy, exists := c.Get("client_id")
	if !exists || createdBy == nil {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	terminal := &domain.Terminal{
		TerminalID: req.TerminalID,
		CreatedBy:  createdBy.(string),
	}

	err := h.terminalService.CreateTerminal(c.Request.Context(), namespace, terminal)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertTerminalToResponse(terminal)
	c.JSON(http.StatusCreated, response)
}

// GetTerminal handles GET /v1/namespaces/{id}/terminals/{terminalId}
func (h *TerminalHandler) GetTerminal(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	terminalID := c.Param("terminalId")
	if terminalID == "" {
		h.responseHandler.BadRequest(c, "Terminal ID is required")
		return
	}

	terminal, err := h.terminalService.GetTerminal(c.Request.Context(), namespace, terminalID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertTerminalToResponse(terminal)
	c.JSON(http.StatusOK, response)
}

// ListTerminals handles GET /v1/namespaces/{id}/terminals
func (h *TerminalHandler) ListTerminals(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	terminals, err := h.terminalService.ListTerminals(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := make([]domain.TerminalResponse, 0, len(terminals))
	for _, terminal := range terminals {
		response = append(response, h.responseHandler.ConvertTerminalToResponse(terminal))
	}
	c.JSON(http.StatusOK, response)
}

// DeleteTerminal handles DELETE /v1/namespaces/{id}/terminals/{terminalId}
func (h *TerminalHandler) DeleteTerminal(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace ID is required")
		return
	}

	terminalID := c.Param("terminalId")
	if terminalID == "" {
		h.responseHandler.BadRequest(c, "Terminal ID is required")
		return
	}

	err := h.terminalService.DeleteTerminal(c.Request.Context(), namespace, terminalID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	h.responseHandler.NoContent(c)
}
