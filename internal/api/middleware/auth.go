package middleware

import (
	"net/http"
	"strings"

	"github.com/835-droid/ms-ai-backend/pkg/config"
	tokenpkg "github.com/835-droid/ms-ai-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserRolesKey = "user_roles"
)

// AuthMiddleware verifies a JWT and sets claims in the context.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		tokenStr := parts[1]

		claims, err := tokenpkg.ValidateToken(tokenStr, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Set user id and roles directly from typed claims
		if claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRolesKey, claims.Roles)
		c.Next()
	}
}

// RequireRole returns middleware that checks for a given role in JWT claims.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ContextUserRolesKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		roles, _ := v.([]string)
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
