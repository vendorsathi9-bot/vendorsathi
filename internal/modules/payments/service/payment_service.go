package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"vendorsaathi/internal/config"
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

// InitiateGatewayPayment creates an online payment order on Razorpay and returns the order token to the client.
func (s *PaymentService) InitiateGatewayPayment(ctx context.Context, cfg config.PaymentConfig, actorUserID uuid.UUID, input model.InitiateGatewayPaymentInput) (*model.GatewayOrderResponse, error) {
	if cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "" {
		return nil, shared.NewValidationError("payment_gateway", "Razorpay credentials not configured on server (RAZORPAY_KEY_ID / RAZORPAY_KEY_SECRET)")
	}

	_, customerID, totalPaise, err := s.repo.GetOrderByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if customerID != actorUserID {
		return nil, shared.NewForbidden("initiate payment for this order")
	}

	if input.AmountPaise > totalPaise {
		return nil, shared.NewPaymentLimitExceeded(totalPaise, input.AmountPaise)
	}

	receiptID := fmt.Sprintf("rcpt_%s", input.OrderID.String()[:8])
	payload := map[string]interface{}{
		"amount":          input.AmountPaise, // in paise
		"currency":        "INR",
		"receipt":         receiptID,
		"payment_capture": 1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal razorpay payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/orders", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create razorpay request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.RazorpayKeyID, cfg.RazorpayKeySecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call razorpay api: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read razorpay response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("razorpay order creation failed (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var rzpOrder struct {
		ID       string `json:"id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(respBytes, &rzpOrder); err != nil {
		return nil, fmt.Errorf("unmarshal razorpay response: %w", err)
	}

	return &model.GatewayOrderResponse{
		GatewayOrderID: rzpOrder.ID,
		AmountPaise:    rzpOrder.Amount,
		Currency:       rzpOrder.Currency,
		KeyID:          cfg.RazorpayKeyID, // Only public key ID sent to client
	}, nil
}

// VerifyGatewayPayment cryptographically verifies Razorpay payment signature and records the confirmed payment.
func (s *PaymentService) VerifyGatewayPayment(ctx context.Context, cfg config.PaymentConfig, actorUserID uuid.UUID, input model.VerifyGatewayPaymentInput) (*model.PaymentOutput, error) {
	if cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "" {
		return nil, shared.NewValidationError("payment_gateway", "Razorpay credentials not configured on server")
	}

	// 1. Verify HMAC-SHA256 signature
	dataToSign := input.GatewayOrderID + "|" + input.GatewayPaymentID
	mac := hmac.New(sha256.New, []byte(cfg.RazorpayKeySecret))
	mac.Write([]byte(dataToSign))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !strings.EqualFold(expectedSig, input.GatewaySignature) {
		return nil, shared.NewForbidden("payment signature verification failed")
	}

	// 2. Prevent duplicate processing
	existing, err := s.repo.GetPaymentByReference(ctx, input.GatewayPaymentID)
	if err == nil && existing != nil {
		out := existing.ToOutput()
		return &out, nil // Idempotent success
	}

	// 3. Verify order ownership
	_, customerID, totalPaise, err := s.repo.GetOrderByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if customerID != actorUserID {
		return nil, shared.NewForbidden("verify payment for this order")
	}

	// 4. Record payment with ONLINE method
	ref := input.GatewayPaymentID
	evidence := map[string]interface{}{
		"gateway_order_id":   input.GatewayOrderID,
		"gateway_payment_id": input.GatewayPaymentID,
		"gateway_signature":  input.GatewaySignature,
		"verified_at":        time.Now().UTC().Format(time.RFC3339),
	}

	recordInput := model.RecordPaymentInput{
		OrderID:     input.OrderID,
		Method:      "ONLINE",
		AmountPaise: totalPaise,
		Reference:   &ref,
		Evidence:    evidence,
	}

	payment, err := s.repo.RecordPayment(ctx, recordInput, actorUserID)
	if err != nil {
		return nil, err
	}

	// 5. Mark payment as confirmed & verified
	confirmed, err := s.repo.ConfirmPayment(ctx, payment.ID)
	if err != nil {
		return nil, err
	}

	out := confirmed.ToOutput()
	return &out, nil
}

