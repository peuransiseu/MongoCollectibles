package models

import "time"

// User represents a registered user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"` // Session token
}
