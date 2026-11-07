package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/orders/model"
	"vendorsaathi/internal/modules/orders/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
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
	return uuid.Nil, false
}

func hasRole(c *gin.Context, targetRole string) bool {
	rolesVal, exists := c.Get("roles")
	if !exists {
		return false
	}
	roles, ok := rolesVal.([]string)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == targetRole {
			return true
		}
	}
	return false
}

// Checkout POST /orders
func (h *OrderHandler) Checkout(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.CheckoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	order, err := h.service.Checkout(c.Request.Context(), customerID, input)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondCreated(c, order)
}

// GetOrderByID GET /orders/:id
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	_, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// ListOrders GET /orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid pagination parameters", nil)
		return
	}

	statusFilter := c.Query("status")

	var orders []model.OrderOutput
	var nextCursor string
	var hasMore bool
	var err error

	if hasRole(c, "ADMIN") || hasRole(c, "SUPER_ADMIN") {
		orders, nextCursor, hasMore, err = h.service.GetAllOrders(c.Request.Context(), statusFilter, params.Limit, params.Cursor)
	} else if hasRole(c, "VENDOR") {
		orders, nextCursor, hasMore, err = h.service.GetOrdersByVendor(c.Request.Context(), userID, statusFilter, params.Limit, params.Cursor)
	} else {
		orders, nextCursor, hasMore, err = h.service.GetOrdersByCustomer(c.Request.Context(), userID, params.Limit, params.Cursor)
	}

	if err != nil {
		handleOrderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        orders,
		"orders":      orders,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

// AcceptOrder POST /orders/:id/accept
func (h *OrderHandler) AcceptOrder(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.AcceptOrder(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// RejectOrder POST /orders/:id/reject
func (h *OrderHandler) RejectOrder(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.RejectOrder(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// PrepareOrder POST /orders/:id/prepare
func (h *OrderHandler) PrepareOrder(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.PrepareOrder(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// ReadyForPickup POST /orders/:id/ready
func (h *OrderHandler) ReadyForPickup(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.MarkReadyForPickup(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// OutForDelivery POST /orders/:id/out-for-delivery
func (h *OrderHandler) OutForDelivery(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.MarkOutForDelivery(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// CompleteOrder POST /orders/:id/complete
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.OrderStatusUpdateInput
	_ = c.ShouldBindJSON(&input)

	order, err := h.service.CompleteOrder(c.Request.Context(), id, vendorUserID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// RequestCancellation POST /orders/:id/request-cancellation
func (h *OrderHandler) RequestCancellation(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.CancellationRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	order, err := h.service.RequestCancellation(c.Request.Context(), id, customerID, input.Reason)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

// DecideCancellation POST /orders/:id/decide-cancellation
func (h *OrderHandler) DecideCancellation(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.CancellationDecisionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	order, err := h.service.DecideCancellation(c.Request.Context(), id, vendorUserID, input)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	httpx.RespondOK(c, order)
}

func handleOrderError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrInsufficientStock):
			httpx.ErrorUnprocessable(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrInvalidTransition):
			httpx.ErrorUnprocessable(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrIdempotencyConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		httpx.ErrorNotFound(c, "order not found")
	case errors.Is(err, shared.ErrForbidden):
		httpx.ErrorForbidden(c, "forbidden")
	case errors.Is(err, shared.ErrInsufficientStock):
		httpx.ErrorUnprocessable(c, "insufficient stock available", nil)
	case errors.Is(err, shared.ErrInvalidTransition):
		httpx.ErrorUnprocessable(c, "invalid order state transition", nil)
	case errors.Is(err, shared.ErrConflict):
		httpx.ErrorConflict(c, err.Error(), nil)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
