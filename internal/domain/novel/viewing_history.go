package novel

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadingProgress represents a user's reading progress for a novel.
type ReadingProgress struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NovelID         primitive.ObjectID `bson:"novel_id" json:"novel_id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	LastReadChapter primitive.ObjectID `bson:"last_read_chapter" json:"last_read_chapter"`
	LastReadAt      time.Time          `bson:"last_read_at" json:"last_read_at"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// ViewingHistory represents a user's viewing history for a novel.
type ViewingHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	NovelID   primitive.ObjectID `bson:"novel_id" json:"novel_id"`
	ChapterID primitive.ObjectID `bson:"chapter_id,omitempty" json:"chapter_id,omitempty"`
	ViewedAt  time.Time          `bson:"viewed_at" json:"viewed_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
