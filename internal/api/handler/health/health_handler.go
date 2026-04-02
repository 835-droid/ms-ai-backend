// ----- START OF FILE: backend/MS-AI/internal/api/handler/health/health_handler.go -----
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthChecker defines an interface for database health checks
type HealthChecker interface {
	Healthy(ctx context.Context) error
}

type Handler struct {
	mongoStore    *mongo.MongoStore
	postgresStore *postgres.PostgresStore
	lastCheck     time.Time
	isReady       bool
}

// NewHandler creates a new health handler
func NewHandler(m *mongo.MongoStore, p *postgres.PostgresStore) *Handler {
	return &Handler{
		mongoStore:    m,
		postgresStore: p,
		isReady:       false,
	}
}

func (h *Handler) LivenessCheck(c *gin.Context) {
	response.SuccessResp(c, http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) ReadinessCheck(c *gin.Context) {
	// Use cached result if checked recently
	if !h.lastCheck.IsZero() && time.Since(h.lastCheck) < 5*time.Second {
		if h.isReady {
			response.SuccessResp(c, http.StatusOK, gin.H{
				"status":    "ready",
				"cached":    true,
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}
		response.ErrorResp(c, http.StatusServiceUnavailable, "Service not ready")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var err error
	// Check whichever database is available
	if h.mongoStore != nil {
		err = h.mongoStore.Healthy(ctx)
	} else if h.postgresStore != nil {
		err = h.postgresStore.Healthy(ctx)
	} else {
		err = nil // no database configured? treat as ready? depends on your need
	}

	h.lastCheck = time.Now()
	h.isReady = err == nil

	if !h.isReady {
		response.ErrorResp(c, http.StatusServiceUnavailable, "Database not ready")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": "connected",
		},
		"timestamp": h.lastCheck.Format(time.RFC3339),
	})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if h.mongoStore != nil {
		stats := h.mongoStore.GetStats()
		metrics["mongo"] = gin.H{
			"stats": stats,
		}
	}
	if h.postgresStore != nil {
		// PostgreSQL stats could be added here
		metrics["postgres"] = gin.H{
			"connected": true,
		}
	}
	response.SuccessResp(c, http.StatusOK, metrics)
}

func (h *Handler) DebugDBCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if h.mongoStore != nil {
		if err := h.mongoStore.Healthy(ctx); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.SuccessResp(c, http.StatusOK, gin.H{"status": "connected", "database": "mongodb"})
	} else if h.postgresStore != nil {
		if err := h.postgresStore.Healthy(ctx); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.SuccessResp(c, http.StatusOK, gin.H{"status": "connected", "database": "postgres"})
	} else {
		response.ErrorResp(c, http.StatusInternalServerError, "no database configured")
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/health/health_handler.go -----
