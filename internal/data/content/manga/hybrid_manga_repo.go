package manga

// ----- START OF FILE: backend/MS-AI/internal/data/content/manga/hybrid_manga_repo.go -----

import (
	"context"
	"errors"

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
	if r.primary != nil {
		manga, err := r.primary.GetMangaByID(ctx, id)
		if err == nil && manga != nil {
			return manga, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, err
		}
	}
	if r.secondary != nil {
		return r.secondary.GetMangaByID(ctx, id)
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridMangaRepository) GetMangaBySlug(ctx context.Context, slug string) (*coremanga.Manga, error) {
	if r.primary != nil {
		manga, err := r.primary.GetMangaBySlug(ctx, slug)
		if err == nil && manga != nil {
			return manga, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, err
		}
	}
	if r.secondary != nil {
		return r.secondary.GetMangaBySlug(ctx, slug)
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridMangaRepository) ListMangas(ctx context.Context, skip, limit int64) ([]*coremanga.Manga, int64, error) {
	if r.primary != nil {
		mangas, total, err := r.primary.ListMangas(ctx, skip, limit)
		if err == nil {
			return mangas, total, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.ListMangas(ctx, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
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

func (r *HybridMangaRepository) SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType coremanga.ReactionType) (string, error) {
	var reaction string
	var errPrimary, errSecondary error
	if r.primary != nil {
		reaction, errPrimary = r.primary.SetReaction(ctx, mangaID, userID, reactionType)
	}
	if errPrimary == nil && r.secondary == nil {
		return reaction, nil
	}
	if r.secondary != nil {
		_, errSecondary = r.secondary.SetReaction(ctx, mangaID, userID, reactionType)
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return "", errors.New("hybrid set reaction failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return "", errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return "", errSecondary
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
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary add favorite failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.AddFavorite(ctx, mangaID, userID)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary add favorite failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid add favorite failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
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
		if err == nil && len(mangas) > 0 {
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
		if errPrimary != nil && r.log != nil {
			r.log.Error("hybrid primary add manga comment failed", map[string]interface{}{"error": errPrimary.Error()})
		}
	}
	if r.secondary != nil {
		errSecondary = r.secondary.AddMangaComment(ctx, comment)
		if errSecondary != nil && r.log != nil {
			r.log.Error("hybrid secondary add manga comment failed", map[string]interface{}{"error": errSecondary.Error()})
		}
	}
	if errPrimary != nil && r.primary != nil && errSecondary != nil && r.secondary != nil {
		return errors.New("hybrid add manga comment failed on all backends")
	}
	if errPrimary != nil && r.primary != nil && r.secondary == nil {
		return errPrimary
	}
	if errSecondary != nil && r.primary == nil && r.secondary != nil {
		return errSecondary
	}
	return nil
}

func (r *HybridMangaRepository) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*coremanga.MangaComment, int64, error) {
	if r.primary != nil {
		comments, total, err := r.primary.ListMangaComments(ctx, mangaID, skip, limit)
		if err == nil && len(comments) > 0 {
			return comments, total, nil
		}
		if err != nil && !errors.Is(err, corecommon.ErrNotFound) {
			return nil, 0, err
		}
	}
	if r.secondary != nil {
		return r.secondary.ListMangaComments(ctx, mangaID, skip, limit)
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
