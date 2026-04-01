package container

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

type Container struct {
	Config   *config.Config
	Logger   *logger.Logger
	MongoDB  *mongo.MongoStore
	UserRepo coreuser.Repository
	Handlers *handler.Container
}

type RepoBundle struct {
	User         coreuser.Repository
	Manga        coremanga.MangaRepository
	MangaChapter coremanga.MangaChapterRepository
}

type serviceBundle struct {
	Auth    coreauth.AuthService
	Admin   coreadmin.Service
	Manga   coremanga.MangaService
	Chapter coremanga.MangaChapterService
}
