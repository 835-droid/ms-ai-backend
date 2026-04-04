// Package novel defines the core domain entities and repository interfaces for novel content.
package novel

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NovelRepository defines the data access operations for novel entities.
// This interface follows the Repository pattern and is part of the domain layer.
type NovelRepository interface {
	// Novel CRUD operations
	CreateNovel(ctx context.Context, novel *Novel) error
	GetNovelByID(ctx context.Context, id primitive.ObjectID) (*Novel, error)
	GetNovelBySlug(ctx context.Context, slug string) (*Novel, error)
	ListNovels(ctx context.Context, skip, limit int64) ([]*Novel, int64, error)
	UpdateNovel(ctx context.Context, novel *Novel) error
	DeleteNovel(ctx context.Context, id primitive.ObjectID) error

	// Novel views
	IncrementViews(ctx context.Context, novelID primitive.ObjectID) error
	LogView(ctx context.Context, novelID primitive.ObjectID) error
	ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*RankedNovel, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Novel, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Novel, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*Novel, error)

	// Novel reactions
	SetReaction(ctx context.Context, novelID, userID primitive.ObjectID, reactionType ReactionType) (reaction string, err error)
	GetUserReaction(ctx context.Context, novelID, userID primitive.ObjectID) (string, error)
	ListLikedNovels(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error)

	// Novel ratings
	AddRating(ctx context.Context, rating *NovelRating) (newAverage float64, err error)
	HasUserRated(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error)

	// Favorites
	AddFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, novelID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, novelID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Novel, int64, error)

	// Novel comments
	AddNovelComment(ctx context.Context, comment *NovelComment) error
	ListNovelComments(ctx context.Context, novelID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*NovelComment, int64, error)
	DeleteNovelComment(ctx context.Context, commentID, userID primitive.ObjectID) error
}

// NovelChapterRepository defines the data access operations for novel chapters.
type NovelChapterRepository interface {
	// Chapter CRUD operations
	CreateNovelChapter(ctx context.Context, chapter *NovelChapter) error
	GetNovelChapterByID(ctx context.Context, id primitive.ObjectID) (*NovelChapter, error)
	ListNovelChaptersByNovel(ctx context.Context, novelID primitive.ObjectID, skip, limit int64) ([]*NovelChapter, int64, error)
	UpdateNovelChapter(ctx context.Context, chapter *NovelChapter) error
	DeleteNovelChapter(ctx context.Context, id primitive.ObjectID) error

	// Chapter views
	IncrementChapterViews(ctx context.Context, chapterID, novelID primitive.ObjectID) error
	TrackChapterView(ctx context.Context, chapterID, novelID, userID primitive.ObjectID) error
	HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	TrackAndIncrementChapterView(ctx context.Context, chapterID, novelID, userID primitive.ObjectID) (bool, error)

	// Chapter ratings
	AddChapterRating(ctx context.Context, rating *ChapterRating) (newAverage float64, count int64, userScore float64, err error)
	HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error)

	// Chapter comments
	AddChapterComment(ctx context.Context, comment *ChapterComment) error
	ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*ChapterComment, int64, error)
	DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error

	// Chapter comment reactions
	AddChapterCommentReaction(ctx context.Context, reaction *ChapterCommentReaction) error
	RemoveChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) error
	GetUserChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error)
}

// NovelFavoriteListRepository defines the data access operations for novel favorite lists.
type NovelFavoriteListRepository interface {
	// List operations
	CreateList(ctx context.Context, list *FavoriteList) error
	GetListByID(ctx context.Context, listID string) (*FavoriteList, error)
	GetListByName(ctx context.Context, userID, name string) (*FavoriteList, error)
	ListUserLists(ctx context.Context, userID string, skip, limit int64) ([]*FavoriteList, int64, error)
	UpdateList(ctx context.Context, list *FavoriteList) error
	DeleteList(ctx context.Context, listID, userID string) error
	GetListNovelCount(ctx context.Context, listID string) (int64, error)

	// List item operations
	AddNovelToList(ctx context.Context, item *FavoriteListItem) error
	RemoveNovelFromList(ctx context.Context, listID, novelID string) error
	IsNovelInList(ctx context.Context, listID, novelID string) (bool, error)
	ListNovelInList(ctx context.Context, listID string, skip, limit int64) ([]*FavoriteListItem, int64, error)
	UpdateItemNotes(ctx context.Context, listID, novelID, notes string) error
	MoveNovelToList(ctx context.Context, fromListID, toListID, novelID string) error
	UpdateItemSortOrder(ctx context.Context, listID, novelID string, sortOrder int) error

	// Cross-list operations
	GetUserNovelLists(ctx context.Context, userID, novelID string) ([]*FavoriteList, error)
	GetPublicListNovels(ctx context.Context, listID string, skip, limit int64) ([]*Novel, int64, error)
}

// NovelReadingProgressRepository defines the data access operations for reading progress.
type NovelReadingProgressRepository interface {
	// SaveProgress saves or updates a user's reading progress for a novel.
	SaveProgress(ctx context.Context, progress *ReadingProgress) error

	// GetProgress gets a user's reading progress for a specific novel.
	GetProgress(ctx context.Context, novelID, userID primitive.ObjectID) (*ReadingProgress, error)

	// GetProgressForNovels gets reading progress for multiple novels.
	GetProgressForNovels(ctx context.Context, novelIDs []primitive.ObjectID, userID primitive.ObjectID) (map[string]*ReadingProgress, error)

	// DeleteProgress deletes a user's reading progress for a novel.
	DeleteProgress(ctx context.Context, novelID, userID primitive.ObjectID) error
}

// NovelViewingHistoryRepository defines the data access operations for viewing history.
type NovelViewingHistoryRepository interface {
	// TrackView records a view in the history.
	TrackView(ctx context.Context, history *ViewingHistory) error

	// GetUserHistory retrieves a user's viewing history.
	GetUserHistory(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*ViewingHistory, int64, error)

	// GetRecentHistory retrieves recent viewing history.
	GetRecentHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*ViewingHistory, error)

	// DeleteHistoryItem deletes a specific history item.
	DeleteHistoryItem(ctx context.Context, id, userID primitive.ObjectID) error

	// DeleteHistoryByNovel deletes all history for a specific novel.
	DeleteHistoryByNovel(ctx context.Context, userID, novelID primitive.ObjectID) error

	// CleanOldHistory deletes old history entries.
	CleanOldHistory(ctx context.Context, userID primitive.ObjectID, days int) (int64, error)
}
