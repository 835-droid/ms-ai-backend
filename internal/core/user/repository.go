package user

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines user-related storage operations
type Repository interface {
	// Auth & User Management
	Create(ctx context.Context, user *User, details *UserDetails) error    // تعديل لتمرير التفاصيل أيضاً
	GetNextSequence(ctx context.Context, sequenceName string) (int, error) // إضافة الدالة هنا
	//Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, page, limit int) ([]*User, error)

	// Token Management
	UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error
	InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*User, error)

	// Invite Code Management
	FindCode(ctx context.Context, code string) (*InviteCode, error)
	UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error
	CreateInvite(ctx context.Context, invite *InviteCode) error
	FindAllInvites(ctx context.Context, page, limit int) ([]*InviteCode, error)
	DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error
}
