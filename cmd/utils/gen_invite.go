package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/container"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	plog "github.com/835-droid/ms-ai-backend/pkg/logger"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	fmt.Println("Starting gen_invite...")
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/utils/gen_invite.go <code>")
		os.Exit(1)
	}
	code := os.Args[1]

	fmt.Printf("Code: %s\n", code)

	fmt.Println("Before LoadConfig")
	// 1. إعدادات الاتصال
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("After LoadConfig")
	logger := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)

	fmt.Printf("DB_TYPE: %s\n", cfg.DBType)
	fmt.Printf("POSTGRES_DSN: %s\n", cfg.PostgresDSN)

	// Check for DEV_NO_DB
	if os.Getenv("DEV_NO_DB") == "1" {
		fmt.Println("DEV_NO_DB mode: skipping database operations")
		fmt.Printf("✅ تم إنشاء الكود بنجاح (simulated): %s\n", code)
		fmt.Println("End")
		return
	}

	// Initialize databases based on DB_TYPE
	mStore, pStore, err := container.InitializeDatabases(cfg, logger)
	if err != nil {
		fmt.Printf("Error initializing databases: %v\n", err)
		os.Exit(1)
	}

	// Initialize repositories
	repos := container.InitializeRepositories(cfg, logger, mStore, pStore)
	ctx := context.Background()

	fmt.Printf("User repo type: %T\n", repos.User)

	fmt.Println("--- إنشاء رمز دعوة ---")

	// 3. بناء الكائن باستخدام النوع الذي يفهمه الـ userRepo
	newInvite := &user.InviteCode{
		ID:        primitive.NewObjectID(),
		Code:      code,
		IsUsed:    false,
		CreatedAt: time.Now(),
		// زدنا المدة لسنة لتجنب مشكلة انتهاء الصلاحية التي واجهتها
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}

	// 4. الحفظ المباشر
	err = repos.User.CreateInvite(ctx, newInvite)
	if err != nil {
		fmt.Printf("Error creating invite: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ تم إنشاء الكود بنجاح: %s\n", code)
	fmt.Println("End")
	os.Exit(0)
}
