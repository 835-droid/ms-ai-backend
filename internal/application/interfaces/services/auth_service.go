// Package services defines the service interfaces (ports) for application logic.
package services

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines the authentication and authorization operations.
type AuthService interface {
	// Register creates a new user account.
	Register(ctx context.Context, email, password, username string, inviteCode *string) (*user.User, error)

	// Login authenticates a user and returns a token.
	Login(ctx context.Context, email, password string) (token string, user *user.User, err error)

	// Logout invalidates a user's token.
	Logout(ctx context.Context, token string) error

	// ValidateToken validates a user's token and returns the user ID.
	ValidateToken(ctx context.Context, token string) (primitive.ObjectID, error)

	// RefreshToken generates a new token for an existing session.
	RefreshToken(ctx context.Context, token string) (string, error)

	// ChangePassword changes a user's password.
	ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error

	// GenerateInviteCode creates a new invite code.
	GenerateInviteCode(ctx context.Context, createdBy primitive.ObjectID) (*user.InviteCode, error)

	// ValidateInviteCode checks if an invite code is valid.
	ValidateInviteCode(ctx context.Context, code string) (bool, error)
}
