package model

import (
	"time"

	"github.com/google/uuid"
)

type MovementType string

const (
	MovementTypeAdjustment  MovementType = "ADJUSTMENT"
	MovementTypeReservation MovementType = "RESERVATION"
	MovementTypeRelease     MovementType = "RELEASE"
	MovementTypeFulfilment  MovementType = "FULFILMENT"
	MovementTypeCorrection  MovementType = "CORRECTION"
)

type Inventory struct {
	ProductID   uuid.UUID `db:"product_id"`
	OnHandQty   int64     `db:"on_hand_qty"`
	ReservedQty int64     `db:"reserved_qty"`
	Version     int64     `db:"version"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (i *Inventory) AvailableQty() int64 {
	return i.OnHandQty - i.ReservedQty
}

type InventoryMovement struct {
	ID             uuid.UUID    `db:"id"`
	ProductID      uuid.UUID    `db:"product_id"`
	MovementType   MovementType `db:"movement_type"`
	QtyDelta       int64        `db:"qty_delta"`
	Reason         *string      `db:"reason"`
	ActorUserID    *uuid.UUID   `db:"actor_user_id"`
	ReferenceType  *string      `db:"reference_type"`
	ReferenceID    *uuid.UUID   `db:"reference_id"`
	IdempotencyKey *string      `db:"idempotency_key"`
	CreatedAt      time.Time    `db:"created_at"`
}

type AdjustStockInput struct {
	ProductID      uuid.UUID `json:"product_id" binding:"required"`
	QtyDelta       int64     `json:"qty_delta" binding:"required"`
	Reason         *string   `json:"reason,omitempty"`
	IdempotencyKey *string   `json:"idempotency_key,omitempty"`
}

type ReserveStockInput struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	Qty            int64      `json:"qty" binding:"required,min=1"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	ReferenceType  *string    `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID `json:"reference_id,omitempty"`
}

type ReleaseStockInput struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	Qty            int64      `json:"qty" binding:"required,min=1"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	ReferenceType  *string    `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID `json:"reference_id,omitempty"`
}

type FulfilStockInput struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	Qty            int64      `json:"qty" binding:"required,min=1"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	ReferenceType  *string    `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID `json:"reference_id,omitempty"`
}

type InventoryOutput struct {
	ProductID    uuid.UUID `json:"product_id"`
	OnHandQty    int64     `json:"on_hand_qty"`
	ReservedQty  int64     `json:"reserved_qty"`
	AvailableQty int64     `json:"available_qty"`
	Version      int64     `json:"version"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InventoryMovementOutput struct {
	ID             uuid.UUID    `json:"id"`
	ProductID      uuid.UUID    `json:"product_id"`
	MovementType   MovementType `json:"movement_type"`
	QtyDelta       int64        `json:"qty_delta"`
	Reason         *string      `json:"reason,omitempty"`
	ActorUserID    *uuid.UUID   `json:"actor_user_id,omitempty"`
	ReferenceType  *string      `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID   `json:"reference_id,omitempty"`
	IdempotencyKey *string      `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

func (i *Inventory) ToOutput() InventoryOutput {
	return InventoryOutput{
		ProductID:    i.ProductID,
		OnHandQty:    i.OnHandQty,
		ReservedQty:  i.ReservedQty,
		AvailableQty: i.OnHandQty - i.ReservedQty,
		Version:      i.Version,
		UpdatedAt:    i.UpdatedAt,
	}
}

func (m *InventoryMovement) ToOutput() InventoryMovementOutput {
	return InventoryMovementOutput{
		ID:             m.ID,
		ProductID:      m.ProductID,
		MovementType:   m.MovementType,
		QtyDelta:       m.QtyDelta,
		Reason:         m.Reason,
		ActorUserID:    m.ActorUserID,
		ReferenceType:  m.ReferenceType,
		ReferenceID:    m.ReferenceID,
		IdempotencyKey: m.IdempotencyKey,
		CreatedAt:      m.CreatedAt,
	}
}

type InventoryMovementListOutput struct {
	Data       []InventoryMovementOutput `json:"data"`
	NextCursor string                    `json:"next_cursor,omitempty"`
	HasMore    bool                      `json:"has_more"`
}
