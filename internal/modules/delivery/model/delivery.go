package model

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryStatusAssigned       DeliveryStatus = "ASSIGNED"
	DeliveryStatusAccepted       DeliveryStatus = "ACCEPTED"
	DeliveryStatusPickedUp       DeliveryStatus = "PICKED_UP"
	DeliveryStatusOutForDelivery DeliveryStatus = "OUT_FOR_DELIVERY"
	DeliveryStatusDelivered      DeliveryStatus = "DELIVERED"
)

// IsValidTransition validates allowed status progressions
func (s DeliveryStatus) IsValidTransition(next DeliveryStatus) bool {
	allowed := map[DeliveryStatus][]DeliveryStatus{
		DeliveryStatusAssigned:       {DeliveryStatusAccepted},
		DeliveryStatusAccepted:       {DeliveryStatusPickedUp},
		DeliveryStatusPickedUp:       {DeliveryStatusOutForDelivery},
		DeliveryStatusOutForDelivery: {DeliveryStatusDelivered},
		DeliveryStatusDelivered:      {},
	}
	for _, a := range allowed[s] {
		if a == next {
			return true
		}
	}
	return false
}

type DeliveryAssignment struct {
	ID          uuid.UUID      `db:"id"`
	OrderID     uuid.UUID      `db:"order_id"`
	AgentUserID uuid.UUID      `db:"agent_user_id"`
	Status      DeliveryStatus `db:"status"`
	AssignedBy  *uuid.UUID     `db:"assigned_by"`
	AssignedAt  time.Time      `db:"assigned_at"`
	AcceptedAt  *time.Time     `db:"accepted_at"`
	PickedUpAt  *time.Time     `db:"picked_up_at"`
	DeliveredAt *time.Time     `db:"delivered_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type DeliveryAssignmentWithOrder struct {
	DeliveryAssignment
	CustomerName         string  `db:"customer_name"`
	CustomerPhone        string  `db:"customer_phone"`
	DeliveryAddressLine1 string  `db:"delivery_address_line1"`
	DeliveryCity         string  `db:"delivery_city"`
	DeliveryLat          float64 `db:"delivery_lat"`
	DeliveryLng          float64 `db:"delivery_lng"`
	OrderTotalPaise      int64   `db:"order_total_paise"`
	OrderStatus          string  `db:"order_status"`
	VendorName           string  `db:"vendor_name"`
	VendorPhone          string  `db:"vendor_phone"`
	VendorLat            float64 `db:"vendor_lat"`
	VendorLng            float64 `db:"vendor_lng"`
}

type DeliveryTracking struct {
	ID          uuid.UUID `db:"id"`
	OrderID     uuid.UUID `db:"order_id"`
	AgentUserID uuid.UUID `db:"agent_user_id"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// --- Input types ---

type AssignDeliveryInput struct {
	AgentUserID uuid.UUID `json:"agent_user_id" binding:"required"`
}

type UpdateDeliveryStatusInput struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED PICKED_UP OUT_FOR_DELIVERY DELIVERED"`
}

type UpdateLocationInput struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// --- Output types ---

type DeliveryAssignmentOutput struct {
	ID          uuid.UUID      `json:"id"`
	OrderID     uuid.UUID      `json:"order_id"`
	AgentUserID uuid.UUID      `json:"agent_user_id"`
	Status      DeliveryStatus `json:"status"`
	AssignedBy  *uuid.UUID     `json:"assigned_by,omitempty"`
	AssignedAt  time.Time      `json:"assigned_at"`
	AcceptedAt  *time.Time     `json:"accepted_at,omitempty"`
	PickedUpAt  *time.Time     `json:"picked_up_at,omitempty"`
	DeliveredAt *time.Time     `json:"delivered_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type DeliveryAssignmentWithOrderOutput struct {
	DeliveryAssignmentOutput
	CustomerName         string  `json:"customer_name"`
	CustomerPhone        string  `json:"customer_phone"`
	DeliveryAddressLine1 string  `json:"delivery_address_line1"`
	DeliveryCity         string  `json:"delivery_city"`
	DeliveryLat          float64 `json:"delivery_lat"`
	DeliveryLng          float64 `json:"delivery_lng"`
	OrderTotalPaise      int64   `json:"order_total_paise"`
	OrderStatus          string  `json:"order_status"`
	VendorName           string  `json:"vendor_name"`
	VendorPhone          string  `json:"vendor_phone"`
	VendorLat            float64 `json:"vendor_lat"`
	VendorLng            float64 `json:"vendor_lng"`
}

type DeliveryTrackingOutput struct {
	OrderID     uuid.UUID `json:"order_id"`
	AgentUserID uuid.UUID `json:"agent_user_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a *DeliveryAssignment) ToOutput() DeliveryAssignmentOutput {
	return DeliveryAssignmentOutput{
		ID:          a.ID,
		OrderID:     a.OrderID,
		AgentUserID: a.AgentUserID,
		Status:      a.Status,
		AssignedBy:  a.AssignedBy,
		AssignedAt:  a.AssignedAt,
		AcceptedAt:  a.AcceptedAt,
		PickedUpAt:  a.PickedUpAt,
		DeliveredAt: a.DeliveredAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func (a *DeliveryAssignmentWithOrder) ToOutput() DeliveryAssignmentWithOrderOutput {
	return DeliveryAssignmentWithOrderOutput{
		DeliveryAssignmentOutput: a.DeliveryAssignment.ToOutput(),
		CustomerName:             a.CustomerName,
		CustomerPhone:            a.CustomerPhone,
		DeliveryAddressLine1:     a.DeliveryAddressLine1,
		DeliveryCity:             a.DeliveryCity,
		DeliveryLat:              a.DeliveryLat,
		DeliveryLng:              a.DeliveryLng,
		OrderTotalPaise:          a.OrderTotalPaise,
		OrderStatus:              a.OrderStatus,
		VendorName:               a.VendorName,
		VendorPhone:              a.VendorPhone,
		VendorLat:                a.VendorLat,
		VendorLng:                a.VendorLng,
	}
}

func (t *DeliveryTracking) ToOutput() DeliveryTrackingOutput {
	return DeliveryTrackingOutput{
		OrderID:     t.OrderID,
		AgentUserID: t.AgentUserID,
		Latitude:    t.Latitude,
		Longitude:   t.Longitude,
		UpdatedAt:   t.UpdatedAt,
	}
}
