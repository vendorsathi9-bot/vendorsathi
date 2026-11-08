package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/vendors/model"
	"vendorsaathi/internal/modules/vendors/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type VendorHandler struct {
	vendorService *service.VendorService
}

func NewVendorHandler(vendorService *service.VendorService) *VendorHandler {
	return &VendorHandler{vendorService: vendorService}
}

func (h *VendorHandler) CreateProfile(c *gin.Context) {
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

	var input model.CreateVendorProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.CreateProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *VendorHandler) GetMyProfile(c *gin.Context) {
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

	output, err := h.vendorService.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) UpdateMyProfile(c *gin.Context) {
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

	var input model.UpdateVendorProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.UpdateMyProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) GetProfileByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid vendor id", nil)
		return
	}

	output, err := h.vendorService.GetProfileByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) ListVendors(c *gin.Context) {
	vendors, err := h.vendorService.ListVendors(c.Request.Context())
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}
	httpx.RespondOK(c, vendors)
}

func (h *VendorHandler) UpdateProfileStatus(c *gin.Context) {
	actorIDStr, _ := c.Get("user_id")
	actorID, _ := uuid.Parse(actorIDStr.(string))

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isAdmin := false
	for _, r := range userRoles {
		if r == "ADMIN" || r == "SUPER_ADMIN" {
			isAdmin = true
			break
		}
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid vendor id", nil)
		return
	}

	var input model.UpdateVendorStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.UpdateProfileStatus(c.Request.Context(), actorID, isAdmin, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) CreateLocation(c *gin.Context) {
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

	var input model.CreateVendorLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.CreateLocation(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *VendorHandler) GetMyLocations(c *gin.Context) {
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

	output, err := h.vendorService.GetMyLocations(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) GetLocationByID(c *gin.Context) {
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

	locationIDStr := c.Param("locationId")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid location id", nil)
		return
	}

	output, err := h.vendorService.GetLocationByID(c.Request.Context(), userID, locationID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) UpdateLocation(c *gin.Context) {
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

	locationIDStr := c.Param("locationId")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid location id", nil)
		return
	}

	var input model.UpdateVendorLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.UpdateLocation(c.Request.Context(), userID, locationID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) DeleteLocation(c *gin.Context) {
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

	locationIDStr := c.Param("locationId")
	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid location id", nil)
		return
	}

	if err := h.vendorService.DeleteLocation(c.Request.Context(), userID, locationID); err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

func (h *VendorHandler) SetBusinessHours(c *gin.Context) {
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

	var input model.SetBusinessHoursInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.vendorService.SetBusinessHours(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) GetBusinessHours(c *gin.Context) {
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

	output, err := h.vendorService.GetBusinessHours(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) GetVendorDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid vendor id", nil)
		return
	}

	output, err := h.vendorService.GetVendorDetail(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *VendorHandler) handleError(c *gin.Context, err error) {
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
		httpx.ErrorNotFound(c, "vendor not found")
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
