// internal/core/auth_service.go
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/config"
	tokenpkg "github.com/835-droid/ms-ai-backend/pkg/jwt"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"github.com/835-droid/ms-ai-backend/pkg/validator"

	corecommon "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AuthResult contains authenticated user and tokens.
type AuthResult struct {
	User         *coreuser.User
	AccessToken  string
	RefreshToken string
}

// AuthService defines authentication flows.
type AuthService interface {
	SignUp(ctx context.Context, username, password, inviteCode string) (*AuthResult, error)
	Login(ctx context.Context, username, password string) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, userID primitive.ObjectID) error
	CreateInviteCode(ctx context.Context, length int) (*coreuser.InviteCode, error)
}

// DefaultAuthService implements AuthService.
type DefaultAuthService struct {
	repo coreuser.Repository
	cfg  *config.Config
	log  *zerolog.Logger
}

// NewAuthService constructs a DefaultAuthService.
func NewAuthService(repo coreuser.Repository, cfg *config.Config, log *zerolog.Logger) *DefaultAuthService {
	if log == nil {
		l := zerolog.Nop()
		log = &l
	}
	return &DefaultAuthService{repo: repo, cfg: cfg, log: log}
}

// SignUp registers a user after validating invite code and credentials.
func (s *DefaultAuthService) SignUp(ctx context.Context, username, password, inviteCode string) (*AuthResult, error) {
	// normalize username to avoid case-sensitivity issues in login
	username = strings.TrimSpace(strings.ToLower(username))

	s.log.Debug().
		Str("username", username).
		Str("inviteCode", inviteCode).
		Msg("attempting user signup")

	if err := validator.ValidateUsername(username); err != nil {
		s.log.Debug().
			Str("username", username).
			Err(err).
			Msg("invalid username")
		return nil, err
	}
	if err := validator.ValidatePassword(password); err != nil {
		s.log.Debug().
			Str("username", username).
			Err(err).
			Msg("invalid password")
		return nil, err
	}

	existing, _ := s.repo.FindByUsername(ctx, username)
	if existing != nil {
		return nil, corecommon.ErrUserExists
	}

	invite, err := s.repo.FindCode(ctx, inviteCode)
	if err != nil || invite == nil {
		return nil, corecommon.ErrInvalidInviteCode
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error().
			Str("username", username).
			Err(err).
			Msg("failed to hash password")
		return nil, corecommon.ErrInternalServer
	}

	// 1. احصل على الرقم التسلسلي التالي
	seq, _ := s.repo.GetNextSequence(ctx, "user_id_counter")

	// 2. حوله لصيغة نصية جميلة
	publicUserID := fmt.Sprintf("User-%d", seq)

	newOID := primitive.NewObjectID() // توليد ID جديد هنا

	now := time.Now()
	newUser := &coreuser.User{
		ID:        newOID,
		UUID:      utils.GenerateUUID(),
		UserID:    publicUserID,
		Username:  username,
		Password:  string(hashed),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	newDetails := &coreuser.UserDetails{
		UUID:      utils.GenerateUUID(),
		UserID:    publicUserID,
		Status:    "active",
		Roles:     []string{"user"},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// copy roles from details to user struct so that tokens generation works
	newUser.Roles = newDetails.Roles

	if err := s.repo.Create(ctx, newUser, newDetails); err != nil {
		s.log.Error().
			Str("username", username).
			Err(err).
			Msg("failed to create user")
		return nil, err
	}

	// mark invite used (best effort; repository should support transactions)
	if err := s.repo.UseCode(ctx, invite.ID, newUser.ID); err != nil {
		s.log.Error().
			Str("userId", newUser.ID.Hex()).
			Str("inviteId", invite.ID.Hex()).
			Err(err).
			Msg("failed to mark invite code as used")
	}

	// generate tokens
	access, err := s.generateAccessToken(newUser)
	if err != nil {
		return nil, corecommon.ErrInternalServer
	}
	refresh, err := s.generateAndStoreRefreshToken(ctx, newUser)
	if err != nil {
		return nil, corecommon.ErrInternalServer
	}

	return &AuthResult{User: newUser, AccessToken: access, RefreshToken: refresh}, nil
}

func (s *DefaultAuthService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	// normalize username to avoid case-sensitivity issues
	username = strings.TrimSpace(strings.ToLower(username))

	u, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		s.log.Warn().Str("username", username).Err(err).Msg("login failed: user not found")
		return nil, corecommon.ErrInvalidCredentials
	}

	// قارن الباسورد المدخل مع المشفر في الداتابيز
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		s.log.Warn().Str("username", username).Msg("login failed: password mismatch")
		return nil, corecommon.ErrInvalidCredentials
	}

	// update last login time (best effort)
	now := time.Now()
	u.LastLoginAt = &now
	if err := s.repo.Update(ctx, u); err != nil {
		s.log.Warn().Err(err).Msg("failed to update last login timestamp")
	}

	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken validates a refresh token and issues a new access token.
func (s *DefaultAuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	s.log.Debug().Msg("attempting token refresh")

	if refreshToken == "" {
		s.log.Debug().Msg("empty refresh token")
		return nil, corecommon.ErrInvalidToken
	}

	// use a bounded context for DB operations
	dbCtx, cancel := context.WithTimeout(ctx, s.cfg.DBTimeout)
	defer cancel()

	user, err := s.repo.FindByRefreshToken(dbCtx, refreshToken)
	if err != nil {
		s.log.Debug().
			Err(err).
			Msg("failed to find user by refresh token")
		return nil, corecommon.ErrInvalidToken
	}
	if user == nil {
		s.log.Debug().Msg("user not found for refresh token")
		return nil, corecommon.ErrInvalidToken
	}

	// rotate refresh token: generate a new one and persist it
	s.log.Debug().
		Str("userId", user.ID.Hex()).
		Msg("rotating refresh token")

	newRefresh, err := tokenpkg.GenerateRefreshToken(32)
	if err != nil {
		s.log.Error().
			Str("userId", user.ID.Hex()).
			Err(err).
			Msg("failed to generate refresh token")
		return nil, corecommon.ErrInternalServer
	}

	expires := primitive.NewDateTimeFromTime(time.Now().Add(s.cfg.JWTRefreshExpiry))
	if err := s.repo.UpdateRefreshToken(dbCtx, user.ID, newRefresh, expires); err != nil {
		s.log.Error().
			Str("userId", user.ID.Hex()).
			Err(err).
			Msg("failed to persist rotated refresh token")
		return nil, corecommon.ErrInternalServer
	}

	access, err := s.generateAccessToken(user)
	if err != nil {
		s.log.Error().
			Str("userId", user.ID.Hex()).
			Err(err).
			Msg("failed to generate access token")
		return nil, corecommon.ErrInternalServer
	}

	s.log.Debug().
		Str("userId", user.ID.Hex()).
		Msg("successfully refreshed tokens")

	return &AuthResult{User: user, AccessToken: access, RefreshToken: newRefresh}, nil
}

// Logout invalidates the refresh token for a user.
func (s *DefaultAuthService) Logout(ctx context.Context, userID primitive.ObjectID) error {
	return s.repo.InvalidateRefreshToken(ctx, userID)
}

// CreateInviteCode generates a secure invite code and stores it.
func (s *DefaultAuthService) CreateInviteCode(ctx context.Context, length int) (*coreuser.InviteCode, error) {
	if length <= 0 {
		length = 12
	}
	code, err := utils.GenerateRandomCode(length)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	inv := &coreuser.InviteCode{Code: code, IsUsed: false, CreatedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour)}
	if err := s.repo.CreateInvite(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// --- helpers ---
func (s *DefaultAuthService) generateAccessToken(u *coreuser.User) (string, error) {
	s.log.Debug().
		Str("userId", u.ID.Hex()).
		Str("username", u.Username).
		Strs("roles", u.Roles).
		Dur("duration", s.cfg.JWTAccessExpiry).
		Msg("generating access token")

	token, err := tokenpkg.GenerateAccessToken(u.ID.Hex(), u.Roles, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		s.log.Error().
			Str("userId", u.ID.Hex()).
			Err(err).
			Msg("failed to generate access token")
		return "", err
	}
	return token, nil
}

func (s *DefaultAuthService) generateAndStoreRefreshToken(ctx context.Context, u *coreuser.User) (string, error) {
	s.log.Debug().
		Str("userId", u.ID.Hex()).
		Dur("duration", s.cfg.JWTRefreshExpiry).
		Msg("generating refresh token")

	t, err := tokenpkg.GenerateRefreshToken(32)
	if err != nil {
		s.log.Error().
			Str("userId", u.ID.Hex()).
			Err(err).
			Msg("failed to generate refresh token")
		return "", err
	}

	expires := primitive.NewDateTimeFromTime(time.Now().Add(s.cfg.JWTRefreshExpiry))
	if err := s.repo.UpdateRefreshToken(ctx, u.ID, t, expires); err != nil {
		s.log.Error().
			Str("userId", u.ID.Hex()).
			Err(err).
			Msg("failed to store refresh token")
		return "", err
	}

	s.log.Info().
		Str("userId", u.ID.Hex()).
		Msg("refresh token stored")
	return t, nil
}
