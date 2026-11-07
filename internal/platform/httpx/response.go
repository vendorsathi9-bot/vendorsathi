package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ErrorDetail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type ErrorResponse struct {
	Error ErrorEnvelope `json:"error"`
}

type ErrorEnvelope struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	RequestID string        `json:"request_id"`
}

func Respond(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func RespondOK(c *gin.Context, data interface{}) {
	Respond(c, http.StatusOK, data)
}

func RespondCreated(c *gin.Context, data interface{}) {
	Respond(c, http.StatusCreated, data)
}

func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func getRequestID(c *gin.Context) string {
	if rid, exists := c.Get("request_id"); exists {
		if s, ok := rid.(string); ok {
			return s
		}
	}
	return uuid.New().String()
}

func Error(c *gin.Context, status int, code, message string, details []ErrorDetail) {
	requestID := getRequestID(c)
	c.JSON(status, ErrorResponse{
		Error: ErrorEnvelope{
			Code:      code,
			Message:   message,
			Details:   details,
			RequestID: requestID,
		},
	})
}

func ErrorBadRequest(c *gin.Context, message string, details []ErrorDetail) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func ErrorUnauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func ErrorForbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func ErrorNotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func ErrorConflict(c *gin.Context, message string, details []ErrorDetail) {
	Error(c, http.StatusConflict, "CONFLICT", message, details)
}

func ErrorUnprocessable(c *gin.Context, message string, details []ErrorDetail) {
	Error(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message, details)
}

func ErrorTooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, "RATE_LIMITED", message, nil)
}

func ErrorInternal(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}

func ErrorServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, nil)
}

type PaginationParams struct {
	Limit  int    `form:"limit,default=20" binding:"min=1,max=100"`
	Cursor string `form:"cursor"`
}

type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func RespondPaginated[T any](c *gin.Context, data []T, nextCursor string, hasMore bool) {
	RespondOK(c, PaginatedResponse[T]{
		Data:       data,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}
