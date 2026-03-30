package router

import (
	auth "github.com/835-droid/ms-ai-backend/internal/api/handler/auth"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(engine *gin.Engine, authHandler *auth.Handler, cfg *config.Config) {
	// Auth routes with rate limiting at 5% of global limit
	authLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.05)

	auth := engine.Group("/api/auth")
	{
		auth.Use(authLimit)
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		// Protected route: requires a valid access token
		auth.POST("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)
		auth.GET("/verify", middleware.AuthMiddleware(cfg), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "authenticated"})
		})
	}
}
