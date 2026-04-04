package repositories

import (
	"context"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/domain/manga"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaRepository defines the data access operations for manga.
type MangaRepository interface {
	// Manga CRUD operations
	CreateManga(ctx context.Context, manga *manga.Manga) error
	GetMangaByID(ctx context.Context, id primitive.ObjectID) (*manga.Manga, error)
	GetMangaBySlug(ctx context.Context, slug string) (*manga.Manga, error)
	ListMangas(ctx context.Context, skip, limit int64) ([]*manga.Manga, int64, error)
	UpdateManga(ctx context.Context, manga *manga.Manga) error
	DeleteManga(ctx context.Context, id primitive.ObjectID) error

	// Manga views
	IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error
	LogView(ctx context.Context, mangaID primitive.ObjectID) error
	ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*manga.RankedManga, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*manga.Manga, error)

	// Manga reactions
	SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType manga.ReactionType) (reaction string, err error)
	GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error)
	ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.Manga, int64, error)

	// Manga ratings
	AddRating(ctx context.Context, rating *manga.MangaRating) (newAverage float64, err error)
	HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)

	// Favorites
	AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.Manga, int64, error)

	// Manga comments
	AddMangaComment(ctx context.Context, comment *manga.MangaComment) error
	ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*manga.MangaComment, int64, error)
	DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error
}

// MangaChapterRepository defines the data access operations for manga chapters.
type MangaChapterRepository interface {
	// Chapter CRUD operations
	CreateMangaChapter(ctx context.Context, chapter *manga.MangaChapter) error
	GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*manga.MangaChapter, error)
	ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*manga.MangaChapter, int64, error)
	UpdateMangaChapter(ctx context.Context, chapter *manga.MangaChapter) error
	DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error

	// Chapter views
	IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error
	TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error
	HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error)

	// Chapter ratings
	AddChapterRating(ctx context.Context, rating *manga.ChapterRating) (newAverage float64, count int64, userScore float64, err error)
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

// FavoriteListRepository defines the data access operations for favorite lists.
type FavoriteListRepository interface {
	// List operations
	CreateList(ctx context.Context, list *manga.FavoriteList) error
	GetListByID(ctx context.Context, listID string) (*manga.FavoriteList, error)
	GetListByName(ctx context.Context, userID, name string) (*manga.FavoriteList, error)
	ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*manga.FavoriteList, int64, error)
	UpdateList(ctx context.Context, list *manga.FavoriteList) error
	DeleteList(ctx context.Context, listID, userID string) error
	GetListMangaCount(ctx context.Context, listID string) (int64, error)

	// List item operations
	AddMangaToList(ctx context.Context, item *manga.FavoriteListItem) error
	RemoveMangaFromList(ctx context.Context, listID, mangaID string) error
	IsMangaInList(ctx context.Context, listID, mangaID string) (bool, error)
	ListMangaInList(ctx context.Context, listID string, skip, limit int64) ([]*manga.FavoriteListItem, int64, error)
	UpdateListItemNotes(ctx context.Context, listID, mangaID, notes string) error
	MoveMangaToList(ctx context.Context, fromListID, toListID, mangaID string) error
	UpdateItemSortOrder(ctx context.Context, listID, mangaID string, sortOrder int) error

	// Cross-list operations
	GetUserMangaLists(ctx context.Context, userID, mangaID string) ([]*manga.FavoriteList, error)
	GetPublicListManga(ctx context.Context, listID string, skip, limit int64) ([]*manga.Manga, int64, error)
}

// ReadingProgressRepository defines the data access operations for reading progress.
type ReadingProgressRepository interface {
	// SaveProgress saves or updates a user's reading progress for a manga.
	SaveProgress(ctx context.Context, progress *manga.ReadingProgress) error

	// GetProgress gets a user's reading progress for a specific manga.
	GetProgress(ctx context.Context, mangaID, userID primitive.ObjectID) (*manga.ReadingProgress, error)

	// GetProgressForMangas gets reading progress for multiple mangas.
	GetProgressForMangas(ctx context.Context, mangaIDs []primitive.ObjectID, userID primitive.ObjectID) (map[string]*manga.ReadingProgress, error)

	// DeleteProgress deletes a user's reading progress for a manga.
	DeleteProgress(ctx context.Context, mangaID, userID primitive.ObjectID) error
}

// ViewingHistoryRepository defines the data access operations for viewing history.
type ViewingHistoryRepository interface {
	// TrackView records a view in the history.
	TrackView(ctx context.Context, history *manga.ViewingHistory) error

	// GetUserHistory retrieves a user's viewing history.
	GetUserHistory(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*manga.ViewingHistory, int64, error)

	// GetRecentHistory retrieves recent viewing history.
	GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*manga.ViewingHistory, error)

	// DeleteHistoryItem deletes a specific history item.
	DeleteHistoryItem(ctx context.Context, id, userID primitive.ObjectID) error

	// DeleteHistoryByManga deletes all history for a specific manga.
	DeleteHistoryByManga(ctx context.Context, userID, mangaID primitive.ObjectID) error

	// CleanOldHistory deletes old history entries.
	CleanOldHistory(ctx context.Context, userID primitive.ObjectID, days int) (int64, error)
}
