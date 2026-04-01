package container

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	mangarepo "github.com/835-droid/ms-ai-backend/internal/data/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
	datauser "github.com/835-droid/ms-ai-backend/internal/data/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func initializeDatabases(cfg *config.Config, log *logger.Logger) (*mongo.MongoStore, *postgres.PostgresStore, error) {
	var m *mongo.MongoStore
	var p *postgres.PostgresStore
	var err error

	if cfg.DBType == "mongo" || cfg.DBType == "hybrid" {
		m, err = mongo.NewMongoStore(cfg, log)
		if err != nil {
			log.Warn("failed to initialize mongo store", map[string]interface{}{"error": err.Error()})
			if cfg.DBType == "mongo" {
				return nil, nil, err
			}
			// For hybrid, continue without mongo
		}
	}
	if cfg.DBType == "postgres" || cfg.DBType == "hybrid" {
		p, err = postgres.NewPostgresStore(cfg, log)
		if err != nil {
			return nil, nil, err
		}
	}
	return m, p, nil
}

func initializeRepositories(cfg *config.Config, log *logger.Logger, m *mongo.MongoStore, p *postgres.PostgresStore) *RepoBundle {
	var uRepo coreuser.Repository

	mongoUserRepo := datauser.NewMongoUserRepository(m)

	switch cfg.DBType {
	case "postgres":
		uRepo = postgres.NewUserRepository(p)
	case "hybrid":
		pgRepo := postgres.NewUserRepository(p)
		uRepo = datauser.NewHybridUserRepository(pgRepo, mongoUserRepo, log)
	default:
		uRepo = mongoUserRepo
	}

	// Manga repositories - always use MongoDB for now (could be extended)
	var mangaRepo coremanga.MangaRepository
	var chapterRepo coremanga.MangaChapterRepository
	if m != nil {
		mangaRepo = mangarepo.NewMongoMangaRepository(m)
		chapterRepo = mangarepo.NewMongoMangaChapterRepository(m)
	}

	return &RepoBundle{
		User:         uRepo,
		Manga:        mangaRepo,
		MangaChapter: chapterRepo,
	}
}

func initializeServices(cfg *config.Config, log *logger.Logger, repos *RepoBundle, m *mongo.MongoStore, p *postgres.PostgresStore) *serviceBundle {
	var adminSvc coreadmin.Service
	if cfg.DBType == "postgres" && p != nil {
		// For PostgreSQL, we need an admin repository that uses PostgreSQL
		adminRepo := postgres.NewAdminRepository(p) // we'll create this file next
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else if m != nil {
		// Use MongoDB admin repository
		adminRepo := coreadmin.NewMongoRepository(m.Client.Database(m.DBName))
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	}

	// Manga services
	var mangaSvc coremanga.MangaService
	var chapterSvc coremanga.MangaChapterService
	if repos.Manga != nil && repos.MangaChapter != nil {
		mangaSvc = coremanga.NewMangaService(repos.Manga, log)
		chapterSvc = coremanga.NewMangaChapterService(repos.MangaChapter, repos.Manga, log)
	}

	return &serviceBundle{
		Auth:    coreauth.NewAuthService(repos.User, cfg, log.GetZerologLogger()),
		Admin:   adminSvc,
		Manga:   mangaSvc,
		Chapter: chapterSvc,
	}
}

func initializeInitialData(ctx context.Context, repos *RepoBundle, log *logger.Logger) {
	admin, _ := repos.User.FindByUsername(ctx, "admin")
	if admin == nil {
		hp, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		newUser := &coreuser.User{
			Username: "admin",
			Password: string(hp),
			Roles:    []string{"admin"},
			IsActive: true,
		}
		_ = repos.User.Create(ctx, newUser, &coreuser.UserDetails{
			Roles:    []string{"admin"},
			IsActive: true,
			Status:   "active",
		})
	}
}

func initializeHandlers(svcs *serviceBundle, m *mongo.MongoStore, p *postgres.PostgresStore) *handler.Container {
	return handler.NewContainer(svcs.Auth, svcs.Manga, svcs.Chapter, svcs.Admin, m, p)
}
