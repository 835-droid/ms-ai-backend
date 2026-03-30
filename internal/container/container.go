// internal/container/container.go
package container

import (
	"context"
	"fmt"

	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	dmanga "github.com/835-droid/ms-ai-backend/internal/data/content/manga"
	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	datauser "github.com/835-droid/ms-ai-backend/internal/data/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/i18n"
	"github.com/835-droid/ms-ai-backend/pkg/jwt"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

// Container holds all application dependencies
type Container struct {
	// Configuration
	Config *config.Config

	// Infrastructure
	Logger     *logger.Logger
	MongoDB    *mongo.MongoStore
	JWT        jwt.Service
	Translator *i18n.Translator

	// Repositories
	UserRepo         coreuser.Repository
	MangaRepo        coremanga.MangaRepository
	MangaChapterRepo coremanga.MangaChapterRepository

	// Services
	AuthService         coreauth.AuthService
	MangaService        coremanga.MangaService
	MangaChapterService coremanga.MangaChapterService
	AdminService        coreadmin.Service

	// Handlers
	Handlers *handler.Container
}

// NewContainer creates and initializes all application dependencies
func NewContainer(cfg *config.Config) (*Container, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel, cfg.Environment, nil)
	log.Info("initializing application container", map[string]interface{}{"environment": cfg.Environment})

	// Initialize MongoDB
	mongoStore, err := mongo.NewMongoStore(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	log.Info("MongoDB connection established", nil)

	// Initialize JWT service
	jwtService := jwt.NewService(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// Initialize translator
	translator := i18n.NewTranslator(i18n.LanguageEnglish)

	// Initialize repositories
	userRepo := datauser.NewMongoUserRepository(mongoStore)
	mangaRepo := dmanga.NewMongoMangaRepository(mongoStore)
	mangaChapterRepo := dmanga.NewMongoMangaChapterRepository(mongoStore)

	// Initialize services
	authService := coreauth.NewAuthService(userRepo, cfg, log.GetZerologLogger())
	mangaService := coremanga.NewMangaService(mangaRepo, log)
	mangaChapterService := coremanga.NewMangaChapterService(mangaChapterRepo, mangaRepo, log)
	adminService := coreadmin.NewAdminService(userRepo.AsAdminRepository(), log)

	// Initialize handlers
	handlers := handler.NewContainer(
		authService,
		mangaService,
		mangaChapterService,
		adminService,
	)

	container := &Container{
		Config:              cfg,
		Logger:              log,
		MongoDB:             mongoStore,
		JWT:                 jwtService,
		Translator:          translator,
		UserRepo:            userRepo,
		MangaRepo:           mangaRepo,
		MangaChapterRepo:    mangaChapterRepo,
		AuthService:         authService,
		MangaService:        mangaService,
		MangaChapterService: mangaChapterService,
		AdminService:        adminService,
		Handlers:            handlers,
	}

	log.Info("application container initialized successfully", nil)
	return container, nil
}

// Close gracefully shuts down all resources
func (c *Container) Close(ctx context.Context) error {
	c.Logger.Info("shutting down application container", nil)

	if err := c.MongoDB.Close(ctx); err != nil {
		c.Logger.Error("error closing MongoDB connection", map[string]interface{}{"error": err.Error()})
		return err
	}

	c.Logger.Info("application container shut down successfully", nil)
	return nil
}
