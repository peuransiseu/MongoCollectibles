package models

import "time"

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundPending   RefundStatus = "PENDING"
	RefundProcessed RefundStatus = "PROCESSED"
	RefundFailed    RefundStatus = "FAILED"
)

// Refund represents a refund for an order
type Refund struct {
	ID        string       `json:"id"`
	OrderID   string       `json:"order_id"`
	UserID    string       `json:"user_id"`
	Amount    float64      `json:"amount"`
	Reason    string       `json:"reason"`
	Status    RefundStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}
