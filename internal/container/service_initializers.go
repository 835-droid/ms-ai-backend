// internal/container/service_initializers.go
package container

import (
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	admindata "github.com/835-droid/ms-ai-backend/internal/data/admin"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

func initializeServices(cfg *config.Config, log *logger.Logger, repos *RepoBundle, m *mongoinfra.MongoStore, p *pginfra.PostgresStore) *serviceBundle {
	var adminSvc coreadmin.Service

	// Choose admin repository: use Mongo admin repo in hybrid mode to follow user repo failover semantics; Postgres for postgres-only.
	if cfg.DBType == "hybrid" && m != nil {
		adminRepo := admindata.NewAdminRepositoryAdapter(repos.User)
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else if p != nil {
		adminRepo := admindata.NewPostgresAdminRepository(p)
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else if m != nil {
		adminRepo := admindata.NewAdminRepositoryAdapter(repos.User)
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else {
		log.Warn("No database available for admin repository, admin service will be nil", nil)
	}

	// Manga services
	var mangaSvc coremanga.MangaService
	var chapterSvc coremanga.MangaChapterService
	var favListSvc coremanga.FavoriteListService
	if repos.Manga != nil && repos.MangaChapter != nil {
		mangaSvc = coremanga.NewMangaService(repos.Manga, log)
		chapterSvc = coremanga.NewMangaChapterService(repos.MangaChapter, repos.Manga, log)
	} else {
		if cfg.DBType == "hybrid" {
			log.Warn("MongoDB is not available, manga services will be disabled", nil)
		}
	}

	// Favorite List service - requires both FavList repo and Manga repo
	if repos.FavList != nil && repos.Manga != nil {
		favListSvc = coremanga.NewFavoriteListService(repos.FavList, repos.Manga, log)
	}

	// Viewing History service - requires ViewingHistory repo and Manga repo
	var viewingHistorySvc coremanga.ViewingHistoryService
	if repos.ViewingHistory != nil && repos.Manga != nil {
		viewingHistorySvc = coremanga.NewViewingHistoryService(repos.ViewingHistory, repos.Manga)
	}

	return &serviceBundle{
		Auth:           coreauth.NewAuthService(repos.User, cfg, log.GetZerologLogger()),
		Admin:          adminSvc,
		Manga:          mangaSvc,
		FavList:        favListSvc,
		Chapter:        chapterSvc,
		ViewingHistory: viewingHistorySvc,
	}
}
