package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/modules/mobility/model"
	"vendorsaathi/internal/modules/mobility/service"
	"vendorsaathi/internal/platform/httpx"
)

type MobilityHandler struct {
	service *service.MobilityService
}

func NewMobilityHandler(svc *service.MobilityService) *MobilityHandler {
	return &MobilityHandler{service: svc}
}

// FindNearbyVendors GET /mobility/nearby-vendors
func (h *MobilityHandler) FindNearbyVendors(c *gin.Context) {
	var query model.NearbyVendorQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	vendors, err := h.service.FindNearbyVendors(c.Request.Context(), query.Latitude, query.Longitude, query.RadiusKm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	httpx.RespondOK(c, vendors)
}
