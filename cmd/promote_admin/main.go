package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
	"github.com/835-droid/ms-ai-backend/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: promote_admin <username>")
		fmt.Println("This tool promotes a user to admin role in MongoDB")
		os.Exit(1)
	}

	username := os.Args[1]

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := mongo.NewMongoStore(cfg, nil)
	if err != nil {
		fmt.Printf("Error connecting to MongoDB: %v\n", err)
		os.Exit(1)
	}

	usersColl := store.GetCollection("users")

	// Find the user
	var existingUser user.User
	err = usersColl.FindOne(ctx, bson.M{"username": username}).Decode(&existingUser)
	if err != nil {
		fmt.Printf("Error: User '%s' not found\n", username)
		os.Exit(1)
	}

	// Check if already admin
	isAdmin := false
	for _, role := range existingUser.Roles {
		if role == user.RoleAdmin {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		fmt.Printf("User '%s' is already an admin\n", username)
	} else {
		// Add admin role
		newRoles := append(existingUser.Roles, user.RoleAdmin)

		_, err = usersColl.UpdateOne(
			ctx,
			bson.M{"_id": existingUser.ID},
			bson.M{
				"$set": bson.M{
					"roles":      newRoles,
					"updated_at": time.Now(),
					"is_active":  true,
				},
			},
		)
		if err != nil {
			fmt.Printf("Error updating user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Successfully promoted user '%s' to admin\n", username)
	}

	// Verify the update
	var updatedUser user.User
	err = usersColl.FindOne(ctx, bson.M{"_id": existingUser.ID}).Decode(&updatedUser)
	if err == nil {
		fmt.Printf("User roles: %v\n", updatedUser.Roles)
		fmt.Printf("User active: %v\n", updatedUser.IsActive)
	}

	fmt.Println("\nNote: The user needs to log in again to get a new token with admin privileges.")
	fmt.Println("Or wait for the next token refresh.")
}
