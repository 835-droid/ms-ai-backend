package novel

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NovelChapter represents a chapter in a novel.
type NovelChapter struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NovelID       primitive.ObjectID `bson:"novel_id" json:"novel_id"`
	Title         string             `bson:"title" json:"title" validate:"required,min=1,max=256"`
	Content       string             `bson:"content" json:"content" validate:"required,min=1"`
	Number        int                `bson:"number" json:"number" validate:"required,min=1"`
	WordCount     int64              `bson:"word_count" json:"word_count"`
	ViewsCount    int64              `bson:"views_count" json:"views_count"`
	RatingSum     float64            `bson:"rating_sum" json:"rating_sum"`
	RatingCount   int64              `bson:"rating_count" json:"rating_count"`
	AverageRating float64            `bson:"average_rating" json:"average_rating"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	HasUserViewed *bool              `json:"has_user_viewed,omitempty"`
}
