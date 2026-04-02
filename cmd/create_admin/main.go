package main1

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/835-droid/ms-ai-backend/internal/container"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	plog "github.com/835-droid/ms-ai-backend/pkg/logger"
	"github.com/835-droid/ms-ai-backend/pkg/utils"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         MS-AI Admin User Creator         ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// 1. تحميل الـ config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}
	fmt.Printf("[✓] Config loaded successfully (DB_TYPE: %s)\n", cfg.DBType)

	// 2. إنشاء logger
	log := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)

	// 3. الاتصال بقواعد البيانات
	mongoStore, postgresStore, err := container.InitializeDatabases(cfg, log)
	if err != nil {
		fmt.Printf("❌ Failed to initialize databases: %v\n", err)
		return
	}

	if mongoStore != nil {
		fmt.Println("[✓] MongoDB connected successfully")
	}
	if postgresStore != nil {
		fmt.Println("[✓] PostgreSQL connected successfully")
	}

	// 4. إنشاء الـ repositories
	repos := container.InitializeRepositories(cfg, log, mongoStore, postgresStore)
	fmt.Println("[✓] Repositories initialized")

	// 5. إنشاء context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 6. التحقق من وجود admin مسبقاً
	fmt.Println()
	fmt.Println("[!] Checking if admin user already exists...")
	existingUser, err := repos.User.FindByUsername(ctx, "admin")
	if err != nil && err != core.ErrUserNotFound {
		fmt.Printf("❌ Failed to check admin user: %v\n", err)
		return
	}

	if existingUser != nil {
		fmt.Println("[!] Admin user already exists. Skipping creation.")
		fmt.Println()
		fmt.Println("══════════════════════════════════════════")
		fmt.Println("  Admin user already exists!")
		fmt.Println("══════════════════════════════════════════")
		return
	}

	// 7. توليد UUID و UserID
	userUUID := utils.GenerateUUID()
	userID := utils.GenerateUUID()

	// 8. Hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Failed to hash password: %v\n", err)
		return
	}

	// 9. بناء adminUser
	adminUser := &coreuser.User{
		UUID:     userUUID,
		UserID:   userID,
		Username: "admin",
		Password: string(hashedPass),
		UserBase: coreuser.UserBase{
			Roles:    coreuser.Roles{coreuser.RoleAdmin},
			IsActive: true,
		},
	}

	// 10. بناء adminDetails
	adminDetails := &coreuser.UserDetails{
		UUID:   userUUID,
		UserID: userID,
		UserBase: coreuser.UserBase{
			Roles:    coreuser.Roles{coreuser.RoleAdmin},
			IsActive: true,
		},
		Status: "active",
	}

	// 11. إنشاء المستخدم
	if err := repos.User.Create(ctx, adminUser, adminDetails); err != nil {
		if err == core.ErrUserExists {
			fmt.Println("[!] Admin user already exists. Skipping creation.")
			fmt.Println()
			fmt.Println("══════════════════════════════════════════")
			fmt.Println("  Admin user already exists!")
			fmt.Println("══════════════════════════════════════════")
			return
		}
		fmt.Printf("❌ Failed to create admin user: %v\n", err)
		return
	}

	fmt.Println("[✓] Admin user created successfully")
	fmt.Println()
	fmt.Printf("    Username : admin\n")
	fmt.Printf("    Password : admin123\n")
	fmt.Printf("    Roles    : [admin]\n")
	fmt.Printf("    UUID     : %s\n", userUUID)
	fmt.Println()

	// طباعة رسائل الـ DB المحفوظة حسب cfg.DBType
	if cfg.DBType == "hybrid" || cfg.DBType == "mongo" {
		fmt.Println("[✓] Saved to MongoDB")
	}
	if cfg.DBType == "hybrid" || cfg.DBType == "postgres" {
		fmt.Println("[✓] Saved to PostgreSQL")
	}

	// Invite code creation is currently disabled in this setup.
	fmt.Println("[⚠] Invite code generation is skipped (use admin dashboard or API to create invite codes).")

	// 14. إغلاق اتصال MongoDB
	if mongoStore != nil {
		if err := mongoStore.Close(ctx); err != nil {
			fmt.Printf("⚠️  Warning: Failed to close MongoDB connection: %v\n", err)
		}
	}

	fmt.Println("══════════════════════════════════════════")
	fmt.Println("  Done! You can now login with admin/admin123")
	fmt.Println("══════════════════════════════════════════")
}
