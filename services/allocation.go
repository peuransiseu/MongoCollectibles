package services

import (
	"errors"
	"math"

	"github.com/mongocollectibles/rental-system/models"
)

// AllocationService handles warehouse allocation logic
type AllocationService struct {
	warehouses map[string][]models.Warehouse // collectibleID -> warehouses
}

// NewAllocationService creates a new allocation service
func NewAllocationService() *AllocationService {
	return &AllocationService{
		warehouses: make(map[string][]models.Warehouse),
	}
}

// SetWarehouses sets the warehouse data for the service
func (s *AllocationService) SetWarehouses(warehouses map[string][]models.Warehouse) {
	s.warehouses = warehouses
}

// AllocateWarehouse finds the nearest available warehouse for a collectible at a given store
// Returns the warehouse ID and distance, or an error if no warehouse is available
func (s *AllocationService) AllocateWarehouse(collectibleID string, storeIndex int) (string, int, error) {
	warehouses, exists := s.warehouses[collectibleID]
	if !exists {
		return "", 0, errors.New("collectible not found")
	}

	var nearestWarehouse *models.Warehouse
	minDistance := math.MaxInt32

	// Find the nearest available warehouse
	for i := range warehouses {
		warehouse := &warehouses[i]
		
		// Skip unavailable warehouses
		if !warehouse.Available {
			continue
		}

		// Check if store index is valid
		if storeIndex >= len(warehouse.DistancesToStores) {
			continue
		}

		distance := warehouse.DistancesToStores[storeIndex]
		
		// Update if this warehouse is closer
		if distance < minDistance {
			minDistance = distance
			nearestWarehouse = warehouse
		}
	}

	if nearestWarehouse == nil {
		return "", 0, errors.New("no available warehouse found for this collectible")
	}

	return nearestWarehouse.ID, minDistance, nil
}

// MarkWarehouseUnavailable marks a warehouse as unavailable after allocation
func (s *AllocationService) MarkWarehouseUnavailable(collectibleID, warehouseID string) error {
	warehouses, exists := s.warehouses[collectibleID]
	if !exists {
		return errors.New("collectible not found")
	}

	for i := range warehouses {
		if warehouses[i].ID == warehouseID {
			warehouses[i].Available = false
			return nil
		}
	}

	return errors.New("warehouse not found")
}

// MarkWarehouseAvailable marks a warehouse as available (for returns/cancellations)
func (s *AllocationService) MarkWarehouseAvailable(collectibleID, warehouseID string) error {
	warehouses, exists := s.warehouses[collectibleID]
	if !exists {
		return errors.New("collectible not found")
	}

	for i := range warehouses {
		if warehouses[i].ID == warehouseID {
			warehouses[i].Available = true
			return nil
		}
	}

	return errors.New("warehouse not found")
}
