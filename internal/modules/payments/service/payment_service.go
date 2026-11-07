package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/payments/model"
	"vendorsaathi/internal/modules/payments/repository"
	"vendorsaathi/internal/shared"
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// RecordPayment records a new payment for an order.
// The actor must be a vendor whose profile matches the order's vendor_id.
// Overpayment (total paid + new amount > order.total_paise) is rejected.
func (s *PaymentService) RecordPayment(ctx context.Context, actorUserID uuid.UUID, input model.RecordPaymentInput) (*model.PaymentOutput, error) {
	// Resolve vendor profile for the actor.
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewForbidden("record payment: actor is not a vendor")
		}
		return nil, err
	}

	// Fetch order and validate ownership.
	orderVendorID, _, totalPaise, err := s.repo.GetOrderByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if orderVendorID != vendorID {
		return nil, shared.NewForbidden("record payment for this order")
	}

	// Check for overpayment.
	alreadyPaid, err := s.repo.GetTotalPaidForOrder(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	outstanding := totalPaise - alreadyPaid
	if input.AmountPaise > outstanding {
		return nil, shared.NewPaymentLimitExceeded(outstanding, input.AmountPaise)
	}

	payment, err := s.repo.RecordPayment(ctx, input, actorUserID)
	if err != nil {
		return nil, err
	}

	out := payment.ToOutput()
	return &out, nil
}

// VerifyPayment verifies (or rejects) a payment. Only the vendor who owns the order may verify.
func (s *PaymentService) VerifyPayment(ctx context.Context, actorUserID uuid.UUID, paymentID uuid.UUID, input model.VerifyPaymentInput) (*model.PaymentOutput, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewForbidden("verify payment: actor is not a vendor")
		}
		return nil, err
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Confirm order belongs to this vendor.
	orderVendorID, _, _, err := s.repo.GetOrderByID(ctx, payment.OrderID)
	if err != nil {
		return nil, err
	}
	if orderVendorID != vendorID {
		return nil, shared.NewForbidden("verify payment for this order")
	}

	updated, err := s.repo.VerifyPayment(ctx, paymentID, input.VerificationStatus)
	if err != nil {
		return nil, err
	}

	out := updated.ToOutput()
	return &out, nil
}

// GetOrderPayments returns all payments for an order. The actor must be the vendor or customer of the order.
func (s *PaymentService) GetOrderPayments(ctx context.Context, actorUserID uuid.UUID, orderID uuid.UUID) ([]model.PaymentOutput, error) {
	orderVendorID, orderCustomerID, _, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Check access: actor is the vendor owner or the customer.
	vendorID, vendorErr := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	isVendor := vendorErr == nil && vendorID == orderVendorID
	isCustomer := actorUserID == orderCustomerID

	if !isVendor && !isCustomer {
		return nil, shared.NewForbidden("view payments for this order")
	}

	payments, err := s.repo.GetPaymentsByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.PaymentOutput, len(payments))
	for i, p := range payments {
		outputs[i] = p.ToOutput()
	}
	return outputs, nil
}

// RecordRefund records a refund against a confirmed payment. Only the vendor may issue refunds.
func (s *PaymentService) RecordRefund(ctx context.Context, actorUserID uuid.UUID, paymentID uuid.UUID, input model.RecordRefundInput) (*model.RefundOutput, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewForbidden("record refund: actor is not a vendor")
		}
		return nil, err
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Validate order ownership.
	orderVendorID, _, _, err := s.repo.GetOrderByID(ctx, payment.OrderID)
	if err != nil {
		return nil, err
	}
	if orderVendorID != vendorID {
		return nil, shared.NewForbidden("record refund for this payment")
	}

	// Check refund limit.
	alreadyRefunded, err := s.repo.GetTotalRefundedForPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	refundable := payment.AmountPaise - alreadyRefunded
	if input.AmountPaise > refundable {
		return nil, shared.NewRefundLimitExceeded(refundable, input.AmountPaise)
	}

	refund, err := s.repo.RecordRefund(ctx, paymentID, input)
	if err != nil {
		return nil, err
	}

	out := refund.ToOutput()
	return &out, nil
}

// GetRefunds returns all refunds for a payment. Only the vendor may list refunds.
func (s *PaymentService) GetRefunds(ctx context.Context, actorUserID uuid.UUID, paymentID uuid.UUID) ([]model.RefundOutput, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewForbidden("view refunds: actor is not a vendor")
		}
		return nil, err
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	orderVendorID, _, _, err := s.repo.GetOrderByID(ctx, payment.OrderID)
	if err != nil {
		return nil, err
	}
	if orderVendorID != vendorID {
		return nil, shared.NewForbidden("view refunds for this payment")
	}

	refunds, err := s.repo.GetRefundsByPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.RefundOutput, len(refunds))
	for i, ref := range refunds {
		outputs[i] = ref.ToOutput()
	}
	return outputs, nil
}
