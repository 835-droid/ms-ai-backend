// internal/api/handler/container.go
package handler

import (
	coreadmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreauth "github.com/835-droid/ms-ai-backend/internal/core/auth"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
)

// Container holds all HTTP handlers
type Container struct {
	AuthHandler         *AuthHandler
	AdminHandler        *AdminHandler
	HealthHandler       *HealthHandler
	MangaHandler        *MangaHandler
	MangaChapterHandler *MangaChapterHandler
}

// NewContainer creates and initializes all HTTP handlers
func NewContainer(
	authService coreauth.AuthService,
	mangaService coremanga.MangaService,
	mangaChapterService coremanga.MangaChapterService,
	adminService coreadmin.Service,
) *Container {
	return &Container{
		AuthHandler:         NewAuthHandler(authService),
		AdminHandler:        NewAdminHandler(adminService),
		HealthHandler:       NewHealthHandler(nil), // TODO: pass mongo store
		MangaHandler:        NewMangaHandler(mangaService),
		MangaChapterHandler: NewMangaChapterHandler(mangaChapterService),
	}
}
