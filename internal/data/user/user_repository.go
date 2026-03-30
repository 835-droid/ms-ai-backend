// internal/data/user_repository.go
package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	coreAdmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"

	// تم التصحيح: إضافة استيراد حزمة mongo
	mongoData "github.com/835-droid/ms-ai-backend/internal/data/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements core.UserRepository backed by MongoDB.
type MongoUserRepository struct {
	// تم التصحيح: استخدام النوع المصدر من الحزمة
	store       *mongoData.MongoStore // كان: *MongoStore
	usersColl   *mongo.Collection
	invitesColl *mongo.Collection
}
type Counter struct {
	ID    string `bson:"_id"` // اسم العداد مثلا "user_id_counter"
	Value int    `bson:"seq"` // القيمة الحالية
}

// ensureInitialized validates that the repository and required collections are available
func (r *MongoUserRepository) ensureInitialized() error {
	if r == nil || r.store == nil || r.usersColl == nil || r.invitesColl == nil {
		return fmt.Errorf("mongo user repository not initialized")
	}
	return nil
}

// تم التصحيح: استخدام النوع المصدر في دالة الإنشاء
func NewMongoUserRepository(s *mongoData.MongoStore) *MongoUserRepository {
	return &MongoUserRepository{
		store:       s,
		usersColl:   s.GetCollection("users"),
		invitesColl: s.GetCollection("invite_codes"),
	}
}

func (r *MongoUserRepository) CreateUser(ctx context.Context, user *coreUser.User) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true
	if user.Roles == nil {
		user.Roles = []string{"user"}
	}

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Check if username exists
		count, err := r.usersColl.CountDocuments(sessCtx, bson.M{"username": user.Username})
		if err != nil {
			return fmt.Errorf("check username: %w", err)
		}
		if count > 0 {
			return core.ErrUserExists
		}

		// Insert the user
		res, err := r.usersColl.InsertOne(sessCtx, user)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return core.ErrUserExists
			}
			return fmt.Errorf("insert user: %w", err)
		}
		user.ID = res.InsertedID.(primitive.ObjectID)
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("create user failed", map[string]interface{}{
			"username": user.Username,
			"error":    err.Error(),
		})
		return err
	}

	r.store.Log.Info("user created", map[string]interface{}{
		"username": user.Username,
		"id":       user.ID.Hex(),
	})
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

	// Perform case-insensitive match to avoid login issues due to character casing.
	// We normalize the username in service layer, but existing DB values may differ.
	escaped := regexp.QuoteMeta(username)
	filter := bson.M{
		"$and": []bson.M{
			{
				"username": bson.M{"$regex": fmt.Sprintf("^%s$", escaped), "$options": "i"},
			},
			{
				"$or": []bson.M{
					{"is_active": true},
					{"is_active": bson.M{"$exists": false}},
				},
			},
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

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Replace user atomically; filter includes is_active so absent/inactive users are not matched.
		result, err := r.usersColl.ReplaceOne(sessCtx, bson.M{"_id": user.ID, "is_active": true}, user)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		// MatchedCount==0 indicates the document was not found (absent or inactive).
		if result.MatchedCount == 0 {
			return core.ErrUserNotFound
		}
		// ModifiedCount can be 0 when the content is identical; don't treat as not found.
		return nil
	}, nil)

	if err != nil {
		r.store.Log.Error("update user failed", map[string]interface{}{
			"id":    user.ID.Hex(),
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("user updated", map[string]interface{}{
		"id": user.ID.Hex(),
	})
	return nil
}

func (r *MongoUserRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "write")
	defer cancel()

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Soft delete: set is_active to false
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

	r.store.Log.Info("user deleted (soft)", map[string]interface{}{
		"id": id.Hex(),
	})
	return nil
}

func (r *MongoUserRepository) ListUsers(ctx context.Context, skip, limit int64) ([]*coreUser.User, int64, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, 0, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "users", "read")
	defer cancel()

	if limit == 0 {
		// تم التصحيح: استخدام الثابت المصدر
		limit = mongoData.DefaultLimit // كان: defaultLimit
	}
	if limit > 100 {
		limit = 100 // Cap limit
	}

	// First, get the total count
	total, err := r.usersColl.CountDocuments(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1}) // Sort by newest first

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

// Delete adapts DeleteUser to match core.Repository interface name.
func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.DeleteUser(ctx, id)
}

// Update adapts UpdateUser to match core.Repository interface name.
func (r *MongoUserRepository) Update(ctx context.Context, user *coreUser.User) error {
	return r.UpdateUser(ctx, user)
}

// FindAll adapts ListUsers to match core.Repository interface name.
func (r *MongoUserRepository) FindAll(ctx context.Context, page, limit int) ([]*coreUser.User, error) {
	// Convert page to skip
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
			r.store.Log.Debug("refresh token not found or expired", map[string]interface{}{
				"token": refreshToken,
			})
			return nil, core.ErrInvalidToken
		}
		r.store.Log.Error("find by refresh token failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("find by refresh token: %w", err)
	}

	return &user, nil
}

// FindCode returns an invite code by its string code value.
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

// UseCode marks an invite code as used by a user (by object id).
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

// ---- Implementations of coreAdmin.InviteRepository ----

