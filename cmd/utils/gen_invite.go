// ----- START OF FILE: backend/MS-AI/cmd/utils/gen_invite.go -----
// cmd/utils/gen_invite.go
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

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	logger := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)

	// DEV_NO_DB mode: skip DB operations
	if os.Getenv("DEV_NO_DB") == "1" {
		fmt.Println("DEV_NO_DB mode: skipping database operations")
		fmt.Printf("✅ تم إنشاء الكود بنجاح (simulated): %s\n", code)
		fmt.Println("End")
		return
	}

	// Create application container (handles DB initialization, repos, services)
	appContainer, err := container.NewContainer(cfg, logger)
	if err != nil {
		fmt.Printf("Error creating container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := appContainer.Close(shutdownCtx); err != nil {
			fmt.Printf("Warning: error closing container: %v\n", err)
		}
	}()

	ctx := context.Background()
	userRepo := appContainer.UserRepo

	fmt.Println("--- إنشاء رمز دعوة ---")

	newInvite := &user.InviteCode{
		ID:        primitive.NewObjectID(),
		Code:      code,
		IsUsed:    false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}

	if err := userRepo.CreateInvite(ctx, newInvite); err != nil {
		fmt.Printf("Error creating invite: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ تم إنشاء الكود بنجاح: %s\n", code)
	fmt.Println("End")
}

// ----- END OF FILE: backend/MS-AI/cmd/utils/gen_invite.go -----
