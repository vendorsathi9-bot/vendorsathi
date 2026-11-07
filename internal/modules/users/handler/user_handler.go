package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/users/model"
	"vendorsaathi/internal/modules/users/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Me(c *gin.Context) {
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

	output, err := h.userService.GetMe(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) GetUser(c *gin.Context) {
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

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid user id", nil)
		return
	}

	output, err := h.userService.GetUser(c.Request.Context(), actorID, targetID, isAdmin)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
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

	var input model.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.userService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
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

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, input.CurrentPassword, input.NewPassword); err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondNoContent(c)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
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

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid user id", nil)
		return
	}

	var input model.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.userService.UpdateStatus(c.Request.Context(), actorID, targetID, isAdmin, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) AssignRole(c *gin.Context) {
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

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid user id", nil)
		return
	}

	var input model.AssignRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.userService.AssignRole(c.Request.Context(), actorID, targetID, isAdmin, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) RevokeRole(c *gin.Context) {
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

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid user id", nil)
		return
	}

	var input model.RevokeRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.userService.RevokeRole(c.Request.Context(), actorID, targetID, isAdmin, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isAdmin := false
	for _, r := range userRoles {
		if r == "ADMIN" || r == "SUPER_ADMIN" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		httpx.ErrorForbidden(c, "admin access required")
		return
	}

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid query parameters", nil)
		return
	}

	var status *model.UserStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.UserStatus(statusStr)
		status = &s
	}

	output, err := h.userService.ListUsers(c.Request.Context(), params.Limit, params.Cursor, status)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondPaginated(c, output.Users, output.NextCursor, output.HasMore)
}

func (h *UserHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "NOT_FOUND":
			httpx.ErrorNotFound(c, appErr.Message)
		case "FORBIDDEN":
			httpx.ErrorForbidden(c, appErr.Message)
		case "CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "INVALID_PASSWORD":
			httpx.ErrorUnauthorized(c, appErr.Message)
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
		httpx.ErrorNotFound(c, "user not found")
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
