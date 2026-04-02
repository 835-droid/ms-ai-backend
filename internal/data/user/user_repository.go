// ----- START OF FILE: backend/MS-AI/internal/data/user/user_repository.go -----
// internal/data/user/user_repository.go
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"
	mongoData "github.com/835-droid/ms-ai-backend/internal/data/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements core.UserRepository backed by MongoDB.
type MongoUserRepository struct {
	store       *mongoData.MongoStore
	usersColl   *mongo.Collection
	invitesColl *mongo.Collection
}

// ensureInitialized validates that the repository and required collections are available
func (r *MongoUserRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.usersColl == nil || r.invitesColl == nil {
		return fmt.Errorf("mongo user repository not initialized")
	}
	return nil
}

// NewMongoUserRepository creates a new repository instance
func NewMongoUserRepository(s *mongoData.MongoStore) *MongoUserRepository {
	return &MongoUserRepository{
		store:       s,
		usersColl:   s.GetCollection("users"),
		invitesColl: s.GetCollection("invite_codes"),
	}
}

// Create creates a new user and its associated details in a transaction.
func (r *MongoUserRepository) Create(ctx context.Context, u *coreUser.User, details *coreUser.UserDetails) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsActive = true
	if u.Roles == nil {
		u.Roles = coreUser.FromStrings([]string{"user"})
	}

	// Copy relevant fields to details
	details.UserBase = u.UserBase
	details.UUID = u.UUID
	details.UserID = u.UserID

	var err error

	// Check if MongoDB supports transactions (replica set)
	isReplicaSet := r.store.IsReplicaSet(ctx)

	if isReplicaSet {
		// Use transaction for replica sets
		err = r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
			return r.createUserInSession(sessCtx, u, details)
		}, nil)
	} else {
		// For standalone MongoDB, perform operations without transaction
		err = r.createUserWithoutTransaction(ctx, u, details)
	}

	if err != nil {
		r.store.Log.Error("create user failed", map[string]interface{}{
			"username": u.Username,
			"error":    err.Error(),
		})
		return err
	}

	r.store.Log.Info("user created", map[string]interface{}{
		"username":    u.Username,
		"id":          u.ID.Hex(),
		"replica_set": isReplicaSet,
	})
	return nil
}

// createUserInSession creates user within a transaction session
func (r *MongoUserRepository) createUserInSession(sessCtx mongo.SessionContext, u *coreUser.User, details *coreUser.UserDetails) error {
	// Check if username already exists
	count, err := r.usersColl.CountDocuments(sessCtx, bson.M{"username": u.Username})
	if err != nil {
		return fmt.Errorf("check username: %w", err)
	}
	if count > 0 {
		return core.ErrUserExists
	}

	// Insert user
	res, err := r.usersColl.InsertOne(sessCtx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return core.ErrUserExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	u.ID = res.InsertedID.(primitive.ObjectID)

	// Insert user details
	detailsColl := r.usersColl.Database().Collection("user_details")
	if _, err := detailsColl.InsertOne(sessCtx, details); err != nil {
		return fmt.Errorf("insert user details: %w", err)
	}

	return nil
}

// createUserWithoutTransaction creates user without transaction (for standalone MongoDB)
func (r *MongoUserRepository) createUserWithoutTransaction(ctx context.Context, u *coreUser.User, details *coreUser.UserDetails) error {
	// Check if username already exists
	count, err := r.usersColl.CountDocuments(ctx, bson.M{"username": u.Username})
	if err != nil {
		return fmt.Errorf("check username: %w", err)
	}
	if count > 0 {
		return core.ErrUserExists
	}

	// Insert user
	res, err := r.usersColl.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return core.ErrUserExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	u.ID = res.InsertedID.(primitive.ObjectID)

	// Insert user details
	detailsColl := r.usersColl.Database().Collection("user_details")
	if _, err := detailsColl.InsertOne(ctx, details); err != nil {
		return fmt.Errorf("insert user details: %w", err)
	}

	return nil
}

// ---- Implementations of coreUser.UserRepository ----

func (r *MongoUserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*coreUser.User, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	var user coreUser.User
	err := r.usersColl.FindOne(ctx, bson.M{"_id": id, "is_active": true}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *MongoUserRepository) FindByUsername(ctx context.Context, username string) (*coreUser.User, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	filter := bson.M{
		"username": username,
		"$or": []bson.M{
			{"is_active": true},
			{"is_active": bson.M{"$exists": false}},
		},
	}

	var user coreUser.User
	err := r.usersColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &user, nil
}

// FindByID adapts GetUserByID to match core.Repository interface name.
func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*coreUser.User, error) {
	return r.GetUserByID(ctx, id)
}

func (r *MongoUserRepository) UpdateUser(ctx context.Context, user *coreUser.User) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	user.UpdatedAt = time.Now()

	result, err := r.usersColl.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{"$set": bson.M{
		"last_login_at":            user.LastLoginAt,
		"updated_at":               user.UpdatedAt,
		"password":                 user.Password,
		"refresh_token":            user.RefreshToken,
		"refresh_token_expires_at": user.RefreshTokenExpiresAt,
		"roles":                    user.Roles,
		"is_active":                user.IsActive,
	}})
	if err != nil {
		r.store.Log.Error("update user failed", map[string]interface{}{
			"username": user.Username,
			"error":    err.Error(),
		})
		return fmt.Errorf("update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (r *MongoUserRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		result, err := r.usersColl.UpdateByID(sessCtx, id, bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}})
		if err != nil {
			return fmt.Errorf("soft delete user: %w", err)
		}
		if result.MatchedCount == 0 {
			return core.ErrUserNotFound
		}
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("delete user failed", map[string]interface{}{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return err
	}
	return nil
}

func (r *MongoUserRepository) ListUsers(ctx context.Context, skip, limit int64) ([]*coreUser.User, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	if limit == 0 {
		limit = mongoData.DefaultLimit
	}
	if limit > 100 {
		limit = 100
	}

	total, err := r.usersColl.CountDocuments(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.usersColl.Find(ctx, bson.M{"is_active": true}, opts)
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

func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.DeleteUser(ctx, id)
}

func (r *MongoUserRepository) Update(ctx context.Context, user *coreUser.User) error {
	return r.UpdateUser(ctx, user)
}

func (r *MongoUserRepository) FindAll(ctx context.Context, page, limit int) ([]*coreUser.User, error) {
	if page < 1 {
		page = 1
	}
	skip := int64((page - 1) * limit)
	users, _, err := r.ListUsers(ctx, skip, int64(limit))
	return users, err
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
	skip := int64((page - 1) * limit)
	invites, _, err := r.ListInviteCodes(ctx, skip, int64(limit))
	return invites, err
}

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
func (r *MongoUserRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role string, add bool) error {
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

// ----- END OF FILE: backend/MS-AI/internal/data/user/user_repository.go -----
