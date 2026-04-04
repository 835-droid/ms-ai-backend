package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// ChapterCommentReaction represents a like/dislike on a chapter comment.
type ChapterCommentReaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommentID primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	ChapterID primitive.ObjectID `bson:"chapter_id" json:"chapter_id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"` // "like" or "dislike"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
