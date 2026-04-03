package user

import (
	"context"
	"fmt"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *PostgresUserRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role user.Role, add bool) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Get current user
	u, err := r.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return core.ErrUserNotFound
	}

	// Update roles
	if add {
		u.Roles = u.Roles.Add(role)
	} else {
		u.Roles = u.Roles.Remove(role)
	}

	// Update user
	err = r.Update(ctx, u)
	if err != nil {
		return err
	}

	return tx.Commit()
}
