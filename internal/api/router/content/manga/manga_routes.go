package router

import (
	manga "github.com/835-droid/ms-ai-backend/internal/api/handler/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupMangaRoutes(engine *gin.Engine, mangaHandler *manga.MangaHandler, chapterHandler *manga.MangaChapterHandler, cfg *config.Config) {
	// Manga routes
	mangas := engine.Group("/api/mangas")
	{
		// Read operations - 100% rate limit
		mangas.GET("", mangaHandler.ListMangas)
		mangas.GET("/:mangaID", mangaHandler.GetManga)

		// Write operations - 25% rate limit
		writeLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.25)
		authGroup := mangas.Group("")
		{
			authGroup.Use(middleware.AuthMiddleware(cfg))
			authGroup.Use(writeLimit)
			authGroup.POST("", mangaHandler.CreateManga)
			authGroup.PUT("/:mangaID", mangaHandler.UpdateManga)
			authGroup.DELETE("/:mangaID", mangaHandler.DeleteManga)
		}

		// Chapters
		chapters := mangas.Group("/:mangaID/chapters")
		{
			// Read operations - 50% rate limit
			readLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.50)
			chapters.Use(readLimit)
			chapters.GET("", chapterHandler.ListChapters)
			chapters.GET("/:chapter_number", chapterHandler.GetChapter)

			// Write operations - 25% rate limit
			writeGroup := chapters.Group("")
			{
				writeGroup.Use(middleware.AuthMiddleware(cfg))
				writeGroup.Use(writeLimit)
				writeGroup.POST("", chapterHandler.CreateChapter)
				writeGroup.PUT("/:chapter_number", chapterHandler.UpdateChapter)
				writeGroup.DELETE("/:chapter_number", chapterHandler.DeleteChapter)
			}
		}
	}
}
