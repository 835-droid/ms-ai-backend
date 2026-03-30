package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/core/user" // الـ Struct الذي يستخدمه الـ Repo
	"github.com/835-droid/ms-ai-backend/internal/data/mongo"
	datauser "github.com/835-droid/ms-ai-backend/internal/data/user" // تأكد من المسار
	"github.com/835-droid/ms-ai-backend/pkg/config"
	plog "github.com/835-droid/ms-ai-backend/pkg/logger"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// 1. إعدادات الاتصال
	cfg, _ := config.LoadConfig()
	logger := plog.NewLogger(cfg.LogLevel, cfg.Environment, nil)
	store, err := mongo.NewMongoStore(cfg, logger)
	if err != nil {
		log.Fatalf("❌ فشل الاتصال: %v", err)
	}

	// 2. الوصول المباشر للـ Repository لتجنب تعارض الـ Interface
	userRepo := datauser.NewMongoUserRepository(store)
	ctx := context.Background()

	fmt.Println("--- إنشاء رمز دعوة (تجاوز تعارض الأنواع) ---")
	fmt.Print(" (8خانات على الاقل)أدخل الرمز: ")
	var code string
	fmt.Scanln(&code)

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
	err = userRepo.CreateInvite(ctx, newInvite)
	if err != nil {
		log.Fatalf("❌ فشل الحفظ: %v", err)
	}

	fmt.Printf("✅ تم إنشاء الكود بنجاح: %s\n", code)
}
