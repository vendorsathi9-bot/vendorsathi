package middleware

import (
	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/platform/logger"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		logger.Logger.Error().
			Str("request_id", requestID.(string)).
			Interface("panic", recovered).
			Msg("panic recovered")

		httpx.ErrorInternal(c, "internal server error")
		c.Abort()
	})
}
