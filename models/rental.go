package models

import "time"

// PaymentMethod represents the available payment options
type PaymentMethod string

const (
	PaymentCard        PaymentMethod = "card"
	PaymentGCash       PaymentMethod = "gcash"
	PaymentGrabPay     PaymentMethod = "grabpay"
	PaymentBPI         PaymentMethod = "bpi"
	PaymentUBP         PaymentMethod = "ubp"
)

// PaymentStatus represents the current status of a payment
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
)

// Customer represents customer information
type Customer struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	PostalCode   string `json:"postal_code"`
}

// Rental represents a rental transaction
type Rental struct {
	ID              string        `json:"id"`
	CollectibleID   string        `json:"collectible_id"`
	CollectibleName string        `json:"collectible_name"`
	StoreID         string        `json:"store_id"`
	WarehouseID     string        `json:"warehouse_id"`
	Customer        Customer      `json:"customer"`
	Duration        int           `json:"duration"` // in days
	DailyRate       float64       `json:"daily_rate"`
	TotalFee        float64       `json:"total_fee"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	PaymentID       string        `json:"payment_id"`
	PaymentURL      string        `json:"payment_url"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// RentalQuoteRequest represents a request for rental fee calculation
type RentalQuoteRequest struct {
	CollectibleID string `json:"collectible_id"`
	Duration      int    `json:"duration"`
}

// RentalQuoteResponse represents the calculated rental quote
type RentalQuoteResponse struct {
	CollectibleID   string  `json:"collectible_id"`
	CollectibleName string  `json:"collectible_name"`
	Size            Size    `json:"size"`
	Duration        int     `json:"duration"`
	DailyRate       float64 `json:"daily_rate"`
	TotalFee        float64 `json:"total_fee"`
	IsSpecialRate   bool    `json:"is_special_rate"`
}

// CheckoutRequest represents a checkout request
type CheckoutRequest struct {
	CollectibleID string        `json:"collectible_id"`
	StoreID       string        `json:"store_id"`
	Duration      int           `json:"duration"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Customer      Customer      `json:"customer"`
}

// CheckoutResponse represents the checkout response
type CheckoutResponse struct {
	RentalID   string  `json:"rental_id"`
	TotalFee   float64 `json:"total_fee"`
	PaymentURL string  `json:"payment_url"`
	Message    string  `json:"message"`
}
