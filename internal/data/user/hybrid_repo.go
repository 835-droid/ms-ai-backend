// ----- START OF FILE: backend/MS-AI/internal/data/user/hybrid_repo.go -----
package user

import (
	"context"
	"errors"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HybridUserRepository struct {
	primary   user.Repository
	secondary user.Repository
	log       *logger.Logger
}

func NewHybridUserRepository(primary, secondary user.Repository, log *logger.Logger) user.Repository {
	return &HybridUserRepository{
		primary:   primary,
		secondary: secondary,
		log:       log,
	}
}

func (r *HybridUserRepository) Create(ctx context.Context, u *user.User, d *user.UserDetails) error {
	if r.primary != nil {
		if err := r.primary.Create(ctx, u, d); err != nil {
			r.log.Error("primary create failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.Create(ctx, u, d); err != nil {
			r.log.Error("secondary create failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err // if no primary, secondary failure is fatal
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	if r.primary != nil {
		u, err := r.primary.FindByUsername(ctx, username)
		if err == nil && u != nil {
			return u, nil
		}
		// If primary returns an error that's not "user not found", return it
		if err != nil && !errors.Is(err, core.ErrUserNotFound) {
			return nil, err
		}
	}
	if r.secondary != nil {
		u, err := r.secondary.FindByUsername(ctx, username)
		if err != nil {
			// If secondary returns "user not found", treat as not found
			if errors.Is(err, core.ErrUserNotFound) {
				return nil, nil
			}
			return nil, err
		}
		if u != nil && r.primary != nil {
			go func() {
				ctxBg := context.Background()
				builtDetails := &user.UserDetails{
					UserBase: u.UserBase,
					UUID:     u.UUID,
					UserID:   u.UserID,
					Status:   "active",
				}
				if err := r.primary.Create(ctxBg, u, builtDetails); err != nil {
					r.log.Error("lazy migration failed", map[string]interface{}{"username": username, "error": err.Error()})
				}
			}()
		}
		return u, nil
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error) {
	if r.primary != nil {
		u, err := r.primary.FindByID(ctx, id)
		if err == nil && u != nil {
			return u, nil
		}
		if err != nil && !errors.Is(err, core.ErrUserNotFound) {
			return nil, err
		}
	}
	if r.secondary != nil {
		u, err := r.secondary.FindByID(ctx, id)
		if err != nil {
			if errors.Is(err, core.ErrUserNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return u, nil
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) Update(ctx context.Context, u *user.User) error {
	if r.primary != nil {
		if err := r.primary.Update(ctx, u); err != nil {
			if errors.Is(err, core.ErrUserNotFound) {
				r.log.Warn("primary update failed", map[string]interface{}{"error": err.Error()})
			} else {
				r.log.Error("primary update failed", map[string]interface{}{"error": err.Error()})
			}
		}
	}
	if r.secondary != nil {
		if err := r.secondary.Update(ctx, u); err != nil {
			if errors.Is(err, core.ErrUserNotFound) {
				r.log.Debug("secondary update: user not yet synced", map[string]interface{}{"error": err.Error()})
			} else {
				r.log.Warn("secondary update failed", map[string]interface{}{"error": err.Error()})
			}
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if r.primary != nil {
		if err := r.primary.Delete(ctx, id); err != nil {
			r.log.Error("primary delete failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.Delete(ctx, id); err != nil {
			r.log.Error("secondary delete failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) CreateInvite(ctx context.Context, invite *user.InviteCode) error {
	if r.primary != nil {
		if err := r.primary.CreateInvite(ctx, invite); err != nil {
			r.log.Error("primary create invite failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.CreateInvite(ctx, invite); err != nil {
			r.log.Error("secondary create invite failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) FindCode(ctx context.Context, code string) (*user.InviteCode, error) {
	if r.primary != nil {
		invite, err := r.primary.FindCode(ctx, code)
		if err == nil && invite != nil {
			return invite, nil
		}
		// Note: FindCode returns nil, nil if not found, not an error
		if err != nil && !errors.Is(err, core.ErrInvalidInviteCode) {
			return nil, err
		}
	}
	if r.secondary != nil {
		invite, err := r.secondary.FindCode(ctx, code)
		if err != nil {
			if errors.Is(err, core.ErrInvalidInviteCode) {
				return nil, nil
			}
			return nil, err
		}
		if invite != nil && r.primary != nil {
			go func() {
				ctxBg := context.Background()
				if err := r.primary.CreateInvite(ctxBg, invite); err != nil {
					r.log.Error("lazy migration invite failed", map[string]interface{}{"code": code, "error": err.Error()})
				}
			}()
		}
		return invite, nil
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error {
	if r.primary != nil {
		if err := r.primary.UseCode(ctx, codeID, userID); err != nil {
			r.log.Error("primary use code failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.UseCode(ctx, codeID, userID); err != nil {
			r.log.Error("secondary use code failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*user.InviteCode, error) {
	if r.secondary != nil {
		return r.secondary.FindAllInvites(ctx, page, limit)
	}
	if r.primary != nil {
		return r.primary.FindAllInvites(ctx, page, limit)
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error {
	if r.primary != nil {
		if err := r.primary.DeleteInvite(ctx, codeID); err != nil {
			r.log.Error("primary delete invite failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.DeleteInvite(ctx, codeID); err != nil {
			r.log.Error("secondary delete invite failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error {
	if r.primary != nil {
		if err := r.primary.UpdateRefreshToken(ctx, userID, token, expiresAt); err != nil {
			r.log.Error("primary update refresh token failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.UpdateRefreshToken(ctx, userID, token, expiresAt); err != nil {
			r.log.Error("secondary update refresh token failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	if r.primary != nil {
		if err := r.primary.InvalidateRefreshToken(ctx, userID); err != nil {
			r.log.Error("primary invalidate refresh token failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.InvalidateRefreshToken(ctx, userID); err != nil {
			r.log.Error("secondary invalidate refresh token failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*user.User, error) {
	if r.primary != nil {
		u, err := r.primary.FindByRefreshToken(ctx, refreshToken)
		if err == nil && u != nil {
			return u, nil
		}
		if err != nil && !errors.Is(err, core.ErrUserNotFound) {
			return nil, err
		}
	}
	if r.secondary != nil {
		u, err := r.secondary.FindByRefreshToken(ctx, refreshToken)
		if err != nil {
			if errors.Is(err, core.ErrUserNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return u, nil
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) GetNextSequence(ctx context.Context, sequenceName string) (int, error) {
	if r.primary != nil {
		return r.primary.GetNextSequence(ctx, sequenceName)
	}
	if r.secondary != nil {
		return r.secondary.GetNextSequence(ctx, sequenceName)
	}
	return 0, errors.New("no repositories available")
}

func (r *HybridUserRepository) FindAll(ctx context.Context, page, limit int) ([]*user.User, error) {
	if r.primary != nil {
		return r.primary.FindAll(ctx, page, limit)
	}
	if r.secondary != nil {
		return r.secondary.FindAll(ctx, page, limit)
	}
	return nil, errors.New("no repositories available")
}

func (r *HybridUserRepository) FindAllUsers(ctx context.Context, skip, limit int64) ([]*user.User, int64, error) {
	if r.primary != nil {
		return r.primary.FindAllUsers(ctx, skip, limit)
	}
	if r.secondary != nil {
		return r.secondary.FindAllUsers(ctx, skip, limit)
	}
	return nil, 0, errors.New("no repositories available")
}

func (r *HybridUserRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role string, add bool) error {
	if r.primary != nil {
		if err := r.primary.UpdateUserRole(ctx, userID, role, add); err != nil {
			r.log.Error("primary update role failed", map[string]interface{}{"error": err.Error()})
		}
	}
	if r.secondary != nil {
		if err := r.secondary.UpdateUserRole(ctx, userID, role, add); err != nil {
			r.log.Error("secondary update role failed", map[string]interface{}{"error": err.Error()})
			if r.primary == nil {
				return err
			}
		}
	} else if r.primary == nil {
		return errors.New("no repositories available")
	}
	return nil
}

func (r *HybridUserRepository) CreateUserWithInvite(ctx context.Context, u *user.User, d *user.UserDetails, inviteCode string) error {
	// Try primary first
	var primaryErr, secondaryErr error
	if r.primary != nil {
		primaryErr = r.primary.CreateUserWithInvite(ctx, u, d, inviteCode)
	}
	// If primary succeeded, propagate to secondary if possible
	if primaryErr == nil {
		if r.secondary != nil {
			// best effort: do not fail if secondary fails
			if err := r.secondary.CreateUserWithInvite(ctx, u, d, inviteCode); err != nil {
				r.log.Error("secondary create user with invite failed after primary success", map[string]interface{}{"error": err.Error()})
			}
		}
		return nil
	}

	// If primary failed, try secondary only if we have no primary
	if r.secondary != nil && r.primary == nil {
		secondaryErr = r.secondary.CreateUserWithInvite(ctx, u, d, inviteCode)
		if secondaryErr == nil {
			return nil
		}
	}

	// If both failed, return primary error (or secondary if primary missing)
	if r.primary != nil {
		return primaryErr
	}
	return secondaryErr
}

// ----- END OF FILE: backend/MS-AI/internal/data/user/hybrid_repo.go -----
