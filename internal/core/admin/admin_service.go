// ----- START OF FILE: backend/MS-AI/internal/core/admin/admin_service.go -----
// internal/core/admin/admin_service.go
package admin

import (
	"context"
	"fmt"
	"time"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"

	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
)

// DefaultAdminService is a concrete implementation of admin service
type DefaultAdminService struct {
	repo      Repository
	userRepo  coreuser.Repository
	log       *logger.Logger
	startTime time.Time
}

// NewAdminService creates a new admin service
func NewAdminService(userRepo coreuser.Repository, repo Repository, log *logger.Logger) *DefaultAdminService {
	return &DefaultAdminService{
		repo:      repo,
		userRepo:  userRepo,
		log:       log,
		startTime: time.Now(),
	}
}

// GetMetrics returns admin metrics
func (s *DefaultAdminService) GetMetrics(ctx context.Context) map[string]interface{} {
	_, totalUsers, userErr := s.userRepo.FindAllUsers(ctx, 0, 1)
	_, inviteTotal, inviteErr := s.repo.ListInvites(ctx, 1, 1)

	return map[string]interface{}{
		"uptime": time.Since(s.startTime).String(),
		"total_users": func() int64 {
			if userErr != nil {
				return -1
			}
			return totalUsers
		}(),
		"total_invites": func() int64 {
			if inviteErr != nil {
				return -1
			}
			return inviteTotal
		}(),
		"errors":      map[string]string{"users": fmt.Sprintf("%v", userErr), "invites": fmt.Sprintf("%v", inviteErr)},
		"server_time": time.Now().UTC().Format(time.RFC3339),
	}
}

// GetDBMetrics returns DB metrics
func (s *DefaultAdminService) GetDBMetrics(ctx context.Context) map[string]interface{} {
	metrics := map[string]interface{}{"user_repo_status": "unknown", "invite_repo_status": "unknown"}

	_, _, userErr := s.userRepo.FindAllUsers(ctx, 0, 1)
	if userErr != nil {
		metrics["user_repo_status"] = userErr.Error()
	} else {
		metrics["user_repo_status"] = "ok"
	}

	_, _, inviteErr := s.repo.ListInvites(ctx, 1, 1)
	if inviteErr != nil {
		metrics["invite_repo_status"] = inviteErr.Error()
	} else {
		metrics["invite_repo_status"] = "ok"
	}

	return metrics
}

func (s *DefaultAdminService) CreateInviteCode(ctx context.Context, length int) (*coreuser.InviteCode, error) {
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
	inv := &coreuser.InviteCode{
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

func (s *DefaultAdminService) ListInviteCodes(ctx context.Context, skip, limit int64) ([]*coreuser.InviteCode, int64, error) {
	return s.repo.ListInvites(ctx, skip, limit)
}

func (s *DefaultAdminService) DeleteInviteCode(ctx context.Context, id string) error {
	return s.repo.DeleteInvite(ctx, id)
}

// CreateCustomInviteCode creates an invite code with a specific code string
func (s *DefaultAdminService) CreateCustomInviteCode(ctx context.Context, code string, daysValid int) (*coreuser.InviteCode, error) {
	now := time.Now()
	if daysValid <= 0 {
		daysValid = 30
	}

	inv := &coreuser.InviteCode{
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
			Roles:     u.Roles.ToStrings(),
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
	return s.userRepo.UpdateUserRole(ctx, oid, user.RoleAdmin, true)
}

// DemoteToUser demotes an admin to regular user role
func (s *DefaultAdminService) DemoteToUser(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return s.userRepo.UpdateUserRole(ctx, oid, user.RoleAdmin, false)
}

func (s *DefaultAdminService) DeactivateUser(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil {
		return err
	}
	if user == nil {
		return corecommon.ErrUserNotFound
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *DefaultAdminService) DeleteUser(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return s.userRepo.Delete(ctx, oid)
}

// ----- END OF FILE: backend/MS-AI/internal/core/admin/admin_service.go -----
