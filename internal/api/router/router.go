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

// Setup initializes the gin engine, routes, and middleware.
func Setup(
	cfg *config.Config,
	log *plog.Logger,
	// store was removed — route setup doesn't require the data store directly
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	healthHandler *handler.HealthHandler,
	mangaHandler *handler.MangaHandler,
	mangaChapterHandler *handler.MangaChapterHandler,
) *gin.Engine {
	r := gin.New()

	// global middleware
	r.Use(i18n.LanguageMiddleware())
	r.Use(middleware.RecoveryMiddleware(log))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.CORSMiddleware(cfg))
	if mangaHandler != nil && mangaChapterHandler != nil {
		mangarouter.SetupMangaRoutes(r, mangaHandler, mangaChapterHandler, cfg)
	}

	// health endpoints (no auth)
	r.GET("/livez", healthHandler.LivenessCheck)
	r.GET("/readyz", healthHandler.ReadinessCheck)
	// Debug endpoint for DB connectivity (exposes raw driver errors — disable in production)
	r.GET("/debug/db", healthHandler.DebugDBCheck)

	// Wire sub-route groups using route helpers
	authrouter.SetupAuthRoutes(r, authHandler, cfg)
	adminrouter.SetupAdminRoutes(r, adminHandler, cfg)

	// Static file serving for web UI
	r.Static("/web", "./cmd/web")

	// Redirect root to web UI
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/index.html")
	})

	return r
}
