package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/sync/model"
	"vendorsaathi/internal/modules/sync/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type SyncHandler struct {
	service *service.SyncService
}

func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{service: svc}
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

// SubmitOperation POST /sync/operations
func (h *SyncHandler) SubmitOperation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.SubmitSyncOperationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	output, err := h.service.SubmitOperation(c.Request.Context(), userID, input)
	if err != nil {
		handleSyncError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

// GetMyOperations GET /sync/operations
func (h *SyncHandler) GetMyOperations(c *gin.Context) {
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

	deviceID := c.Query("device_id")

	output, err := h.service.GetMyOperations(c.Request.Context(), userID, deviceID, params.Limit, params.Cursor)
	if err != nil {
		handleSyncError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func handleSyncError(c *gin.Context, err error) {
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
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
