package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/rule-engine/internal/execution"
)

// EngineInterface defines the interface for the execution engine
type EngineInterface interface {
	GetCacheInfo(namespace string) (*execution.CacheInfo, error)
	ExecuteRule(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error)
	ExecuteWorkflow(ctx context.Context, req *domain.ExecutionRequest) (*domain.ExecutionResponse, error)
}

// AdminHandler handles admin-only HTTP requests
type AdminHandler struct {
	engine          EngineInterface
	responseHandler *ResponseHandler
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(engine EngineInterface) *AdminHandler {
	return &AdminHandler{
		engine:          engine,
		responseHandler: NewResponseHandler(),
	}
}

// CacheStatsResponse represents the response for cache stats
type CacheStatsResponse struct {
	Namespace     string    `json:"namespace"`
	Checksum      string    `json:"checksum"`
	LastRefresh   time.Time `json:"lastRefresh"`
	StalenessSecs float64   `json:"stalenessSeconds"`
	CacheStatus   string    `json:"cacheStatus"`
	IsStale       bool      `json:"isStale"`
}

// GetCacheStats returns cache statistics for a namespace
func (h *AdminHandler) GetCacheStats(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		h.responseHandler.BadRequest(c, "Namespace is required")
		return
	}

	// Get cache info from engine
	cacheInfo, err := h.engine.GetCacheInfo(namespace)
	if err != nil {
		h.responseHandler.MapDomainErrorToResponse(c, err)
		return
	}

	// Calculate staleness
	staleness := time.Since(cacheInfo.LastRefresh).Seconds()
	isStale := staleness > 300 // Consider stale if older than 5 minutes

	response := CacheStatsResponse{
		Namespace:     namespace,
		Checksum:      cacheInfo.Checksum,
		LastRefresh:   cacheInfo.LastRefresh,
		StalenessSecs: staleness,
		CacheStatus:   h.getCacheStatus(cacheInfo, isStale),
		IsStale:       isStale,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// getCacheStatus determines the cache status based on staleness and other factors
func (h *AdminHandler) getCacheStatus(cacheInfo *execution.CacheInfo, isStale bool) string {
	if cacheInfo.LastRefresh.IsZero() {
		return "not_initialized"
	}
	if isStale {
		return "stale"
	}
	return "fresh"
}
