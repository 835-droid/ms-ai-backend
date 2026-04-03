// ----- START OF FILE: backend/MS-AI/internal/core/user/user.go -----
// internal/core/user.go
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserBase contains common fields shared between User and UserDetails
type UserBase struct {
	Roles                 Roles      `bson:"roles" json:"roles"`
	IsActive              bool       `bson:"is_active" json:"is_active"`
	LastLoginAt           *time.Time `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	RefreshToken          string     `bson:"refresh_token,omitempty" json:"-"`
	RefreshTokenExpiresAt *time.Time `bson:"refresh_token_expires_at,omitempty" json:"-"`
	CreatedAt             time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `bson:"updated_at" json:"updated_at"`
}

// User يمثل هيكل المستخدم في التطبيق وقاعدة البيانات.
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UUID     string             `bson:"uuid" json:"uuid"`
	UserID   string             `bson:"user_id" json:"user_id"`
	Username string             `bson:"username" json:"username" validate:"required"` // يوزر وليس إيميل
	Password string             `bson:"password" json:"-" validate:"required"`        // hashed password (hidden)
	UserBase                    // embedded
}

type UserDetails struct {
	UUID                  string      `bson:"uuid" json:"uuid"`
	UserID                string      `bson:"user_id" json:"user_id"`
	Profile               UserProfile `bson:"profile" json:"profile"`
	Status                string      `bson:"status" json:"status"` // (active, banned, pending)
	LastLoginAt           *time.Time  `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	RefreshToken          string      `bson:"refresh_token,omitempty" json:"-"`
	RefreshTokenExpiresAt *time.Time  `bson:"refresh_token_expires_at,omitempty" json:"-"`
	CreatedAt             time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time   `bson:"updated_at" json:"updated_at"`
}

type UserProfile struct {
	DisplayName string `bson:"display_name" json:"display_name"`
	AvatarURL   string `bson:"avatar_url" json:"avatar_url"`
	Bio         string `bson:"bio" json:"bio"`
}

// HasRole returns true if user has the given role.
func (u *User) HasRole(role Role) bool {
	return u.Roles.Has(role)
}

// IsAdmin checks whether the user has admin privileges.
func (u *User) IsAdmin() bool {
	return u.Roles.IsAdmin()
}

// IsModerator checks whether the user has moderator privileges.
func (u *User) IsModerator() bool {
	return u.Roles.IsModerator()
}

// CollectionName هو اسم Collection في MongoDB لهذا النموذج.
const UserCollectionName = "users"
const UserDetailsCollectionName = "user_details"

// ----- END OF FILE: backend/MS-AI/internal/core/user/user.go -----