func (r *MongoUserRepository) CreateInviteCode(ctx context.Context, invite *coreUser.InviteCode) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	now := time.Now()
	invite.CreatedAt = now
	invite.ExpiresAt = now.Add(30 * 24 * time.Hour) // Default 30 days expiry
	invite.IsUsed = false

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
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

	if err != nil {
		r.store.Log.Error("create invite code failed", map[string]interface{}{
			"code":  invite.Code,
			"error": err.Error(),
		})
		return err
	}

	r.store.Log.Info("invite code created", map[string]interface{}{
		"code": invite.Code,
		"id":   invite.ID.Hex(),
	})
	return nil
}

func (r *MongoUserRepository) UseInviteCode(ctx context.Context, code string) (*coreUser.InviteCode, error) {
	if err := r.ensureInitialized(); err != nil {
		return nil, err
	}
	ctx, cancel := r.store.WithCollectionTimeout(ctx, "invite_codes", "write")
	defer cancel()

	var invite *coreUser.InviteCode

	err := r.store.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// 1. Find the invite code
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

		// 2. Mark it as used
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
			// This should not happen if FindOne succeeded, but good for safety
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
		limit = 100 // Cap limit
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

// ---- Adapter methods to satisfy admin.Repository ----

// CreateInvite implements user.Repository.CreateInvite
func (r *MongoUserRepository) CreateInvite(ctx context.Context, invite *coreUser.InviteCode) error {
	return r.CreateInviteCode(ctx, invite)
}

// CreateAdminInvite implements admin.Repository.CreateInvite
func (r *MongoUserRepository) CreateAdminInvite(ctx context.Context, invite *coreAdmin.InviteCode) error {
	// adapt admin.InviteCode -> user.InviteCode
	u := &coreUser.InviteCode{
		ID:        invite.ID,
		Code:      invite.Code,
		CreatedAt: invite.CreatedAt,
		ExpiresAt: invite.ExpiresAt,
		IsUsed:    invite.IsUsed,
		UsedBy:    invite.UsedBy,
	}
	return r.CreateInviteCode(ctx, u)
}

// adminRepoAdapter implements admin.Repository
type adminRepoAdapter struct {
	r *MongoUserRepository
}

// AsAdminRepository returns an adapter that implements admin.Repository
func (r *MongoUserRepository) AsAdminRepository() coreAdmin.Repository {
	return &adminRepoAdapter{r: r}
}

// CreateInvite implements admin.Repository.CreateInvite
func (a *adminRepoAdapter) CreateInvite(ctx context.Context, invite *coreAdmin.InviteCode) error {
	// Adapt admin.InviteCode to user.InviteCode
	u := &coreUser.InviteCode{
		ID:        invite.ID,
		Code:      invite.Code,
		CreatedAt: invite.CreatedAt,
		ExpiresAt: invite.ExpiresAt,
		IsUsed:    invite.IsUsed,
		UsedBy:    invite.UsedBy,
	}
	return a.r.CreateInviteCode(ctx, u)
}

// ListInvites implements admin.Repository.ListInvites
func (a *adminRepoAdapter) ListInvites(ctx context.Context, skip, limit int64) ([]*coreAdmin.InviteCode, int64, error) {
	invites, total, err := a.r.ListInviteCodes(ctx, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	var out []*coreAdmin.InviteCode
	for _, v := range invites {
		out = append(out, &coreAdmin.InviteCode{
			ID:        v.ID,
			Code:      v.Code,
			CreatedAt: v.CreatedAt,
			ExpiresAt: v.ExpiresAt,
			IsUsed:    v.IsUsed,
			UsedBy:    v.UsedBy,
		})
	}
	return out, total, nil
}

// DeleteInvite implements admin.Repository.DeleteInvite
func (a *adminRepoAdapter) DeleteInvite(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid invite code id")
	}
	return a.r.DeleteInviteCode(ctx, oid)
}

// DeleteInvite adapts to admin.Repository.DeleteInvite

// DeleteInvite implements Repository.DeleteInvite
func (r *MongoUserRepository) DeleteInvite(ctx context.Context, id primitive.ObjectID) error {
	return r.DeleteInviteCode(ctx, id)
}

// FindAllInvites implements user.Repository.FindAllInvites
func (r *MongoUserRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*coreUser.InviteCode, error) {
	if page < 1 {
		page = 1
	}
	skip := int64((page - 1) * limit)
	invites, _, err := r.ListInviteCodes(ctx, skip, int64(limit))
	if err != nil {
		return nil, err
	}
	return invites, nil
}

func (r *MongoUserRepository) Create(ctx context.Context, u *coreUser.User, details *coreUser.UserDetails) error {
	if err := r.ensureInitialized(); err != nil {
		return err
	}

	// استخدام الـ Transaction لضمان حفظ الجدولين معاً أو لا شيء
	return mongoData.ExecuteTransaction(ctx, r.store, func(sessCtx mongo.SessionContext) error {
		// 1. حفظ المستخدم الأساسي
		if _, err := r.usersColl.InsertOne(sessCtx, u); err != nil {
			return err
		}

		// 2. حفظ تفاصيل المستخدم في الـ Collection الثانية
		detailsColl := r.usersColl.Database().Collection(coreUser.UserDetailsCollectionName)
		if _, err := detailsColl.InsertOne(sessCtx, details); err != nil {
			return err
		}

		return nil
	}, nil)
}

// دالة العداد التصاعدي مع إصلاح مرجع الـ Struct
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
