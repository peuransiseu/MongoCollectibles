package models

import "time"

// CartStatus represents the status of a shopping cart
type CartStatus string

const (
	CartActive     CartStatus = "ACTIVE"
	CartCheckedOut CartStatus = "CHECKED_OUT"
	CartAbandoned  CartStatus = "ABANDONED"
)

// Cart represents a user's shopping cart
type Cart struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Status    CartStatus `json:"status"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CartItem represents an item in the cart
// NOTE: This is INTENT ONLY - no allocation or stock reservation
type CartItem struct {
	CollectibleID string `json:"collectible_id"`
	StoreID       string `json:"store_id"`
	RentalDays    int    `json:"rental_days"`
	Quantity      int    `json:"quantity"` // Desired quantity, not reserved
}

// AddToCartRequest represents a request to add an item to cart
type AddToCartRequest struct {
	CollectibleID string `json:"collectible_id"`
	StoreID       string `json:"store_id"`
	RentalDays    int    `json:"rental_days"`
	Quantity      int    `json:"quantity"`
}

// UpdateCartItemRequest represents a request to update a cart item
type UpdateCartItemRequest struct {
	RentalDays int `json:"rental_days"`
	Quantity   int `json:"quantity"`
}
