// ----- START OF FILE: backend/MS-AI/internal/api/handler/container.go -----
package handler

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler/health"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
)

type Container struct {
	AuthHandler         *AuthHandler
	AdminHandler        *AdminHandler
	HealthHandler       *health.Handler
	MangaHandler        *MangaHandler
	MangaChapterHandler *MangaChapterHandler
}

func NewContainer(
	authService coreauth.AuthService,
	mangaService coremanga.MangaService,
	mangaChapterService coremanga.MangaChapterService,
	adminService coreadmin.Service,
	mongoStore *mongoinfra.MongoStore,
	postgresStore *pginfra.PostgresStore,
) *Container {
	var adminHandler *AdminHandler
	if adminService != nil {
		adminHandler = NewAdminHandler(adminService)
	}

	var mangaHandler *MangaHandler
	var chapterHandler *MangaChapterHandler
	if mangaService != nil {
		mangaHandler = NewMangaHandler(mangaService)
	}
	if mangaChapterService != nil {
		chapterHandler = NewMangaChapterHandler(mangaChapterService)
	}

	return &Container{
		AuthHandler:         NewAuthHandler(authService),
		AdminHandler:        adminHandler,
		HealthHandler:       health.NewHandler(mongoStore, postgresStore),
		MangaHandler:        mangaHandler,
		MangaChapterHandler: chapterHandler,
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/container.go -----
