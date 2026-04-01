package main

import (
"context"
"fmt"
"time"

"golang.org/x/crypto/bcrypt"

coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
"github.com/835-droid/ms-ai-backend/internal/data/mongo"
datauser "github.com/835-droid/ms-ai-backend/internal/data/user"
"github.com/835-droid/ms-ai-backend/pkg/config"
plog "github.com/835-droid/ms-ai-backend/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("load config failed:", err)
		return
	}
	log := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)
	store, err := mongo.NewMongoStore(cfg, log)
	if err != nil {
		log.Fatal("connect mongo failed", map[string]interface{}{"error": err.Error()})
	}
	repo := datauser.NewMongoUserRepository(store)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("failed to hash password: %v\n", err)
		return
	}

	adminUser := &coreuser.User{
		Username: "admin",
		Password: string(hashedPass),
		Roles:    []string{"admin"},
		IsActive: true,
	}

	adminDetails := &coreuser.UserDetails{
		Roles:    []string{"admin"},
		IsActive: true,
		Status:   "active",
	}

	if err := repo.Create(ctx, adminUser, adminDetails); err != nil {
		fmt.Printf("failed to create admin user: %v\n", err)
		return
	}

	fmt.Println("created admin user successfully")

	// Create invite code
	invite := &coreuser.InviteCode{
		Code:      "ADMINCODE123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(1, 0, 0), // 1 year
	}

	if err := repo.CreateInviteCode(ctx, invite); err != nil {
		fmt.Printf("failed to create invite code: %v\n", err)
		return
	}

	fmt.Println("created invite code: ADMINCODE123")
}
