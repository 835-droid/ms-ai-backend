// Package health provides health check endpointspackage health

package health

import (
	"context"
	"net/http"
	"time"

	datamongo "github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler provides health check endpoints
type Handler struct {
	mongoStore *datamongo.MongoStore
	lastCheck  time.Time
	isReady    bool
}

// NewHandler creates a new health handler backed by MongoStore for richer metrics
func NewHandler(store *datamongo.MongoStore) *Handler {
	return &Handler{
		mongoStore: store,
		lastCheck:  time.Time{},
		isReady:    false,
	}
}

// LivenessCheck godoc
// @Summary Liveness probe
// @Description Check if the application is live
// @Tags health
// @Produce json
// @Success 200 {object} response.Response
// @Router /health/live [get]
func (h *Handler) LivenessCheck(c *gin.Context) {
	response.SuccessResp(c, http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ReadinessCheck godoc
// @Summary Readiness probe
// @Description Check if the application is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} response.Response
// @Failure 503 {object} response.ErrorResponse
// @Router /health/ready [get]
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

	// Check MongoDB connection via store helper
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	err := h.mongoStore.Healthy(ctx)
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

// GetMetrics godoc
// @Summary Application metrics
// @Description Get basic application metrics
// @Tags health
// @Produce json
// @Success 200 {object} response.Response
// @Router /metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	stats := h.mongoStore.GetStats()
	detailed, _ := h.mongoStore.GetDetailedHealth(c.Request.Context())
	response.SuccessResp(c, http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics": gin.H{
			"uptime": time.Since(h.lastCheck).String(),
			"health": gin.H{
				"database": h.isReady,
			},
			"mongo": gin.H{
				"stats":    stats,
				"detailed": detailed,
			},
		},
	})
}

// DebugDBCheck attempts a direct ping to MongoDB and returns the raw driver error (useful for debugging connectivity issues).
// WARNING: This endpoint exposes internal error messages and should be disabled or protected in production.
func (h *Handler) DebugDBCheck(c *gin.Context) {
	// short timeout for debugging
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if h.mongoStore == nil || h.mongoStore.Client == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "mongo store or client not initialized")
		return
	}

	// Try a ping with primary read preference
	if err := h.mongoStore.Client.Ping(ctx, nil); err != nil {
		// return raw error to aid debugging
		response.ErrorResp(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"status": "connected"})
}
