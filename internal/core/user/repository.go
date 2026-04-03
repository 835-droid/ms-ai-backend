// ----- START OF FILE: backend/MS-AI/internal/core/user/repository.go -----
package user

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	Create(ctx context.Context, user *User, details *UserDetails) error
	CreateUserWithInvite(ctx context.Context, user *User, details *UserDetails, inviteCode string) error // جديد
	GetNextSequence(ctx context.Context, sequenceName string) (int, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, page, limit int) ([]*User, error)

	UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error
	InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*User, error)

	FindCode(ctx context.Context, code string) (*InviteCode, error)
	UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error
	CreateInvite(ctx context.Context, invite *InviteCode) error
	FindAllInvites(ctx context.Context, page, limit int) ([]*InviteCode, error)
	FindAllInvitesWithTotal(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error)
	DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error

	FindAllUsers(ctx context.Context, skip, limit int64) ([]*User, int64, error)
	UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role Role, add bool) error
}

// ----- END OF FILE: backend/MS-AI/internal/core/user/repository.go -----
