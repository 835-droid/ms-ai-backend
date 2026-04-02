// ----- START OF FILE: backend/MS-AI/internal/api/middleware/recovery.go -----
package middleware

import (
	"net/http"
	"runtime/debug"

	plog "github.com/835-droid/ms-ai-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(log *plog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", map[string]interface{}{"panic": r, "stack": string(debug.Stack())})
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/middleware/recovery.go -----
