package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/notifications/model"
	"vendorsaathi/internal/modules/notifications/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
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

// GetMyNotifications GET /notifications
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid query parameters", nil)
		return
	}

	list, err := h.service.GetMyNotifications(c.Request.Context(), userID, params.Limit, params.Cursor)
	if err != nil {
		handleNotificationError(c, err)
		return
	}

	httpx.RespondOK(c, list)
}

// MarkRead POST /notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid notification id", nil)
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), id, userID); err != nil {
		handleNotificationError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

// MarkAllRead POST /notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		handleNotificationError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

// RegisterDeviceToken POST /notifications/device-tokens
func (h *NotificationHandler) RegisterDeviceToken(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.RegisterDeviceTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	if _, err := h.service.RegisterDeviceToken(c.Request.Context(), userID, input.Token, model.Platform(input.Platform)); err != nil {
		handleNotificationError(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{"status": "token registered"})
}

// RevokeDeviceToken DELETE /notifications/device-tokens/:token
func (h *NotificationHandler) RevokeDeviceToken(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	token := c.Param("token")
	if token == "" {
		httpx.ErrorBadRequest(c, "token parameter is required", nil)
		return
	}

	if err := h.service.RevokeDeviceToken(c.Request.Context(), userID, token); err != nil {
		handleNotificationError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

func handleNotificationError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	if errors.Is(err, shared.ErrNotFound) {
		httpx.ErrorNotFound(c, "resource not found")
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
