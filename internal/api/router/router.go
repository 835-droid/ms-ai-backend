// ----- START OF FILE: backend/MS-AI/internal/api/router/router.go -----
package router

import (
	"net/http"

	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
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
	userRepo user.Repository,
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
		mangarouter.SetupMangaRoutes(r, mangaHandler, mangaChapterHandler, cfg, userRepo)
		r.GET("/api/assets/image-proxy", mangaChapterHandler.ProxyImage)
	}

	r.GET("/health", healthHandler.LivenessCheck)
	r.GET("/livez", healthHandler.LivenessCheck)
	r.GET("/readyz", healthHandler.ReadinessCheck)
	r.GET("/debug/db", healthHandler.DebugDBCheck)

	authrouter.SetupAuthRoutes(r, authHandler, cfg, userRepo)
	if adminHandler != nil {
		adminrouter.SetupAdminRoutes(r, adminHandler, cfg, userRepo)
	}

	r.Static("/web", "./cmd/web")
	r.Static("/uploads", "./uploads")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/index.html")
	})

	return r
}

// ----- END OF FILE: backend/MS-AI/internal/api/router/router.go -----
