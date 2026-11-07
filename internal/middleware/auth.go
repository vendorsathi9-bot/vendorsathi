package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/platform/httpx"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpx.ErrorUnauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			httpx.ErrorUnauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			if errors.Is(err, auth.ErrExpiredToken) {
				httpx.ErrorUnauthorized(c, "access token expired")
			} else {
				httpx.ErrorUnauthorized(c, "invalid access token")
			}
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("roles")
		if !exists {
			httpx.ErrorForbidden(c, "role information not found")
			c.Abort()
			return
		}

		userRoles, ok := rolesVal.([]string)
		if !ok {
			httpx.ErrorForbidden(c, "invalid role data")
			c.Abort()
			return
		}

		allowed := make(map[string]bool)
		for _, r := range allowedRoles {
			allowed[r] = true
		}

		hasRole := false
		for _, r := range userRoles {
			if allowed[r] {
				hasRole = true
				break
			}
		}

		if !hasRole {
			httpx.ErrorForbidden(c, "insufficient role permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequirePermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsVal, exists := c.Get("permissions")
		if !exists {
			httpx.ErrorForbidden(c, "permission information not found")
			c.Abort()
			return
		}

		userPerms, ok := permsVal.([]string)
		if !ok {
			httpx.ErrorForbidden(c, "invalid permission data")
			c.Abort()
			return
		}

		userPermSet := make(map[string]bool)
		for _, p := range userPerms {
			userPermSet[p] = true
		}

		for _, required := range requiredPermissions {
			if !userPermSet[required] {
				httpx.ErrorForbidden(c, "missing required permission: "+required)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
