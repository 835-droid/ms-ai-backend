// ----- START OF FILE: backend/MS-AI/internal/api/router/admin/admin_routes.go -----
package router

import (
	admin "github.com/835-droid/ms-ai-backend/internal/api/handler/admin"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(engine *gin.Engine, adminHandler *admin.Handler, cfg *config.Config, userRepo coreuser.Repository) {
	// Use 5% of rate limit for admin endpoints (more permissive for dashboard usage)
	adminLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.05)

	adminGroup := engine.Group("/api/admin")
	{
		adminGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
		adminGroup.Use(middleware.RequireRole(string(coreuser.RoleAdmin)))
		adminGroup.Use(adminLimit)

		adminGroup.GET("/check", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "admin"})
		})

		adminGroup.POST("/invite", adminHandler.CreateInvite)
		adminGroup.GET("/invites", adminHandler.ListInvites)
		adminGroup.DELETE("/invite/:id", adminHandler.DeleteInvite)

		adminGroup.GET("/metrics", adminHandler.GetMetrics)
		adminGroup.GET("/metrics/db", adminHandler.GetDBMetrics)

		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.PUT("/users/:id/promote", adminHandler.PromoteUser)
		adminGroup.PUT("/users/:id/demote", adminHandler.DemoteUser)
		adminGroup.PUT("/users/:id/password", adminHandler.ChangePassword)
		adminGroup.PUT("/users/:id/deactivate", adminHandler.DeactivateUser)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/router/admin/admin_routes.go -----
