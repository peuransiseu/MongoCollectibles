package services

import (
	"context"
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/mongocollectibles/rental-system/models"
)

// AllocationManager handles the allocation of specific units to customers
type AllocationManager struct {
	inventory  []*models.CollectibleUnit
	warehouses map[string]models.WarehouseNode
	mu         sync.Mutex // Protects inventory from race conditions
}

// NewAllocationManager creates a new instance
func NewAllocationManager(inventory []*models.CollectibleUnit, warehouses []models.WarehouseNode) *AllocationManager {
	// Index warehouses for faster lookup
	whMap := make(map[string]models.WarehouseNode)
	for _, wh := range warehouses {
		whMap[wh.ID] = wh
	}

	return &AllocationManager{
		inventory:  inventory,
		warehouses: whMap,
	}
}

// Allocate selects the best available unit for a customer
// filtering by collectible type and finding the nearest warehouse.
// rentalID is used to track which rental reserved this unit.
func (am *AllocationManager) Allocate(collectibleID string, storeID string, rentalID string) (*models.CollectibleUnit, int, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// First pass: Release expired reservations before searching
	am.releaseExpiredReservationsUnsafe()

	log.Printf("[Allocation] Starting allocation for Collectible: %s at Store ID: %s (Rental: %s)", collectibleID, storeID, rentalID)

	var bestUnit *models.CollectibleUnit
	minDistance := math.MaxInt32
	found := false

	// Iterate through all units to find candidates
	for _, unit := range am.inventory {
		// Filter 1: Must match the requested collectible type
		if unit.CollectibleID != collectibleID {
			continue
		}

		// Filter 2: Must be available
		if !unit.IsAvailable {
			log.Printf("[Allocation] Skipping Unit %s (Warehouse %s): Already reserved", unit.ID, unit.WarehouseID)
			continue
		}

		// Find the warehouse for this unit
		warehouse, exists := am.warehouses[unit.WarehouseID]
		if !exists {
			log.Printf("[Allocation] Critical Error: Unit %s linked to unknown Warehouse %s", unit.ID, unit.WarehouseID)
			continue
		}

		// Validate store ID and get distance
		dist, ok := warehouse.Distances[storeID]
		if !ok {
			log.Printf("[Allocation] Error: Store ID %s not found for Warehouse %s", storeID, warehouse.ID)
			continue // Skip this warehouse if it doesn't serve the requested store
		}
		log.Printf("[Allocation] Candidate: Unit %s (Warehouse %s) - Distance: %d km", unit.ID, unit.WarehouseID, dist)

		// Select if this is the closest valid option found so far
		if dist < minDistance {
			minDistance = dist
			bestUnit = unit
			found = true
			log.Printf("[Allocation] -> New Best Candidate: %s", unit.ID)
		}
	}

	if !found {
		log.Printf("[Allocation] Failed: No available units found for Collectible %s", collectibleID)
		return nil, 0, errors.New("no available units found for the selected collectible")
	}

	// Reservation: Mark as unavailable and set reservation metadata
	now := time.Now()
	bestUnit.IsAvailable = false
	bestUnit.ReservedAt = &now
	bestUnit.ReservationID = rentalID
	log.Printf("[Allocation] Success: Allocated Unit %s from Warehouse %s (Distance: %d km, Rental: %s)", bestUnit.ID, bestUnit.WarehouseID, minDistance, rentalID)

	return bestUnit, minDistance, nil
}

// releaseExpiredReservationsUnsafe releases units with expired reservations
// Must be called with mutex already locked
func (am *AllocationManager) releaseExpiredReservationsUnsafe() {
	const reservationTimeout = 15 * time.Minute
	cutoff := time.Now().Add(-reservationTimeout)

	for _, unit := range am.inventory {
		if !unit.IsAvailable && unit.ReservedAt != nil {
			if unit.ReservedAt.Before(cutoff) {
				log.Printf("[Allocation] Auto-releasing expired reservation: Unit %s (Reserved at %v, Rental %s)",
					unit.ID, unit.ReservedAt, unit.ReservationID)
				unit.IsAvailable = true
				unit.ReservedAt = nil
				unit.ReservationID = ""
			}
		}
	}
}

// GetTotalStock returns the number of available units for a collectible
func (am *AllocationManager) GetTotalStock(collectibleID string) int {
	am.mu.Lock()
	defer am.mu.Unlock()

	count := 0
	for _, unit := range am.inventory {
		if unit.CollectibleID == collectibleID && unit.IsAvailable {
			count++
		}
	}
	log.Printf("[Allocation] Stock query for %s: %d available", collectibleID, count)
	return count
}

// GetETA calculates the estimated delivery time (minimum distance) for a collectible to a store
func (am *AllocationManager) GetETA(collectibleID string, storeID string) (int, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	minDistance := math.MaxInt32
	found := false

	for _, unit := range am.inventory {
		// Filter: Must be correct type and available
		if unit.CollectibleID != collectibleID || !unit.IsAvailable {
			continue
		}

		warehouse, exists := am.warehouses[unit.WarehouseID]
		if !exists {
			continue
		}

		dist, ok := warehouse.Distances[storeID]
		if !ok {
			continue
		}

		if dist < minDistance {
			minDistance = dist
			found = true
		}
	}

	if !found {
		log.Printf("[Allocation] ETA query failed for %s: no units available", collectibleID)
		return 0, errors.New("no units available for ETA calculation")
	}

	log.Printf("[Allocation] ETA query for %s at Store %s: %d days", collectibleID, storeID, minDistance)
	return minDistance, nil
}

// ReleaseUnit marks a unit as available again (e.g., when payment fails or is cancelled)
func (am *AllocationManager) ReleaseUnit(collectibleID string, warehouseID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, unit := range am.inventory {
		if unit.CollectibleID == collectibleID && unit.WarehouseID == warehouseID && !unit.IsAvailable {
			unit.IsAvailable = true
			unit.ReservedAt = nil
			unit.ReservationID = ""
			log.Printf("[Allocation] Released Unit %s from Warehouse %s back to inventory", unit.ID, warehouseID)
			return nil
		}
	}

	log.Printf("[Allocation] Warning: Could not find unavailable unit for Collectible %s at Warehouse %s", collectibleID, warehouseID)
	return errors.New("unit not found or already available")
}

// CleanupExpiredReservations releases all units with expired reservations
func (am *AllocationManager) CleanupExpiredReservations() {
	am.mu.Lock()
	defer am.mu.Unlock()

	const reservationTimeout = 15 * time.Minute
	cutoff := time.Now().Add(-reservationTimeout)
	releasedCount := 0

	for _, unit := range am.inventory {
		if !unit.IsAvailable && unit.ReservedAt != nil {
			if unit.ReservedAt.Before(cutoff) {
				log.Printf("[Cleanup] Releasing expired reservation: Unit %s (Reserved at %v, Rental %s)",
					unit.ID, unit.ReservedAt, unit.ReservationID)
				unit.IsAvailable = true
				unit.ReservedAt = nil
				unit.ReservationID = ""
				releasedCount++
			}
		}
	}

	if releasedCount > 0 {
		log.Printf("[Cleanup] Released %d expired reservations", releasedCount)
	}
}

// StartCleanupJob starts a background goroutine to periodically clean up expired reservations
func (am *AllocationManager) StartCleanupJob(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("[Allocation] Starting background cleanup job (runs every 5 minutes)")

	for {
		select {
		case <-ticker.C:
			am.CleanupExpiredReservations()
		case <-ctx.Done():
			log.Println("[Allocation] Cleanup job stopped")
			return
		}
	}
}
