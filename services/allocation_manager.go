package services

import (
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
func (am *AllocationManager) Allocate(collectibleID string, storeID string) (*models.CollectibleUnit, int, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	log.Printf("[Allocation] Starting allocation for Collectible: %s at Store ID: %s", collectibleID, storeID)

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

	// Reservation: Mark as unavailable immediately with timestamp
	bestUnit.IsAvailable = false
	now := time.Now()
	bestUnit.ReservedAt = &now
	// reservationID isn't passed here in the current signature, we can add it later or rely on IsAvailable=false + timestamp
	// Ideally, Allocate should take a reservationID or return one. For now, we rely on the timestamp.

	log.Printf("[Reservation] Temporary reservation created for Unit %s (Expires in 10m)", bestUnit.ID)
	log.Printf("[Allocation] Success: Allocated Unit %s from Warehouse %s (Distance: %d km)", bestUnit.ID, bestUnit.WarehouseID, minDistance)

	return bestUnit, minDistance, nil
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
			log.Printf("[Allocation] Released Unit %s from Warehouse %s back to inventory", unit.ID, warehouseID)
			return nil
		}
	}

	log.Printf("[Allocation] Warning: Could not find unavailable unit for Collectible %s at Warehouse %s", collectibleID, warehouseID)
	return errors.New("unit not found or already available")
}

// CleanupExpiredReservations releases units that have been reserved longer than the timeout
func (am *AllocationManager) CleanupExpiredReservations(timeout time.Duration) {
	am.mu.Lock()
	defer am.mu.Unlock()

	cutoff := time.Now().Add(-timeout)
	count := 0

	for _, unit := range am.inventory {
		if !unit.IsAvailable && unit.ReservedAt != nil {
			if unit.ReservedAt.Before(cutoff) {
				unit.IsAvailable = true
				unit.ReservedAt = nil
				unit.ReservationID = ""
				count++
				log.Printf("[Cleanup] Released expired reservation for unit %s", unit.ID)
			}
		}
	}
	if count > 0 {
		log.Printf("[Cleanup] Released %d expired reservations", count)
	}
}

// StartCleanupJob starts a background goroutine to clean up expired reservations
func (am *AllocationManager) StartCleanupJob(interval time.Duration, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			am.CleanupExpiredReservations(timeout)
		}
	}()
	log.Printf("[Allocation] Started cleanup job (Interval: %v, Timeout: %v)", interval, timeout)
}
