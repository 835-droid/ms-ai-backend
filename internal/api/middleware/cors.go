// ----- START OF FILE: backend/MS-AI/internal/api/middleware/cors.go -----
package middleware

import (
	"net/http"
	"strings"

	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a simple CORS middleware.
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.ToLower(strings.TrimSpace(c.GetHeader("Origin")))
		if origin == "" || origin == "file://" {
			// No Origin header or file:// - allow for development
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			allowAll := cfg.CORSOrigins == "*"
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

// ----- END OF FILE: backend/MS-AI/internal/api/middleware/cors.go -----
