package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/inventory/model"
	"vendorsaathi/internal/modules/inventory/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	output, err := h.inventoryService.GetInventory(c.Request.Context(), productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *InventoryHandler) GetMyInventory(c *gin.Context) {
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

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	output, err := h.inventoryService.GetInventoryByVendor(c.Request.Context(), vendorID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *InventoryHandler) AdjustStock(c *gin.Context) {
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

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	var input model.AdjustStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	invOutput, movOutput, err := h.inventoryService.AdjustStock(c.Request.Context(), vendorID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{
		"inventory": invOutput,
		"movement":  movOutput,
	})
}

func (h *InventoryHandler) ReserveStock(c *gin.Context) {
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

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	var input model.ReserveStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	invOutput, movOutput, err := h.inventoryService.ReserveStock(c.Request.Context(), vendorID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{
		"inventory": invOutput,
		"movement":  movOutput,
	})
}

func (h *InventoryHandler) ReleaseStock(c *gin.Context) {
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

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	var input model.ReleaseStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	invOutput, movOutput, err := h.inventoryService.ReleaseStock(c.Request.Context(), vendorID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{
		"inventory": invOutput,
		"movement":  movOutput,
	})
}

func (h *InventoryHandler) FulfilStock(c *gin.Context) {
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

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	var input model.FulfilStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	invOutput, movOutput, err := h.inventoryService.FulfilStock(c.Request.Context(), vendorID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, gin.H{
		"inventory": invOutput,
		"movement":  movOutput,
	})
}

func (h *InventoryHandler) GetMovements(c *gin.Context) {
	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid query parameters", nil)
		return
	}

	output, err := h.inventoryService.GetMovements(c.Request.Context(), productID, params.Limit, params.Cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondPaginated(c, output.Data, output.NextCursor, output.HasMore)
}

func (h *InventoryHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "NOT_FOUND":
			httpx.ErrorNotFound(c, appErr.Message)
		case "FORBIDDEN":
			httpx.ErrorForbidden(c, appErr.Message)
		case "CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "INSUFFICIENT_STOCK":
			httpx.ErrorUnprocessable(c, appErr.Message, nil)
		case "IDEMPOTENCY_CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "OPTIMISTIC_LOCK":
			httpx.ErrorConflict(c, "concurrent modification, please retry", nil)
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
	if errors.Is(err, shared.ErrInsufficientStock) {
		httpx.ErrorUnprocessable(c, "insufficient stock available", nil)
		return
	}
	if errors.Is(err, shared.ErrIdempotencyConflict) {
		httpx.ErrorConflict(c, "idempotency key already used with different payload", nil)
		return
	}
	if errors.Is(err, shared.ErrOptimisticLock) {
		httpx.ErrorConflict(c, "concurrent modification, please retry", nil)
		return
	}
	if errors.Is(err, shared.ErrValidationFailed) {
		httpx.ErrorUnprocessable(c, "validation failed", nil)
		return
	}

	httpx.ErrorInternal(c, "internal server error")
}
