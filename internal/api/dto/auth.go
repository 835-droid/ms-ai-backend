// ----- START OF FILE: backend/MS-AI/internal/api/dto/auth.go -----
// internal/api/dto/auth.go
package dto

// SignUpRequestDTO represents the signup request payload
type SignUpRequestDTO struct {
	Username   string `json:"username" binding:"required,min=3,max=30,alphanum"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
	InviteCode string `json:"invite_code" binding:"required,len=8"`
}

// LoginRequestDTO represents the login request payload
type LoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponseDTO represents the authentication response
type AuthResponseDTO struct {
	Success      bool     `json:"success"`
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	User         *UserDTO `json:"user,omitempty"`
	Error        string   `json:"error,omitempty"`
	RequestID    string   `json:"request_id"`
}

// UserDTO represents user data in responses
type UserDTO struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

//
