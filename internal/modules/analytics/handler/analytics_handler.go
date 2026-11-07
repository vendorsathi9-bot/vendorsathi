package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/analytics/service"
	"vendorsaathi/internal/platform/httpx"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: svc}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	if id, ok := val.(uuid.UUID); ok {
		return id, true
	}
	if str, ok := val.(string); ok {
		id, err := uuid.Parse(str)
		if err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

// GetVendorSummary GET /analytics/vendor/summary (Vendor only)
func (h *AnalyticsHandler) GetVendorSummary(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	summary, err := h.service.GetVendorSummary(c.Request.Context(), vendorUserID)
	if err != nil {
		httpx.ErrorForbidden(c, err.Error())
		return
	}

	httpx.RespondOK(c, summary)
}

// GetPlatformOverview GET /analytics/platform/overview (Admin only)
func (h *AnalyticsHandler) GetPlatformOverview(c *gin.Context) {
	overview, err := h.service.GetPlatformOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	httpx.RespondOK(c, overview)
}
