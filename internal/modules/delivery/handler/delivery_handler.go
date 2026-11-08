package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/delivery/model"
	"vendorsaathi/internal/modules/delivery/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type DeliveryHandler struct {
	svc *service.DeliveryService
}

func NewDeliveryHandler(svc *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{svc: svc}
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

func getUserRoles(c *gin.Context) []string {
	rolesVal, exists := c.Get("roles")
	if !exists {
		return nil
	}
	if roles, ok := rolesVal.([]string); ok {
		return roles
	}
	return nil
}

func handleDeliveryError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrInvalidTransition):
			httpx.ErrorUnprocessable(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrValidationFailed):
			httpx.ErrorBadRequest(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		httpx.ErrorNotFound(c, "resource not found")
	case errors.Is(err, shared.ErrForbidden):
		httpx.ErrorForbidden(c, "action forbidden")
	default:
		httpx.ErrorInternal(c, "internal server error")
	}
}

// AssignDelivery POST /delivery/orders/:orderId/assign — Admin only
func (h *DeliveryHandler) AssignDelivery(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.AssignDeliveryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid input: "+err.Error(), nil)
		return
	}

	out, err := h.svc.AssignDelivery(c.Request.Context(), orderID, input.AgentUserID, adminID)
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// GetMyAssignments GET /delivery/agent/assignments — Delivery agent only
func (h *DeliveryHandler) GetMyAssignments(c *gin.Context) {
	agentID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	out, err := h.svc.GetMyAssignments(c.Request.Context(), agentID)
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	if out == nil {
		out = []model.DeliveryAssignmentWithOrderOutput{}
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// GetAssignment GET /delivery/orders/:orderId/assignment
func (h *DeliveryHandler) GetAssignment(c *gin.Context) {
	actorID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	roles := getUserRoles(c)
	out, err := h.svc.GetAssignmentByOrderID(c.Request.Context(), orderID, actorID, roles)
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// UpdateDeliveryStatus PATCH /delivery/orders/:orderId/status — Delivery agent only
func (h *DeliveryHandler) UpdateDeliveryStatus(c *gin.Context) {
	agentID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.UpdateDeliveryStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid input: "+err.Error(), nil)
		return
	}

	out, err := h.svc.UpdateDeliveryStatus(c.Request.Context(), orderID, agentID, model.DeliveryStatus(input.Status))
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// UpdateLocation POST /delivery/orders/:orderId/location — Delivery agent only
func (h *DeliveryHandler) UpdateLocation(c *gin.Context) {
	agentID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	var input model.UpdateLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid location: "+err.Error(), nil)
		return
	}

	out, err := h.svc.UpdateLocation(c.Request.Context(), orderID, agentID, input.Latitude, input.Longitude)
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// GetLocation GET /delivery/orders/:orderId/location
func (h *DeliveryHandler) GetLocation(c *gin.Context) {
	actorID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid order id", nil)
		return
	}

	roles := getUserRoles(c)
	out, err := h.svc.GetLocation(c.Request.Context(), orderID, actorID, roles)
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// ListAllAssignments GET /delivery/admin/assignments — Admin only
func (h *DeliveryHandler) ListAllAssignments(c *gin.Context) {
	out, err := h.svc.ListAllAssignments(c.Request.Context())
	if err != nil {
		handleDeliveryError(c, err)
		return
	}

	if out == nil {
		out = []model.DeliveryAssignmentWithOrderOutput{}
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
