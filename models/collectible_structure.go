package models

import (
	"time"
)

// CollectibleType represents the abstract product type a customer chooses
// type CollectibleType struct {
// 	ID   string
// 	Name string
// }

// CollectibleUnit represents a specific physical item in a warehouse
// This struct is internal-facing; the customer never sees the Unit ID.
type CollectibleUnit struct {
	ID            string
	CollectibleID string // Links to CollectibleType
	WarehouseID   string // Links to WarehouseNode
	IsAvailable   bool
	ReservedAt    *time.Time // When unit was reserved (nil if not reserved)
	ReservationID string     // Rental ID that reserved this unit
}
