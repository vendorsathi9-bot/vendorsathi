package model

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	OrderID      *uuid.UUID `db:"order_id" json:"order_id,omitempty"`
	CustomerID   uuid.UUID  `db:"customer_id" json:"customer_id"`
	VendorID     uuid.UUID  `db:"vendor_id" json:"vendor_id"`
	CustomerName string     `json:"customer_name,omitempty"`
	Rating       int        `db:"rating" json:"rating"`
	Comment      *string    `db:"comment" json:"comment,omitempty"`
	IsHidden     bool       `db:"is_hidden" json:"is_hidden"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type CreateReviewInput struct {
	OrderID  *uuid.UUID `json:"order_id,omitempty"`
	VendorID uuid.UUID  `json:"vendor_id" binding:"required"`
	Rating   int        `json:"rating" binding:"required,min=1,max=5"`
	Comment  *string    `json:"comment,omitempty"`
}

type ModerateReviewInput struct {
	IsHidden bool `json:"is_hidden"`
}

type VendorReviewSummary struct {
	AverageRating float64  `json:"average_rating"`
	TotalReviews  int      `json:"total_reviews"`
	Reviews       []Review `json:"reviews"`
}

type LoyaltyAccount struct {
	ID            uuid.UUID `db:"id" json:"id"`
	CustomerID    uuid.UUID `db:"customer_id" json:"customer_id"`
	PointsBalance int       `db:"points_balance" json:"points_balance"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type LoyaltyTransaction struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AccountID       uuid.UUID  `db:"account_id" json:"account_id"`
	OrderID         *uuid.UUID `db:"order_id" json:"order_id,omitempty"`
	Points          int        `db:"points" json:"points"`
	TransactionType string     `db:"transaction_type" json:"transaction_type"`
	Description     *string    `db:"description" json:"description,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

type LoyaltySummaryOutput struct {
	PointsBalance int                  `json:"points_balance"`
	Transactions  []LoyaltyTransaction `json:"transactions"`
}
