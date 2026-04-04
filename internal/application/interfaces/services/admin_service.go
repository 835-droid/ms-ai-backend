package services

import (
	"context"

	"github.com/835-droid/ms-ai-backend/internal/domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminService defines the admin operations.
type AdminService interface {
	// CreateAdmin creates a new admin user.
	CreateAdmin(ctx context.Context, email, password, username string) (*user.User, error)

	// ListUsers retrieves a paginated list of users.
	ListUsers(ctx context.Context, skip, limit int64) ([]*user.User, int64, error)

	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, id primitive.ObjectID) (*user.User, error)

	// UpdateUser updates a user's information.
	UpdateUser(ctx context.Context, user *user.User) (*user.User, error)

	// DeleteUser deletes a user by ID.
	DeleteUser(ctx context.Context, id primitive.ObjectID) error

	// BanUser bans a user account.
	BanUser(ctx context.Context, id primitive.ObjectID, reason string) error

	// UnbanUser unbans a user account.
	UnbanUser(ctx context.Context, id primitive.ObjectID) error

	// MakeAdmin promotes a user to admin role.
	MakeAdmin(ctx context.Context, id primitive.ObjectID) error

	// RemoveAdmin demotes an admin to regular user.
	RemoveAdmin(ctx context.Context, id primitive.ObjectID) error
}
