package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/customers/model"
	"vendorsaathi/internal/modules/customers/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

func (h *CustomerHandler) GetMyProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	output, err := h.customerService.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) UpdateMyProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	var input model.UpdateCustomerProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.customerService.UpdateMyProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) CreateAddress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	var input model.CreateCustomerAddressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.customerService.CreateAddress(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *CustomerHandler) GetMyAddresses(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	output, err := h.customerService.GetMyAddresses(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) GetDefaultAddress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	output, err := h.customerService.GetDefaultAddress(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) GetAddressByID(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	addressIDStr := c.Param("addressId")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid address id", nil)
		return
	}

	output, err := h.customerService.GetAddressByID(c.Request.Context(), userID, addressID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) UpdateAddress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	addressIDStr := c.Param("addressId")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid address id", nil)
		return
	}

	var input model.UpdateCustomerAddressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.customerService.UpdateAddress(c.Request.Context(), userID, addressID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *CustomerHandler) DeleteAddress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	addressIDStr := c.Param("addressId")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid address id", nil)
		return
	}

	if err := h.customerService.DeleteAddress(c.Request.Context(), userID, addressID); err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

func (h *CustomerHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "NOT_FOUND":
			httpx.ErrorNotFound(c, appErr.Message)
		case "FORBIDDEN":
			httpx.ErrorForbidden(c, appErr.Message)
		case "CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "VALIDATION_ERROR":
			var details []httpx.ErrorDetail
			for field, reason := range appErr.Details {
				details = append(details, httpx.ErrorDetail{Field: field, Reason: reason})
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
	if errors.Is(err, shared.ErrValidationFailed) {
		httpx.ErrorUnprocessable(c, "validation failed", nil)
		return
	}

	httpx.ErrorInternal(c, "internal server error")
}
