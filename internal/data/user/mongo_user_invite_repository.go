// internal/data/user/user_invite_repository.go
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"
	mongoData "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---- Invite Code Management ----

func (r *MongoUserRepository) FindCode(ctx context.Context, code string) (*coreUser.InviteCode, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "read")
	defer cancel()

	var invite coreUser.InviteCode
	err := r.invitesColl.FindOne(ctx, bson.M{"code": code}).Decode(&invite)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("find invite code: %w", err)
	}
	return &invite, nil
}

func (r *MongoUserRepository) UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_used":    true,
			"used_at":    time.Now(),
			"used_by":    userID,
			"updated_at": time.Now(),
		},
	}
	result, err := r.invitesColl.UpdateByID(ctx, codeID, update)
	if err != nil {
		return fmt.Errorf("use invite code: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrInviteCodeNotFound
	}
	return nil
}

// CreateInviteCode creates a new invite code.
func (r *MongoUserRepository) CreateInviteCode(ctx context.Context, invite *coreUser.InviteCode) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	now := time.Now()
	invite.CreatedAt = now
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = now.Add(30 * 24 * time.Hour)
	}
	invite.IsUsed = false

	// تحقق مما إذا كان MongoDB يدعم المعاملات (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	var err error
	if isReplicaSet {
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			res, err := r.invitesColl.InsertOne(sessCtx, invite)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return core.ErrInviteCodeExists
				}
				return fmt.Errorf("insert invite code: %w", err)
			}
			invite.ID = res.InsertedID.(primitive.ObjectID)
			return nil
		}, nil)
	} else {
		// في حالة standalone، نقوم بالإدراج بدون معاملة
		res, err := r.invitesColl.InsertOne(ctx, invite)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return core.ErrInviteCodeExists
			}
			return fmt.Errorf("insert invite code: %w", err)
		}
		invite.ID = res.InsertedID.(primitive.ObjectID)
	}

	if err != nil {
		r.store.Log.Error("create invite code failed", map[string]interface{}{
			"code":  invite.Code,
			"error": err.Error(),
		})
		return err
	}
	return nil
}

// CreateInvite is an alias for CreateInviteCode (to satisfy coreUser.Repository)
func (r *MongoUserRepository) CreateInvite(ctx context.Context, invite *coreUser.InviteCode) error {
	return r.CreateInviteCode(ctx, invite)
}

func (r *MongoUserRepository) UseInviteCode(ctx context.Context, code string) (*coreUser.InviteCode, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	var invite *coreUser.InviteCode

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		filter := bson.M{
			"code":       code,
			"is_used":    false,
			"expires_at": bson.M{"$gt": time.Now()},
		}
		if err := r.invitesColl.FindOne(sessCtx, filter).Decode(&invite); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return core.ErrInvalidInviteCode
			}
			return fmt.Errorf("find invite code: %w", err)
		}

		update := bson.M{
			"$set": bson.M{
				"is_used":    true,
				"used_at":    time.Now(),
				"updated_at": time.Now(),
			},
		}
		result, err := r.invitesColl.UpdateByID(sessCtx, invite.ID, update)
		if err != nil {
			return fmt.Errorf("update invite code: %w", err)
		}
		if result.MatchedCount == 0 {
			return errors.New("invite code disappeared during transaction")
		}
		invite.IsUsed = true
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("use invite code failed", map[string]interface{}{
			"code":  code,
			"error": err.Error(),
		})
		return nil, err
	}
	return invite, nil
}

func (r *MongoUserRepository) ListInviteCodes(ctx context.Context, skip, limit int64) ([]*coreUser.InviteCode, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "read")
	defer cancel()

	if limit == 0 {
		limit = mongoData.DefaultLimit
	}
	if limit > 100 {
		limit = 100
	}

	total, err := r.invitesColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("count invites: %w", err)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.invitesColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find invites: %w", err)
	}
	defer cursor.Close(ctx)

	var invites []*coreUser.InviteCode
	if err := cursor.All(ctx, &invites); err != nil {
		return nil, 0, fmt.Errorf("decode invites: %w", err)
	}
	return invites, total, nil
}

func (r *MongoUserRepository) DeleteInviteCode(ctx context.Context, codeID primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	result, err := r.invitesColl.DeleteOne(ctx, bson.M{"_id": codeID})
	if err != nil {
		return fmt.Errorf("delete invite code: %w", err)
	}
	if result.DeletedCount == 0 {
		return core.ErrInviteCodeNotFound
	}
	return nil
}

// DeleteInvite implements coreuser.Repository.DeleteInvite
func (r *MongoUserRepository) DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error {
	return r.DeleteInviteCode(ctx, codeID)
}

func (r *MongoUserRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*coreUser.InviteCode, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	invites, _, err := r.ListInviteCodes(ctx, skip, int64(limit))
	return invites, err
}

func (r *MongoUserRepository) FindAllInvitesWithTotal(ctx context.Context, skip, limit int64) ([]*coreUser.InviteCode, int64, error) {
	if skip < 0 {
		skip = 0
	}
	if limit <= 0 {
		limit = 20
	}
	invites, total, err := r.ListInviteCodes(ctx, skip, limit)
	return invites, total, err
}

// helper functions for session
func (r *MongoUserRepository) findCodeInSession(sessCtx mongo.SessionContext, code string) (*coreUser.InviteCode, error) {
	var inv coreUser.InviteCode
	err := r.invitesColl.FindOne(sessCtx, bson.M{"code": code}).Decode(&inv)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("find invite: %w", err)
	}
	return &inv, nil
}

func (r *MongoUserRepository) markInviteUsedInSession(sessCtx mongo.SessionContext, inviteID, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_used": true,
			"used_by": userID,
			"used_at": time.Now(),
		},
	}
	result, err := r.invitesColl.UpdateByID(sessCtx, inviteID, update)
	if err != nil {
		return fmt.Errorf("mark invite used: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrInviteCodeNotFound
	}
	return nil
}
