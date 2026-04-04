// Package handler provides the dependency injection container for handlers.
package handler

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/api/handler/content/novel"
	"github.com/835-droid/ms-ai-backend/internal/api/handler/health"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	corenovel "github.com/835-droid/ms-ai-backend/internal/core/content/novel"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
)

// Container holds all handler instances for dependency injection.
type Container struct {
	AuthHandler           *AuthHandler
	AdminHandler          *AdminHandler
	HealthHandler         *health.Handler
	MangaHandler          *MangaHandler
	MangaChapterHandler   *MangaChapterHandler
	ViewingHistoryHandler *manga.ViewingHistoryHandler
	NovelHandler          *novel.NovelHandler
}

func NewContainer(
	authService coreauth.AuthService,
	mangaService coremanga.MangaService,
	favListService coremanga.FavoriteListService,
	mangaChapterService coremanga.MangaChapterService,
	adminService coreadmin.Service,
	viewingHistoryService coremanga.ViewingHistoryService,
	novelService corenovel.NovelService,
	mongoStore *mongoinfra.MongoStore,
	postgresStore *pginfra.PostgresStore,
) *Container {
	var adminHandler *AdminHandler
	if adminService != nil {
		adminHandler = NewAdminHandler(adminService)
	}

	var mangaHandler *MangaHandler
	var chapterHandler *MangaChapterHandler
	if mangaService != nil && favListService != nil {
		mangaHandler = NewMangaHandler(mangaService, favListService)
	}
	if mangaChapterService != nil {
		chapterHandler = NewMangaChapterHandler(mangaChapterService)
	}

	var viewingHistoryHandler *manga.ViewingHistoryHandler
	if viewingHistoryService != nil && mangaService != nil {
		viewingHistoryHandler = manga.NewViewingHistoryHandler(viewingHistoryService, mangaService)
	}

	var novelHandler *novel.NovelHandler
	if novelService != nil {
		novelHandler = novel.NewNovelHandler(novelService)
	}

	return &Container{
		AuthHandler:           NewAuthHandler(authService),
		AdminHandler:          adminHandler,
		HealthHandler:         health.NewHandler(mongoStore, postgresStore),
		MangaHandler:          mangaHandler,
		MangaChapterHandler:   chapterHandler,
		ViewingHistoryHandler: viewingHistoryHandler,
		NovelHandler:          novelHandler,
	}
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/container.go -----
