// internal/container/seed_initializers.go
package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func initializeInitialData(ctx context.Context, cfg *config.Config, repos *RepoBundle, log *logger.Logger) {
	if repos == nil || repos.User == nil {
		log.Warn("user repository is unavailable, skipping initial data setup", nil)
		return
	}

	// In production, skip default admin creation unless explicitly allowed
	environment := os.Getenv("ENVIRONMENT")
	if cfg != nil && cfg.Environment != "" {
		environment = cfg.Environment
	}
	if strings.ToLower(environment) == "production" && os.Getenv("DISABLE_DEFAULT_ADMIN") != "false" {
		log.Info("production mode: skipping default admin creation", nil)
		return
	}

	admin, err := repos.User.FindByUsername(ctx, "admin")
	if err != nil && !errors.Is(err, core.ErrUserNotFound) {
		log.Warn("failed to check for admin user existence", map[string]interface{}{"error": err.Error()})
		return
	}
	if admin == nil {
		hp, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		if err != nil {
			log.Warn("failed to hash default admin password", map[string]interface{}{"error": err.Error()})
			return
		}
		newUser := &coreuser.User{
			ID:       primitive.NewObjectID(),
			UUID:     utils.GenerateUUID(),
			UserID:   fmt.Sprintf("User-%d", 1),
			Username: "admin",
			Password: string(hp),
			UserBase: coreuser.UserBase{
				Roles:     coreuser.FromStrings([]string{"admin"}),
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		if err := repos.User.Create(ctx, newUser, &coreuser.UserDetails{
			UUID:                  utils.GenerateUUID(),
			UserID:                newUser.ID.Hex(),
			Status:                "active",
			LastLoginAt:           nil,
			RefreshToken:          "",
			RefreshTokenExpiresAt: nil,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}); err != nil {
			log.Warn("failed to create admin user", map[string]interface{}{"error": err.Error()})
		} else {
			log.Info("admin user created successfully", nil)
		}
	} else {
		log.Info("admin user already exists, skipping creation", nil)
	}
}
