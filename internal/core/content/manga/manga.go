// ----- START OF FILE: backend/MS-AI/internal/core/content/manga/manga.go -----
package manga

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionType represents the type of reaction a user can give to a manga.
type ReactionType string

const (
	ReactionUpvote    ReactionType = "upvote"
	ReactionFunny     ReactionType = "funny"
	ReactionLove      ReactionType = "love"
	ReactionSurprised ReactionType = "surprised"
	ReactionAngry     ReactionType = "angry"
	ReactionSad       ReactionType = "sad"
)

// ReactionTypeFromString converts a string to a ReactionType
func ReactionTypeFromString(s string) (ReactionType, error) {
	switch s {
	case "upvote":
		return ReactionUpvote, nil
	case "funny":
		return ReactionFunny, nil
	case "love":
		return ReactionLove, nil
	case "surprised":
		return ReactionSurprised, nil
	case "angry":
		return ReactionAngry, nil
	case "sad":
		return ReactionSad, nil
	default:
		return "", fmt.Errorf("invalid reaction type: %s", s)
	}
}

// Manga represents a manga in the system.
type Manga struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title          string             `bson:"title" json:"title" validate:"required,min=1,max=256"`
	Slug           string             `bson:"slug" json:"slug"`
	Description    string             `bson:"description" json:"description" validate:"required,min=1"`
	AuthorID       primitive.ObjectID `bson:"author_id" json:"author_id"`
	Tags           []string           `bson:"tags" json:"tags"`
	CoverImage     string             `bson:"cover_image" json:"cover_image"`
	IsPublished    bool               `bson:"is_published" json:"is_published"`
	PublishedAt    *time.Time         `bson:"published_at,omitempty" json:"published_at,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	ViewsCount     int64              `bson:"views_count" json:"views_count"`
	LikesCount     int64              `bson:"likes_count" json:"likes_count"`
	FavoritesCount int64              `bson:"favorites_count" json:"favorites_count"`
	ReactionsCount map[string]int64   `bson:"reactions_count" json:"reactions_count"`
	RatingSum      float64            `bson:"rating_sum" json:"rating_sum"`
	RatingCount    int64              `bson:"rating_count" json:"rating_count"`
	AverageRating  float64            `bson:"average_rating" json:"average_rating"`
}

// MangaRating represents a user's rating for a manga.
type MangaRating struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Score     float64            `bson:"score" json:"score" validate:"min=1,max=10"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// MangaReaction represents a user's reaction to a manga.
type MangaReaction struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      ReactionType       `bson:"type" json:"type"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// MangaChapter represents a chapter in a manga.
type MangaChapter struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID       primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	Title         string             `bson:"title" json:"title" validate:"required,min=1,max=256"`
	Pages         []string           `bson:"pages" json:"pages" validate:"required,min=1"`
	Number        int                `bson:"number" json:"number" validate:"required,min=1"`
	ViewsCount    int64              `bson:"views_count" json:"views_count"`
	RatingSum     float64            `bson:"rating_sum" json:"rating_sum"`
	RatingCount   int64              `bson:"rating_count" json:"rating_count"`
	AverageRating float64            `bson:"average_rating" json:"average_rating"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	HasUserViewed *bool              `json:"has_user_viewed,omitempty"`
}

// UserFavorite represents a user's favorite manga.
type UserFavorite struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// ChapterRating represents a user's rating for a chapter (1-10).
type ChapterRating struct {
	ChapterID primitive.ObjectID `bson:"chapter_id" json:"chapter_id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Score     float64            `bson:"score" json:"score" validate:"min=1,max=10"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MangaComment represents a comment on a manga.
type MangaComment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Content   string             `bson:"content" json:"content" validate:"required,min=1,max=1000"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChapterComment represents a comment on a chapter.
type ChapterComment struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChapterID    primitive.ObjectID `bson:"chapter_id" json:"chapter_id"`
	MangaID      primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username     string             `bson:"username" json:"username"`
	UserAvatar   string             `bson:"user_avatar" json:"user_avatar"`
	Content      string             `bson:"content" json:"content" validate:"required,min=1,max=1000"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	LikeCount    int64              `bson:"like_count" json:"like_count"`
	DislikeCount int64              `bson:"dislike_count" json:"dislike_count"`
}

