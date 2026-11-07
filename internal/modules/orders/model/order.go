package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending               OrderStatus = "PENDING"
	OrderStatusAccepted              OrderStatus = "ACCEPTED"
	OrderStatusPreparing             OrderStatus = "PREPARING"
	OrderStatusReadyForPickup        OrderStatus = "READY_FOR_PICKUP"
	OrderStatusOutForDelivery        OrderStatus = "OUT_FOR_DELIVERY"
	OrderStatusCompleted             OrderStatus = "COMPLETED"
	OrderStatusRejected              OrderStatus = "REJECTED"
	OrderStatusCancelled             OrderStatus = "CANCELLED"
	OrderStatusCancellationRequested OrderStatus = "CANCELLATION_REQUESTED"
)

type FulfilmentType string

const (
	FulfilmentTypePickup   FulfilmentType = "PICKUP"
	FulfilmentTypeDelivery FulfilmentType = "DELIVERY"
)

type CancellationDecision string

const (
	CancellationDecisionApproved CancellationDecision = "APPROVED"
	CancellationDecisionDeclined CancellationDecision = "DECLINED"
)

// Domain models

type Order struct {
	ID                      uuid.UUID      `db:"id"`
	CustomerID              uuid.UUID      `db:"customer_id"`
	VendorID                uuid.UUID      `db:"vendor_id"`
	FulfilmentType          FulfilmentType `db:"fulfilment_type"`
	Status                  OrderStatus    `db:"status"`
	SubtotalPaise           int64          `db:"subtotal_paise"`
	TotalPaise              int64          `db:"total_paise"`
	DeliveryAddressSnapshot *string        `db:"delivery_address_snapshot"`
	PickupAddressSnapshot   *string        `db:"pickup_address_snapshot"`
	IdempotencyKey          *string        `db:"idempotency_key"`
	CreatedAt               time.Time      `db:"created_at"`
	UpdatedAt               time.Time      `db:"updated_at"`
}

type OrderItem struct {
	ID                  uuid.UUID  `db:"id"`
	OrderID             uuid.UUID  `db:"order_id"`
	ProductID           *uuid.UUID `db:"product_id"`
	ProductNameSnapshot string     `db:"product_name_snapshot"`
	UnitPricePaise      int64      `db:"unit_price_paise"`
	Quantity            int64      `db:"quantity"`
	LineTotalPaise      int64      `db:"line_total_paise"`
	CreatedAt           time.Time  `db:"created_at"`
}

type OrderStatusHistory struct {
	ID          uuid.UUID  `db:"id"`
	OrderID     uuid.UUID  `db:"order_id"`
	FromStatus  *string    `db:"from_status"`
	ToStatus    string     `db:"to_status"`
	ActorUserID *uuid.UUID `db:"actor_user_id"`
	Reason      *string    `db:"reason"`
	CreatedAt   time.Time  `db:"created_at"`
}

type OrderCancellation struct {
	ID             uuid.UUID             `db:"id"`
	OrderID        uuid.UUID             `db:"order_id"`
	RequestedBy    uuid.UUID             `db:"requested_by"`
	Reason         string                `db:"reason"`
	PreviousStatus string                `db:"previous_status"`
	Decision       *CancellationDecision `db:"decision"`
	DecidedBy      *uuid.UUID            `db:"decided_by"`
	DecidedAt      *time.Time            `db:"decided_at"`
	CreatedAt      time.Time             `db:"created_at"`
}

// Input models

type CheckoutInput struct {
	FulfilmentType          FulfilmentType `json:"fulfilment_type" binding:"required,oneof=PICKUP DELIVERY"`
	IdempotencyKey          string         `json:"idempotency_key"`
	DeliveryAddressSnapshot *string        `json:"delivery_address_snapshot,omitempty"`
}

type OrderStatusUpdateInput struct {
	Reason *string `json:"reason,omitempty"`
}

type CancellationRequestInput struct {
	Reason string `json:"reason" binding:"required"`
}

type CancellationDecisionInput struct {
	Decision CancellationDecision `json:"decision" binding:"required,oneof=APPROVED DECLINED"`
	Reason   *string              `json:"reason,omitempty"`
}

// Output models

type OrderItemOutput struct {
	ID                  uuid.UUID  `json:"id"`
	OrderID             uuid.UUID  `json:"order_id"`
	ProductID           *uuid.UUID `json:"product_id,omitempty"`
	ProductNameSnapshot string     `json:"product_name_snapshot"`
	UnitPricePaise      int64      `json:"unit_price_paise"`
	Quantity            int64      `json:"quantity"`
	LineTotalPaise      int64      `json:"line_total_paise"`
	CreatedAt           time.Time  `json:"created_at"`
}

type OrderOutput struct {
	ID                      uuid.UUID         `json:"id"`
	CustomerID              uuid.UUID         `json:"customer_id"`
	VendorID                uuid.UUID         `json:"vendor_id"`
	FulfilmentType          FulfilmentType    `json:"fulfilment_type"`
	Status                  OrderStatus       `json:"status"`
	SubtotalPaise           int64             `json:"subtotal_paise"`
	TotalPaise              int64             `json:"total_paise"`
	DeliveryAddressSnapshot *string           `json:"delivery_address_snapshot,omitempty"`
	PickupAddressSnapshot   *string           `json:"pickup_address_snapshot,omitempty"`
	Items                   []OrderItemOutput `json:"items"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

type OrderListOutput struct {
	Orders     []OrderOutput `json:"orders"`
	NextCursor string        `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more"`
}
