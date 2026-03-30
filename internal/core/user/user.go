// internal/core/user.go
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User يمثل هيكل المستخدم في التطبيق وقاعدة البيانات.
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UUID     string             `bson:"uuid" json:"uuid"`
	UserID   string             `bson:"user_id" json:"user_id"`
	Username string             `bson:"username" json:"username" validate:"required"` // يوزر وليس إيميل
	Password string             `bson:"password" json:"-" validate:"required"`        // hashed password (hidden)
	//Gem     int                `bson:"gem" json:"gem"`

	// metadata fields that used to live on user_details after restructuring
	Roles                 []string   `bson:"roles" json:"roles"`
	IsActive              bool       `bson:"is_active" json:"is_active"` // active flag
	LastLoginAt           *time.Time `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	RefreshToken          string     `bson:"refresh_token,omitempty" json:"-"`
	RefreshTokenExpiresAt *time.Time `bson:"refresh_token_expires_at,omitempty" json:"-"`
	CreatedAt             time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `bson:"updated_at" json:"updated_at"`
}

type UserDetails struct {
	UUID     string      `bson:"uuid" json:"uuid"`
	UserID   string      `bson:"user_id" json:"user_id"`
	Profile  UserProfile `bson:"profile" json:"profile"`
	Status   string      `bson:"status" json:"status"` // (active, banned, pending)
	Roles    []string    `bson:"roles" json:"roles"`
	IsActive bool        `bson:"is_active" json:"is_active"` // active flag

	CreatedAt             time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `bson:"updated_at" json:"updated_at"`
	LastLoginAt           *time.Time `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	RefreshToken          string     `bson:"refresh_token,omitempty" json:"-"`
	RefreshTokenExpiresAt *time.Time `bson:"refresh_token_expires_at,omitempty" json:"-"`
}

type UserProfile struct {
	DisplayName string `bson:"display_name" json:"display_name"`
	AvatarURL   string `bson:"avatar_url" json:"avatar_url"`
	Bio         string `bson:"bio" json:"bio"`
}

// CollectionName هو اسم Collection في MongoDB لهذا النموذج.
const UserCollectionName = "users"
const UserDetailsCollectionName = "user_details"

// HasRole returns true if user has the given role.
func (u *UserDetails) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin checks whether the user has admin privileges.
func (u *UserDetails) IsAdmin() bool {
	return u.HasRole("admin")
}
