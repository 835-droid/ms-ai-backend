// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_chapter_repo.go -----
package manga

import (
	"context"
	"errors"
	"fmt"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HybridMangaChapterRepository struct {
	primary   coremanga.MangaChapterRepository
	secondary coremanga.MangaChapterRepository
	log       *logger.Logger
}

func NewHybridChapterRepository(primary, secondary coremanga.MangaChapterRepository, log *logger.Logger) coremanga.MangaChapterRepository {
	return &HybridMangaChapterRepository{primary: primary, secondary: secondary, log: log}
}

func (r *HybridMangaChapterRepository) CreateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.CreateMangaChapter(ctx, chapter)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary create chapter failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.CreateMangaChapter(ctx, chapter)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary create chapter failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid create chapter failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*coremanga.MangaChapter, error) {
	if r.primary != nil {
		chapter, err := r.primary.GetMangaChapterByID(ctx, id)
		if err == nil && chapter != nil {
			return chapter, nil
		}
		if err != nil {
			return nil, err
		}
	}
	if r.secondary != nil {
		return r.secondary.GetMangaChapterByID(ctx, id)
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaChapter, int64, error) {
	if r.primary != nil {
		chapters, total, err := r.primary.ListMangaChaptersByManga(ctx, mangaID, skip, limit)
		if err == nil {
			return chapters, total, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.ListMangaChaptersByManga(ctx, mangaID, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) UpdateMangaChapter(ctx context.Context, chapter *coremanga.MangaChapter) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.UpdateMangaChapter(ctx, chapter)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary update chapter failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.UpdateMangaChapter(ctx, chapter)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary update chapter failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid update chapter failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.DeleteMangaChapter(ctx, id)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary delete chapter failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.DeleteMangaChapter(ctx, id)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary delete chapter failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid delete chapter failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

// ========== ENGAGEMENT METHODS ==========

func (r *HybridMangaChapterRepository) IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.IncrementChapterViews(ctx, chapterID, mangaID)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary increment chapter views failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.IncrementChapterViews(ctx, chapterID, mangaID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary increment chapter views failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid increment chapter views failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (float64, error) {
	var primaryErr, secondaryErr error
	var primaryAvg, secondaryAvg float64

	// Try primary repository
	if r.primary != nil {
		avg, err := r.primary.AddChapterRating(ctx, rating)
		if err == nil {
			primaryAvg = avg
		} else {
			primaryErr = err
			if r.log != nil {
				r.log.Error("hybrid primary add chapter rating failed", map[string]interface{}{"error": err.Error()})
			}
		}
	}

	// Try secondary repository
	if r.secondary != nil {
		avg, err := r.secondary.AddChapterRating(ctx, rating)
		if err == nil {
			secondaryAvg = avg
		} else {
			secondaryErr = err
			if r.log != nil {
				r.log.Error("hybrid secondary add chapter rating failed", map[string]interface{}{"error": err.Error()})
			}
		}
	}

	// Enforce explicit success criteria
	if primaryErr != nil && secondaryErr != nil {
		// Both repositories failed
		return 0, fmt.Errorf("both repositories failed: primary=%v, secondary=%v", primaryErr, secondaryErr)
	}

	// Return the successful result (prefer primary if both succeeded)
	if primaryErr == nil {
		return primaryAvg, nil
	}
	return secondaryAvg, nil
}

func (r *HybridMangaChapterRepository) HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if r.primary != nil {
		v, err := r.primary.HasUserRatedChapter(ctx, chapterID, userID)
		if err == nil {
			return v, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.HasUserRatedChapter(ctx, chapterID, userID)
	}
	return false, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) AddChapterComment(ctx context.Context, comment *coremanga.ChapterComment) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.AddChapterComment(ctx, comment)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary add chapter comment failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.AddChapterComment(ctx, comment)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary add chapter comment failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid add chapter comment failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64) ([]*coremanga.ChapterComment, int64, error) {
	if r.primary != nil {
		comments, total, err := r.primary.ListChapterComments(ctx, chapterID, skip, limit)
		if err == nil && len(comments) > 0 {
			return comments, total, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, 0, err
		}
	}
	if r.secondary != nil {
		return r.secondary.ListChapterComments(ctx, chapterID, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.DeleteChapterComment(ctx, commentID, userID)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary delete chapter comment failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.DeleteChapterComment(ctx, commentID, userID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary delete chapter comment failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid delete chapter comment failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_chapter_repo.go -----
