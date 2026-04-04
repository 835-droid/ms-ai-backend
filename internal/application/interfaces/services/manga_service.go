package services

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/domain/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Note: Repository interfaces are now defined in the domain layer.
// Services depend on domain repository interfaces, not application interfaces.
// This follows Clean Architecture principles where high-level modules define the interfaces.

// MangaService defines the manga business logic operations.
type MangaService interface {
	// Manga CRUD operations
	CreateManga(ctx context.Context, manga *manga.Manga) (*manga.Manga, error)
	GetManga(ctx context.Context, id primitive.ObjectID) (*manga.Manga, error)
	ListMangas(ctx context.Context, skip, limit int64) ([]*manga.Manga, int64, error)
	IncrementViews(ctx context.Context, id primitive.ObjectID) (*manga.Manga, error)
	ListMostViewed(ctx context.Context, period string, skip, limit int64) ([]*manga.RankedManga, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)
	UpdateManga(ctx context.Context, manga *manga.Manga, callerID primitive.ObjectID, roles []string) error
	DeleteManga(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error

	// Manga reactions
	SetReaction(ctx context.Context, id, userID primitive.ObjectID, reactionType manga.ReactionType) (*manga.Manga, string, error)
	GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error)
	ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.Manga, int64, error)

	// Manga ratings
	AddRating(ctx context.Context, id, userID primitive.ObjectID, score float64) (*manga.Manga, error)

	// Manga favorites (engagement)
	AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.Manga, int64, error)

	// Manga comments
	AddMangaComment(ctx context.Context, comment *manga.MangaComment) error
	ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*manga.MangaComment, int64, error)
	DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error
}

// MangaChapterService defines the manga chapter business logic operations.
type MangaChapterService interface {
	// Chapter CRUD operations
	CreateChapter(ctx context.Context, chapter *manga.MangaChapter) (*manga.MangaChapter, error)
	GetChapter(ctx context.Context, id primitive.ObjectID) (*manga.MangaChapter, error)
	ListChapters(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*manga.MangaChapter, int64, error)
	UpdateChapter(ctx context.Context, chapter *manga.MangaChapter, callerID primitive.ObjectID, roles []string) error
	DeleteChapter(ctx context.Context, id primitive.ObjectID, callerID primitive.ObjectID, roles []string) error

	// Chapter views
	IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error
	TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error
	HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error)

	// Chapter ratings
	AddChapterRating(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID, score float64) (newAverage float64, count int64, userScore float64, err error)
	HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error)

	// Chapter comments
	AddChapterComment(ctx context.Context, comment *manga.ChapterComment) error
	ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*manga.ChapterComment, int64, error)
	DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error

	// Chapter comment reactions
	AddChapterCommentReaction(ctx context.Context, reaction *manga.ChapterCommentReaction) error
	RemoveChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) error
	GetUserChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error)
}

// FavoriteListService defines the favorite list business logic operations.
type FavoriteListService interface {
	// List operations
	CreateList(ctx context.Context, userID, name, description string, isPublic bool) (*manga.FavoriteList, error)
	GetList(ctx context.Context, listID, userID string) (*manga.FavoriteList, error)
	ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*manga.FavoriteList, int64, error)
	UpdateList(ctx context.Context, listID, userID, name, description string, isPublic bool) (*manga.FavoriteList, error)
	DeleteList(ctx context.Context, listID, userID string) error

	// List item operations
	AddMangaToList(ctx context.Context, listID, userID, mangaID, notes string) error
	RemoveMangaFromList(ctx context.Context, listID, userID, mangaID string) error
	ListMangaInList(ctx context.Context, listID, userID string, skip, limit int64) ([]*manga.FavoriteListItem, int64, error)
	GetUserMangaLists(ctx context.Context, userID, mangaID string) ([]*manga.FavoriteList, error)
}

// ViewingHistoryService defines the viewing history business logic operations.
type ViewingHistoryService interface {
	// TrackView records a view in the history and returns the history entry.
	TrackView(ctx context.Context, userID, mangaID primitive.ObjectID, chapterID primitive.ObjectID, page int) (*manga.ViewingHistory, error)

	// GetUserHistory retrieves a user's viewing history.
	GetUserHistory(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.ViewingHistory, int64, error)

	// GetRecentHistory retrieves recent viewing history.
	GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*manga.ViewingHistory, error)

	// GetUserStats retrieves statistics about the user's viewing history.
	GetUserStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error)

	// DeleteHistoryItem deletes a specific history item.
	DeleteHistoryItem(ctx context.Context, id, userID primitive.ObjectID) error

	// DeleteHistoryByManga deletes all history for a specific manga.
	DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error

	// CleanOldHistory deletes old history entries.
	CleanOldHistory(ctx context.Context, userID primitive.ObjectID, days int) (int64, error)
}
