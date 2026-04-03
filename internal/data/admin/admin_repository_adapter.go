// ----- START OF FILE: backend/MS-AI/internal/data/admin/admin_repository_adapter.go -----
// internal/data/admin/admin_repository_adapter.go
package admin

import (
	"context"
	"errors"

	coreAdmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// adminRepositoryAdapter implements coreAdmin.Repository by delegating to coreUser.Repository
type adminRepositoryAdapter struct {
	r coreUser.Repository
}

// NewAdminRepositoryAdapter creates a new admin repository adapter
func NewAdminRepositoryAdapter(r coreUser.Repository) coreAdmin.Repository {
	return &adminRepositoryAdapter{r: r}
}

// CreateInvite implements coreAdmin.Repository.CreateInvite
func (a *adminRepositoryAdapter) CreateInvite(ctx context.Context, invite *coreUser.InviteCode) error {
	return a.r.CreateInvite(ctx, invite)
}

// ListInvites implements coreAdmin.Repository.ListInvites
func (a *adminRepositoryAdapter) ListInvites(ctx context.Context, skip, limit int64) ([]*coreUser.InviteCode, int64, error) {
	if limit <= 0 {
		return nil, 0, errors.New("limit must be positive")
	}
	if a.r == nil {
		return nil, 0, errors.New("user repository is nil")
	}
	// Prefer direct total from repository if supported.
	if res, total, err := a.r.FindAllInvitesWithTotal(ctx, skip, limit); err == nil {
		return res, total, nil
	}

	// Fallback: convert skip/limit to page and use existing list method, but we may lose total.
	page := int(skip/limit) + 1
	invites, err := a.r.FindAllInvites(ctx, page, int(limit))
	if err != nil {
		return nil, 0, err
	}

	// If the repository does not support total, count by querying page 1 and last page could be inaccurate; still, this is fallback.
	_, total, err := a.r.FindAllInvitesWithTotal(ctx, 0, 1)
	if err != nil {
		return invites, int64(len(invites)), nil
	}
	return invites, total, nil
}

// DeleteInvite implements coreAdmin.Repository.DeleteInvite
func (a *adminRepositoryAdapter) DeleteInvite(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid invite code id")
	}
	return a.r.DeleteInvite(ctx, oid)
}

// ----- END OF FILE: backend/MS-AI/internal/data/admin/admin_repository_adapter.go -----
