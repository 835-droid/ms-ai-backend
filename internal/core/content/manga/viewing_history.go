package manga

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewingHistory represents a user's viewing history for manga/chapters
type ViewingHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	ChapterID primitive.ObjectID `bson:"chapter_id,omitempty" json:"chapter_id,omitempty"`
	Page      int                `bson:"page,omitempty" json:"page,omitempty"`
	ViewedAt  time.Time          `bson:"viewed_at" json:"viewed_at"`
	Duration  int                `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// HistoryStats represents statistics about a user's viewing history
type HistoryStats struct {
	TotalViews     int64            `json:"total_views"`
	UniqueManga    int64            `json:"unique_manga"`
	UniqueChapters int64            `json:"unique_chapters"`
	TotalDuration  int64            `json:"total_duration"` // in seconds
	RecentManga    []*Manga         `json:"recent_manga,omitempty"`
	ByPeriod       map[string]int64 `json:"by_period,omitempty"`
}

// ViewingHistoryRepository defines the viewing history data operations
type ViewingHistoryRepository interface {
	// Create/Update operations
	CreateHistory(ctx context.Context, history *ViewingHistory) error
	UpdateHistory(ctx context.Context, history *ViewingHistory) error

	// Query operations
	GetHistoryByUser(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*ViewingHistory, int64, error)
	GetHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) (*ViewingHistory, error)
	GetHistoryByChapter(ctx context.Context, userID, chapterID primitive.ObjectID) (*ViewingHistory, error)

	// Get recent history for a user (for quick access)
	GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*ViewingHistory, error)

	// Stats
	GetUserStats(ctx context.Context, userID primitive.ObjectID) (*HistoryStats, error)

	// Delete operations
	DeleteHistory(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error
	DeleteHistoryByChapter(ctx context.Context, userID, chapterID primitive.ObjectID) error

	// Bulk operations
	DeleteOlderThan(ctx context.Context, userID primitive.ObjectID, before time.Time) (int64, error)
}

// ViewingHistoryService defines the viewing history business logic
type ViewingHistoryService interface {
	// Track viewing
	TrackView(ctx context.Context, userID, mangaID primitive.ObjectID, chapterID primitive.ObjectID, page int) (*ViewingHistory, error)

	// Get history
	GetUserHistory(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*ViewingHistoryWithManga, int64, error)
	GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*ViewingHistoryWithManga, error)

	// Get stats
	GetUserStats(ctx context.Context, userID primitive.ObjectID) (*HistoryStats, error)

	// Delete
	DeleteHistoryItem(ctx context.Context, historyID, userID primitive.ObjectID) error
	DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error
	CleanOldHistory(ctx context.Context, userID primitive.ObjectID, days int) (int64, error)
}

// ViewingHistoryWithManga represents a history entry with manga details
type ViewingHistoryWithManga struct {
	*ViewingHistory
	Manga *Manga `json:"manga,omitempty"`
}

// DefaultViewingHistoryService implements ViewingHistoryService
type DefaultViewingHistoryService struct {
	repo      ViewingHistoryRepository
	mangaRepo MangaRepository
}

// NewViewingHistoryService creates a new viewing history service
func NewViewingHistoryService(repo ViewingHistoryRepository, mangaRepo MangaRepository) *DefaultViewingHistoryService {
	return &DefaultViewingHistoryService{repo: repo, mangaRepo: mangaRepo}
}

// TrackView tracks a manga/chapter view
func (s *DefaultViewingHistoryService) TrackView(ctx context.Context, userID, mangaID primitive.ObjectID, chapterID primitive.ObjectID, page int) (*ViewingHistory, error) {
	// Check if history exists for this manga
	existing, err := s.repo.GetHistoryByManga(ctx, userID, mangaID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if existing != nil {
		// Update existing
		existing.ChapterID = chapterID
		existing.Page = page
		existing.ViewedAt = now
		existing.UpdatedAt = now
		if chapterID.IsZero() {
			existing.Duration = 0 // Reset duration for manga-level view
		}

		if err := s.repo.UpdateHistory(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new
	history := &ViewingHistory{
		UserID:    userID,
		MangaID:   mangaID,
		ChapterID: chapterID,
		Page:      page,
		ViewedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateHistory(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

// GetUserHistory retrieves a user's viewing history with manga details
func (s *DefaultViewingHistoryService) GetUserHistory(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*ViewingHistoryWithManga, int64, error) {
	histories, total, err := s.repo.GetHistoryByUser(ctx, userID, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*ViewingHistoryWithManga, 0, len(histories))
	for _, h := range histories {
		manga, err := s.mangaRepo.GetMangaByID(ctx, h.MangaID)
		if err != nil {
			continue // Skip if manga not found
		}
		result = append(result, &ViewingHistoryWithManga{
			ViewingHistory: h,
			Manga:          manga,
		})
	}

	return result, total, nil
}

// GetRecentHistory retrieves recent viewing history
func (s *DefaultViewingHistoryService) GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*ViewingHistoryWithManga, error) {
	histories, err := s.repo.GetRecentHistory(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*ViewingHistoryWithManga, 0, len(histories))
	for _, h := range histories {
		manga, err := s.mangaRepo.GetMangaByID(ctx, h.MangaID)
		if err != nil {
			continue
		}
		result = append(result, &ViewingHistoryWithManga{
			ViewingHistory: h,
			Manga:          manga,
		})
	}

	return result, nil
}

// GetUserStats retrieves statistics about a user's viewing history
func (s *DefaultViewingHistoryService) GetUserStats(ctx context.Context, userID primitive.ObjectID) (*HistoryStats, error) {
	return s.repo.GetUserStats(ctx, userID)
}

// DeleteHistoryItem deletes a specific history item
func (s *DefaultViewingHistoryService) DeleteHistoryItem(ctx context.Context, historyID, userID primitive.ObjectID) error {
	return s.repo.DeleteHistory(ctx, historyID, userID)
}

// DeleteHistoryByManga deletes all history for a specific manga
func (s *DefaultViewingHistoryService) DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error {
	return s.repo.DeleteHistoryByManga(ctx, userID, mangaID)
}

// CleanOldHistory deletes history older than specified days
func (s *DefaultViewingHistoryService) CleanOldHistory(ctx context.Context, userID primitive.ObjectID, days int) (int64, error) {
	if days <= 0 {
		days = 90 // Default to 90 days
	}
	before := time.Now().AddDate(0, 0, -days)
	return s.repo.DeleteOlderThan(ctx, userID, before)
}
