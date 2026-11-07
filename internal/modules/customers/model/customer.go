package model

import (
	"time"

	"github.com/google/uuid"
)

type CustomerProfile struct {
	UserID             uuid.UUID `db:"user_id"`
	DisplayName        *string   `db:"display_name"`
	ContactPreferences []byte    `db:"contact_preferences"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type CustomerAddress struct {
	ID           uuid.UUID `db:"id"`
	CustomerID   uuid.UUID `db:"customer_id"`
	AddressLine1 string    `db:"address_line1"`
	AddressLine2 *string   `db:"address_line2"`
	City         string    `db:"city"`
	State        string    `db:"state"`
	PostalCode   *string   `db:"postal_code"`
	Country      string    `db:"country"`
	Latitude     *float64  `db:"latitude"`
	Longitude    *float64  `db:"longitude"`
	Label        *string   `db:"label"`
	IsDefault    bool      `db:"is_default"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type CreateCustomerProfileInput struct {
	DisplayName *string `json:"display_name,omitempty"`
}

type UpdateCustomerProfileInput struct {
	DisplayName        *string `json:"display_name,omitempty"`
	ContactPreferences []byte  `json:"contact_preferences,omitempty"`
}

type CreateCustomerAddressInput struct {
	AddressLine1 string   `json:"address_line1" binding:"required"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         string   `json:"city" binding:"required"`
	State        string   `json:"state" binding:"required"`
	PostalCode   *string  `json:"postal_code,omitempty"`
	Country      *string  `json:"country,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Label        *string  `json:"label,omitempty"`
	IsDefault    *bool    `json:"is_default,omitempty"`
}

type UpdateCustomerAddressInput struct {
	AddressLine1 *string  `json:"address_line1,omitempty"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         *string  `json:"city,omitempty"`
	State        *string  `json:"state,omitempty"`
	PostalCode   *string  `json:"postal_code,omitempty"`
	Country      *string  `json:"country,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Label        *string  `json:"label,omitempty"`
	IsDefault    *bool    `json:"is_default,omitempty"`
}

type CustomerProfileOutput struct {
	UserID             uuid.UUID `json:"user_id"`
	DisplayName        *string   `json:"display_name,omitempty"`
	ContactPreferences []byte    `json:"contact_preferences,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CustomerAddressOutput struct {
	ID           uuid.UUID `json:"id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 *string   `json:"address_line2,omitempty"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	PostalCode   *string   `json:"postal_code,omitempty"`
	Country      string    `json:"country"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	Label        *string   `json:"label,omitempty"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (p *CustomerProfile) ToOutput() CustomerProfileOutput {
	return CustomerProfileOutput{
		UserID:             p.UserID,
		DisplayName:        p.DisplayName,
		ContactPreferences: p.ContactPreferences,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func (a *CustomerAddress) ToOutput() CustomerAddressOutput {
	return CustomerAddressOutput{
		ID:           a.ID,
		CustomerID:   a.CustomerID,
		AddressLine1: a.AddressLine1,
		AddressLine2: a.AddressLine2,
		City:         a.City,
		State:        a.State,
		PostalCode:   a.PostalCode,
		Country:      a.Country,
		Latitude:     a.Latitude,
		Longitude:    a.Longitude,
		Label:        a.Label,
		IsDefault:    a.IsDefault,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}
