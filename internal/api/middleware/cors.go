package middleware

import (
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a simple CORS middleware.
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.ToLower(strings.TrimSpace(c.GetHeader("Origin")))
		if origin == "" {
			// No Origin header - skip CORS headers
			c.Next()
			return
		}

		allowAll := cfg.CORSOrigins == "*"
		// Note: Cannot use wildcard (*) with Access-Control-Allow-Credentials
		// Browsers will reject this combination for security reasons
		if allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			if _, ok := cfg.AllowedOriginsSet[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Origin not in allowlist - skip CORS headers
				c.Next()
				return
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
