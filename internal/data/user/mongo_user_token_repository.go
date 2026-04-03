// internal/data/user/user_token_repository.go
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoUserRepository) UpdateRefreshToken(ctx context.Context, id primitive.ObjectID, refreshToken string, expiresAt primitive.DateTime) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"refresh_token":            refreshToken,
			"refresh_token_expires_at": expiresAt.Time(),
			"updated_at":               time.Now(),
		},
	}

	result, err := r.usersColl.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (r *MongoUserRepository) InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	update := bson.M{
		"$unset": bson.M{
			"refresh_token":            "",
			"refresh_token_expires_at": "",
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := r.usersColl.UpdateByID(ctx, userID, update)
	if err != nil {
		return fmt.Errorf("invalidate refresh token: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (r *MongoUserRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*coreUser.User, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	var user coreUser.User
	filter := bson.M{
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": bson.M{"$gt": time.Now()},
		"is_active":                true,
	}

	err := r.usersColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrInvalidToken
		}
		return nil, fmt.Errorf("find by refresh token: %w", err)
	}
	return &user, nil
}
