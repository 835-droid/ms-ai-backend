package novel

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserFavorite represents a user's favorite novel.
type UserFavorite struct {
	NovelID   primitive.ObjectID `bson:"novel_id" json:"novel_id"`
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
	NovelCount  int64     `json:"novel_count,omitempty"` // Computed field
}

// FavoriteListItem represents a novel in a favorite list.
type FavoriteListItem struct {
	ListID    string    `json:"list_id"`
	NovelID   string    `json:"novel_id"`
	Notes     string    `json:"notes,omitempty"`
	AddedAt   time.Time `json:"added_at"`
	SortOrder int       `json:"sort_order"`
	Novel     *Novel    `json:"novel,omitempty"` // Computed field when joining
}
