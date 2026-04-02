// ----- START OF FILE: backend/MS-AI/internal/api/router/content/manga/manga_routes.go -----
package router

import (
	manga "github.com/835-droid/ms-ai-backend/internal/api/handler/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func SetupMangaRoutes(engine *gin.Engine, mangaHandler *manga.MangaHandler, chapterHandler *manga.MangaChapterHandler, cfg *config.Config, userRepo coreuser.Repository) {
	mangas := engine.Group("/api/mangas")
	{
		mangas.GET("", mangaHandler.ListMangas)
		mangas.GET("/:mangaID", mangaHandler.GetManga)

		writeLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.25)
		engagementLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.15)
		authGroup := mangas.Group("")
		{
			authGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
			authGroup.Use(middleware.RequireRole("admin"))
			authGroup.Use(writeLimit)
			authGroup.POST("", mangaHandler.CreateManga)
			authGroup.PUT("/:mangaID", mangaHandler.UpdateManga)
			authGroup.DELETE("/:mangaID", mangaHandler.DeleteManga)
		}

		engagementGroup := mangas.Group("")
		{
			engagementGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
			engagementGroup.Use(engagementLimit)
			engagementGroup.GET("/favorites", mangaHandler.ListLikedMangas)
			engagementGroup.POST("/:mangaID/view", mangaHandler.IncrementViews)
			engagementGroup.POST("/:mangaID/react", mangaHandler.SetReaction)
			engagementGroup.GET("/:mangaID/my-reaction", mangaHandler.GetUserReaction)
			// engagementGroup.POST("/:mangaID/rate", mangaHandler.RateManga) // Disabled: rating moved to chapters only
			// New engagement routes
			engagementGroup.POST("/:mangaID/favorite", mangaHandler.AddFavorite)
			engagementGroup.DELETE("/:mangaID/favorite", mangaHandler.RemoveFavorite)
			engagementGroup.GET("/:mangaID/favorite", mangaHandler.IsFavorite)
			engagementGroup.GET("/favorites/list", mangaHandler.ListFavorites)
			engagementGroup.POST("/:mangaID/comments", mangaHandler.AddMangaComment)
			engagementGroup.GET("/:mangaID/comments", mangaHandler.ListMangaComments)
			engagementGroup.DELETE("/:mangaID/comments/:comment_id", mangaHandler.DeleteMangaComment)
		}

		chapters := mangas.Group("/:mangaID/chapters")
		{
			readLimit := middleware.RateLimitMiddlewareFromConfig(cfg, 0.50)
			chapters.Use(readLimit)
			chapters.GET("", chapterHandler.ListChapters)
			chapters.GET("/:chapterID", chapterHandler.GetChapter)

			// New chapter engagement routes
			engagementGroup := chapters.Group("")
			{
				engagementGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
				engagementGroup.Use(engagementLimit)
				engagementGroup.POST("/:chapterID/view", chapterHandler.IncrementChapterViews)
				engagementGroup.POST("/:chapterID/rate", chapterHandler.AddChapterRating)
				engagementGroup.POST("/:chapterID/comments", chapterHandler.AddChapterComment)
				engagementGroup.GET("/:chapterID/comments", chapterHandler.ListChapterComments)
				engagementGroup.DELETE("/:chapterID/comments/:comment_id", chapterHandler.DeleteChapterComment)
			}

			writeGroup := chapters.Group("")
			{
				writeGroup.Use(middleware.AuthMiddleware(cfg, userRepo))
				writeGroup.Use(middleware.RequireRole("admin"))
				writeGroup.Use(writeLimit)
				writeGroup.POST("/upload-images", chapterHandler.UploadChapterImages)
				writeGroup.POST("", chapterHandler.CreateChapter)
				writeGroup.PUT("/:chapterID", chapterHandler.UpdateChapter)
				writeGroup.DELETE("/:chapterID", chapterHandler.DeleteChapter)
			}
		}
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/router/content/manga/manga_routes.go -----
