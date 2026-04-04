// Package repositories defines the repository interfaces (ports) for data access.
package repositories

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the data access operations for users.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *user.User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id primitive.ObjectID) (*user.User, error)

	// GetByEmail retrieves a user by their email.
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *user.User) error

	// Delete deletes a user by their ID.
	Delete(ctx context.Context, id primitive.ObjectID) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists.
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// List retrieves a paginated list of users.
	List(ctx context.Context, skip, limit int64) ([]*user.User, int64, error)
}

// InviteCodeRepository defines the data access operations for invite codes.
type InviteCodeRepository interface {
	// Create creates a new invite code.
	Create(ctx context.Context, code *user.InviteCode) error

	// GetByCode retrieves an invite code by its code string.
	GetByCode(ctx context.Context, code string) (*user.InviteCode, error)

	// Use marks an invite code as used.
	Use(ctx context.Context, code string, usedBy string) error

	// Validate checks if an invite code is valid and active.
	Validate(ctx context.Context, code string) (bool, error)
}

// UserTokenRepository defines the data access operations for user tokens.
type UserTokenRepository interface {
	// Create creates a new user token.
	Create(ctx context.Context, userID primitive.ObjectID, token string, expiresAt int64) error

	// GetByToken retrieves a user token by its token string.
	GetByToken(ctx context.Context, token string) (string, error)

	// Delete removes a user token.
	Delete(ctx context.Context, token string) error

	// DeleteExpired removes all expired tokens.
	DeleteExpired(ctx context.Context) error
}
