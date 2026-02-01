package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
)

// CartHandler handles cart operations
type CartHandler struct {
	repo        *data.Repository
	authHandler *AuthHandler
}

// NewCartHandler creates a new cart handler
func NewCartHandler(repo *data.Repository, authHandler *AuthHandler) *CartHandler {
	return &CartHandler{
		repo:        repo,
		authHandler: authHandler,
	}
}

// GetCart returns the user's active cart
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get or create active cart
	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		// Create new cart if none exists
		cart = &models.Cart{
			ID:        uuid.New().String(),
			UserID:    userID,
			Status:    models.CartActive,
			Items:     []models.CartItem{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		h.repo.CreateCart(cart)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    cart,
	})
}

// AddToCart adds an item to the cart
// NOTE: No stock validation - cart is intent only
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	var req models.AddToCartRequest
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
	if req.CollectibleID == "" || req.StoreID == "" || req.RentalDays <= 0 || req.Quantity <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid cart item data",
		})
		return
	}

	// Get or create cart
	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		cart = &models.Cart{
			ID:        uuid.New().String(),
			UserID:    userID,
			Status:    models.CartActive,
			Items:     []models.CartItem{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		h.repo.CreateCart(cart)
	}

	// Add item (no stock check - intent only)
	item := models.CartItem{
		CollectibleID: req.CollectibleID,
		StoreID:       req.StoreID,
		RentalDays:    req.RentalDays,
		Quantity:      req.Quantity,
	}

	h.repo.AddCartItem(cart.ID, item)
	cart.UpdatedAt = time.Now()
	h.repo.UpdateCart(cart)

	log.Printf("[Cart] Added item to cart: User %s, Collectible %s, Quantity %d", userID, req.CollectibleID, req.Quantity)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Item added to cart",
		"data":    cart,
	})
}

// UpdateCartItem updates a cart item
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	vars := mux.Vars(r)
	collectibleID := vars["collectible_id"]

	var req models.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cart not found",
		})
		return
	}

	// Find and update item
	var updatedItem models.CartItem
	for _, item := range cart.Items {
		if item.CollectibleID == collectibleID {
			updatedItem = models.CartItem{
				CollectibleID: item.CollectibleID,
				StoreID:       item.StoreID,
				RentalDays:    req.RentalDays,
				Quantity:      req.Quantity,
			}
			break
		}
	}

	if updatedItem.CollectibleID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Item not found in cart",
		})
		return
	}

	h.repo.UpdateCartItem(cart.ID, collectibleID, updatedItem)
	cart.UpdatedAt = time.Now()
	h.repo.UpdateCart(cart)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cart item updated",
		"data":    cart,
	})
}

// RemoveFromCart removes an item from the cart
func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	vars := mux.Vars(r)
	collectibleID := vars["collectible_id"]

	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cart not found",
		})
		return
	}

	if err := h.repo.RemoveCartItem(cart.ID, collectibleID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	cart.UpdatedAt = time.Now()
	h.repo.UpdateCart(cart)

	log.Printf("[Cart] Removed item from cart: User %s, Collectible %s", userID, collectibleID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Item removed from cart",
		"data":    cart,
	})
}

// ClearCart clears all items from the cart
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cart not found",
		})
		return
	}

	cart.Items = []models.CartItem{}
	cart.UpdatedAt = time.Now()
	h.repo.UpdateCart(cart)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cart cleared",
		"data":    cart,
	})
}
