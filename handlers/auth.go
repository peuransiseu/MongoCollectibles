package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	repo        *data.Repository
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(repo *data.Repository, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Email and password are required",
		})
		return
	}

	// Hash password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to process registration",
		})
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.repo.CreateUser(user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Generate session token
	token, err := h.authService.GenerateToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	// Create session
	h.repo.CreateSession(user.ID, token)

	log.Printf("[Auth] User registered: %s", user.Email)

	// Return response
	response := models.AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Token:  token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Get user by email
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Check password
	if !h.authService.CheckPassword(req.Password, user.PasswordHash) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Generate session token
	token, err := h.authService.GenerateToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	// Create session
	h.repo.CreateSession(user.ID, token)

	log.Printf("[Auth] User logged in: %s", user.Email)

	// Return response
	response := models.AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Token:  token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from header
	token := r.Header.Get("Authorization")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No authorization token provided",
		})
		return
	}

	// Delete session
	h.repo.DeleteSession(token)

	log.Printf("[Auth] User logged out")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetUserFromToken is a helper to extract user ID from request
func (h *AuthHandler) GetUserFromToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", http.ErrNoCookie
	}

	return h.repo.GetUserByToken(token)
}
