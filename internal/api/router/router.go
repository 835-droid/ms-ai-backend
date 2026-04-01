package router

import (
	"net/http"

	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/i18n"
	plog "github.com/835-droid/ms-ai-backend/pkg/logger"

	adminrouter "github.com/835-droid/ms-ai-backend/internal/api/router/admin"
	authrouter "github.com/835-droid/ms-ai-backend/internal/api/router/auth"
	mangarouter "github.com/835-droid/ms-ai-backend/internal/api/router/content/manga"

	"github.com/gin-gonic/gin"
)

func Setup(
	cfg *config.Config,
	log *plog.Logger,
	container *handler.Container,
) *gin.Engine {
	r := gin.New()

	authHandler := container.AuthHandler
	adminHandler := container.AdminHandler
	healthHandler := container.HealthHandler
	mangaHandler := container.MangaHandler
	mangaChapterHandler := container.MangaChapterHandler

	r.Use(i18n.LanguageMiddleware())
	r.Use(middleware.RecoveryMiddleware(log))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.CORSMiddleware(cfg))

	if mangaHandler != nil && mangaChapterHandler != nil {
		mangarouter.SetupMangaRoutes(r, mangaHandler, mangaChapterHandler, cfg)
	}

	r.GET("/livez", healthHandler.LivenessCheck)
	r.GET("/readyz", healthHandler.ReadinessCheck)
	r.GET("/debug/db", healthHandler.DebugDBCheck)

	authrouter.SetupAuthRoutes(r, authHandler, cfg)
	if adminHandler != nil {
		adminrouter.SetupAdminRoutes(r, adminHandler, cfg)
	}

	r.Static("/web", "./cmd/web")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/index.html")
	})

	return r
}
