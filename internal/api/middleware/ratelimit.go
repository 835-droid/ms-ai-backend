package middleware

import (
	"net/http"
	"sync"

	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu      sync.Mutex
	clients map[string]*rate.Limiter
	r       rate.Limit
	b       int
}

func NewIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{clients: make(map[string]*rate.Limiter), r: r, b: b}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()
	limiter, exists := il.clients[ip]
	if !exists {
		limiter = rate.NewLimiter(il.r, il.b)
		il.clients[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware returns a gin middleware enforcing per-IP rate limits.
func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	il := NewIPLimiter(r, b)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := il.get(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

// RateLimitMiddlewareFromConfig builds a rate limiter using configuration values.
// portion is a multiplier of the configured RATE_LIMIT_REQUESTS (0.0-1.0). For example,
// portion=0.25 uses 25% of the configured requests for write operations.
func RateLimitMiddlewareFromConfig(cfg *config.Config, portion float64) gin.HandlerFunc {
	if cfg == nil {
		return RateLimitMiddleware(1, 1)
	}
	if portion <= 0 || portion > 1 {
		portion = 1
	}
	total := cfg.RateLimitRequests
	if total <= 0 {
		total = 100
	}
	window := cfg.RateLimitWindow
	if window <= 0 {
		window = cfg.RateLimitWindow
	}
	// requests per second
	rps := (float64(total) / window.Seconds()) * portion
	if rps <= 0 {
		rps = 1
	}
	burst := int(float64(total) * portion)
	if burst < 1 {
		burst = 1
	}
	return RateLimitMiddleware(rate.Limit(rps), burst)
}
