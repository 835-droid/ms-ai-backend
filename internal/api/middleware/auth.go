// ----- START OF FILE: backend/MS-AI/internal/api/middleware/auth.go -----
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	tokenpkg "github.com/835-droid/ms-ai-backend/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserRolesKey = "user_roles"
)

// AuthMiddleware verifies JWT and optionally checks user activity.
func AuthMiddleware(cfg *config.Config, userRepo user.Repository) gin.HandlerFunc {
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

		if claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRolesKey, claims.Roles)

		// Check if user is still active (only if userRepo provided)
		if userRepo != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			objID, err := primitive.ObjectIDFromHex(claims.UserID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id format"})
				return
			}
			u, err := userRepo.FindByID(ctx, objID)
			if err != nil || u == nil || !u.IsActive {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user inactive or not found"})
				return
			}
		}
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

// RequireAnyRole returns middleware that passes if user has at least one of the given roles.
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ContextUserRolesKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		userRoles, _ := v.([]string)
		for _, requiredRole := range roles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/middleware/auth.go -----
