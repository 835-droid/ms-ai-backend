// ----- START OF FILE: backend/MS-AI/internal/container/initializers.go -----
package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	admindata "github.com/835-droid/ms-ai-backend/internal/data/admin"
	mangarepo "github.com/835-droid/ms-ai-backend/internal/data/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
	datauser "github.com/835-droid/ms-ai-backend/internal/data/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// initializeDatabases attempts to connect to databases according to DB_TYPE.
// If the primary database fails, it attempts to failover to the other if available.
// It may modify cfg.DBType to reflect the actual working database.
func initializeDatabases(cfg *config.Config, log *logger.Logger) (*mongo.MongoStore, *postgres.PostgresStore, error) {
	var m *mongo.MongoStore
	var p *postgres.PostgresStore
	var err error

	switch cfg.DBType {
	case "mongo":
		// Try MongoDB
		m, err = mongo.NewMongoStore(cfg, log)
		if err != nil {
			log.Warn("MongoDB connection failed, attempting failover to PostgreSQL",
				map[string]interface{}{"error": err.Error()})
			// Attempt to fallback to PostgreSQL if DSN exists
			if cfg.PostgresDSN != "" {
				p, err2 := postgres.NewPostgresStore(cfg, log)
				if err2 == nil {
					log.Info("Failover to PostgreSQL successful. Switching DB_TYPE to postgres", nil)
					cfg.DBType = "postgres"
					return nil, p, nil
				}
				log.Warn("PostgreSQL failover also failed", map[string]interface{}{"error": err2.Error()})
			}
			// No fallback or fallback failed
			return nil, nil, fmt.Errorf("mongo init failed and no working fallback: %w", err)
		}
		// MongoDB worked
		return m, nil, nil

	case "postgres":
		// Try PostgreSQL
		p, err = postgres.NewPostgresStore(cfg, log)
		if err != nil {
			log.Warn("PostgreSQL connection failed, attempting failover to MongoDB",
				map[string]interface{}{"error": err.Error()})
			// Attempt to fallback to MongoDB if URI exists
			if cfg.MongoURI != "" {
				m, err2 := mongo.NewMongoStore(cfg, log)
				if err2 == nil {
					log.Info("Failover to MongoDB successful. Switching DB_TYPE to mongo", nil)
					cfg.DBType = "mongo"
					return m, nil, nil
				}
				log.Warn("MongoDB failover also failed", map[string]interface{}{"error": err2.Error()})
			}
			// No fallback or fallback failed
			return nil, nil, fmt.Errorf("postgres init failed and no working fallback: %w", err)
		}
		// PostgreSQL worked
		return nil, p, nil

	case "hybrid":
		// Try both, but continue even if one fails
		var mErr, pErr error

		// Try MongoDB
		if cfg.MongoURI != "" {
			m, mErr = mongo.NewMongoStore(cfg, log)
			if mErr != nil {
				log.Warn("MongoDB connection failed in hybrid mode",
					map[string]interface{}{"error": mErr.Error()})
			}
		} else {
			log.Warn("MONGO_URI not set, skipping MongoDB in hybrid mode", nil)
			mErr = errors.New("MONGO_URI not set")
		}

		// Try PostgreSQL
		if cfg.PostgresDSN != "" {
			p, pErr = postgres.NewPostgresStore(cfg, log)
			if pErr != nil {
				log.Warn("PostgreSQL connection failed in hybrid mode",
					map[string]interface{}{"error": pErr.Error()})
			}
		} else {
			log.Warn("POSTGRES_DSN not set, skipping PostgreSQL in hybrid mode", nil)
			pErr = errors.New("POSTGRES_DSN not set")
		}

		// Determine if we have at least one working DB
		if m == nil && p == nil {
			return nil, nil, fmt.Errorf("hybrid: both databases failed: mongo=%v, postgres=%v", mErr, pErr)
		}

		// Adjust DBType to reflect what's actually available
		if m != nil && p != nil {
			// Both work, keep hybrid
			cfg.DBType = "hybrid"
			log.Info("Hybrid mode: both MongoDB and PostgreSQL are active", nil)
		} else if m != nil {
			cfg.DBType = "mongo"
			log.Warn("Hybrid mode degraded: only MongoDB is active", map[string]interface{}{
				"postgres_error": pErr,
			})
		} else if p != nil {
			cfg.DBType = "postgres"
			log.Warn("Hybrid mode degraded: only PostgreSQL is active", map[string]interface{}{
				"mongo_error": mErr,
			})
		}

		return m, p, nil

	default:
		return nil, nil, fmt.Errorf("unknown DB_TYPE: %s", cfg.DBType)
	}
}

