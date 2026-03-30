package utils

import "github.com/google/uuid"

// GenerateUUID يولد معرفاً فريداً عشوائياً طويل جداً
func GenerateUUID() string {
	return uuid.New().String()
}
