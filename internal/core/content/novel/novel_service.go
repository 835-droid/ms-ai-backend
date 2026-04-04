// Package novel defines novel business logic operations.
package novel

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

// NovelService defines novel business logic operations.
type NovelService interface {
	CreateNovel(ctx context.Context, novel *Novel) (*Novel, error)
	GetNovel(ctx context.Context, id primitive.ObjectID) (*Novel, error)
	ListNovels(ctx context.Context, skip, limit int64) ([]*Novel, int64, error)
	IncrementViews(ctx context.Context, id primitive.ObjectID) (*Novel, error)
	ListMostViewed(ctx context.Context, period string, skip, limit int64) ([]*RankedNovel, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Novel, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Novel, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*Novel, error)
	SetReaction(ctx context.Context, id, userID primitive.ObjectID, reactionType ReactionType) (*Novel, string, error)
	GetUserReaction(ctx context.Context, novelID, userID primitive.ObjectID) (string, error)
	ListLikedNovels(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error)
	AddRating(ctx context.Context, id, userID primitive.ObjectID, score float64) (*Novel, error)
	UpdateNovel(ctx context.Context, novel *Novel, callerID primitive.ObjectID, roles []string) error
	DeleteNovel(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error
	// Engagement methods
	AddFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error)
	AddNovelComment(ctx context.Context, comment *NovelComment) error
	ListNovelComments(ctx context.Context, novelID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*NovelComment, int64, error)
	DeleteNovelComment(ctx context.Context, commentID, userID primitive.ObjectID) error
}

// DefaultNovelService implements NovelService.
type DefaultNovelService struct {
	repo NovelRepository
	log  *logger.Logger
}

// NewNovelService creates a new DefaultNovelService.
func NewNovelService(repo NovelRepository, log *logger.Logger) *DefaultNovelService {
	return &DefaultNovelService{repo: repo, log: log}
}

