// internal/core/admin/admin_service.go
package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"

	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
)

// DefaultAdminService is a concrete implementation of admin service
type DefaultAdminService struct {
	repo     Repository
	userRepo coreuser.Repository
	log      *logger.Logger
}

// NewAdminService creates a new admin service
func NewAdminService(userRepo coreuser.Repository, repo Repository, log *logger.Logger) *DefaultAdminService {
	return &DefaultAdminService{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
	}
}

// GetMetrics returns admin metrics (placeholder)
func (s *DefaultAdminService) GetMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{}
}

// GetDBMetrics returns DB metrics (placeholder)
func (s *DefaultAdminService) GetDBMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{}
}

func (s *DefaultAdminService) CreateInviteCode(ctx context.Context, length int) (*InviteCode, error) {
	if s.log != nil {
		s.log.Debug("generating invite code", map[string]interface{}{"length": length})
	}
	if length <= 0 {
		length = 12
	}
	code, err := utils.GenerateRandomCode(length)
	if err != nil {
		if s.log != nil {
			s.log.Error("failed to generate random code", map[string]interface{}{"error": err.Error()})
		}
		return nil, err
	}
	now := time.Now()
	inv := &InviteCode{
		Code:      code,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	if err := s.repo.CreateInvite(ctx, inv); err != nil {
		if s.log != nil {
			s.log.Error("failed to create invite code", map[string]interface{}{"error": err.Error()})
		}
		return nil, err
	}
	return inv, nil
}

func (s *DefaultAdminService) ListInviteCodes(ctx context.Context, skip, limit int64) ([]*InviteCode, int64, error) {
	return s.repo.ListInvites(ctx, skip, limit)
}

func (s *DefaultAdminService) DeleteInviteCode(ctx context.Context, id string) error {
	return s.repo.DeleteInvite(ctx, id)
}

// CreateCustomInviteCode creates an invite code with a specific code string
func (s *DefaultAdminService) CreateCustomInviteCode(ctx context.Context, code string, daysValid int) (*InviteCode, error) {
	now := time.Now()
	if daysValid <= 0 {
		daysValid = 30
	}

	inv := &InviteCode{
		ID:        primitive.NewObjectID(),
		Code:      code,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(daysValid) * 24 * time.Hour),
		IsUsed:    false,
	}

	if err := s.repo.CreateInvite(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// ListUsers returns a list of all users (for admin)
func (s *DefaultAdminService) ListUsers(ctx context.Context, page, limit int) ([]*UserInfo, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := int64((page - 1) * limit)

	users, total, err := s.userRepo.FindAllUsers(ctx, skip, int64(limit))
	if err != nil {
		return nil, 0, err
	}

	result := make([]*UserInfo, len(users))
	for i, u := range users {
		result[i] = &UserInfo{
			ID:        u.ID.Hex(),
			Username:  u.Username,
			Roles:     u.Roles,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt,
		}
	}
	return result, total, nil
}

// PromoteToAdmin promotes a user to admin role
func (s *DefaultAdminService) PromoteToAdmin(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return s.userRepo.UpdateUserRole(ctx, oid, "admin", true)
}
