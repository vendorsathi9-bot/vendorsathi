package middleware

import (
	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/platform/logger"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Never panic inside the panic handler: request_id may be absent or
		// stored under an unexpected type.
		requestID := c.GetString("request_id")
		logger.Logger.Error().
			Str("request_id", requestID).
			Interface("panic", recovered).
			Str("route", c.FullPath()).
			Str("method", c.Request.Method).
			Msg("panic recovered")

		httpx.ErrorInternal(c, "internal server error")
		c.Abort()
	})
}
