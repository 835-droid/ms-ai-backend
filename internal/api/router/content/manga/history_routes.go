package router

import (
	manga "github.com/835-droid/ms-ai-backend/internal/api/handler/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupHistoryRoutes(engine *gin.Engine, historyHandler *manga.ViewingHistoryHandler, cfg *config.Config, userRepo coreuser.Repository) {
	history := engine.Group("/api/mangas/history")
	{
		// All history routes require authentication
		history.Use(middleware.AuthMiddleware(cfg, userRepo))

		// Track view (can be called from any page)
		history.POST("/track", historyHandler.TrackView)

		// Get user's history
		history.GET("", historyHandler.GetUserHistory)
		history.GET("/recent", historyHandler.GetRecentHistory)
		history.GET("/stats", historyHandler.GetUserStats)

		// Delete operations
		history.DELETE("/:id", historyHandler.DeleteHistoryItem)
		history.DELETE("/manga/:mangaID", historyHandler.DeleteHistoryByManga)
		history.DELETE("/clean", historyHandler.CleanOldHistory)
	}
}
