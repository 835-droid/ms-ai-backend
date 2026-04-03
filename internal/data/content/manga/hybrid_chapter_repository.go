// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_chapter_repository.go -----
package manga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

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
	var primaryChapter, secondaryChapter *coremanga.MangaChapter
	var primaryErr, secondaryErr error

	// Query primary repository
	if r.primary != nil {
		primaryChapter, primaryErr = r.primary.GetMangaChapterByID(ctx, id)
	}

	// Query secondary repository
	if r.secondary != nil {
		secondaryChapter, secondaryErr = r.secondary.GetMangaChapterByID(ctx, id)
	}

	// Return primary result if successful
	if primaryErr == nil && primaryChapter != nil {
		if secondaryErr != nil && r.log != nil {
			r.log.Error("hybrid secondary get chapter by id failed", map[string]interface{}{"error": secondaryErr.Error()})
		}
		return primaryChapter, nil
	}

	// Return secondary result if primary failed but secondary succeeded
	if secondaryErr == nil && secondaryChapter != nil {
		if primaryErr != nil && r.log != nil {
			r.log.Error("hybrid primary get chapter by id failed", map[string]interface{}{"error": primaryErr.Error()})
		}
		return secondaryChapter, nil
	}

	// Both failed - return primary error if available, otherwise secondary error
	if primaryErr != nil {
		return nil, primaryErr
	}
	if secondaryErr != nil {
		return nil, secondaryErr
	}

	return nil, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaChapter, int64, error) {
	chapterMap := make(map[string]*coremanga.MangaChapter)
	var primarySuccess, secondarySuccess bool

	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication and ensure minimum coverage
	fetchLimit := skip + limit + 500
	if fetchLimit < 2000 {
		fetchLimit = 2000
	}

	// Query primary repository
	if r.primary != nil {
		chapters, _, err := r.primary.ListMangaChaptersByManga(ctx, mangaID, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, chapter := range chapters {
				chapterMap[chapter.ID.Hex()] = chapter
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list chapters failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge unique items
	if r.secondary != nil {
		chapters, _, err := r.secondary.ListMangaChaptersByManga(ctx, mangaID, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, chapter := range chapters {
				if _, exists := chapterMap[chapter.ID.Hex()]; !exists {
					chapterMap[chapter.ID.Hex()] = chapter
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list chapters failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, 0, errors.New("no repositories available or all failed")
	}

	// Convert map to slice and sort by number
	result := make([]*coremanga.MangaChapter, 0, len(chapterMap))
	for _, chapter := range chapterMap {
		result = append(result, chapter)
	}

	// Sort by chapter number ascending using sort.Slice for efficiency
	sort.Slice(result, func(i, j int) bool {
		return result[i].Number < result[j].Number
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.MangaChapter{}, total, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], total, nil
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

func (r *HybridMangaChapterRepository) AddChapterRating(ctx context.Context, rating *coremanga.ChapterRating) (float64, int64, float64, error) {
	var avg float64
	var count int64
	var userScore float64
	var errPrimary, errSecondary error

	// Try primary repository
	if r.primary != nil {
		avg, count, userScore, errPrimary = r.primary.AddChapterRating(ctx, rating)
	}

	// Try secondary repository (tolerant - only log failures)
	if r.secondary != nil {
		_, _, _, errSecondary = r.secondary.AddChapterRating(ctx, rating)
	}

	// Return success if primary succeeded, log secondary failures
	if errPrimary != nil {
		if r.log != nil {
			r.log.Error("hybrid primary add chapter rating failed", map[string]interface{}{"error": errPrimary.Error()})
		}
		return 0, 0, 0, errPrimary
	}

	if errSecondary != nil && r.log != nil {
		r.log.Error("hybrid secondary add chapter rating failed", map[string]interface{}{"error": errSecondary.Error()})
	}

	// Return the successful primary result
	return avg, count, userScore, nil
}

func (r *HybridMangaChapterRepository) TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.TrackChapterView(ctx, chapterID, mangaID, userID)
	}
	if r.secondary != nil {
		errSecondary = r.secondary.TrackChapterView(ctx, chapterID, mangaID, userID)
	}
	if errPrimary != nil || errSecondary != nil {
		if r.log != nil {
			if errPrimary != nil {
				r.log.Error("hybrid primary track chapter view failed", map[string]interface{}{"error": errPrimary.Error()})
			}
			if errSecondary != nil {
				r.log.Error("hybrid secondary track chapter view failed", map[string]interface{}{"error": errSecondary.Error()})
			}
		}
		return fmt.Errorf("dual-write requirement failed: primary=%v, secondary=%v", errPrimary, errSecondary)
	}
	return nil
}

func (r *HybridMangaChapterRepository) HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error) {
	if r.primary != nil {
		v, err := r.primary.HasUserViewedChapter(ctx, chapterID, userID)
		if err == nil {
			return v, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.HasUserViewedChapter(ctx, chapterID, userID)
	}
	return false, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error) {
	var isNewView bool
	var errPrimary, errSecondary error

	// Try primary repository
	if r.primary != nil {
		isNewView, errPrimary = r.primary.TrackAndIncrementChapterView(ctx, chapterID, mangaID, userID)
	}

	// Try secondary repository (tolerant - only log failures)
	if r.secondary != nil {
		_, errSecondary = r.secondary.TrackAndIncrementChapterView(ctx, chapterID, mangaID, userID)
	}

	// Return success if primary succeeded, log secondary failures
	if errPrimary != nil {
		if r.log != nil {
			r.log.Error("hybrid primary track and increment chapter view failed", map[string]interface{}{"error": errPrimary.Error()})
		}
		return false, errPrimary
	}

	if errSecondary != nil && r.log != nil {
		r.log.Error("hybrid secondary track and increment chapter view failed", map[string]interface{}{"error": errSecondary.Error()})
	}

	return isNewView, nil
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

func (r *HybridMangaChapterRepository) GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error) {
	if r.primary != nil {
		v, has, err := r.primary.GetUserChapterRating(ctx, chapterID, userID)
		if err == nil {
			return v, has, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.GetUserChapterRating(ctx, chapterID, userID)
	}
	return 0, false, errors.New("no repositories available")
}

func (r *HybridMangaChapterRepository) AddChapterComment(ctx context.Context, comment *coremanga.ChapterComment) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.AddChapterComment(ctx, comment)
	}
	if r.secondary != nil {
		// Create a deep copy for secondary to prevent ID/timestamp conflicts
		commentJSON, _ := json.Marshal(comment)
		var commentCopy coremanga.ChapterComment
		json.Unmarshal(commentJSON, &commentCopy)
		errSecondary = r.secondary.AddChapterComment(ctx, &commentCopy)
	}
	if errPrimary != nil || errSecondary != nil {
		if r.log != nil {
			if errPrimary != nil {
				r.log.Error("hybrid primary add chapter comment failed", map[string]interface{}{"error": errPrimary.Error()})
			}
			if errSecondary != nil {
				r.log.Error("hybrid secondary add chapter comment failed", map[string]interface{}{"error": errSecondary.Error()})
			}
		}
		return fmt.Errorf("dual-write requirement failed: primary=%v, secondary=%v", errPrimary, errSecondary)
	}
	return nil
}

func (r *HybridMangaChapterRepository) ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.ChapterComment, int64, error) {
	if r.primary != nil {
		comments, total, err := r.primary.ListChapterComments(ctx, chapterID, skip, limit, sortOrder)
		if err == nil {
			return comments, total, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, 0, err
		}
	}
	if r.secondary != nil {
		return r.secondary.ListChapterComments(ctx, chapterID, skip, limit, sortOrder)
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

// ========== CHAPTER COMMENT REACTIONS ==========

func (r *HybridMangaChapterRepository) AddChapterCommentReaction(ctx context.Context, reaction *coremanga.ChapterCommentReaction) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.AddChapterCommentReaction(ctx, reaction)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary add chapter comment reaction failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.AddChapterCommentReaction(ctx, reaction)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary add chapter comment reaction failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid add chapter comment reaction failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) RemoveChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.RemoveChapterCommentReaction(ctx, commentID, userID)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary remove chapter comment reaction failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.RemoveChapterCommentReaction(ctx, commentID, userID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary remove chapter comment reaction failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid remove chapter comment reaction failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaChapterRepository) GetUserChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error) {
	if r.primary != nil {
		reaction, err := r.primary.GetUserChapterCommentReaction(ctx, commentID, userID)
		if err == nil {
			return reaction, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.GetUserChapterCommentReaction(ctx, commentID, userID)
	}
	return "", errors.New("no repositories available")
}

// ----- END OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_chapter_repo.go -----
