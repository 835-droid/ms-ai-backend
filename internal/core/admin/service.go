// ----- START OF FILE: backend/MS-AI/internal/core/admin/service.go -----
// internal/core/admin/service.go
package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

// AdminService implements Service
type AdminService struct {
	repo Repository
}

// CreateInviteCode generates a random invite code
func (s *AdminService) CreateInviteCode(ctx context.Context, length int) (*coreuser.InviteCode, error) {
	if length <= 0 {
		length = 12
	}
	buf := make([]byte, (length*6+7)/8)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(buf)[:length]

	now := time.Now()
	invite := &coreuser.InviteCode{
		Code:      code,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite code: %w", err)
	}
	return invite, nil
}

// CreateCustomInviteCode creates an invite code with a specific code string
func (s *AdminService) CreateCustomInviteCode(ctx context.Context, code string, daysValid int) (*coreuser.InviteCode, error) {
	now := time.Now()
	if daysValid <= 0 {
		daysValid = 30
	}
	invite := &coreuser.InviteCode{
		ID:        primitive.NewObjectID(),
		Code:      code,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(daysValid) * 24 * time.Hour),
		IsUsed:    false,
	}
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create custom invite: %w", err)
	}
	return invite, nil
}

// ListInviteCodes returns paginated invite codes
func (s *AdminService) ListInviteCodes(ctx context.Context, skip, limit int64) ([]*coreuser.InviteCode, int64, error) {
	return s.repo.ListInvites(ctx, skip, limit)
}

// DeleteInviteCode deletes an invite code by ID
func (s *AdminService) DeleteInviteCode(ctx context.Context, id string) error {
	return s.repo.DeleteInvite(ctx, id)
}

// GetMetrics returns system metrics
func (s *AdminService) GetMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"uptime": time.Since(time.Now()).String(),
	}
}

// GetDBMetrics returns database metrics
func (s *AdminService) GetDBMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"status": "connected",
	}
}

// ----- END OF FILE: backend/MS-AI/internal/core/admin/service.go -----
