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
		origin := c.GetHeader("Origin")

		// Handle empty origin or file:// for development
		if origin == "" || origin == "file://" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			allowAll := cfg.CORSOrigins == "*"
			if allowAll {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Check if origin is in allowed list (case-sensitive)
				if _, ok := cfg.AllowedOriginsSet[origin]; ok {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					// For development, allow localhost variations
					if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
						c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					} else {
						// Origin not allowed - still set CORS headers but don't allow
						c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
					}
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, Accept, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/middleware/cors.go -----
