package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/audit/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type AuditHandler struct {
	service *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{service: svc}
}

// ListAuditLogs GET /audit-logs (Admin only)
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid pagination parameters", nil)
		return
	}

	var actorUserID *uuid.UUID
	if actorStr := c.Query("actor_user_id"); actorStr != "" {
		parsed, err := uuid.Parse(actorStr)
		if err == nil {
			actorUserID = &parsed
		}
	}

	var resourceType *string
	if rType := c.Query("resource_type"); rType != "" {
		resourceType = &rType
	}

	var resourceID *string
	if rID := c.Query("resource_id"); rID != "" {
		resourceID = &rID
	}

	output, err := h.service.List(c.Request.Context(), actorUserID, resourceType, resourceID, params.Limit, params.Cursor)
	if err != nil {
		handleAuditError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func handleAuditError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
