package manga

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserFavorite represents a user's favorite manga.
type UserFavorite struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// FavoriteList represents a custom favorite list created by a user.
type FavoriteList struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsPublic    bool      `json:"is_public"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MangaCount  int64     `json:"manga_count,omitempty"` // Computed field
}

// FavoriteListItem represents a manga in a favorite list.
type FavoriteListItem struct {
	ListID    string    `json:"list_id"`
	MangaID   string    `json:"manga_id"`
	Notes     string    `json:"notes,omitempty"`
	AddedAt   time.Time `json:"added_at"`
	SortOrder int       `json:"sort_order"`
	Manga     *Manga    `json:"manga,omitempty"` // Computed field when joining
}
