// ----- START OF FILE: backend/MS-AI/internal/data/admin/mongo_admin_repo.go -----
// internal/data/admin/mongo_admin_repo.go
package admin

import (
	"context"
	"errors"

	coreAdmin "github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreUser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mongoAdminRepo implements coreAdmin.Repository using MongoDB
type mongoAdminRepo struct {
	r coreUser.Repository
}

// NewMongoAdminRepository creates a new MongoDB admin repository
func NewMongoAdminRepository(r coreUser.Repository) coreAdmin.Repository {
	return &mongoAdminRepo{r: r}
}

// CreateInvite implements coreAdmin.Repository.CreateInvite
func (a *mongoAdminRepo) CreateInvite(ctx context.Context, invite *coreUser.InviteCode) error {
	return a.r.CreateInvite(ctx, invite)
}

// ListInvites implements coreAdmin.Repository.ListInvites
func (a *mongoAdminRepo) ListInvites(ctx context.Context, skip, limit int64) ([]*coreUser.InviteCode, int64, error) {
	// Note: datauser.Repository has FindAllInvites, but it takes page, limit int, not skip, limit int64
	// We need to convert
	page := int(skip/limit) + 1
	invites, err := a.r.FindAllInvites(ctx, page, int(limit))
	if err != nil {
		return nil, 0, err
	}
	// For simplicity, return without total count, or we need to add a method
	return invites, int64(len(invites)), nil
}

// DeleteInvite implements coreAdmin.Repository.DeleteInvite
func (a *mongoAdminRepo) DeleteInvite(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid invite code id")
	}
	return a.r.DeleteInvite(ctx, oid)
}

// ----- END OF FILE: backend/MS-AI/internal/data/admin/mongo_admin_repo.go -----
