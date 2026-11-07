package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/carts/model"
	"vendorsaathi/internal/modules/carts/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type CartHandler struct {
	service *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{service: svc}
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

// GetMyCart GET /carts/me
func (h *CartHandler) GetMyCart(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	cart, err := h.service.GetMyCart(c.Request.Context(), customerID)
	if err != nil {
		handleCartError(c, err)
		return
	}
	httpx.RespondOK(c, cart)
}

// AddItem POST /carts/items
func (h *CartHandler) AddItem(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.AddCartItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	cart, err := h.service.AddItem(c.Request.Context(), customerID, input.ProductID, input.Quantity)
	if err != nil {
		handleCartError(c, err)
		return
	}
	c.JSON(http.StatusOK, cart)
}

// RemoveItem DELETE /carts/items/:productId
func (h *CartHandler) RemoveItem(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	cart, err := h.service.RemoveItem(c.Request.Context(), customerID, productID)
	if err != nil {
		handleCartError(c, err)
		return
	}
	httpx.RespondOK(c, cart)
}

// UpdateItemQuantity PATCH /carts/items/:productId
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	var input model.UpdateCartItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	cart, err := h.service.UpdateItemQuantity(c.Request.Context(), customerID, productID, input.Quantity)
	if err != nil {
		handleCartError(c, err)
		return
	}
	httpx.RespondOK(c, cart)
}

// ClearCart DELETE /carts/me
func (h *CartHandler) ClearCart(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	if err := h.service.ClearCart(c.Request.Context(), customerID); err != nil {
		handleCartError(c, err)
		return
	}
	httpx.RespondNoContent(c)
}

// GetCart GET /carts/:id
func (h *CartHandler) GetCart(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid cart id", nil)
		return
	}

	cart, err := h.service.GetCart(c.Request.Context(), customerID, cartID)
	if err != nil {
		handleCartError(c, err)
		return
	}
	httpx.RespondOK(c, cart)
}

func handleCartError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}
	if errors.Is(err, shared.ErrNotFound) {
		httpx.ErrorNotFound(c, "resource not found")
		return
	}
	httpx.ErrorInternal(c, "internal server error")
}
