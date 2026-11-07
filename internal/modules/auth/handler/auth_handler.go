package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/auth/model"
	"vendorsaathi/internal/modules/auth/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input model.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input model.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input model.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.authService.Refresh(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input model.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), input.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	uidStr, ok := userID.(string)
	if !ok {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id format")
		return
	}

	output, err := h.authService.GetMe(c.Request.Context(), uid)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "NOT_FOUND":
			httpx.ErrorNotFound(c, appErr.Message)
		case "FORBIDDEN":
			httpx.ErrorForbidden(c, appErr.Message)
		case "CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "INVALID_TRANSITION":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "INSUFFICIENT_STOCK":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "IDEMPOTENCY_CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "VALIDATION_ERROR":
			var details []httpx.ErrorDetail
			for field, reason := range appErr.Details {
				details = append(details, httpx.ErrorDetail{Field: field, Reason: reason})
			}
			httpx.ErrorUnprocessable(c, appErr.Message, details)
		case "INVALID_CREDENTIALS", "TOKEN_REVOKED", "TOKEN_EXPIRED", "ACCOUNT_INACTIVE":
			httpx.ErrorUnauthorized(c, appErr.Message)
		case "INVALID_REFRESH_TOKEN":
			httpx.ErrorUnauthorized(c, appErr.Message)
		case "PAYMENT_LIMIT_EXCEEDED", "REFUND_LIMIT_EXCEEDED", "UDHAAR_LIMIT_EXCEEDED":
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
	if errors.Is(err, shared.ErrForbidden) {
		httpx.ErrorForbidden(c, "access forbidden")
		return
	}
	if errors.Is(err, shared.ErrConflict) {
		httpx.ErrorConflict(c, err.Error(), nil)
		return
	}
	if errors.Is(err, shared.ErrUnauthorized) {
		httpx.ErrorUnauthorized(c, "unauthorized")
		return
	}
	if errors.Is(err, shared.ErrValidationFailed) {
		httpx.ErrorUnprocessable(c, "validation failed", nil)
		return
	}

	httpx.ErrorInternal(c, "internal server error")
}
