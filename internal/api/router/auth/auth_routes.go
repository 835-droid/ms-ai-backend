// ----- START OF FILE: backend/MS-AI/internal/api/router/auth/auth_routes.go -----
package router

import (
	auth "github.com/835-droid/ms-ai-backend/internal/api/handler/auth"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(engine *gin.Engine, authHandler *auth.Handler, cfg *config.Config, userRepo coreuser.Repository) {
	authLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.05)

	auth := engine.Group("/api/auth")
	{
		auth.Use(authLimit)
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(cfg, userRepo), authHandler.Logout)
		auth.PUT("/password", middleware.AuthMiddleware(cfg, userRepo), authHandler.ChangePassword)
		auth.GET("/verify", middleware.AuthMiddleware(cfg, userRepo), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "authenticated"})
		})
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/router/auth/auth_routes.go -----
