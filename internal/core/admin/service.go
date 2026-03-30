package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service provides admin functionality
type Service interface {
	CreateInviteCode(ctx context.Context, length int) (*InviteCode, error)
	ListInviteCodes(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error)
	DeleteInviteCode(ctx context.Context, id string) error
	GetMetrics(ctx context.Context) map[string]interface{}
	GetDBMetrics(ctx context.Context) map[string]interface{}
}

// AdminService implements Service
type AdminService struct {
	repo Repository
}

// Deprecated: Use admin.NewAdminService with logger (implemented in admin_service.go)
// NewAdminService kept for compatibility but should not be used.
// func NewAdminService(repo Repository) *AdminService {
//     return &AdminService{repo: repo}
// }

// InviteCode represents an invite code
type InviteCode struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Code      string             `bson:"code" json:"code"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	IsUsed    bool               `bson:"is_used" json:"is_used"`
	UsedBy    primitive.ObjectID `bson:"used_by,omitempty" json:"used_by,omitempty"`
}

func (s *AdminService) CreateInviteCode(ctx context.Context, length int) (*InviteCode, error) {
	if length <= 0 {
		length = 12
	}
	buf := make([]byte, (length*6+7)/8)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(buf)[:length]

	now := time.Now()
	invite := &InviteCode{
		Code:      code,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour), // Expires in 24 hours
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite code: %w", err)
	}
	return invite, nil
}

func (s *AdminService) ListInviteCodes(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error) {
	return s.repo.ListInvites(ctx, skip, limit)
}

func (s *AdminService) DeleteInviteCode(ctx context.Context, id string) error {
	return s.repo.DeleteInvite(ctx, id)
}

func (s *AdminService) GetMetrics(ctx context.Context) map[string]interface{} {
	// TODO: Implement system metrics
	return map[string]interface{}{
		"uptime": time.Since(time.Now()).String(),
	}
}

func (s *AdminService) GetDBMetrics(ctx context.Context) map[string]interface{} {
	// TODO: Implement database metrics
	return map[string]interface{}{
		"status": "connected",
	}
}
