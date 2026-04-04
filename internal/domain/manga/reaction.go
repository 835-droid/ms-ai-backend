package manga

import (
	"fmt"

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

// ReactionTypeFromString converts a string to a ReactionType.
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

// MangaReaction represents a user's reaction to a manga.
type MangaReaction struct {
	MangaID   primitive.ObjectID `bson:"manga_id" json:"manga_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      ReactionType       `bson:"type" json:"type"`
	CreatedAt string             `bson:"created_at" json:"created_at"`
}
