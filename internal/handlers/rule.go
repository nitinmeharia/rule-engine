package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/service"
)

// RuleHandler handles HTTP requests for rules
type RuleHandler struct {
	ruleService     *service.RuleService
	responseHandler *ResponseHandler
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(ruleService *service.RuleService) *RuleHandler {
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
	if !exists {
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
	h.responseHandler.Created(c, response)
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
	h.responseHandler.OK(c, response)
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
	h.responseHandler.OK(c, response)
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
	h.responseHandler.OK(c, response)
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
	h.responseHandler.OK(c, response)
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
	if !exists {
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
	h.responseHandler.OK(c, response)
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
	if !exists {
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
	h.responseHandler.OK(c, response)
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

	h.responseHandler.NoContent(c)
}
