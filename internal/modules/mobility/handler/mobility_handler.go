package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/config"
	"vendorsaathi/internal/modules/mobility/model"
	"vendorsaathi/internal/modules/mobility/service"
	"vendorsaathi/internal/platform/httpx"
)

type MobilityHandler struct {
	service *service.MobilityService
	mapCfg  config.MapConfig
}

func NewMobilityHandler(svc *service.MobilityService, mapCfg config.MapConfig) *MobilityHandler {
	return &MobilityHandler{service: svc, mapCfg: mapCfg}
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

// GetRoute GET /mobility/route
func (h *MobilityHandler) GetRoute(c *gin.Context) {
	var req model.RouteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.ErrorBadRequest(c, "invalid coordinates or query parameters: "+err.Error(), nil)
		return
	}

	route, err := h.service.GetRoute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "routing failed: " + err.Error()})
		return
	}

	httpx.RespondOK(c, route)
}

// GetMapConfig GET /mobility/config
func (h *MobilityHandler) GetMapConfig(c *gin.Context) {
	httpx.RespondOK(c, gin.H{
		"provider":           h.mapCfg.Provider,
		"tile_url":           h.mapCfg.TileURL,
		"routing_provider":   h.mapCfg.RoutingProvider,
		"osrm_base_url":      h.mapCfg.OSRMBaseURL,
		"geocoding_provider": h.mapCfg.GeocodingProvider,
		"geocoding_base_url": h.mapCfg.GeocodingBaseURL,
		"attribution":        "© OpenStreetMap contributors, MapLibre, OSRM",
	})
}

// Geocode GET /mobility/geocode
func (h *MobilityHandler) Geocode(c *gin.Context) {
	var query model.GeocodeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpx.ErrorBadRequest(c, "query parameter 'q' is required", nil)
		return
	}

	res, err := h.service.Geocode(c.Request.Context(), query.Query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	httpx.RespondOK(c, res)
}

// ReverseGeocode GET /mobility/reverse-geocode
func (h *MobilityHandler) ReverseGeocode(c *gin.Context) {
	var query model.ReverseGeocodeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	res, err := h.service.ReverseGeocode(c.Request.Context(), query.Latitude, query.Longitude)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	httpx.RespondOK(c, res)
}
