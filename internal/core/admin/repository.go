package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines the interface for admin data operations
type Repository interface {
	CreateInvite(ctx context.Context, invite *InviteCode) error
	ListInvites(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error)
	DeleteInvite(ctx context.Context, id string) error
}

// MongoRepository implements Repository using MongoDB
type MongoRepository struct {
	db         *mongo.Database
	inviteColl *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		db:         db,
		inviteColl: db.Collection("invite_codes"),
	}
}

func (r *MongoRepository) CreateInvite(ctx context.Context, invite *InviteCode) error {
	if invite.ID.IsZero() {
		invite.ID = primitive.NewObjectID()
	}
	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now()
	}
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = invite.CreatedAt.Add(24 * time.Hour)
	}

	_, err := r.inviteColl.InsertOne(ctx, invite)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("invite code already exists: %w", err)
		}
		return fmt.Errorf("failed to create invite code: %w", err)
	}
	return nil
}

func (r *MongoRepository) ListInvites(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.inviteColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find invite codes: %w", err)
	}
	defer cursor.Close(ctx)

	var invites []*InviteCode
	if err := cursor.All(ctx, &invites); err != nil {
		return nil, 0, fmt.Errorf("failed to decode invite codes: %w", err)
	}

	total, err := r.inviteColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invite codes: %w", err)
	}

	return invites, total, nil
}

func (r *MongoRepository) DeleteInvite(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	result, err := r.inviteColl.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("failed to delete invite code: %w", err)
	}
	if result.DeletedCount == 0 {
		return errors.New("invite code not found")
	}
	return nil
}
