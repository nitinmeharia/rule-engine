package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// RuleHandler handles HTTP requests for rules
type RuleHandler struct {
	ruleService     service.RuleServiceInterface
	responseHandler *ResponseHandler
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(ruleService service.RuleServiceInterface) *RuleHandler {
	return &RuleHandler{
		ruleService:     ruleService,
		responseHandler: NewResponseHandler(),
	}
}

// CreateRule handles POST /v1/namespaces/{namespace}/rules
func (h *RuleHandler) CreateRule(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	var req domain.CreateRuleRequest
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

	rule := &domain.Rule{
		RuleID:     req.ID,
		Logic:      req.Logic,
		Conditions: req.Conditions,
		CreatedBy:  clientID.(string),
	}

	err := h.ruleService.CreateRule(c.Request.Context(), namespace, rule)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := domain.CreateRuleResponse{
		Status: "draft",
		Rule:   h.responseHandler.ConvertRuleToResponse(rule),
	}
	c.JSON(http.StatusCreated, response)
}

// GetRule handles GET /v1/namespaces/{namespace}/rules/{ruleId}
func (h *RuleHandler) GetRule(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")

	if namespace == "" || ruleID == "" {
		h.responseHandler.BadRequest(c, "Namespace and rule ID are required")
		return
	}

	rule, err := h.ruleService.GetRule(c.Request.Context(), namespace, ruleID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertRuleToResponse(rule)
	c.JSON(http.StatusOK, response)
}

// GetDraftRule handles GET /v1/namespaces/{namespace}/rules/{ruleId}/versions/draft
func (h *RuleHandler) GetDraftRule(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")

	if namespace == "" || ruleID == "" {
		h.responseHandler.BadRequest(c, "Namespace and rule ID are required")
		return
	}

	rule, err := h.ruleService.GetDraftRule(c.Request.Context(), namespace, ruleID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertRuleToResponse(rule)
	c.JSON(http.StatusOK, response)
}

// ListRules handles GET /v1/namespaces/{namespace}/rules
func (h *RuleHandler) ListRules(c *gin.Context) {
	namespace := c.Param("id")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	rules, err := h.ruleService.ListRules(c.Request.Context(), namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertRulesToResponse(rules)
	c.JSON(http.StatusOK, response)
}

// ListRuleVersions handles GET /v1/namespaces/{namespace}/rules/{ruleId}/history
func (h *RuleHandler) ListRuleVersions(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")

	if namespace == "" || ruleID == "" {
		h.responseHandler.BadRequest(c, "Namespace and rule ID are required")
		return
	}

	rules, err := h.ruleService.ListRuleVersions(c.Request.Context(), namespace, ruleID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertRuleVersionsToResponse(rules)
	c.JSON(http.StatusOK, response)
}

// UpdateRule handles PUT /v1/namespaces/{namespace}/rules/{ruleId}/versions/draft
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")

	if namespace == "" || ruleID == "" {
		h.responseHandler.BadRequest(c, "Namespace and rule ID are required")
		return
	}

	var req domain.UpdateRuleRequest
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

	rule := &domain.Rule{
		Logic:      req.Logic,
		Conditions: req.Conditions,
		CreatedBy:  clientID.(string),
	}

	err := h.ruleService.UpdateRule(c.Request.Context(), namespace, ruleID, rule)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	// Get the updated rule
	updatedRule, err := h.ruleService.GetDraftRule(c.Request.Context(), namespace, ruleID)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := h.responseHandler.ConvertRuleToResponse(updatedRule)
	c.JSON(http.StatusOK, response)
}

// PublishRule handles POST /v1/namespaces/{namespace}/rules/{ruleId}/publish
func (h *RuleHandler) PublishRule(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")

	if namespace == "" || ruleID == "" {
		h.responseHandler.BadRequest(c, "Namespace and rule ID are required")
		return
	}

	// Get client ID from context
	clientID, exists := c.Get("client_id")
	if !exists || clientID == nil {
		h.responseHandler.Unauthorized(c, "Client ID not found")
		return
	}

	err := h.ruleService.PublishRule(c.Request.Context(), namespace, ruleID, clientID.(string))
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	response := domain.PublishRuleResponse{
		Status: "active",
	}
	c.JSON(http.StatusOK, response)
}

// DeleteRule handles DELETE /v1/namespaces/{namespace}/rules/{ruleId}/versions/{version}
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	namespace := c.Param("id")
	ruleID := c.Param("ruleId")
	versionStr := c.Param("version")

	if namespace == "" || ruleID == "" || versionStr == "" {
		h.responseHandler.BadRequest(c, "Namespace, rule ID, and version are required")
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		h.responseHandler.BadRequest(c, "Invalid version number")
		return
	}

	err = h.ruleService.DeleteRule(c.Request.Context(), namespace, ruleID, int32(version))
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
