package model

import (
	"time"

	"github.com/google/uuid"
)

type VendorStatus string

const (
	VendorStatusPending   VendorStatus = "PENDING"
	VendorStatusActive    VendorStatus = "ACTIVE"
	VendorStatusSuspended VendorStatus = "SUSPENDED"
	VendorStatusClosed    VendorStatus = "CLOSED"
)

type AvailabilityStatus string

const (
	AvailabilityOpen                   AvailabilityStatus = "OPEN"
	AvailabilityClosed                 AvailabilityStatus = "CLOSED"
	AvailabilityTemporarilyUnavailable AvailabilityStatus = "TEMPORARILY_UNAVAILABLE"
)

type VendorProfile struct {
	ID                 uuid.UUID          `db:"id"`
	OwnerUserID        uuid.UUID          `db:"owner_user_id"`
	BusinessName       string             `db:"business_name"`
	Description        *string            `db:"description"`
	Phone              string             `db:"phone"`
	Status             VendorStatus       `db:"status"`
	AvailabilityStatus AvailabilityStatus `db:"availability_status"`
	CreatedAt          time.Time          `db:"created_at"`
	UpdatedAt          time.Time          `db:"updated_at"`
}

type VendorLocation struct {
	ID           uuid.UUID `db:"id"`
	VendorID     uuid.UUID `db:"vendor_id"`
	AddressLine1 string    `db:"address_line1"`
	AddressLine2 *string   `db:"address_line2"`
	City         string    `db:"city"`
	State        string    `db:"state"`
	PostalCode   *string   `db:"postal_code"`
	Country      string    `db:"country"`
	Latitude     float64   `db:"latitude"`
	Longitude    float64   `db:"longitude"`
	IsPrimary    bool      `db:"is_primary"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type BusinessHour struct {
	ID        uuid.UUID `db:"id"`
	VendorID  uuid.UUID `db:"vendor_id"`
	Weekday   int       `db:"weekday"`
	OpenTime  *string   `db:"open_time"`
	CloseTime *string   `db:"close_time"`
	IsClosed  bool      `db:"is_closed"`
}

type CreateVendorProfileInput struct {
	BusinessName string  `json:"business_name" binding:"required"`
	Description  *string `json:"description,omitempty"`
	Phone        string  `json:"phone" binding:"required"`
}

type UpdateVendorProfileInput struct {
	BusinessName       *string             `json:"business_name,omitempty"`
	Description        *string             `json:"description,omitempty"`
	Phone              *string             `json:"phone,omitempty"`
	AvailabilityStatus *AvailabilityStatus `json:"availability_status,omitempty"`
}

type UpdateVendorStatusInput struct {
	Status VendorStatus `json:"status" binding:"required,oneof=PENDING ACTIVE SUSPENDED CLOSED"`
}

type CreateVendorLocationInput struct {
	AddressLine1 string  `json:"address_line1" binding:"required"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city" binding:"required"`
	State        string  `json:"state" binding:"required"`
	PostalCode   *string `json:"postal_code,omitempty"`
	Country      *string `json:"country,omitempty"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	IsPrimary    *bool   `json:"is_primary,omitempty"`
}

type UpdateVendorLocationInput struct {
	AddressLine1 *string  `json:"address_line1,omitempty"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         *string  `json:"city,omitempty"`
	State        *string  `json:"state,omitempty"`
	PostalCode   *string  `json:"postal_code,omitempty"`
	Country      *string  `json:"country,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	IsPrimary    *bool    `json:"is_primary,omitempty"`
}

type BusinessHourInput struct {
	Weekday   int     `json:"weekday" binding:"required,min=0,max=6"`
	OpenTime  *string `json:"open_time,omitempty"`
	CloseTime *string `json:"close_time,omitempty"`
	IsClosed  *bool   `json:"is_closed,omitempty"`
}

type SetBusinessHoursInput struct {
	Hours []BusinessHourInput `json:"hours" binding:"required,min=1,max=7"`
}

type VendorProfileOutput struct {
	ID                 uuid.UUID          `json:"id"`
	OwnerUserID        uuid.UUID          `json:"owner_user_id"`
	BusinessName       string             `json:"business_name"`
	Description        *string            `json:"description,omitempty"`
	Phone              string             `json:"phone"`
	Status             VendorStatus       `json:"status"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type VendorLocationOutput struct {
	ID           uuid.UUID `json:"id"`
	VendorID     uuid.UUID `json:"vendor_id"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 *string   `json:"address_line2,omitempty"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	PostalCode   *string   `json:"postal_code,omitempty"`
	Country      string    `json:"country"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	IsPrimary    bool      `json:"is_primary"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BusinessHourOutput struct {
	ID        uuid.UUID `json:"id"`
	VendorID  uuid.UUID `json:"vendor_id"`
	Weekday   int       `json:"weekday"`
	OpenTime  *string   `json:"open_time,omitempty"`
	CloseTime *string   `json:"close_time,omitempty"`
	IsClosed  bool      `json:"is_closed"`
}

type VendorDetailOutput struct {
	Profile       VendorProfileOutput    `json:"profile"`
	Locations     []VendorLocationOutput `json:"locations"`
	BusinessHours []BusinessHourOutput   `json:"business_hours"`
}

func (v *VendorProfile) ToOutput() VendorProfileOutput {
	return VendorProfileOutput{
		ID:                 v.ID,
		OwnerUserID:        v.OwnerUserID,
		BusinessName:       v.BusinessName,
		Description:        v.Description,
		Phone:              v.Phone,
		Status:             v.Status,
		AvailabilityStatus: v.AvailabilityStatus,
		CreatedAt:          v.CreatedAt,
		UpdatedAt:          v.UpdatedAt,
	}
}

func (l *VendorLocation) ToOutput() VendorLocationOutput {
	return VendorLocationOutput{
		ID:           l.ID,
		VendorID:     l.VendorID,
		AddressLine1: l.AddressLine1,
		AddressLine2: l.AddressLine2,
		City:         l.City,
		State:        l.State,
		PostalCode:   l.PostalCode,
		Country:      l.Country,
		Latitude:     l.Latitude,
		Longitude:    l.Longitude,
		IsPrimary:    l.IsPrimary,
		UpdatedAt:    l.UpdatedAt,
	}
}

func (bh *BusinessHour) ToOutput() BusinessHourOutput {
	return BusinessHourOutput{
		ID:        bh.ID,
		VendorID:  bh.VendorID,
		Weekday:   bh.Weekday,
		OpenTime:  bh.OpenTime,
		CloseTime: bh.CloseTime,
		IsClosed:  bh.IsClosed,
	}
}
