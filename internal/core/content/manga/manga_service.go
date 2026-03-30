package manga

import (
	"context"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"github.com/835-droid/ms-ai-backend/pkg/validator"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaService defines manga business logic operations.
type MangaService interface {
	CreateManga(ctx context.Context, manga *Manga) (*Manga, error)
	GetManga(ctx context.Context, id primitive.ObjectID) (*Manga, error)
	ListMangas(ctx context.Context, skip, limit int64) ([]*Manga, int64, error)
	UpdateManga(ctx context.Context, manga *Manga, callerID primitive.ObjectID, roles []string) error
	DeleteManga(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error
}

// DefaultMangaService implements MangaService.
type DefaultMangaService struct {
	repo MangaRepository
	log  *logger.Logger
}

// NewMangaService creates a new DefaultMangaService.
func NewMangaService(repo MangaRepository, log *logger.Logger) *DefaultMangaService {
	return &DefaultMangaService{repo: repo, log: log}
}

// CreateManga creates a new manga.
func (s *DefaultMangaService) CreateManga(ctx context.Context, manga *Manga) (*Manga, error) {
	s.log.Debug("creating manga", map[string]interface{}{
		"title":     manga.Title,
		"author_id": manga.AuthorID.Hex(),
	})

	if err := validator.Validate(manga); err != nil {
		s.log.Error("manga validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	manga.Slug = utils.Slugify(manga.Title)
	now := time.Now()
	manga.CreatedAt = now
	manga.UpdatedAt = now

	if err := s.repo.CreateManga(ctx, manga); err != nil {
		s.log.Error("manga creation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	s.log.Info("manga created", map[string]interface{}{
		"id":    manga.ID.Hex(),
		"title": manga.Title,
	})
	return manga, nil
}

// GetManga retrieves a manga by ID.
func (s *DefaultMangaService) GetManga(ctx context.Context, id primitive.ObjectID) (*Manga, error) {
	s.log.Debug("getting manga", map[string]interface{}{
		"id": id.Hex(),
	})

	manga, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		s.log.Error("get manga failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return nil, err
	}

	s.log.Debug("manga retrieved", map[string]interface{}{
		"id":    manga.ID.Hex(),
		"title": manga.Title,
	})
	return manga, nil
}

// ListMangas retrieves a paginated list of manga.
func (s *DefaultMangaService) ListMangas(ctx context.Context, skip, limit int64) ([]*Manga, int64, error) {
	s.log.Debug("listing mangas", map[string]interface{}{
		"skip":  skip,
		"limit": limit,
	})

	mangas, total, err := s.repo.ListMangas(ctx, skip, limit)
	if err != nil {
		s.log.Error("list mangas failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, err
	}

	s.log.Debug("mangas listed", map[string]interface{}{
		"count": len(mangas),
		"total": total,
	})
	return mangas, total, nil
}

// UpdateManga updates a manga.
func (s *DefaultMangaService) UpdateManga(ctx context.Context, manga *Manga, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("updating manga", map[string]interface{}{
		"id":        manga.ID.Hex(),
		"title":     manga.Title,
		"caller_id": callerID.Hex(),
	})

	// Verify ownership or admin role
	if manga.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized manga update attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"manga_id":  manga.ID.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := validator.Validate(manga); err != nil {
		s.log.Error("manga validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	manga.Slug = utils.Slugify(manga.Title)
	manga.UpdatedAt = time.Now()

	if err := s.repo.UpdateManga(ctx, manga); err != nil {
		s.log.Error("manga update failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.log.Info("manga updated", map[string]interface{}{
		"id":    manga.ID.Hex(),
		"title": manga.Title,
	})
	return nil
}

// DeleteManga deletes a manga.
func (s *DefaultMangaService) DeleteManga(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("deleting manga", map[string]interface{}{
		"id":        id.Hex(),
		"caller_id": callerID.Hex(),
	})

	manga, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		if err == core.ErrNotFound {
			return err
		}
		s.log.Error("get manga for deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	if manga.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized manga deletion attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"manga_id":  id.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := s.repo.DeleteManga(ctx, id); err != nil {
		s.log.Error("manga deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	s.log.Info("manga deleted", map[string]interface{}{
		"id": id.Hex(),
	})
	return nil
}

// Helper function to check if a role exists in a slice of roles.
func hasRole(roles []string, role string) bool {
	role = strings.ToLower(role)
	for _, r := range roles {
		if strings.ToLower(r) == role {
			return true
		}
	}
	return false
}