// CreateNovel creates a new novel.
func (s *DefaultNovelService) CreateNovel(ctx context.Context, novel *Novel) (*Novel, error) {
	s.log.Debug("creating novel", map[string]interface{}{
		"title":     novel.Title,
		"author_id": novel.AuthorID.Hex(),
	})

	if err := validator.Validate(novel); err != nil {
		s.log.Error("novel validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	novel.Slug = utils.Slugify(novel.Title)
	now := time.Now()
	novel.CreatedAt = now
	novel.UpdatedAt = now

	if err := s.repo.CreateNovel(ctx, novel); err != nil {
		s.log.Error("novel creation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	s.log.Info("novel created", map[string]interface{}{
		"id":    novel.ID.Hex(),
		"title": novel.Title,
	})
	return novel, nil
}

// GetNovel retrieves a novel by ID.
func (s *DefaultNovelService) GetNovel(ctx context.Context, id primitive.ObjectID) (*Novel, error) {
	s.log.Debug("getting novel", map[string]interface{}{
		"id": id.Hex(),
	})

	novel, err := s.repo.GetNovelByID(ctx, id)
	if err != nil {
		s.log.Error("get novel failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return nil, err
	}

	if novel == nil {
		s.log.Error("novel not found", map[string]interface{}{
			"id": id.Hex(),
		})
		return nil, core.ErrNotFound
	}

	s.log.Debug("novel retrieved", map[string]interface{}{
		"id":    novel.ID.Hex(),
		"title": novel.Title,
	})
	return novel, nil
}

// ListNovels retrieves a paginated list of novels.
func (s *DefaultNovelService) ListNovels(ctx context.Context, skip, limit int64) ([]*Novel, int64, error) {
	s.log.Debug("listing novels", map[string]interface{}{
		"skip":  skip,
		"limit": limit,
	})

	novels, total, err := s.repo.ListNovels(ctx, skip, limit)
	if err != nil {
		s.log.Error("list novels failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, err
	}

	s.log.Debug("novels listed", map[string]interface{}{
		"count": len(novels),
		"total": total,
	})
	return novels, total, nil
}

// IncrementViews increases the novel view count and returns the updated novel.
func (s *DefaultNovelService) IncrementViews(ctx context.Context, id primitive.ObjectID) (*Novel, error) {
	if err := s.repo.LogView(ctx, id); err != nil {
		s.log.Error("increment novel views failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, err
	}
	return s.repo.GetNovelByID(ctx, id)
}

func (s *DefaultNovelService) ListMostViewed(ctx context.Context, period string, skip, limit int64) ([]*RankedNovel, error) {
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

func (s *DefaultNovelService) ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Novel, error) {
	return s.repo.ListRecentlyUpdated(ctx, skip, limit)
}

func (s *DefaultNovelService) ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Novel, error) {
	return s.repo.ListMostFollowed(ctx, skip, limit)
}

func (s *DefaultNovelService) ListTopRated(ctx context.Context, skip, limit int64) ([]*Novel, error) {
	return s.repo.ListTopRated(ctx, skip, limit)
}

// SetReaction sets or toggles a reaction for a novel and returns the updated novel.
func (s *DefaultNovelService) SetReaction(ctx context.Context, id, userID primitive.ObjectID, reactionType ReactionType) (*Novel, string, error) {
	reaction, err := s.repo.SetReaction(ctx, id, userID, reactionType)
	if err != nil {
		s.log.Error("set novel reaction failed", map[string]interface{}{
			"id":            id.Hex(),
			"user_id":       userID.Hex(),
			"reaction_type": string(reactionType),
			"error":         err.Error(),
		})
		return nil, "", err
	}

	novel, err := s.repo.GetNovelByID(ctx, id)
	if err != nil {
		s.log.Error("get novel after reaction set failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, reaction, err
	}

	return novel, reaction, nil
}

// GetUserReaction gets the current reaction type for a user on a novel.
func (s *DefaultNovelService) GetUserReaction(ctx context.Context, novelID, userID primitive.ObjectID) (string, error) {
	return s.repo.GetUserReaction(ctx, novelID, userID)
}

// ListLikedNovels returns novels liked by a user.
func (s *DefaultNovelService) ListLikedNovels(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error) {
	novels, total, err := s.repo.ListLikedNovels(ctx, userID, skip, limit)
	if err != nil {
		s.log.Error("list liked novels failed", map[string]interface{}{
			"user_id": userID.Hex(),
			"error":   err.Error(),
		})
		return nil, 0, err
	}
	return novels, total, nil
}

// AddRating stores a user rating and returns the updated novel.
func (s *DefaultNovelService) AddRating(ctx context.Context, id, userID primitive.ObjectID, score float64) (*Novel, error) {
	_, err := s.repo.AddRating(ctx, &NovelRating{
		NovelID: id,
		UserID:  userID,
		Score:   score,
	})
	if err != nil {
		s.log.Error("add novel rating failed", map[string]interface{}{
			"id":      id.Hex(),
			"user_id": userID.Hex(),
			"score":   score,
			"error":   err.Error(),
		})
		return nil, err
	}

	return s.repo.GetNovelByID(ctx, id)
}

// UpdateNovel updates a novel.
func (s *DefaultNovelService) UpdateNovel(ctx context.Context, novel *Novel, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("updating novel", map[string]interface{}{
		"id":        novel.ID.Hex(),
		"title":     novel.Title,
		"caller_id": callerID.Hex(),
	})

	// Verify ownership or admin role
	if novel.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized novel update attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"novel_id":  novel.ID.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := validator.Validate(novel); err != nil {
		s.log.Error("novel validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	novel.Slug = utils.Slugify(novel.Title)
	novel.UpdatedAt = time.Now()

	if err := s.repo.UpdateNovel(ctx, novel); err != nil {
		s.log.Error("novel update failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.log.Info("novel updated", map[string]interface{}{
		"id":    novel.ID.Hex(),
		"title": novel.Title,
	})
	return nil
}

// DeleteNovel deletes a novel.
func (s *DefaultNovelService) DeleteNovel(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error {
	s.log.Debug("deleting novel", map[string]interface{}{
		"id":        id.Hex(),
		"caller_id": callerID.Hex(),
	})

	novel, err := s.repo.GetNovelByID(ctx, id)
	if err != nil {
		s.log.Error("get novel for deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	if novel.AuthorID != callerID && !hasRole(roles, "admin") {
		s.log.Error("unauthorized novel deletion attempt", map[string]interface{}{
			"caller_id": callerID.Hex(),
			"novel_id":  id.Hex(),
		})
		return core.ErrUnauthorized
	}

	if err := s.repo.DeleteNovel(ctx, id); err != nil {
		s.log.Error("novel deletion failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id.Hex(),
		})
		return err
	}

	s.log.Info("novel deleted", map[string]interface{}{
		"id": id.Hex(),
	})
	return nil
}

// ========== ENGAGEMENT METHODS ==========

// AddFavorite adds a novel to user's favorites.
func (s *DefaultNovelService) AddFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	s.log.Debug("adding favorite", map[string]interface{}{
		"novel_id": novelID.Hex(),
		"user_id":  userID.Hex(),
	})

	if err := s.repo.AddFavorite(ctx, novelID, userID); err != nil {
		s.log.Error("add favorite failed", map[string]interface{}{
			"error":    err.Error(),
			"novel_id": novelID.Hex(),
			"user_id":  userID.Hex(),
		})
		return err
	}

	s.log.Info("favorite added", map[string]interface{}{
		"novel_id": novelID.Hex(),
		"user_id":  userID.Hex(),
	})
	return nil
}

// RemoveFavorite removes a novel from user's favorites.
func (s *DefaultNovelService) RemoveFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error {
	s.log.Debug("removing favorite", map[string]interface{}{
		"novel_id": novelID.Hex(),
		"user_id":  userID.Hex(),
	})

	if err := s.repo.RemoveFavorite(ctx, novelID, userID); err != nil {
		s.log.Error("remove favorite failed", map[string]interface{}{
			"error":    err.Error(),
			"novel_id": novelID.Hex(),
			"user_id":  userID.Hex(),
		})
		return err
	}

	s.log.Info("favorite removed", map[string]interface{}{
		"novel_id": novelID.Hex(),
		"user_id":  userID.Hex(),
	})
	return nil
}

// IsFavorite checks if a novel is in user's favorites.
func (s *DefaultNovelService) IsFavorite(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error) {
	return s.repo.IsFavorite(ctx, novelID, userID)
}

// ListFavorites retrieves a user's favorite novels.
func (s *DefaultNovelService) ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error) {
	s.log.Debug("listing favorites", map[string]interface{}{
		"user_id": userID.Hex(),
		"skip":    skip,
		"limit":   limit,
	})

	novels, total, err := s.repo.ListFavorites(ctx, userID, skip, limit)
	if err != nil {
		s.log.Error("list favorites failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("favorites listed", map[string]interface{}{
		"user_id": userID.Hex(),
		"count":   len(novels),
		"total":   total,
	})
	return novels, total, nil
}

// AddNovelComment adds a comment to a novel.
func (s *DefaultNovelService) AddNovelComment(ctx context.Context, comment *NovelComment) error {
	s.log.Debug("adding novel comment", map[string]interface{}{
		"novel_id": comment.NovelID.Hex(),
		"user_id":  comment.UserID.Hex(),
	})

	if err := validator.Validate(comment); err != nil {
		s.log.Error("novel comment validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if err := s.repo.AddNovelComment(ctx, comment); err != nil {
		s.log.Error("add novel comment failed", map[string]interface{}{
			"error":    err.Error(),
			"novel_id": comment.NovelID.Hex(),
		})
		return err
	}

	s.log.Info("novel comment added", map[string]interface{}{
		"id":       comment.ID.Hex(),
		"novel_id": comment.NovelID.Hex(),
		"user_id":  comment.UserID.Hex(),
	})
	return nil
}

// ListNovelComments retrieves comments for a novel.
func (s *DefaultNovelService) ListNovelComments(ctx context.Context, novelID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*NovelComment, int64, error) {
	s.log.Debug("listing novel comments", map[string]interface{}{
		"novel_id":  novelID.Hex(),
		"skip":      skip,
		"limit":     limit,
		"sortOrder": sortOrder,
	})

	comments, total, err := s.repo.ListNovelComments(ctx, novelID, skip, limit, sortOrder)
	if err != nil {
		s.log.Error("list novel comments failed", map[string]interface{}{
			"error":    err.Error(),
			"novel_id": novelID.Hex(),
		})
		return nil, 0, err
	}

	s.log.Debug("novel comments listed", map[string]interface{}{
		"novel_id": novelID.Hex(),
		"count":    len(comments),
		"total":    total,
	})
	return comments, total, nil
}

// DeleteNovelComment deletes a novel comment.
func (s *DefaultNovelService) DeleteNovelComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	s.log.Debug("deleting novel comment", map[string]interface{}{
		"comment_id": commentID.Hex(),
		"user_id":    userID.Hex(),
	})

	if err := s.repo.DeleteNovelComment(ctx, commentID, userID); err != nil {
		s.log.Error("delete novel comment failed", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID.Hex(),
			"user_id":    userID.Hex(),
		})
		return err
	}

	s.log.Info("novel comment deleted", map[string]interface{}{
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
