// internal/container/repo_initializers.go
package container

import (
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	mangarepo "github.com/835-droid/ms-ai-backend/internal/data/content/manga"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	datauser "github.com/835-droid/ms-ai-backend/internal/data/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

func initializeRepositories(cfg *config.Config, log *logger.Logger, m *mongoinfra.MongoStore, p *pginfra.PostgresStore) *RepoBundle {
	var uRepo coreuser.Repository
	var mangaRepo coremanga.MangaRepository
	var chapterRepo coremanga.MangaChapterRepository

	// Conditionally create mongo user repository only if m is not nil
	var mongoUserRepo coreuser.Repository
	if m != nil {
		mongoUserRepo = datauser.NewMongoUserRepository(m)
	}

	switch cfg.DBType {
	case "postgres":
		if p == nil {
			log.Fatal("PostgreSQL store is nil but DB_TYPE is postgres", nil)
		}
		uRepo = datauser.NewPostgresUserRepository(p)
	case "hybrid":
		if p == nil {
			log.Warn("PostgreSQL store is nil in hybrid mode, using only MongoDB", nil)
			uRepo = mongoUserRepo
		} else if m == nil {
			log.Warn("MongoDB store is nil in hybrid mode, using only PostgreSQL", nil)
			uRepo = datauser.NewPostgresUserRepository(p)
		} else {
			pgRepo := datauser.NewPostgresUserRepository(p)
			uRepo = datauser.NewHybridUserRepository(pgRepo, mongoUserRepo, log)
		}
	default: // mongo
		if m == nil {
			log.Fatal("MongoDB store is nil but DB_TYPE is mongo", nil)
		}
		uRepo = mongoUserRepo
	}

	// Manga repositories - Postgres support and hybrid support
	switch cfg.DBType {
	case "postgres":
		if p == nil {
			log.Fatal("PostgreSQL store is nil but DB_TYPE is postgres", nil)
		}
		mangaRepo = mangarepo.NewPostgresMangaRepository(p)
		chapterRepo = mangarepo.NewPostgresChapterRepository(p)
	case "hybrid":
		if p != nil && m != nil {
			pgMangaRepo := mangarepo.NewPostgresMangaRepository(p)
			mongoMangaRepo := mangarepo.NewMongoMangaRepository(m)
			mangaRepo = mangarepo.NewHybridMangaRepository(pgMangaRepo, mongoMangaRepo, log)
			chapterRepo = mangarepo.NewHybridChapterRepository(mangarepo.NewPostgresChapterRepository(p), mangarepo.NewMongoMangaChapterRepository(m), log)
		} else if p != nil {
			mangaRepo = mangarepo.NewPostgresMangaRepository(p)
			chapterRepo = mangarepo.NewPostgresChapterRepository(p)
		} else if m != nil {
			mangaRepo = mangarepo.NewMongoMangaRepository(m)
			chapterRepo = mangarepo.NewMongoMangaChapterRepository(m)
		} else {
			log.Warn("Hybrid mode with no databases available for manga", nil)
		}
	default:
		if m != nil {
			mangaRepo = mangarepo.NewMongoMangaRepository(m)
			chapterRepo = mangarepo.NewMongoMangaChapterRepository(m)
		} else {
			if cfg.DBType == "hybrid" {
				log.Warn("MongoDB is not available, manga features will be disabled", nil)
			} else {
				log.Warn("MongoDB store is nil, manga repositories will be nil", nil)
			}
		}
	}

	return &RepoBundle{
		User:         uRepo,
		Manga:        mangaRepo,
		MangaChapter: chapterRepo,
	}
}
