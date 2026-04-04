package user

import "time"

// InviteCode represents an invitation code for user registration.
type InviteCode struct {
	Code      string    `json:"code"`
	CreatedBy string    `json:"created_by"`
	UsedBy    string    `json:"used_by,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UsedAt    time.Time `json:"used_at,omitempty"`
}
