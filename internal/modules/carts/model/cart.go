package model

import (
	"time"

	"github.com/google/uuid"
)

type CartStatus string

const (
	CartStatusActive    CartStatus = "ACTIVE"
	CartStatusConverted CartStatus = "CONVERTED"
	CartStatusAbandoned CartStatus = "ABANDONED"
)

type FulfilmentType string

const (
	FulfilmentTypePickup   FulfilmentType = "PICKUP"
	FulfilmentTypeDelivery FulfilmentType = "DELIVERY"
)

// Domain models

type Cart struct {
	ID         uuid.UUID  `db:"id"`
	CustomerID uuid.UUID  `db:"customer_id"`
	VendorID   uuid.UUID  `db:"vendor_id"`
	Status     CartStatus `db:"status"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

type CartItem struct {
	ID        uuid.UUID `db:"id"`
	CartID    uuid.UUID `db:"cart_id"`
	ProductID uuid.UUID `db:"product_id"`
	Quantity  int64     `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Input models

type AddCartItemInput struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int64     `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemInput struct {
	Quantity int64 `json:"quantity" binding:"required,min=1"`
}

// Output models

type CartItemOutput struct {
	ID             uuid.UUID `json:"id"`
	CartID         uuid.UUID `json:"cart_id"`
	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	Quantity       int64     `json:"quantity"`
	UnitPricePaise int64     `json:"unit_price_paise"`
	LineTotalPaise int64     `json:"line_total_paise"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CartOutput struct {
	ID         uuid.UUID        `json:"id"`
	CustomerID uuid.UUID        `json:"customer_id"`
	VendorID   uuid.UUID        `json:"vendor_id"`
	Status     CartStatus       `json:"status"`
	Items      []CartItemOutput `json:"items"`
	UpdatedAt  time.Time        `json:"updated_at"`
}
