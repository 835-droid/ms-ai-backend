// Package router sets up novel API routes.
package router

import (
	novel "github.com/835-droid/ms-ai-backend/internal/api/handler/content/novel"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

// SetupNovelRoutes sets up all novel-related API routes.
func SetupNovelRoutes(engine *gin.Engine, novelHandler *novel.NovelHandler, cfg *config.Config, userRepo coreuser.Repository) {
	novels := engine.Group("/api/novels")
	{
		novels.GET("", novelHandler.ListNovels)
		novels.GET("/most-viewed", novelHandler.ListMostViewed)
		novels.GET("/recently-updated", novelHandler.ListRecentlyUpdated)
		novels.GET("/most-followed", novelHandler.ListMostFollowed)
		novels.GET("/top-rated", novelHandler.ListTopRated)
		novels.GET("/:novelID", novelHandler.GetNovel)

		writeLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.25)
		engagementLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.15)

		authGroup := novels.Group("")
		{
			authGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
			authGroup.Use(middleware.RequireRole("admin"))
			authGroup.Use(writeLimit)
			authGroup.POST("", novelHandler.CreateNovel)
			authGroup.PUT("/:novelID", novelHandler.UpdateNovel)
			authGroup.DELETE("/:novelID", novelHandler.DeleteNovel)
		}

		// View endpoint uses optional auth to allow anonymous visits for trending/most-viewed tracking
		viewGroup := novels.Group("")
		{
			viewGroup.Use(middleware.OptionalAuthMiddleware(cfg, userRepo))
			viewGroup.Use(engagementLimit)
			viewGroup.POST("/:novelID/view", novelHandler.IncrementViews)
		}

		engagementGroup := novels.Group("")
		{
			engagementGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
			engagementGroup.Use(engagementLimit)
			engagementGroup.GET("/favorites", novelHandler.ListFavorites)
			engagementGroup.POST("/:novelID/react", novelHandler.SetReaction)
			engagementGroup.GET("/:novelID/my-reaction", novelHandler.GetUserReaction)
			engagementGroup.POST("/:novelID/rate", novelHandler.RateNovel)
			// Engagement routes
			engagementGroup.POST("/:novelID/favorite", novelHandler.AddFavorite)
			engagementGroup.DELETE("/:novelID/favorite", novelHandler.RemoveFavorite)
			engagementGroup.GET("/:novelID/favorite", novelHandler.IsFavorite)
			engagementGroup.POST("/:novelID/comments", novelHandler.AddNovelComment)
			engagementGroup.GET("/:novelID/comments", novelHandler.ListNovelComments)
			engagementGroup.DELETE("/:novelID/comments/:comment_id", novelHandler.DeleteNovelComment)
		}
	}
}
