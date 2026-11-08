package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/analytics/model"
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

// GetDetailedVendorAnalytics GET /analytics/vendor/detailed
func (h *AnalyticsHandler) GetDetailedVendorAnalytics(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	detailed, err := h.service.GetDetailedVendorAnalytics(c.Request.Context(), vendorUserID)
	if err != nil {
		httpx.ErrorForbidden(c, err.Error())
		return
	}

	httpx.RespondOK(c, detailed)
}

// GetDemandForecast GET /analytics/vendor/demand-forecast
func (h *AnalyticsHandler) GetDemandForecast(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	forecast, err := h.service.GetDemandForecast(c.Request.Context(), vendorUserID)
	if err != nil {
		httpx.ErrorForbidden(c, err.Error())
		return
	}

	httpx.RespondOK(c, forecast)
}

// CreateExpense POST /analytics/expenses
func (h *AnalyticsHandler) CreateExpense(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.CreateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid expense payload", nil)
		return
	}

	expense, err := h.service.CreateExpense(c.Request.Context(), vendorUserID, input)
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondCreated(c, expense)
}

// GetExpenses GET /analytics/expenses
func (h *AnalyticsHandler) GetExpenses(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	expenses, err := h.service.GetExpenses(c.Request.Context(), vendorUserID)
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondOK(c, expenses)
}

// DeleteExpense DELETE /analytics/expenses/:id
func (h *AnalyticsHandler) DeleteExpense(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	expenseIDStr := c.Param("id")
	expenseID, err := uuid.Parse(expenseIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid expense id", nil)
		return
	}

	if err := h.service.DeleteExpense(c.Request.Context(), vendorUserID, expenseID); err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondOK(c, gin.H{"status": "deleted"})
}

// GetWeather GET /analytics/weather
func (h *AnalyticsHandler) GetWeather(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	var lat, lng float64
	if latStr != "" {
		lat, _ = strconv.ParseFloat(latStr, 64)
	}
	if lngStr != "" {
		lng, _ = strconv.ParseFloat(lngStr, 64)
	}

	weather := h.service.GetWeather(c.Request.Context(), lat, lng)
	httpx.RespondOK(c, weather)
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
