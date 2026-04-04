// Package dtos defines Data Transfer Objects for API communication.
package dtos

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	Username   string `json:"username" validate:"required,min=3,max=50"`
	InviteCode string `json:"invite_code,omitempty"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents an authentication response.
type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// UserDTO represents a user data transfer object.
type UserDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar,omitempty"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// InviteCodeResponse represents an invite code response.
type InviteCodeResponse struct {
	Code      string `json:"code"`
	CreatedBy string `json:"created_by"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}
