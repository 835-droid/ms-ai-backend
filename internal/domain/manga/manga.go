// Package manga defines the core domain entities for manga content.
package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// RankedManga represents a manga with its ranking information for a specific period.
type RankedManga struct {
	Manga     *Manga `json:"manga"`
	ViewCount int64  `json:"view_count"`
}
