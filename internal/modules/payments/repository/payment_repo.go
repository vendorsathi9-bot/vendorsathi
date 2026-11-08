package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/payments/model"
	"vendorsaathi/internal/shared"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// RecordPayment inserts a new payment record. Handles idempotency via unique(order_id, idempotency_key).
func (r *PaymentRepository) RecordPayment(ctx context.Context, input model.RecordPaymentInput, recordedByUserID uuid.UUID) (*model.Payment, error) {
	var evidenceJSON []byte
	if input.Evidence != nil {
		b, err := json.Marshal(input.Evidence)
		if err != nil {
			return nil, err
		}
		evidenceJSON = b
	}

	id := uuid.New()
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, evidence, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, 'PENDING_VERIFICATION', 'PENDING', $5, $6, $7, $8, NOW())
		RETURNING id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, idempotency_key, created_at
	`, id, input.OrderID, input.Method, input.AmountPaise, recordedByUserID, input.Reference, evidenceJSON, input.IdempotencyKey,
	).Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
		&p.RecordedBy, &p.Reference, &p.IdempotencyKey, &p.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, shared.NewConflict("payment with this idempotency key already recorded for this order")
		}
		return nil, err
	}
	return &p, nil
}

// GetPaymentByID fetches a single payment by primary key.
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, idempotency_key, created_at
		FROM payments WHERE id = $1
	`, id).Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
		&p.RecordedBy, &p.Reference, &p.IdempotencyKey, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("payment")
		}
		return nil, err
	}
	return &p, nil
}

// GetPaymentsByOrder returns all payments for a given order.
func (r *PaymentRepository) GetPaymentsByOrder(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, idempotency_key, created_at
		FROM payments WHERE order_id = $1 ORDER BY created_at ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
			&p.RecordedBy, &p.Reference, &p.IdempotencyKey, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// GetTotalPaidForOrder returns the sum of CONFIRMED and PENDING_VERIFICATION payment amounts for an order.
func (r *PaymentRepository) GetTotalPaidForOrder(ctx context.Context, orderID uuid.UUID) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_paise), 0)
		FROM payments
		WHERE order_id = $1 AND status IN ('CONFIRMED', 'PENDING_VERIFICATION')
	`, orderID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// VerifyPayment updates the verification_status and cascades to payment status.
// VERIFIED -> status = CONFIRMED; REJECTED -> status = REJECTED.
func (r *PaymentRepository) VerifyPayment(ctx context.Context, paymentID uuid.UUID, verificationStatus string) (*model.Payment, error) {
	newPaymentStatus := "REJECTED"
	if verificationStatus == "VERIFIED" {
		newPaymentStatus = "CONFIRMED"
	}

	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		UPDATE payments
		SET verification_status = $2, status = $3
		WHERE id = $1 AND status = 'PENDING_VERIFICATION'
		RETURNING id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, idempotency_key, created_at
	`, paymentID, verificationStatus, newPaymentStatus,
	).Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
		&p.RecordedBy, &p.Reference, &p.IdempotencyKey, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("payment")
		}
		return nil, err
	}
	return &p, nil
}

// RecordRefund inserts a refund record after checking idempotency.
func (r *PaymentRepository) RecordRefund(ctx context.Context, paymentID uuid.UUID, input model.RecordRefundInput) (*model.Refund, error) {
	id := uuid.New()
	var ref model.Refund
	err := r.pool.QueryRow(ctx, `
		INSERT INTO refunds (id, payment_id, amount_paise, status, reason, idempotency_key, created_at)
		VALUES ($1, $2, $3, 'RECORDED', $4, $5, NOW())
		RETURNING id, payment_id, amount_paise, status, reason, idempotency_key, created_at
	`, id, paymentID, input.AmountPaise, input.Reason, input.IdempotencyKey,
	).Scan(&ref.ID, &ref.PaymentID, &ref.AmountPaise, &ref.Status, &ref.Reason, &ref.IdempotencyKey, &ref.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, shared.NewConflict("refund with this idempotency key already recorded for this payment")
		}
		return nil, err
	}
	return &ref, nil
}

// GetRefundsByPayment returns all refunds for a payment.
func (r *PaymentRepository) GetRefundsByPayment(ctx context.Context, paymentID uuid.UUID) ([]model.Refund, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, payment_id, amount_paise, status, reason, idempotency_key, created_at
		FROM refunds WHERE payment_id = $1 ORDER BY created_at ASC
	`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refunds []model.Refund
	for rows.Next() {
		var ref model.Refund
		if err := rows.Scan(&ref.ID, &ref.PaymentID, &ref.AmountPaise, &ref.Status, &ref.Reason, &ref.IdempotencyKey, &ref.CreatedAt); err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, rows.Err()
}

// GetTotalRefundedForPayment returns the sum of RECORDED and PROCESSED refund amounts for a payment.
func (r *PaymentRepository) GetTotalRefundedForPayment(ctx context.Context, paymentID uuid.UUID) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_paise), 0)
		FROM refunds
		WHERE payment_id = $1 AND status IN ('RECORDED', 'PROCESSED')
	`, paymentID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetOrderByID fetches the order's vendor_id, customer_id and total_paise for validation.
func (r *PaymentRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (vendorID uuid.UUID, customerID uuid.UUID, totalPaise int64, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT vendor_id, customer_id, total_paise FROM orders WHERE id = $1
	`, orderID).Scan(&vendorID, &customerID, &totalPaise)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = shared.NewNotFound("order")
			return
		}
		return
	}
	return
}

// GetVendorProfileByOwnerUserID fetches the vendor profile id for the given user.
func (r *PaymentRepository) GetVendorProfileByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (uuid.UUID, error) {
	var vendorID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM vendor_profiles WHERE owner_user_id = $1
	`, ownerUserID).Scan(&vendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, shared.NewNotFound("vendor profile")
		}
		return uuid.Nil, err
	}
	return vendorID, nil
}

// GetPaymentByReference returns a payment by its reference field (e.g. gateway payment ID).
func (r *PaymentRepository) GetPaymentByReference(ctx context.Context, reference string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, evidence, idempotency_key, created_at
		FROM payments WHERE reference = $1 LIMIT 1
	`, reference).Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
		&p.RecordedBy, &p.Reference, &p.Evidence, &p.IdempotencyKey, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// ConfirmPayment marks a payment as confirmed and verified.
func (r *PaymentRepository) ConfirmPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		UPDATE payments
		SET status = 'CONFIRMED', verification_status = 'VERIFIED'
		WHERE id = $1
		RETURNING id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, evidence, idempotency_key, created_at
	`, paymentID).Scan(&p.ID, &p.OrderID, &p.Method, &p.AmountPaise, &p.Status, &p.VerificationStatus,
		&p.RecordedBy, &p.Reference, &p.Evidence, &p.IdempotencyKey, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

