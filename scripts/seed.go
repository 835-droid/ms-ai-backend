package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
		os.Exit(1)
	}
	log := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)
	store, err := mongo.NewMongoStore(cfg, log)
	if err != nil {
		log.Fatal("connect mongo failed", map[string]interface{}{"error": err.Error()})
	}
	repo := datauser.NewMongoUserRepository(store)
	// seed admin user if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	adminName := os.Getenv("SEED_ADMIN_USERNAME")
	if adminName == "" {
		adminName = "admin"
	}
	_, err = repo.FindByUsername(ctx, adminName)
	if err == nil {
		fmt.Println("admin user exists, skipping creation")
	}
	// create default invite codes
	codes := []string{"INITCODE1", "INITCODE2"}
	for _, c := range codes {
		inv := &coreuser.InviteCode{Code: c}
		if err := repo.CreateInviteCode(ctx, inv); err != nil {
			fmt.Printf("seed invite code %s failed: %v\n", c, err)
			continue
		}
		fmt.Printf("seeded invite code: %s\n", c)
	}
	fmt.Println("seeding completed")
}
