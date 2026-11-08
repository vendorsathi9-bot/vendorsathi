package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/config"
	"vendorsaathi/internal/modules/payments/model"
	"vendorsaathi/internal/modules/payments/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type PaymentHandler struct {
	svc *service.PaymentService
	cfg config.PaymentConfig
}

func NewPaymentHandler(svc *service.PaymentService, cfg config.PaymentConfig) *PaymentHandler {
	return &PaymentHandler{svc: svc, cfg: cfg}
}

// parseActorUserID extracts and parses the actor's user ID from gin context.
func parseActorUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return uuid.Nil, false
	}
	if id, ok := val.(uuid.UUID); ok {
		return id, true
	}
	if str, ok := val.(string); ok {
		id, err := uuid.Parse(str)
		if err == nil {
			return id, true
		}
	}
	httpx.ErrorInternal(c, "invalid user id in token")
	return uuid.Nil, false
}

// RecordPayment handles POST /payments
func (h *PaymentHandler) RecordPayment(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	var input model.RecordPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body: "+err.Error(), nil)
		return
	}

	output, err := h.svc.RecordPayment(c.Request.Context(), actorUserID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

// GetOrderPayments handles GET /payments/order/:orderId
func (h *PaymentHandler) GetOrderPayments(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	outputs, err := h.svc.GetOrderPayments(c.Request.Context(), actorUserID, orderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, outputs)
}

// VerifyPayment handles POST /payments/:id/verify
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid payment id", nil)
		return
	}

	var input model.VerifyPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body: "+err.Error(), nil)
		return
	}

	output, err := h.svc.VerifyPayment(c.Request.Context(), actorUserID, paymentID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

// RecordRefund handles POST /payments/:id/refunds
func (h *PaymentHandler) RecordRefund(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid payment id", nil)
		return
	}

	var input model.RecordRefundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body: "+err.Error(), nil)
		return
	}

	output, err := h.svc.RecordRefund(c.Request.Context(), actorUserID, paymentID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

// GetRefunds handles GET /payments/:id/refunds
func (h *PaymentHandler) GetRefunds(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid payment id", nil)
		return
	}

	outputs, err := h.svc.GetRefunds(c.Request.Context(), actorUserID, paymentID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, outputs)
}

// InitiateGatewayPayment handles POST /payments/gateway/initiate
func (h *PaymentHandler) InitiateGatewayPayment(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	var input model.InitiateGatewayPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body: "+err.Error(), nil)
		return
	}

	output, err := h.svc.InitiateGatewayPayment(c.Request.Context(), h.cfg, actorUserID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

// VerifyGatewayPayment handles POST /payments/gateway/verify
func (h *PaymentHandler) VerifyGatewayPayment(c *gin.Context) {
	actorUserID, ok := parseActorUserID(c)
	if !ok {
		return
	}

	var input model.VerifyGatewayPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body: "+err.Error(), nil)
		return
	}

	output, err := h.svc.VerifyGatewayPayment(c.Request.Context(), h.cfg, actorUserID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *PaymentHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrValidationFailed):
			httpx.ErrorBadRequest(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrPaymentLimitExceeded):
			details := make([]httpx.ErrorDetail, 0, len(appErr.Details))
			for k, v := range appErr.Details {
				details = append(details, httpx.ErrorDetail{Field: k, Reason: v})
			}
			httpx.ErrorUnprocessable(c, appErr.Message, details)
		case errors.Is(appErr.Err, shared.ErrRefundLimitExceeded):
			details := make([]httpx.ErrorDetail, 0, len(appErr.Details))
			for k, v := range appErr.Details {
				details = append(details, httpx.ErrorDetail{Field: k, Reason: v})
			}
			httpx.ErrorUnprocessable(c, appErr.Message, details)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	if errors.Is(err, shared.ErrNotFound) {
		httpx.ErrorNotFound(c, "resource not found")
		return
	}
	if errors.Is(err, shared.ErrForbidden) {
		httpx.ErrorForbidden(c, "access forbidden")
		return
	}
	if errors.Is(err, shared.ErrConflict) {
		httpx.ErrorConflict(c, err.Error(), nil)
		return
	}

	httpx.ErrorInternal(c, "internal server error")
}