// ChapterCommentReaction represents a like/dislike on a chapter comment
type ChapterCommentReaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommentID primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	ChapterID primitive.ObjectID `bson:"chapter_id" json:"chapter_id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"` // "like" or "dislike"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// RankedManga represents a manga with its ranking information for a specific period.
type RankedManga struct {
	Manga     *Manga `json:"manga"`
	ViewCount int64  `json:"view_count"`
}

// MangaRepository defines the manga data operations.
type MangaRepository interface {
	CreateManga(ctx context.Context, manga *Manga) error
	GetMangaByID(ctx context.Context, id primitive.ObjectID) (*Manga, error)
	GetMangaBySlug(ctx context.Context, slug string) (*Manga, error)
	ListMangas(ctx context.Context, skip, limit int64) ([]*Manga, int64, error)
	UpdateManga(ctx context.Context, manga *Manga) error
	DeleteManga(ctx context.Context, id primitive.ObjectID) error
	IncrementViews(ctx context.Context, mangaID primitive.ObjectID) error
	LogView(ctx context.Context, mangaID primitive.ObjectID) error
	ListMostViewed(ctx context.Context, since time.Time, skip, limit int64) ([]*RankedManga, error)
	ListRecentlyUpdated(ctx context.Context, skip, limit int64) ([]*Manga, error)
	ListMostFollowed(ctx context.Context, skip, limit int64) ([]*Manga, error)
	ListTopRated(ctx context.Context, skip, limit int64) ([]*Manga, error)
	SetReaction(ctx context.Context, mangaID, userID primitive.ObjectID, reactionType ReactionType) (reaction string, err error)
	GetUserReaction(ctx context.Context, mangaID, userID primitive.ObjectID) (string, error)
	ListLikedMangas(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error)
	AddRating(ctx context.Context, rating *MangaRating) (newAverage float64, err error)
	HasUserRated(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)

	// Favorites
	AddFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	RemoveFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) error
	IsFavorite(ctx context.Context, mangaID, userID primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, skip, limit int64) ([]*Manga, int64, error)

	// Manga Comments
	AddMangaComment(ctx context.Context, comment *MangaComment) error
	ListMangaComments(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*MangaComment, int64, error)
	DeleteMangaComment(ctx context.Context, commentID, userID primitive.ObjectID) error
}

// MangaChapterRepository defines the manga chapter data operations.
type MangaChapterRepository interface {
	CreateMangaChapter(ctx context.Context, chapter *MangaChapter) error
	GetMangaChapterByID(ctx context.Context, id primitive.ObjectID) (*MangaChapter, error)
	ListMangaChaptersByManga(ctx context.Context, mangaID primitive.ObjectID, skip, limit int64) ([]*MangaChapter, int64, error)
	UpdateMangaChapter(ctx context.Context, chapter *MangaChapter) error
	DeleteMangaChapter(ctx context.Context, id primitive.ObjectID) error

	// Chapter Views
	IncrementChapterViews(ctx context.Context, chapterID, mangaID primitive.ObjectID) error
	TrackChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) error
	HasUserViewedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	TrackAndIncrementChapterView(ctx context.Context, chapterID, mangaID, userID primitive.ObjectID) (bool, error)

	// Chapter Ratings
	AddChapterRating(ctx context.Context, rating *ChapterRating) (newAverage float64, count int64, userScore float64, err error)
	HasUserRatedChapter(ctx context.Context, chapterID, userID primitive.ObjectID) (bool, error)
	GetUserChapterRating(ctx context.Context, chapterID, userID primitive.ObjectID) (float64, bool, error)

	// Chapter Comments
	AddChapterComment(ctx context.Context, comment *ChapterComment) error
	ListChapterComments(ctx context.Context, chapterID primitive.ObjectID, skip, limit int64, sortOrder string) ([]*ChapterComment, int64, error)
	DeleteChapterComment(ctx context.Context, commentID, userID primitive.ObjectID) error

	// Chapter Comment Reactions (like/dislike)
	AddChapterCommentReaction(ctx context.Context, reaction *ChapterCommentReaction) error
	RemoveChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) error
	GetUserChapterCommentReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error)
}

// ----- END OF FILE: backend/MS-AI/internal/core/content/manga/manga.go -----
