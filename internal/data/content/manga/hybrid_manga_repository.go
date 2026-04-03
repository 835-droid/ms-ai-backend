package manga

// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_manga_repo.go -----

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coremanga "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HybridMangaRepository struct {
	primary   coremanga.MangaRepository
	secondary coremanga.MangaRepository
	log       *logger.Logger
}

func NewHybridMangaRepository(primary, secondary coremanga.MangaRepository, log *logger.Logger) coremanga.MangaRepository {
	return &HybridMangaRepository{primary: primary, secondary: secondary, log: log}
}

func (r *HybridMangaRepository) CreateManga(ctx context.Context, manga *coremanga.Manga) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.CreateManga(ctx, manga)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary create manga failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.CreateManga(ctx, manga)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary create manga failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid create manga failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) GetMangaByID(ctx context.Context, id primitive.ObjectID) (*coremanga.Manga, error) {
	var primaryManga, secondaryManga *coremanga.Manga
	var primaryErr, secondaryErr error

	// Query primary repository
	if r.primary != nil {
		primaryManga, primaryErr = r.primary.GetMangaByID(ctx, id)
	}

	// Query secondary repository
	if r.secondary != nil {
		secondaryManga, secondaryErr = r.secondary.GetMangaByID(ctx, id)
	}

	// Return primary result if successful
	if primaryErr == nil && primaryManga != nil {
		if secondaryErr != nil && r.log != nil {
			r.log.Error("hybrid secondary get manga by id failed", map[string]interface{}{"error": secondaryErr.Error()})
		}
		return primaryManga, nil
	}

	// Return secondary result if primary failed but secondary succeeded
	if secondaryErr == nil && secondaryManga != nil {
		if primaryErr != nil && r.log != nil {
			r.log.Error("hybrid primary get manga by id failed", map[string]interface{}{"error": primaryErr.Error()})
		}
		return secondaryManga, nil
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

func (r *HybridMangaRepository) GetMangaBySlug(ctx context.Context, slug string) (*coremanga.Manga, error) {
	var primaryManga, secondaryManga *coremanga.Manga
	var primaryErr, secondaryErr error

	// Query primary repository
	if r.primary != nil {
		primaryManga, primaryErr = r.primary.GetMangaBySlug(ctx, slug)
	}

	// Query secondary repository
	if r.secondary != nil {
		secondaryManga, secondaryErr = r.secondary.GetMangaBySlug(ctx, slug)
	}

	// Return primary result if successful
	if primaryErr == nil && primaryManga != nil {
		if secondaryErr != nil && r.log != nil {
			r.log.Error("hybrid secondary get manga by slug failed", map[string]interface{}{"error": secondaryErr.Error()})
		}
		return primaryManga, nil
	}

	// Return secondary result if primary failed but secondary succeeded
	if secondaryErr == nil && secondaryManga != nil {
		if primaryErr != nil && r.log != nil {
			r.log.Error("hybrid primary get manga by slug failed", map[string]interface{}{"error": primaryErr.Error()})
		}
		return secondaryManga, nil
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

func (r *HybridMangaRepository) ListMangas(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	mangaMap := make(map[string]*coremanga.Manga)
	var primarySuccess, secondarySuccess bool

	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication and ensure minimum coverage
	fetchLimit := skip + limit + 1000
	if fetchLimit < 5000 {
		fetchLimit = 5000
	}

	// Query primary repository
	if r.primary != nil {
		mangas, _, err := r.primary.ListMangas(ctx, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, manga := range mangas {
				mangaMap[manga.ID.Hex()] = manga
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list mangas failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge unique items
	if r.secondary != nil {
		mangas, _, err := r.secondary.ListMangas(ctx, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, manga := range mangas {
				if _, exists := mangaMap[manga.ID.Hex()]; !exists {
					mangaMap[manga.ID.Hex()] = manga
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list mangas failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, 0, errors.New("no repositories available or all failed")
	}

	// Convert map to slice and sort by created_at DESC
	result := make([]*coremanga.Manga, 0, len(mangaMap))
	for _, manga := range mangaMap {
		result = append(result, manga)
	}

	// Sort by created_at descending (most recent first) using sort.Slice for efficiency
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.Manga{}, total, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], total, nil
}

func (r *HybridMangaRepository) UpdateManga(ctx context.Context, manga *coremanga.Manga) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.UpdateManga(ctx, manga)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary update manga failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.UpdateManga(ctx, manga)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary update manga failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid update manga failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) DeleteManga(ctx context.Context, id primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.DeleteManga(ctx, id)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary delete manga failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.DeleteManga(ctx, id)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary delete manga failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid delete manga failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.IncrementViews(ctx, mangaID)
	}
	if r.secondary != nil {
		errSecondary = r.secondary.IncrementViews(ctx, mangaID)
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid increment views failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) LogView(ctx context.Context, mangaID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.LogView(ctx, mangaID)
	}
	if r.secondary != nil {
		errSecondary = r.secondary.LogView(ctx, mangaID)
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid log view failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*coremanga.RankedManga, error) {
	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication and merging
	fetchLimit := skip + limit + 100
	if fetchLimit < 500 {
		fetchLimit = 500
	}

	rankedMap := make(map[string]*coremanga.RankedManga)
	var primarySuccess, secondarySuccess bool

	// Query primary repository
	if r.primary != nil {
		ranked, err := r.primary.ListMostViewed(ctx, since, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, rm := range ranked {
				rankedMap[rm.Manga.ID.Hex()] = &coremanga.RankedManga{
					Manga:     rm.Manga,
					ViewCount: rm.ViewCount,
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list most viewed failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge results
	if r.secondary != nil {
		ranked, err := r.secondary.ListMostViewed(ctx, since, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, rm := range ranked {
				idStr := rm.Manga.ID.Hex()
				if existing, exists := rankedMap[idStr]; exists {
					// Merge view counts from both backends for the same manga
					existing.ViewCount += rm.ViewCount
				} else {
					rankedMap[idStr] = &coremanga.RankedManga{
						Manga:     rm.Manga,
						ViewCount: rm.ViewCount,
					}
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list most viewed failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, errors.New("no repositories available or all failed for list most viewed")
	}

	// Convert map to slice and sort by view count descending
	result := make([]*coremanga.RankedManga, 0, len(rankedMap))
	for _, rm := range rankedMap {
		result = append(result, rm)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ViewCount > result[j].ViewCount
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.RankedManga{}, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], nil
}

func (r *HybridMangaRepository) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication
	fetchLimit := skip + limit + 100
	if fetchLimit < 500 {
		fetchLimit = 500
	}

	mangaMap := make(map[string]*coremanga.Manga)
	var primarySuccess, secondarySuccess bool

	// Query primary repository
	if r.primary != nil {
		mangas, err := r.primary.ListRecentlyUpdated(ctx, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, manga := range mangas {
				mangaMap[manga.ID.Hex()] = manga
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list recently updated failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge unique items
	if r.secondary != nil {
		mangas, err := r.secondary.ListRecentlyUpdated(ctx, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, manga := range mangas {
				if _, exists := mangaMap[manga.ID.Hex()]; !exists {
					mangaMap[manga.ID.Hex()] = manga
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list recently updated failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, errors.New("no repositories available or all failed for list recently updated")
	}

	// Convert map to slice and sort by updated_at DESC
	result := make([]*coremanga.Manga, 0, len(mangaMap))
	for _, manga := range mangaMap {
		result = append(result, manga)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.Manga{}, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], nil
}

func (r *HybridMangaRepository) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication
	fetchLimit := skip + limit + 100
	if fetchLimit < 500 {
		fetchLimit = 500
	}

	mangaMap := make(map[string]*coremanga.Manga)
	var primarySuccess, secondarySuccess bool

	// Query primary repository
	if r.primary != nil {
		mangas, err := r.primary.ListMostFollowed(ctx, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, manga := range mangas {
				mangaMap[manga.ID.Hex()] = manga
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list most followed failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge unique items
	if r.secondary != nil {
		mangas, err := r.secondary.ListMostFollowed(ctx, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, manga := range mangas {
				if _, exists := mangaMap[manga.ID.Hex()]; !exists {
					mangaMap[manga.ID.Hex()] = manga
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list most followed failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, errors.New("no repositories available or all failed for list most followed")
	}

	// Convert map to slice and sort by favorites_count DESC
	result := make([]*coremanga.Manga, 0, len(mangaMap))
	for _, manga := range mangaMap {
		result = append(result, manga)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FavoritesCount > result[j].FavoritesCount
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.Manga{}, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], nil
}

func (r *HybridMangaRepository) ListTopRated(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, error) {
	// Calculate fetch limit to ensure we have enough data for the requested page
	// Add buffer for deduplication
	fetchLimit := skip + limit + 100
	if fetchLimit < 500 {
		fetchLimit = 500
	}

	mangaMap := make(map[string]*coremanga.Manga)
	var primarySuccess, secondarySuccess bool

	// Query primary repository
	if r.primary != nil {
		mangas, err := r.primary.ListTopRated(ctx, 0, fetchLimit)
		if err == nil {
			primarySuccess = true
			for _, manga := range mangas {
				mangaMap[manga.ID.Hex()] = manga
			}
		} else if r.log != nil {
			r.log.Error("hybrid primary list top rated failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Query secondary repository and merge unique items
	if r.secondary != nil {
		mangas, err := r.secondary.ListTopRated(ctx, 0, fetchLimit)
		if err == nil {
			secondarySuccess = true
			for _, manga := range mangas {
				if _, exists := mangaMap[manga.ID.Hex()]; !exists {
					mangaMap[manga.ID.Hex()] = manga
				}
			}
		} else if r.log != nil {
			r.log.Error("hybrid secondary list top rated failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Only error if no repositories are available at all
	if !primarySuccess && !secondarySuccess {
		return nil, errors.New("no repositories available or all failed for list top rated")
	}

	// Convert map to slice and sort by average_rating DESC, then rating_count DESC
	result := make([]*coremanga.Manga, 0, len(mangaMap))
	for _, manga := range mangaMap {
		result = append(result, manga)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].AverageRating != result[j].AverageRating {
			return result[i].AverageRating > result[j].AverageRating
		}
		return result[i].RatingCount > result[j].RatingCount
	})

	// Apply pagination after sorting
	total := int64(len(result))
	start := skip
	end := skip + limit

	if start >= total {
		return []*coremanga.Manga{}, nil
	}
	if end > total {
		end = total
	}

	return result[start:end], nil
}

func (r *HybridMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (string, error) {
	var reaction string
	var errPrimary, errSecondary error
	if r.primary != nil {
		reaction, errPrimary = r.primary.SetReaction(ctx, mangaID, userID, reactionType)
		if errPrimary != nil && strings.Contains(errPrimary.Error(), "reaction request already in progress") {
			// Respect the primary lock, do not fall back to secondary in this anti-spam scenario.
			return "", errPrimary
		}
	}
	if r.secondary != nil {
		_, errSecondary = r.secondary.SetReaction(ctx, mangaID, userID, reactionType)
	}
	if errPrimary != nil {
		if r.log != nil {
			r.log.Error("hybrid primary set reaction failed", map[string]interface{}{"error": errPrimary.Error()})
		}
		return "", errPrimary
	}
	if errSecondary != nil && r.log != nil {
		r.log.Error("hybrid secondary set reaction failed", map[string]interface{}{"error": errSecondary.Error()})
	}
	return reaction, nil
}

func (r *HybridMangaRepository) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
	if r.primary != nil {
		reaction, err := r.primary.GetUserReaction(ctx, mangaID, userID)
		if err == nil {
			return reaction, nil
		}
		if !errors.Is(err, core.ErrNotFound) {
			r.log.Warn("primary get user reaction failed", map[string]interface{}{
				"manga_id": mangaID.Hex(),
				"user_id":  userID.Hex(),
				"error":    err.Error(),
			})
		}
	}
	if r.secondary != nil {
		reaction, err := r.secondary.GetUserReaction(ctx, mangaID, userID)
		if err == nil {
			return reaction, nil
		}
		if !errors.Is(err, core.ErrNotFound) {
			r.log.Warn("secondary get user reaction failed", map[string]interface{}{
				"manga_id": mangaID.Hex(),
				"user_id":  userID.Hex(),
				"error":    err.Error(),
			})
		}
	}
	return "", nil
}

func (r *HybridMangaRepository) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r.primary != nil {
		mangas, total, err := r.primary.ListLikedMangas(ctx, userID, skip, limit)
		if err == nil {
			return mangas, total, nil
		}
		if r.log != nil {
			r.log.Error("hybrid primary list liked mangas failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		return r.secondary.ListLikedMangas(ctx, userID, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridMangaRepository) AddRating(ctx context.Context, rating *coremanga.MangaRating) (float64, error) {
	var average float64
	var errPrimary, errSecondary error
	if r.primary != nil {
		average, errPrimary = r.primary.AddRating(ctx, rating)
	}
	if errPrimary == nil && r.secondary == nil {
		return average, nil
	}
	if r.secondary != nil {
		if _, errSecondary = r.secondary.AddRating(ctx, rating); errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary add rating failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return 0, errors.New("hybrid add rating failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return 0, errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return 0, errSecondary
	}
	return average, nil
}

func (r *HybridMangaRepository) HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if r.primary != nil {
		v, err := r.primary.HasUserRated(ctx, mangaID, userID)
		if err == nil {
			return v, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.HasUserRated(ctx, mangaID, userID)
	}
	return false, errors.New("no repositories available")
}

// ========== ENGAGEMENT METHODS ==========

func (r *HybridMangaRepository) AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.AddFavorite(ctx, mangaID, userID)
	}
	if r.secondary != nil {
		errSecondary = r.secondary.AddFavorite(ctx, mangaID, userID)
	}
	if errPrimary != nil || errSecondary != nil {
		if r.log != nil {
			if errPrimary != nil {
				r.log.Error("hybrid primary add favorite failed", map[string]interface{}{"error": errPrimary.Error()})
			}
			if errSecondary != nil {
				r.log.Error("hybrid secondary add favorite failed", map[string]interface{}{"error": errSecondary.Error()})
			}
		}
		return fmt.Errorf("dual-write requirement failed: primary=%v, secondary=%v", errPrimary, errSecondary)
	}
	return nil
}

func (r *HybridMangaRepository) RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.RemoveFavorite(ctx, mangaID, userID)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary remove favorite failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.RemoveFavorite(ctx, mangaID, userID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary remove favorite failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid remove favorite failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	if r.primary != nil {
		v, err := r.primary.IsFavorite(ctx, mangaID, userID)
		if err == nil {
			return v, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.IsFavorite(ctx, mangaID, userID)
	}
	return false, errors.New("no repositories available")
}

func (r *HybridMangaRepository) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r.primary != nil {
		mangas, total, err := r.primary.ListFavorites(ctx, userID, skip, limit)
		if err == nil {
			return mangas, total, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, 0, err
		}
	}
	if r.secondary != nil {
		return r.secondary.ListFavorites(ctx, userID, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridMangaRepository) AddMangaComment(ctx context.Context, comment *coremanga.MangaComment) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.AddMangaComment(ctx, comment)
	}
	if r.secondary != nil {
		// Create a deep copy for secondary to prevent ID/timestamp conflicts
		commentJSON, _ := json.Marshal(comment)
		var commentCopy coremanga.MangaComment
		json.Unmarshal(commentJSON, &commentCopy)
		errSecondary = r.secondary.AddMangaComment(ctx, &commentCopy)
	}
	if errPrimary != nil || errSecondary != nil {
		if r.log != nil {
			if errPrimary != nil {
				r.log.Error("hybrid primary add manga comment failed", map[string]interface{}{"error": errPrimary.Error()})
			}
			if errSecondary != nil {
				r.log.Error("hybrid secondary add manga comment failed", map[string]interface{}{"error": errSecondary.Error()})
			}
		}
		return fmt.Errorf("dual-write requirement failed: primary=%v, secondary=%v", errPrimary, errSecondary)
	}
	return nil
}

func (r *HybridMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*coremanga.MangaComment, int64, error) {
	if r.primary != nil {
		comments, total, err := r.primary.ListMangaComments(ctx, mangaID, skip, limit, sortOrder)
		if err == nil {
			return comments, total, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, 0, err
		}
	}
	if r.secondary != nil {
		return r.secondary.ListMangaComments(ctx, mangaID, skip, limit, sortOrder)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridMangaRepository) DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	var errPrimary, errSecondary error
	if r.primary != nil {
		errPrimary = r.primary.DeleteMangaComment(ctx, commentID, userID)
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary delete manga comment failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.DeleteMangaComment(ctx, commentID, userID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary delete manga comment failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid delete manga comment failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_manga_repo.go -----
