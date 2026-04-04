// ----- START OF FILE: backend/MS-AI/internal/container/types.go -----

package container

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler"
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	mongoinfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
)

type Container struct {
	Config     *config.Config
	Logger     *logger.Logger
	MongoDB    *mongoinfra.MongoStore
	PostgresDB *postgres.PostgresStore
	UserRepo   coreuser.Repository
	Handlers   *handler.Container
}

type RepoBundle struct {
	User           coreuser.Repository
	Manga          coremanga.MangaRepository
	FavList        coremanga.FavoriteListRepository
	MangaChapter   coremanga.MangaChapterRepository
	ViewingHistory coremanga.ViewingHistoryRepository
}

type serviceBundle struct {
	Auth           coreauth.AuthService
	Admin          coreadmin.Service
	Manga          coremanga.MangaService
	FavList        coremanga.FavoriteListService
	Chapter        coremanga.MangaChapterService
	ViewingHistory coremanga.ViewingHistoryService
}

// ----- END OF FILE: backend/MS-AI/internal/container/types.go -----
