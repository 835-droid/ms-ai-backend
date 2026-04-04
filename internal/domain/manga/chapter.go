package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
