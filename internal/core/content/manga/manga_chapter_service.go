// ----- START OF FILE: backend/MS-AI/internal/core/content/manga/manga_chapter_service.go -----
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
	// Engagement methods
	IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error
	AddChapterRating(ctx context.Context, rating *ChapterRating) (float64, error)
	HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	AddChapterComment(ctx context.Context, comment *ChapterComment) error
	ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64) ([]*ChapterComment, int64, error)
	DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error
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

// ========== ENGAGEMENT METHODS ==========

// IncrementChapterViews increments view count for a chapter.
func (s *DefaultMangaChapterService) IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error {
	s.log.Debug("incrementing chapter views", map[string]interface{}{
		"chapter_id": chapterID.Hex(),
		"manga_id":   mangaID.Hex(),
	})

	if err := s.chapterRepo.IncrementChapterViews(ctx, chapterID, mangaID); err != nil {
		s.log.Error("increment chapter views failed", map[string]interface{}{
			"error":      err.Error(),
			"chapter_id": chapterID.Hex(),
			"manga_id":   mangaID.Hex(),
		})
		return err
	}

	s.log.Debug("chapter views incremented", map[string]interface{}{
		"chapter_id": chapterID.Hex(),
		"manga_id":   mangaID.Hex(),
	})
	return nil
}

// AddChapterRating adds a rating to a chapter.
func (s *DefaultMangaChapterService) AddChapterRating(ctx context.Context, rating *ChapterRating) (float64, error) {
	s.log.Debug("adding chapter rating", map[string]interface{}{
		"chapter_id": rating.ChapterID.Hex(),
		"user_id":    rating.UserID.Hex(),
		"score":      rating.Score,
	})

	if err := validator.Validate(rating); err != nil {
		s.log.Error("chapter rating validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return 0, err
	}

	avgRating, err := s.chapterRepo.AddChapterRating(ctx, rating)
	if err != nil {
		s.log.Error("add chapter rating failed", map[string]interface{}{
			"error":      err.Error(),
			"chapter_id": rating.ChapterID.Hex(),
		})
		return 0, err
	}

	s.log.Info("chapter rating added", map[string]interface{}{
		"chapter_id": rating.ChapterID.Hex(),
		"user_id":    rating.UserID.Hex(),
		"avg_rating": avgRating,
	})
	return avgRating, nil
}

// HasUserRatedChapter checks if user has rated a chapter.
func (s *DefaultMangaChapterService) HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	return s.chapterRepo.HasUserRatedChapter(ctx, chapterID, userID)
}

// AddChapterComment adds a comment to a chapter.
func (s *DefaultMangaChapterService) AddChapterComment(ctx context.Context, comment *ChapterComment) error {
	s.log.Debug("adding chapter comment", map[string]interface{}{
		"chapter_id": comment.ChapterID.Hex(),
		"user_id":    comment.UserID.Hex(),
	})

	if err := validator.Validate(comment); err != nil {
		s.log.Error("chapter comment validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if err := s.chapterRepo.AddChapterComment(ctx, comment); err != nil {
		s.log.Error("add chapter comment failed", map[string]interface{}{
			"error":      err.Error(),
			"chapter_id": comment.ChapterID.Hex(),
		})
		return err
	}

	s.log.Info("chapter comment added", map[string]interface{}{
		"id":         comment.ID.Hex(),
		"chapter_id": comment.ChapterID.Hex(),
		"user_id":    comment.UserID.Hex(),
	})
	return nil
}

// ListChapterComments retrieves comments for a chapter.
func (s *DefaultMangaChapterService) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64) ([]*ChapterComment, int64, error) {
	s.log.Debug("listing chapter comments", map[string]interface{}{
		"chapter_id": chapterID.Hex(),
		"skip":       skip,
		"limit":      limit,
	})

	comments, total, err := s.chapterRepo.ListChapterComments(ctx, chapterID, skip, limit)
	if err != nil {
		s.log.Error("list chapter comments failed", map[string]interface{}{
			"error":      err.Error(),
			"chapter_id": chapterID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("chapter comments listed", map[string]interface{}{
		"chapter_id": chapterID.Hex(),
		"count":      len(comments),
		"total":      total,
	})
	return comments, total, nil
}

// DeleteChapterComment deletes a chapter comment.
func (s *DefaultMangaChapterService) DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	s.log.Debug("deleting chapter comment", map[string]interface{}{
		"comment_id": commentID.Hex(),
		"user_id":    userID.Hex(),
	})

	if err := s.chapterRepo.DeleteChapterComment(ctx, commentID, userID); err != nil {
		s.log.Error("delete chapter comment failed", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID.Hex(),
			"user_id":    userID.Hex(),
		})
		return err
	}

	s.log.Info("chapter comment deleted", map[string]interface{}{
		"comment_id": commentID.Hex(),
		"user_id":    userID.Hex(),
	})
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/core/content/manga/manga_chapter_service.go -----
