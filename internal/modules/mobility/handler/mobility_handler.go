package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	attribution := h.mapCfg.Attribution
	if attribution == "" {
		attribution = config.DefaultMapAttribution
	}
	styleURL := h.mapCfg.StyleURL
	if styleURL == "" {
		styleURL = config.DefaultMapStyleURL
	}
	httpx.RespondOK(c, gin.H{
		"provider":           h.mapCfg.Provider,
		"style_url":          styleURL,
		"tile_url":           h.mapCfg.TileURL,
		"attribution":        attribution,
		"routing_provider":   h.mapCfg.RoutingProvider,
		"osrm_base_url":      h.mapCfg.OSRMBaseURL,
		"geocoding_provider": h.mapCfg.GeocodingProvider,
		"geocoding_base_url": h.mapCfg.GeocodingBaseURL,
	})
}

// ProxyTile GET /mobility/tiles/:z/:x/:y
func (h *MobilityHandler) ProxyTile(c *gin.Context) {
	zStr := c.Param("z")
	xStr := c.Param("x")
	yStr := strings.TrimSuffix(c.Param("y"), ".png")

	z, err := strconv.Atoi(zStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoom coordinate"})
		return
	}
	x, err := strconv.Atoi(xStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid x coordinate"})
		return
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid y coordinate"})
		return
	}

	ifNoneMatch := c.GetHeader("If-None-Match")

	data, contentType, etag, notModified, err := h.service.GetTile(c.Request.Context(), z, x, y, ifNoneMatch)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "429") {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if notModified {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")
	c.Header("ETag", etag)
	c.Data(http.StatusOK, contentType, data)
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
