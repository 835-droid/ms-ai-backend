// Package user defines the core domain entities for user management.
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents a user's role in the system.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a user in the system.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username" validate:"required,min=3,max=50"`
	Email     string             `bson:"email" json:"email" validate:"required,email"`
	Password  string             `bson:"password" json:"-"` // Never expose password in JSON
	Avatar    string             `bson:"avatar" json:"avatar,omitempty"`
	Role      Role               `bson:"role" json:"role"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
	return string(u.Role) == role
}

// IsAdmin checks if the user is an admin.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
