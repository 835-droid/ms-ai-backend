// internal/data/user/user_admin_repository.go
package user

import (
	"context"
	"fmt"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetNextSequence returns the next sequence value for a given sequence name.
func (r *MongoUserRepository) GetNextSequence(ctx context.Context, sequenceName string) (int, error) {
	collection := r.usersColl.Database().Collection("counters")
	filter := bson.M{"_id": sequenceName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		ID  string `bson:"_id"`
		Seq int    `bson:"seq"`
	}
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, err
	}
	return counter.Seq, nil
}

// ---- Additional methods for admin user management ----

// FindAllUsers retrieves all users with pagination
func (r *MongoUserRepository) FindAllUsers(ctx context.Context, skip, limit int64) ([]*coreUser.User, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	filter := bson.M{}
	total, err := r.usersColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.usersColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*coreUser.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, fmt.Errorf("decode users: %w", err)
	}
	return users, total, nil
}

// UpdateUserRole adds or removes a role from a user
func (r *MongoUserRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role user.Role, add bool) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	var update bson.M
	if add {
		update = bson.M{"$addToSet": bson.M{"roles": role}}
	} else {
		update = bson.M{"$pull": bson.M{"roles": role}}
	}

	result, err := r.usersColl.UpdateByID(ctx, userID, update)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

//////

// داخل MongoUserRepository

func (r *MongoUserRepository) CreateUserWithInvite(ctx context.Context, user *coreUser.User, details *coreUser.UserDetails, inviteCode string) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	if r.store.IsReplicaSet(ctx) {
		return r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			// Find invite within session
			inv, err := r.findCodeInSession(sessCtx, inviteCode)
			if err != nil {
				return err
			}
			if inv.IsUsed || inv.ExpiresAt.Before(time.Now()) {
				return core.ErrInvalidInviteCode
			}
			// Create user
			if err := r.createUserInSession(sessCtx, user, details); err != nil {
				return err
			}
			// Mark invite used
			return r.markInviteUsedInSession(sessCtx, inv.ID, user.ID)
		}, nil)
	} else {
		// No transaction: do sequentially but not atomic
		inv, err := r.FindCode(ctx, inviteCode)
		if err != nil {
			return err
		}
		if inv.IsUsed || inv.ExpiresAt.Before(time.Now()) {
			return core.ErrInvalidInviteCode
		}
		if err := r.Create(ctx, user, details); err != nil {
			return err
		}
		if err := r.UseCode(ctx, inv.ID, user.ID); err != nil {
			// log but still return error to fail signup
			r.store.Log.Error("failed to mark invite used after user creation", map[string]interface{}{"error": err.Error()})
			return err
		}
		return nil
	}
}
