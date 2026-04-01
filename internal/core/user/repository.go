package user

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	Create(ctx context.Context, user *User, details *UserDetails) error
	GetNextSequence(ctx context.Context, sequenceName string) (int, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, page, limit int) ([]*User, error)

	UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error
	InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*User, error)

	// Invite Code Management
	FindCode(ctx context.Context, code string) (*InviteCode, error) // يجب أن يكون الاسم FindCode
	UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error
	CreateInvite(ctx context.Context, invite *InviteCode) error
	FindAllInvites(ctx context.Context, page, limit int) ([]*InviteCode, error)
	DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error

	// Admin User Management
	FindAllUsers(ctx context.Context, skip, limit int64) ([]*User, int64, error)
	UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role string, add bool) error
}
