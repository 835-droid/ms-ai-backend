package admin

import (
	"context"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DefaultAdminService is a concrete implementation of admin service using the package Repository
type DefaultAdminService struct {
	repo Repository
	log  *logger.Logger
}

func NewAdminService(repo Repository, log *logger.Logger) *DefaultAdminService {
	return &DefaultAdminService{repo: repo, log: log}
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

// تعديل في internal/core/admin/admin_service.go
func (s *DefaultAdminService) CreateCustomInviteCode(ctx context.Context, code string, daysValid int) (*InviteCode, error) {
	now := time.Now()
	// إذا لم يتم تحديد أيام، نجعلها 30 يوماً بدلاً من يوم واحد
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
