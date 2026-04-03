package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/835-droid/ms-ai-backend/internal/container"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	plog "github.com/835-droid/ms-ai-backend/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
	}
	log := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)

	mStore, pStore, err := container.InitializeDatabases(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize databases", map[string]interface{}{"error": err.Error()})
	}

	repos := container.InitializeRepositories(cfg, log, mStore, pStore)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// seed admin user if not exists
	adminName := os.Getenv("SEED_ADMIN_USERNAME")
	if adminName == "" {
		adminName = "admin"
	}
	existing, _ := repos.User.FindByUsername(ctx, adminName)
	if existing != nil {
		fmt.Println("admin user exists, skipping creation")
	} else {
		adminPass := os.Getenv("SEED_ADMIN_PASSWORD")
		if adminPass == "" {
			adminPass = "admin123"
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("failed to hash admin password: %v\n", err)
			os.Exit(1)
		}
		adminUser := &coreuser.User{
			Username: adminName,
			Password: string(hashedPass),
			UserBase: coreuser.UserBase{
				Roles:    coreuser.Roles{coreuser.RoleAdmin},
				IsActive: true,
			},
		}
		adminDetails := &coreuser.UserDetails{
			Status:                "active",
			LastLoginAt:           nil,
			RefreshToken:          "",
			RefreshTokenExpiresAt: nil,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		if err := repos.User.Create(ctx, adminUser, adminDetails); err != nil {
			fmt.Printf("failed to create admin user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("created admin user: %s\n", adminName)
	}

	// create default invite codes
	codes := []string{"INITCODE1", "INITCODE2"}
	for _, c := range codes {
		inv := &coreuser.InviteCode{
			Code:      c,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		}
		if err := repos.User.CreateInvite(ctx, inv); err != nil {
			fmt.Printf("seed invite code %s failed: %v\n", c, err)
			continue
		}
		fmt.Printf("seeded invite code: %s\n", c)
	}
	fmt.Println("seeding completed")
}