func initializeRepositories(cfg *config.Config, log *logger.Logger, m *mongo.MongoStore, p *postgres.PostgresStore) *RepoBundle {
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
		uRepo = postgres.NewUserRepository(p)
	case "hybrid":
		if p == nil {
			log.Warn("PostgreSQL store is nil in hybrid mode, using only MongoDB", nil)
			uRepo = mongoUserRepo
		} else if m == nil {
			log.Warn("MongoDB store is nil in hybrid mode, using only PostgreSQL", nil)
			uRepo = postgres.NewUserRepository(p)
		} else {
			pgRepo := postgres.NewUserRepository(p)
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
		mangaRepo = postgres.NewPostgresMangaRepository(p)
		chapterRepo = postgres.NewPostgresChapterRepository(p)
	case "hybrid":
		if p != nil && m != nil {
			pgMangaRepo := postgres.NewPostgresMangaRepository(p)
			mongoMangaRepo := mangarepo.NewMongoMangaRepository(m)
			mangaRepo = mangarepo.NewHybridMangaRepository(pgMangaRepo, mongoMangaRepo, log)
			chapterRepo = mangarepo.NewHybridChapterRepository(postgres.NewPostgresChapterRepository(p), mangarepo.NewMongoMangaChapterRepository(m), log)
		} else if p != nil {
			mangaRepo = postgres.NewPostgresMangaRepository(p)
			chapterRepo = postgres.NewPostgresChapterRepository(p)
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

func initializeServices(cfg *config.Config, log *logger.Logger, repos *RepoBundle, m *mongo.MongoStore, p *postgres.PostgresStore) *serviceBundle {
	var adminSvc coreadmin.Service

	// Choose admin repository based on available database
	if cfg.DBType == "postgres" && p != nil {
		adminRepo := admindata.NewPostgresAdminRepository(p)
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else if m != nil {
		adminRepo := admindata.NewMongoAdminRepository(repos.User)
		adminSvc = coreadmin.NewAdminService(repos.User, adminRepo, log)
	} else {
		log.Warn("No database available for admin repository, admin service will be nil", nil)
	}

	// Manga services
	var mangaSvc coremanga.MangaService
	var chapterSvc coremanga.MangaChapterService
	if repos.Manga != nil && repos.MangaChapter != nil {
		mangaSvc = coremanga.NewMangaService(repos.Manga, log)
		chapterSvc = coremanga.NewMangaChapterService(repos.MangaChapter, repos.Manga, log)
	} else {
		if cfg.DBType == "hybrid" {
			log.Warn("MongoDB is not available, manga services will be disabled", nil)
		}
	}

	return &serviceBundle{
		Auth:    coreauth.NewAuthService(repos.User, cfg, log.GetZerologLogger()),
		Admin:   adminSvc,
		Manga:   mangaSvc,
		Chapter: chapterSvc,
	}
}

func initializeHandlers(svcs *serviceBundle, m *mongo.MongoStore, p *postgres.PostgresStore) *handler.Container {
	return handler.NewContainer(svcs.Auth, svcs.Manga, svcs.Chapter, svcs.Admin, m, p)
}

func initializeInitialData(ctx context.Context, cfg *config.Config, repos *RepoBundle, log *logger.Logger) {
	if repos == nil || repos.User == nil {
		log.Warn("user repository is unavailable, skipping initial data setup", nil)
		return
	}

	// In production, skip default admin creation unless explicitly allowed
	environment := os.Getenv("ENVIRONMENT")
	if cfg != nil && cfg.Environment != "" {
		environment = cfg.Environment
	}
	if strings.ToLower(environment) == "production" && os.Getenv("DISABLE_DEFAULT_ADMIN") != "false" {
		log.Info("production mode: skipping default admin creation", nil)
		return
	}

	admin, err := repos.User.FindByUsername(ctx, "admin")
	if err != nil && !errors.Is(err, core.ErrUserNotFound) {
		log.Warn("failed to check for admin user existence", map[string]interface{}{"error": err.Error()})
		return
	}
	if admin == nil {
		hp, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		if err != nil {
			log.Warn("failed to hash default admin password", map[string]interface{}{"error": err.Error()})
			return
		}
		newUser := &coreuser.User{
			ID:       primitive.NewObjectID(),
			UUID:     utils.GenerateUUID(),
			UserID:   fmt.Sprintf("User-%d", 1),
			Username: "admin",
			Password: string(hp),
			UserBase: coreuser.UserBase{
				Roles:     coreuser.FromStrings([]string{"admin"}),
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		if err := repos.User.Create(ctx, newUser, &coreuser.UserDetails{
			UserBase: coreuser.UserBase{
				Roles:     coreuser.FromStrings([]string{"admin"}),
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Status: "active",
		}); err != nil {
			log.Warn("failed to create admin user", map[string]interface{}{"error": err.Error()})
		} else {
			log.Info("admin user created successfully", nil)
		}
	} else {
		log.Info("admin user already exists, skipping creation", nil)
	}
}

// ----- END OF FILE: backend/MS-AI/internal/container/initializers.go -----
