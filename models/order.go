package models

import "time"

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderPaid           OrderStatus = "PAID"
	OrderAllocated      OrderStatus = "ALLOCATED"
	OrderInTransit      OrderStatus = "IN_TRANSIT"
	OrderReadyForPickup OrderStatus = "READY_FOR_PICKUP"
	OrderCompleted      OrderStatus = "COMPLETED"
	OrderCancelled      OrderStatus = "CANCELLED"
	OrderRefunded       OrderStatus = "REFUNDED"
)

// Order represents a customer order
type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	StoreID     string      `json:"store_id"`
	Status      OrderStatus `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	PaymentID   string      `json:"payment_id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem represents an item in an order
// This contains the ALLOCATED unit details (not exposed to frontend)
type OrderItem struct {
	CollectibleID   string  `json:"collectible_id"`
	CollectibleName string  `json:"collectible_name"`
	UnitID          string  `json:"-"` // Internal only
	WarehouseID     string  `json:"-"` // Internal only
	RentalDays      int     `json:"rental_days"`
	ETADays         int     `json:"eta_days"`
	Price           float64 `json:"price"`
}
