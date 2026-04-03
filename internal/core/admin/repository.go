// ----- START OF FILE: backend/MS-AI/internal/core/admin/repository.go -----
// internal/core/admin/repository.go
package admin

import (
	"context"

	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
)

// Repository defines the interface for admin data operations
type Repository interface {
	CreateInvite(ctx context.Context, invite *coreuser.InviteCode) error
	ListInvites(ctx context.Context, skip, limit int64) ([]*coreuser.InviteCode, int64, error)
	DeleteInvite(ctx context.Context, id string) error
}

// ----- END OF FILE: backend/MS-AI/internal/core/admin/repository.go -----
