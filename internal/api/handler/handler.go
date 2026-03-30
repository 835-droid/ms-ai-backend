package handler

import (
	"github.com/835-droid/ms-ai-backend/internal/api/handler/admin"
	auth "github.com/835-droid/ms-ai-backend/internal/api/handler/auth"
	manga "github.com/835-droid/ms-ai-backend/internal/api/handler/content/manga"
	health "github.com/835-droid/ms-ai-backend/internal/api/handler/health"

	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"

	datamongo "github.com/835-droid/ms-ai-backend/internal/data/mongo"
)

// Re-export handler types as aliases so higher-level code can import a single package.
type AuthHandler = auth.Handler
type AdminHandler = admin.Handler
type HealthHandler = health.Handler
type MangaHandler = manga.MangaHandler
type MangaChapterHandler = manga.MangaChapterHandler

// Constructors
func NewAuthHandler(s coreauth.AuthService) *auth.Handler          { return auth.NewHandler(s) }
func NewAdminHandler(s coreadmin.Service) *admin.Handler           { return admin.NewHandler(s) }
func NewHealthHandler(s *datamongo.MongoStore) *health.Handler     { return health.NewHandler(s) }
func NewMangaHandler(s coremanga.MangaService) *manga.MangaHandler { return manga.NewMangaHandler(s) }
func NewMangaChapterHandler(s coremanga.MangaChapterService) *manga.MangaChapterHandler {
	return manga.NewMangaChapterHandler(s)
}
