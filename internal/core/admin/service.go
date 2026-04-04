// ----- START OF FILE: backend/MS-AI/internal/core/admin/service.go -----
// internal/core/admin/service.go
package admin

import (
	"context"
	"time"

	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
)

// Service provides admin functionality
type Service interface {
	CreateInviteCode(ctx context.Context, length int) (*coreuser.InviteCode, error)
	CreateCustomInviteCode(ctx context.Context, code string, daysValid int) (*coreuser.InviteCode, error)
	ListInviteCodes(ctx context.Context, skip, limit int64) ([]*coreuser.InviteCode, int64, error)
	DeleteInviteCode(ctx context.Context, id string) error
	GetMetrics(ctx context.Context) map[string]interface{}
	GetDBMetrics(ctx context.Context) map[string]interface{}
	ListUsers(ctx context.Context, page, limit int) ([]*UserInfo, int64, error)
	PromoteToAdmin(ctx context.Context, userID string) error
	DemoteToUser(ctx context.Context, userID string) error
	ChangeUserPassword(ctx context.Context, userID string, newPassword string) error
	DeactivateUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
}

// UserInfo represents user data for admin panel
type UserInfo struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ----- END OF FILE: backend/MS-AI/internal/core/admin/service.go -----
