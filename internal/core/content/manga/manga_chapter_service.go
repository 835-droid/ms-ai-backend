package manga

import (
	"context"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/validator"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaChapterService defines manga chapter business logic operations.
type MangaChapterService interface {
	CreateMangaChapter(ctx context.Context, chapter *MangaChapter, callerID primitive.ObjectID, roles []string) (*MangaChapter, error)
	GetMangaChapter(ctx context.Context, id primitive.ObjectID) (*MangaChapter, error)
	ListMangaChapters(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*MangaChapter, int64, error)
	UpdateMangaChapter(ctx context.Context, chapter *MangaChapter, callerID primitive.ObjectID, roles []string) error
	DeleteMangaChapter(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error
}

// DefaultMangaChapterService implements MangaChapterService.
type DefaultMangaChapterService struct {
	chapterRepo MangaChapterRepository
	mangaRepo   MangaRepository
	log         *logger.Logger
}

// NewMangaChapterService creates a new DefaultMangaChapterService.
func NewMangaChapterService(chapterRepo MangaChapterRepository, mangaRepo MangaRepository, log *logger.Logger) *DefaultMangaChapterService {
	return &DefaultMangaChapterService{
		chapterRepo: chapterRepo,
		mangaRepo:   mangaRepo,
		log:         log,
	}
}

// CreateMangaChapter creates a new manga chapter.
func (s *DefaultMangaChapterService) CreateMangaChapter(ctx context.Context, chapter *MangaChapter, callerID primitive.ObjectID, roles []string) (*MangaChapter, error) {
	s.log.Debug("creating manga chapter", map[string]interface{}{
		"manga_id":  chapter.MangaID.Hex(),
		"title":     chapter.Title,
		"caller_id": callerID.Hex(),
	})

	// Verify manga exists and caller owns it
	manga, err := s.mangaRepo.GetMangaByID(ctx, chapter.MangaID)
	if err != nil {
		s.log.Error("get manga failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": chapter.MangaID.Hex(),
		})
		return nil, err
	}

	if manga.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized chapter creation attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"manga_id":  manga.ID.Hex(),
		})
		return nil, core.ErrUnauthorized
	}

	if err := validator.Validate(chapter); err != nil {
		s.log.Error("chapter validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	now := time.Now()
	chapter.CreatedAt = now
	chapter.UpdatedAt = now

	if err := s.chapterRepo.CreateMangaChapter(ctx, chapter); err != nil {
		s.log.Error("chapter creation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	s.log.Info("manga chapter created", map[string]interface{}{
		"id":        chapter.ID.Hex(),
		"manga_id":  chapter.MangaID.Hex(),
		"title":     chapter.Title,
		"number":    chapter.Number,
		"caller_id": callerID.Hex(),
	})
	return chapter, nil
}

// GetMangaChapter retrieves a manga chapter by ID.
func (s *DefaultMangaChapterService) GetMangaChapter(ctx context.Context, id primitive.ObjectID) (*MangaChapter, error) {
	s.log.Debug("getting manga chapter", map[string]interface{}{
		"id": id.Hex(),
	})

	chapter, err := s.chapterRepo.GetMangaChapterByID(ctx, id)
	if err != nil {
		s.log.Error("get chapter failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return nil, err
	}

	s.log.Debug("manga chapter retrieved", map[string]interface{}{
		"id":       chapter.ID.Hex(),
		"title":    chapter.Title,
		"manga_id": chapter.MangaID.Hex(),
	})
	return chapter, nil
}

// ListMangaChapters retrieves a paginated list of manga chapters.
func (s *DefaultMangaChapterService) ListMangaChapters(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*MangaChapter, int64, error) {
	s.log.Debug("listing manga chapters", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"skip":     skip,
		"limit":    limit,
	})

	chapters, total, err := s.chapterRepo.ListMangaChaptersByManga(ctx, mangaID, skip, limit)
	if err != nil {
		s.log.Error("list chapters failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": mangaID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("manga chapters listed", map[string]interface{}{
		"count":    len(chapters),
		"total":    total,
		"manga_id": mangaID.Hex(),
	})
	return chapters, total, nil
}

// UpdateMangaChapter updates a manga chapter.
func (s *DefaultMangaChapterService) UpdateMangaChapter(ctx context.Context, chapter *MangaChapter, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("updating manga chapter", map[string]interface{}{
		"id":        chapter.ID.Hex(),
		"manga_id":  chapter.MangaID.Hex(),
		"caller_id": callerID.Hex(),
	})

	// Verify manga exists and caller owns it
	manga, err := s.mangaRepo.GetMangaByID(ctx, chapter.MangaID)
	if err != nil {
		s.log.Error("get manga failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": chapter.MangaID.Hex(),
		})
		return err
	}

	if manga.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized chapter update attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"manga_id":  manga.ID.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := validator.Validate(chapter); err != nil {
		s.log.Error("chapter validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	chapter.UpdatedAt = time.Now()

	if err := s.chapterRepo.UpdateMangaChapter(ctx, chapter); err != nil {
		s.log.Error("chapter update failed", map[string]interface{}{
			"error": err.Error(),
			"id":    chapter.ID.Hex(),
		})
		return err
	}

	s.log.Info("manga chapter updated", map[string]interface{}{
		"id":        chapter.ID.Hex(),
		"manga_id":  chapter.MangaID.Hex(),
		"title":     chapter.Title,
		"number":    chapter.Number,
		"caller_id": callerID.Hex(),
	})
	return nil
}

// DeleteMangaChapter deletes a manga chapter.
func (s *DefaultMangaChapterService) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("deleting manga chapter", map[string]interface{}{
		"id":        id.Hex(),
		"caller_id": callerID.Hex(),
	})

	// Get chapter first to verify ownership
	chapter, err := s.chapterRepo.GetMangaChapterByID(ctx, id)
	if err != nil {
		if err == core.ErrNotFound {
			return err
		}
		s.log.Error("get chapter for deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	// Verify manga exists and caller owns it
	manga, err := s.mangaRepo.GetMangaByID(ctx, chapter.MangaID)
	if err != nil {
		s.log.Error("get manga failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": chapter.MangaID.Hex(),
		})
		return err
	}

	if manga.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized chapter deletion attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"manga_id":  manga.ID.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := s.chapterRepo.DeleteMangaChapter(ctx, id); err != nil {
		s.log.Error("chapter deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	s.log.Info("manga chapter deleted", map[string]interface{}{
		"id":        id.Hex(),
		"manga_id":  chapter.MangaID.Hex(),
		"caller_id": callerID.Hex(),
	})
	return nil
}
