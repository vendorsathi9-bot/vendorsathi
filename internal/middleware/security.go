package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets conservative security response headers for an
// API service. These are safe for JSON responses and do not interfere with
// the mobile clients or the admin dashboard.
func SecurityHeadersMiddleware(enableHSTS bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(self), camera=(), microphone=()")
		// The API is intentionally consumed by separate admin/landing origins.
		c.Header("Cross-Origin-Resource-Policy", "cross-origin")
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"font-src 'self'; "+
				"frame-src 'none';")
		if enableHSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		c.Next()
	}
}

// BodyLimitMiddleware rejects request bodies larger than maxBytes before they
// are read by a handler. This bounds memory use for every JSON endpoint.
func BodyLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
