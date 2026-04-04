package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaRating represents a user's rating for a manga.
type MangaRating struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Score     float64            `bson:"score" json:"score" validate:"min=1,max=10"`
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
