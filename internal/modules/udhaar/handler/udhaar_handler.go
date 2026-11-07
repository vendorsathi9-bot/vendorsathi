package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/udhaar/model"
	"vendorsaathi/internal/modules/udhaar/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type UdhaarHandler struct {
	service *service.UdhaarService
}

func NewUdhaarHandler(svc *service.UdhaarService) *UdhaarHandler {
	return &UdhaarHandler{service: svc}
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

func hasRole(c *gin.Context, targetRole string) bool {
	rolesVal, exists := c.Get("roles")
	if !exists {
		return false
	}
	roles, ok := rolesVal.([]string)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == targetRole {
			return true
		}
	}
	return false
}

// CreateAccount POST /udhaar/accounts
func (h *UdhaarHandler) CreateAccount(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.CreateUdhaarAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	acc, err := h.service.CreateAccount(c.Request.Context(), vendorUserID, input)
	if err != nil {
		handleUdhaarError(c, err)
		return
	}

	httpx.RespondCreated(c, acc)
}

// GetAccount GET /udhaar/accounts/:id
func (h *UdhaarHandler) GetAccount(c *gin.Context) {
	actorID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid account id", nil)
		return
	}

	acc, err := h.service.GetAccount(c.Request.Context(), actorID, id)
	if err != nil {
		handleUdhaarError(c, err)
		return
	}

	httpx.RespondOK(c, acc)
}

// ListAccounts GET /udhaar/accounts
func (h *UdhaarHandler) ListAccounts(c *gin.Context) {
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

	if hasRole(c, "VENDOR") {
		accounts, nextCursor, hasMore, err := h.service.GetAccountsByVendor(c.Request.Context(), userID, params.Limit, params.Cursor)
		if err != nil {
			handleUdhaarError(c, err)
			return
		}
		httpx.RespondPaginated(c, accounts, nextCursor, hasMore)
		return
	}

	accounts, nextCursor, hasMore, err := h.service.GetAccountsByCustomer(c.Request.Context(), userID, params.Limit, params.Cursor)
	if err != nil {
		handleUdhaarError(c, err)
		return
	}
	httpx.RespondPaginated(c, accounts, nextCursor, hasMore)
}

// AddEntry POST /udhaar/accounts/:id/entries
func (h *UdhaarHandler) AddEntry(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid account id", nil)
		return
	}

	var input model.AddUdhaarEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	entry, err := h.service.AddEntry(c.Request.Context(), vendorUserID, id, input)
	if err != nil {
		handleUdhaarError(c, err)
		return
	}

	httpx.RespondCreated(c, entry)
}

// GetEntries GET /udhaar/accounts/:id/entries
func (h *UdhaarHandler) GetEntries(c *gin.Context) {
	actorID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid account id", nil)
		return
	}

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid query parameters", nil)
		return
	}

	entries, nextCursor, hasMore, err := h.service.GetEntries(c.Request.Context(), actorID, id, params.Limit, params.Cursor)
	if err != nil {
		handleUdhaarError(c, err)
		return
	}

	httpx.RespondPaginated(c, entries, nextCursor, hasMore)
}

// UpdateAccountStatus PATCH /udhaar/accounts/:id/status
func (h *UdhaarHandler) UpdateAccountStatus(c *gin.Context) {
	vendorUserID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid account id", nil)
		return
	}

	var input model.UpdateUdhaarAccountStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, err.Error(), nil)
		return
	}

	acc, err := h.service.UpdateAccountStatus(c.Request.Context(), vendorUserID, id, model.UdhaarAccountStatus(input.Status))
	if err != nil {
		handleUdhaarError(c, err)
		return
	}

	httpx.RespondOK(c, acc)
}

func handleUdhaarError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, shared.ErrNotFound):
			httpx.ErrorNotFound(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrForbidden):
			httpx.ErrorForbidden(c, appErr.Message)
		case errors.Is(appErr.Err, shared.ErrConflict):
			httpx.ErrorConflict(c, appErr.Message, nil)
		case errors.Is(appErr.Err, shared.ErrUdhaarLimitExceeded):
			httpx.ErrorUnprocessable(c, appErr.Message, nil)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		httpx.ErrorNotFound(c, "not found")
	case errors.Is(err, shared.ErrForbidden):
		httpx.ErrorForbidden(c, "forbidden")
	case errors.Is(err, shared.ErrConflict):
		httpx.ErrorConflict(c, err.Error(), nil)
	case errors.Is(err, shared.ErrUdhaarLimitExceeded):
		httpx.ErrorUnprocessable(c, "repayment exceeds outstanding udhaar balance", nil)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
