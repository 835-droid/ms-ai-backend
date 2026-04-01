package health

import (
	"context"
	"net/http"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	datamongo "github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
	datapostgres "github.com/835-droid/ms-ai-backend/internal/data/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthChecker defines an interface for database health checks
type HealthChecker interface {
	Healthy(ctx context.Context) error
}

type Handler struct {
	mongoStore    *datamongo.MongoStore
	postgresStore *datapostgres.PostgresStore
	lastCheck     time.Time
	isReady       bool
}

// NewHealthHandler ينشئ نسخة جديدة من الـ HealthHandler مع تمرير الاتصالات اللازمة
func NewHealthHandler(m *mongo.MongoStore, p *postgres.PostgresStore, r *redis.Client) *HealthHandler {
	return &HealthHandler{
		mongoStore:    m,
		postgresStore: p,
		redisClient:   r,
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

// NewHealthHandler ينشئ نسخة جديدة من الـ HealthHandler مع تمرير الاتصالات اللازمة
func NewHealthHandler(m *mongo.MongoStore, p *postgres.PostgresStore, r *redis.Client) *HealthHandler {
	return &HealthHandler{
		mongoStore:    m,
		postgresStore: p,
		redisClient:   r,
		isReady:       false,
	}
}

// Liveness فحص بسيط للتأكد من أن تطبيق Go يعمل (Liveness Probe)
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

// ReadinessCheck يقوم بفحص عميق لجميع الاتصالات (Readiness Probe)
// يدعم الوضع الهجين بفحص كل القواعد النشطة
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx := c.Request.Context()
	var errs []error

	// 1. فحص MongoDB إذا كان مفعلاً
	if h.mongoStore != nil {
		if err := h.mongoStore.Healthy(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	// 2. فحص PostgreSQL إذا كان مفعلاً
	if h.postgresStore != nil {
		if err := h.postgresStore.Healthy(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	// 3. فحص Redis إذا كان مفعلاً (التطوير الإضافي الذي طلبته)
	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			errs = append(errs, err)
		}
	}

	// تحديث حالة الجاهزية بناءً على وجود أخطاء من عدمه
	h.mu.Lock()
	h.isReady = len(errs) == 0
	h.mu.Unlock()

	// إذا كانت هناك أخطاء، نرجع حالة 503 مع قائمة الأخطاء
	if !h.isReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "DISCONNECTED",
			"errors": formatErrors(errs),
		})
		return
	}

	// في حالة النجاح
	response.SuccessResp(c, http.StatusOK, gin.H{
		"status": "READY",
		"services": gin.H{
			"mongo":    h.mongoStore != nil,
			"postgres": h.postgresStore != nil,
			"redis":    h.redisClient != nil,
		},
	})
}

// formatErrors دالة مساعدة لتحويل الأخطاء إلى نصوص مفهومة
func formatErrors(errs []error) []string {
	var s []string
	for _, e := range errs {
		s = append(s, e.Error())
	}
	return s
}
