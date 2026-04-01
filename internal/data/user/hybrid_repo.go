package user

import (
	"context"

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
	if err := r.primary.Create(ctx, u, d); err != nil {
		r.log.Error("primary create failed", map[string]interface{}{"error": err.Error()})
	}
	if err := r.secondary.Create(ctx, u, d); err != nil {
		r.log.Error("secondary create failed", map[string]interface{}{"error": err.Error()})
		return err
	}
	return nil
}

func (r *HybridUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	u, err := r.primary.FindByUsername(ctx, username)
	if err == nil && u != nil {
		return u, nil
	}
	u, err = r.secondary.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u != nil {
		go func() {
			ctxBg := context.Background()
			if err := r.primary.Create(ctxBg, u, &user.UserDetails{}); err != nil {
				r.log.Error("lazy migration failed", map[string]interface{}{"username": username, "error": err.Error()})
			}
		}()
	}
	return u, nil
}

func (r *HybridUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error) {
	u, err := r.primary.FindByID(ctx, id)
	if err == nil && u != nil {
		return u, nil
	}
	return r.secondary.FindByID(ctx, id)
}

func (r *HybridUserRepository) Update(ctx context.Context, u *user.User) error {
	if err := r.primary.Update(ctx, u); err != nil {
		r.log.Error("primary update failed", map[string]interface{}{"error": err.Error()})
	}
	if err := r.secondary.Update(ctx, u); err != nil {
		r.log.Error("secondary update failed", map[string]interface{}{"error": err.Error()})
		return err
	}
	return nil
}

func (r *HybridUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if err := r.primary.Delete(ctx, id); err != nil {
		r.log.Error("primary delete failed", map[string]interface{}{"error": err.Error()})
	}
	if err := r.secondary.Delete(ctx, id); err != nil {
		r.log.Error("secondary delete failed", map[string]interface{}{"error": err.Error()})
		return err
	}
	return nil
}

func (r *HybridUserRepository) CreateInvite(ctx context.Context, invite *user.InviteCode) error {
	return r.secondary.CreateInvite(ctx, invite)
}

func (r *HybridUserRepository) FindCode(ctx context.Context, code string) (*user.InviteCode, error) {
	return r.secondary.FindCode(ctx, code)
}

func (r *HybridUserRepository) UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error {
	return r.secondary.UseCode(ctx, codeID, userID)
}

func (r *HybridUserRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*user.InviteCode, error) {
	return r.secondary.FindAllInvites(ctx, page, limit)
}

func (r *HybridUserRepository) DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error {
	return r.secondary.DeleteInvite(ctx, codeID)
}

func (r *HybridUserRepository) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error {
	if err := r.primary.UpdateRefreshToken(ctx, userID, token, expiresAt); err != nil {
		r.log.Error("primary update refresh token failed", map[string]interface{}{"error": err.Error()})
	}
	return r.secondary.UpdateRefreshToken(ctx, userID, token, expiresAt)
}

func (r *HybridUserRepository) InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	if err := r.primary.InvalidateRefreshToken(ctx, userID); err != nil {
		r.log.Error("primary invalidate refresh token failed", map[string]interface{}{"error": err.Error()})
	}
	return r.secondary.InvalidateRefreshToken(ctx, userID)
}

func (r *HybridUserRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*user.User, error) {
	u, err := r.primary.FindByRefreshToken(ctx, refreshToken)
	if err == nil && u != nil {
		return u, nil
	}
	return r.secondary.FindByRefreshToken(ctx, refreshToken)
}

func (r *HybridUserRepository) GetNextSequence(ctx context.Context, sequenceName string) (int, error) {
	return r.primary.GetNextSequence(ctx, sequenceName)
}

func (r *HybridUserRepository) FindAll(ctx context.Context, page, limit int) ([]*user.User, error) {
	return r.primary.FindAll(ctx, page, limit)
}

func (r *HybridUserRepository) FindAllUsers(ctx context.Context, skip, limit int64) ([]*user.User, int64, error) {
	return r.primary.FindAllUsers(ctx, skip, limit)
}

func (r *HybridUserRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role string, add bool) error {
	if err := r.primary.UpdateUserRole(ctx, userID, role, add); err != nil {
		r.log.Error("primary update role failed", map[string]interface{}{"error": err.Error()})
	}
	return r.secondary.UpdateUserRole(ctx, userID, role, add)
}
