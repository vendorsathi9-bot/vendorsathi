package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "CASH"
	PaymentMethodUPI    PaymentMethod = "UPI"
	PaymentMethodOnline PaymentMethod = "ONLINE"
)

type PaymentStatus string

const (
	PaymentStatusPendingVerification PaymentStatus = "PENDING_VERIFICATION"
	PaymentStatusConfirmed           PaymentStatus = "CONFIRMED"
	PaymentStatusRejected            PaymentStatus = "REJECTED"
	PaymentStatusRefundedPartial     PaymentStatus = "REFUNDED_PARTIAL"
	PaymentStatusRefundedFull        PaymentStatus = "REFUNDED_FULL"
)

type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "PENDING"
	VerificationStatusVerified VerificationStatus = "VERIFIED"
	VerificationStatusRejected VerificationStatus = "REJECTED"
)

type RefundStatus string

const (
	RefundStatusRecorded  RefundStatus = "RECORDED"
	RefundStatusProcessed RefundStatus = "PROCESSED"
	RefundStatusFailed    RefundStatus = "FAILED"
)

type Payment struct {
	ID                 uuid.UUID          `db:"id"`
	OrderID            uuid.UUID          `db:"order_id"`
	Method             PaymentMethod      `db:"method"`
	AmountPaise        int64              `db:"amount_paise"`
	Status             PaymentStatus      `db:"status"`
	VerificationStatus VerificationStatus `db:"verification_status"`
	RecordedBy         uuid.UUID          `db:"recorded_by"`
	Reference          *string            `db:"reference"`
	Evidence           []byte             `db:"evidence"` // raw JSONB
	IdempotencyKey     *string            `db:"idempotency_key"`
	CreatedAt          time.Time          `db:"created_at"`
}

type Refund struct {
	ID             uuid.UUID    `db:"id"`
	PaymentID      uuid.UUID    `db:"payment_id"`
	AmountPaise    int64        `db:"amount_paise"`
	Status         RefundStatus `db:"status"`
	Reason         *string      `db:"reason"`
	IdempotencyKey *string      `db:"idempotency_key"`
	CreatedAt      time.Time    `db:"created_at"`
}

// --- Input types ---

type RecordPaymentInput struct {
	OrderID        uuid.UUID              `json:"order_id" binding:"required"`
	Method         string                 `json:"method" binding:"required,oneof=CASH UPI ONLINE"`
	AmountPaise    int64                  `json:"amount_paise" binding:"required,min=1"`
	Reference      *string                `json:"reference,omitempty"`
	Evidence       map[string]interface{} `json:"evidence,omitempty"`
	IdempotencyKey *string                `json:"idempotency_key,omitempty"`
}

type VerifyPaymentInput struct {
	VerificationStatus string  `json:"verification_status" binding:"required,oneof=VERIFIED REJECTED"`
	Note               *string `json:"note,omitempty"`
}

type RecordRefundInput struct {
	AmountPaise    int64   `json:"amount_paise" binding:"required,min=1"`
	Reason         *string `json:"reason,omitempty"`
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

// --- Output types ---

type PaymentOutput struct {
	ID                 uuid.UUID          `json:"id"`
	OrderID            uuid.UUID          `json:"order_id"`
	Method             PaymentMethod      `json:"method"`
	AmountPaise        int64              `json:"amount_paise"`
	Status             PaymentStatus      `json:"status"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	RecordedBy         uuid.UUID          `json:"recorded_by"`
	Reference          *string            `json:"reference,omitempty"`
	IdempotencyKey     *string            `json:"idempotency_key,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
}

type RefundOutput struct {
	ID             uuid.UUID    `json:"id"`
	PaymentID      uuid.UUID    `json:"payment_id"`
	AmountPaise    int64        `json:"amount_paise"`
	Status         RefundStatus `json:"status"`
	Reason         *string      `json:"reason,omitempty"`
	IdempotencyKey *string      `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

func (p *Payment) ToOutput() PaymentOutput {
	return PaymentOutput{
		ID:                 p.ID,
		OrderID:            p.OrderID,
		Method:             p.Method,
		AmountPaise:        p.AmountPaise,
		Status:             p.Status,
		VerificationStatus: p.VerificationStatus,
		RecordedBy:         p.RecordedBy,
		Reference:          p.Reference,
		IdempotencyKey:     p.IdempotencyKey,
		CreatedAt:          p.CreatedAt,
	}
}

func (r *Refund) ToOutput() RefundOutput {
	return RefundOutput{
		ID:             r.ID,
		PaymentID:      r.PaymentID,
		AmountPaise:    r.AmountPaise,
		Status:         r.Status,
		Reason:         r.Reason,
		IdempotencyKey: r.IdempotencyKey,
		CreatedAt:      r.CreatedAt,
	}
}

// --- Gateway payment types ---

type InitiateGatewayPaymentInput struct {
	OrderID     uuid.UUID `json:"order_id" binding:"required"`
	AmountPaise int64     `json:"amount_paise" binding:"required,min=1"`
}

type GatewayOrderResponse struct {
	GatewayOrderID string `json:"gateway_order_id"`
	AmountPaise    int64  `json:"amount_paise"`
	Currency       string `json:"currency"`
	KeyID          string `json:"key_id"` // Public key — safe to send to client
}

type VerifyGatewayPaymentInput struct {
	GatewayOrderID   string    `json:"gateway_order_id" binding:"required"`
	GatewayPaymentID string    `json:"gateway_payment_id" binding:"required"`
	GatewaySignature string    `json:"gateway_signature" binding:"required"`
	OrderID          uuid.UUID `json:"order_id" binding:"required"`
}

