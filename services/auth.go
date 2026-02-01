package services

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct{}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	return &AuthService{}
}

// HashPassword hashes a plain text password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against a hash
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken generates a random session token
func (s *AuthService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
