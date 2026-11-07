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
