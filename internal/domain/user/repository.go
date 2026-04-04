// Package user defines the core domain entities and repository interfaces for user management.
package user

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the data access operations for user entities.
// This interface follows the Repository pattern and is part of the domain layer.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)

	// GetByEmail retrieves a user by their email.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *User) error

	// Delete deletes a user by their ID.
	Delete(ctx context.Context, id primitive.ObjectID) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists.
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// List retrieves a paginated list of users.
	List(ctx context.Context, skip, limit int64) ([]*User, int64, error)
}

// UserAdminRepository defines admin-specific user operations.
type UserAdminRepository interface {
	// SetRole sets a user's role.
	SetRole(ctx context.Context, userID primitive.ObjectID, role string) error

	// GetRole retrieves a user's role.
	GetRole(ctx context.Context, userID primitive.ObjectID) (string, error)

	// HasRole checks if a user has a specific role.
	HasRole(ctx context.Context, userID primitive.ObjectID, role string) (bool, error)

	// AddRole adds a role to a user.
	AddRole(ctx context.Context, userID primitive.ObjectID, role string) error

	// RemoveRole removes a role from a user.
	RemoveRole(ctx context.Context, userID primitive.ObjectID, role string) error

	// ListByRole retrieves users with a specific role.
	ListByRole(ctx context.Context, role string, skip, limit int64) ([]*User, int64, error)
}

// InviteCodeRepository defines the data access operations for invite codes.
type InviteCodeRepository interface {
	// Create creates a new invite code.
	Create(ctx context.Context, code *InviteCode) error

	// GetByCode retrieves an invite code by its code string.
	GetByCode(ctx context.Context, code string) (*InviteCode, error)

	// Use marks an invite code as used.
	Use(ctx context.Context, code string, usedBy string) error

	// Validate checks if an invite code is valid and active.
	Validate(ctx context.Context, code string) (bool, error)

	// ListByUser retrieves all invite codes for a user.
	ListByUser(ctx context.Context, userID string, skip, limit int64) ([]*InviteCode, int64, error)

	// Delete removes an invite code.
	Delete(ctx context.Context, code string) error
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

	// GetByUserID retrieves all tokens for a user.
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]string, error)
}
