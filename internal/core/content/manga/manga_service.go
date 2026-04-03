// ----- START OF FILE: backend/MS-AI/internal/core/content/manga/manga_service.go -----
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
	IncrementViews(ctx context.Context, id primitive.ObjectID) (*Manga, error)
	ListMostViewed(ctx context.Context, period string, skip, limit int64) ([]*RankedManga, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Manga, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Manga, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*Manga, error)
	SetReaction(ctx context.Context, id, userID primitive.ObjectID, reactionType ReactionType) (*Manga, string, error)
	GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error)
	ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error)
	AddRating(ctx context.Context, id, userID primitive.ObjectID, score float64) (*Manga, error)
	UpdateManga(ctx context.Context, manga *Manga, callerID primitive.ObjectID, roles []string) error
	DeleteManga(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error
	// Engagement methods
	AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error)
	AddMangaComment(ctx context.Context, comment *MangaComment) error
	ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*MangaComment, int64, error)
	DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error
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

	if manga == nil {
		s.log.Error("manga not found", map[string]interface{}{
			"id": id.Hex(),
		})
		return nil, core.ErrMangaNotFound
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

// IncrementViews increases the manga view count and returns the updated manga.
func (s *DefaultMangaService) IncrementViews(ctx context.Context, id primitive.ObjectID) (*Manga, error) {
	if err := s.repo.LogView(ctx, id); err != nil {
		s.log.Error("increment manga views failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, err
	}
	return s.repo.GetMangaByID(ctx, id)
}

func (s *DefaultMangaService) ListMostViewed(ctx context.Context, period string, skip, limit int64) ([]*RankedManga, error) {
	var since time.Time
	now := time.Now()
	switch strings.ToLower(period) {
	case "day":
		since = now.AddDate(0, 0, -1)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	case "all":
		since = time.Time{} // beginning of time
	default:
		since = now.AddDate(0, 0, -1) // default to day
	}
	return s.repo.ListMostViewed(ctx, since, skip, limit)
}

func (s *DefaultMangaService) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Manga, error) {
	return s.repo.ListRecentlyUpdated(ctx, skip, limit)
}

func (s *DefaultMangaService) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Manga, error) {
	return s.repo.ListMostFollowed(ctx, skip, limit)
}

func (s *DefaultMangaService) ListTopRated(ctx context.Context, skip, limit int64) ([]*Manga, error) {
	return s.repo.ListTopRated(ctx, skip, limit)
}

// SetReaction sets or toggles a reaction for a manga and returns the updated manga.
func (s *DefaultMangaService) SetReaction(ctx context.Context, id, userID primitive.ObjectID, reactionType ReactionType) (*Manga, string, error) {
	reaction, err := s.repo.SetReaction(ctx, id, userID, reactionType)
	if err != nil {
		s.log.Error("set manga reaction failed", map[string]interface{}{
			"id":            id.Hex(),
			"user_id":       userID.Hex(),
			"reaction_type": string(reactionType),
			"error":         err.Error(),
		})
		return nil, "", err
	}

	manga, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		s.log.Error("get manga after reaction set failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, reaction, err
	}

	return manga, reaction, nil
}

// GetUserReaction gets the current reaction type for a user on a manga.
func (s *DefaultMangaService) GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error) {
	return s.repo.GetUserReaction(ctx, mangaID, userID)
}

// ListLikedMangas returns mangas liked by a user.
func (s *DefaultMangaService) ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error) {
	mangas, total, err := s.repo.ListLikedMangas(ctx, userID, skip, limit)
	if err != nil {
		s.log.Error("list liked mangas failed", map[string]interface{}{
			"user_id": userID.Hex(),
			"error":   err.Error(),
		})
		return nil, 0, err
	}
	return mangas, total, nil
}

