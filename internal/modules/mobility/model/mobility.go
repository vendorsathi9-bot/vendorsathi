package model

import (
	"github.com/google/uuid"
)

type NearbyVendor struct {
	VendorID           uuid.UUID `json:"vendor_id"`
	BusinessName       string    `json:"business_name"`
	Description        *string   `json:"description,omitempty"`
	Phone              string    `json:"phone"`
	Status             string    `json:"status"`
	AvailabilityStatus string    `json:"availability_status"`
	DistanceKm         float64   `json:"distance_km"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	AddressLine1       string    `json:"address_line1"`
	City               string    `json:"city"`
}

type NearbyVendorQuery struct {
	Latitude  float64 `form:"lat" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"lng" binding:"required,min=-180,max=180"`
	RadiusKm  float64 `form:"radius_km,default=5.0" binding:"min=0.1,max=50.0"`
}

// --- Route Generation (OSRM + GeoJSON) ---

type RouteRequest struct {
	OriginLat float64 `form:"origin_lat" binding:"required,min=-90,max=90"`
	OriginLng float64 `form:"origin_lng" binding:"required,min=-180,max=180"`
	DestLat   float64 `form:"dest_lat" binding:"required,min=-90,max=90"`
	DestLng   float64 `form:"dest_lng" binding:"required,min=-180,max=180"`
}

type GeoJSONGeometry struct {
	Type        string      `json:"type"`        // "LineString"
	Coordinates [][]float64 `json:"coordinates"` // [[longitude, latitude], ...]
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteResponse struct {
	RoutingAvailable          bool             `json:"routing_available"`
	Provider                  string           `json:"provider"`
	DistanceType              string           `json:"distance_type"` // "road" or "straight_line_estimate"
	DistanceMeters            float64          `json:"distance_meters"`
	DurationSeconds           float64          `json:"duration_seconds"`
	DistanceKm                float64          `json:"distance_km"`
	DurationMinutes           float64          `json:"duration_minutes"`
	ApproximateStraightLineKm float64          `json:"approximate_straight_line_km,omitempty"`
	Geometry                  *GeoJSONGeometry `json:"geometry,omitempty"`
	Origin                    Coordinates      `json:"origin"`
	Destination               Coordinates      `json:"destination"`
	Attribution               string           `json:"attribution"`
	Message                   string           `json:"message,omitempty"`
	DirectionsURL             string           `json:"directions_url"`
}

// --- Geocoding (Nominatim / OSM) ---

type GeocodeQuery struct {
	Query string `form:"q" binding:"required"`
}

type ReverseGeocodeQuery struct {
	Latitude  float64 `form:"lat" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"lng" binding:"required,min=-180,max=180"`
}

type GeocodeResult struct {
	DisplayName string  `json:"display_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Type        string  `json:"type,omitempty"`
	Attribution string  `json:"attribution"`
}
