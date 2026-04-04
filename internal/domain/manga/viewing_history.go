package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadingProgress represents a user's reading progress for a manga.
type ReadingProgress struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID         primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	LastReadChapter primitive.ObjectID `bson:"last_read_chapter" json:"last_read_chapter"`
	LastReadPage    int                `bson:"last_read_page" json:"last_read_page"`
	LastReadAt      time.Time          `bson:"last_read_at" json:"last_read_at"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// ViewingHistory represents a user's viewing history for a manga.
type ViewingHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	ChapterID primitive.ObjectID `bson:"chapter_id,omitempty" json:"chapter_id,omitempty"`
	Page      int                `bson:"page" json:"page"`
	ViewedAt  time.Time          `bson:"viewed_at" json:"viewed_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