// AddRating stores a user rating and returns the updated manga.
func (s *DefaultMangaService) AddRating(ctx context.Context, id, userID primitive.ObjectID, score float64) (*Manga, error) {
	_, err := s.repo.AddRating(ctx, &MangaRating{
		MangaID: id,
		UserID:  userID,
		Score:   score,
	})
	if err != nil {
		s.log.Error("add manga rating failed", map[string]interface{}{
			"id":      id.Hex(),
			"user_id": userID.Hex(),
			"score":   score,
			"error":   err.Error(),
		})
		return nil, err
	}

	return s.repo.GetMangaByID(ctx, id)
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

// ========== ENGAGEMENT METHODS ==========

// AddFavorite adds a manga to user's favorites.
func (s *DefaultMangaService) AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	s.log.Debug("adding favorite", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"user_id":  userID.Hex(),
	})

	if err := s.repo.AddFavorite(ctx, mangaID, userID); err != nil {
		s.log.Error("add favorite failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": mangaID.Hex(),
			"user_id":  userID.Hex(),
		})
		return err
	}

	s.log.Info("favorite added", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"user_id":  userID.Hex(),
	})
	return nil
}

// RemoveFavorite removes a manga from user's favorites.
func (s *DefaultMangaService) RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error {
	s.log.Debug("removing favorite", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"user_id":  userID.Hex(),
	})

	if err := s.repo.RemoveFavorite(ctx, mangaID, userID); err != nil {
		s.log.Error("remove favorite failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": mangaID.Hex(),
			"user_id":  userID.Hex(),
		})
		return err
	}

	s.log.Info("favorite removed", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"user_id":  userID.Hex(),
	})
	return nil
}

// IsFavorite checks if a manga is in user's favorites.
func (s *DefaultMangaService) IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error) {
	return s.repo.IsFavorite(ctx, mangaID, userID)
}

// ListFavorites retrieves a user's favorite mangas.
func (s *DefaultMangaService) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error) {
	s.log.Debug("listing favorites", map[string]interface{}{
		"user_id": userID.Hex(),
		"skip":    skip,
		"limit":   limit,
	})

	mangas, total, err := s.repo.ListFavorites(ctx, userID, skip, limit)
	if err != nil {
		s.log.Error("list favorites failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("favorites listed", map[string]interface{}{
		"user_id": userID.Hex(),
		"count":   len(mangas),
		"total":   total,
	})
	return mangas, total, nil
}

// AddMangaComment adds a comment to a manga.
func (s *DefaultMangaService) AddMangaComment(ctx context.Context, comment *MangaComment) error {
	s.log.Debug("adding manga comment", map[string]interface{}{
		"manga_id": comment.MangaID.Hex(),
		"user_id":  comment.UserID.Hex(),
	})

	if err := validator.Validate(comment); err != nil {
		s.log.Error("manga comment validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if err := s.repo.AddMangaComment(ctx, comment); err != nil {
		s.log.Error("add manga comment failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": comment.MangaID.Hex(),
		})
		return err
	}

	s.log.Info("manga comment added", map[string]interface{}{
		"id":       comment.ID.Hex(),
		"manga_id": comment.MangaID.Hex(),
		"user_id":  comment.UserID.Hex(),
	})
	return nil
}

// ListMangaComments retrieves comments for a manga.
func (s *DefaultMangaService) ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*MangaComment, int64, error) {
	s.log.Debug("listing manga comments", map[string]interface{}{
		"manga_id":  mangaID.Hex(),
		"skip":      skip,
		"limit":     limit,
		"sortOrder": sortOrder,
	})

	comments, total, err := s.repo.ListMangaComments(ctx, mangaID, skip, limit, sortOrder)
	if err != nil {
		s.log.Error("list manga comments failed", map[string]interface{}{
			"error":    err.Error(),
			"manga_id": mangaID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("manga comments listed", map[string]interface{}{
		"manga_id": mangaID.Hex(),
		"count":    len(comments),
		"total":    total,
	})
	return comments, total, nil
}

// DeleteMangaComment deletes a manga comment.
func (s *DefaultMangaService) DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	s.log.Debug("deleting manga comment", map[string]interface{}{
		"comment_id": commentID.Hex(),
		"user_id":    userID.Hex(),
	})

	if err := s.repo.DeleteMangaComment(ctx, commentID, userID); err != nil {
		s.log.Error("delete manga comment failed", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID.Hex(),
			"user_id":    userID.Hex(),
		})
		return err
	}

	s.log.Info("manga comment deleted", map[string]interface{}{
		"comment_id": commentID.Hex(),
		"user_id":    userID.Hex(),
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

// ----- END OF FILE: backend/MS-AI/internal/core/content/manga/manga_service.go -----
