package router

import (
	admin "github.com/835-droid/ms-ai-backend/internal/api/handler/admin"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(engine *gin.Engine, adminHandler *admin.Handler, cfg *config.Config) {
	// Admin routes with rate limiting at 2% of global limit
	adminLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.02)

	// Admin routes require authentication and admin role
	adminGroup := engine.Group("/api/admin")
	{
		adminGroup.Use(middleware.AuthMiddleware(cfg))
		adminGroup.Use(middleware.RequireRole("admin"))
		adminGroup.Use(adminLimit)

		// Check admin access
		adminGroup.GET("/check", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "admin"})
		})

		// Invite code management
		adminGroup.POST("/invite", adminHandler.CreateInvite)
		adminGroup.GET("/invites", adminHandler.ListInvites)
		adminGroup.DELETE("/invite/:id", adminHandler.DeleteInvite)

		// Metrics and monitoring
		adminGroup.GET("/metrics", adminHandler.GetMetrics)
		adminGroup.GET("/metrics/db", adminHandler.GetDBMetrics)
	}
}
