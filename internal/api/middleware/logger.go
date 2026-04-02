// ----- START OF FILE: backend/MS-AI/internal/api/middleware/logger.go -----
package middleware

import (
	"time"

	plog "github.com/835-droid/ms-ai-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(log *plog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		reqID := ""
		if v, ok := c.Get(RequestIDKey); ok {
			if s, ok := v.(string); ok {
				reqID = s
			}
		}
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     status,
			"latency":    latency.String(),
			"client_ip":  c.ClientIP(),
			"request_id": reqID,
		}
		log.Info("http_request", fields)
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/middleware/logger.go -----
