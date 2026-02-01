package services

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/mongocollectibles/rental-system/models"
)

// WarehouseManager handles the validation and creation of warehouse networks
type WarehouseManager struct{}

// NewWarehouseManager creates a new instance
func NewWarehouseManager() *WarehouseManager {
	return &WarehouseManager{}
}

// ValidateStoreCount ensures we have the required minimum number of stores
func (wm *WarehouseManager) ValidateStoreCount(stores []models.StoreNode) error {
	const minStores = 3
	if len(stores) < minStores {
		return fmt.Errorf("insufficient stores: expected at least %d, got %d", minStores, len(stores))
	}
	return nil
}

// ValidateDistanceMap ensures all active stores have valid distance entries
func (wm *WarehouseManager) ValidateDistanceMap(distances map[string]int, stores []models.Store) error {
	// Ensure all active stores have distance entries
	for _, store := range stores {
		dist, exists := distances[store.ID]
		if !exists {
			return fmt.Errorf("missing distance for store %s", store.ID)
		}
		if dist < 0 {
			return fmt.Errorf("negative distance %d for store %s", dist, store.ID)
		}
	}
	return nil
}

// ConvertDistances transforms a distance tuple into a map using the provided store IDs
func (wm *WarehouseManager) ConvertDistances(tuple []int, stores []models.Store) (map[string]int, error) {
	if len(tuple) != len(stores) {
		return nil, fmt.Errorf("tuple length %d does not match store count %d", len(tuple), len(stores))
	}

	distMap := make(map[string]int)
	for i, dist := range tuple {
		distMap[stores[i].ID] = dist
	}
	return distMap, nil
}

// BuildWarehouses constructs warehouse entities from raw distance tuples using store definitions
// This function maintains backward compatibility by accepting tuples and converting them to maps
func (wm *WarehouseManager) BuildWarehouses(distanceTuples [][]int, stores []models.Store) ([]models.WarehouseNode, error) {
	log.Printf("[WarehouseManager] Building network for %d stores with %d warehouse tuples defined", len(stores), len(distanceTuples))
	warehouses := make([]models.WarehouseNode, 0, len(distanceTuples))

	for i, distances := range distanceTuples {
		// Convert tuple to map
		distMap, err := wm.ConvertDistances(distances, stores)
		if err != nil {
			log.Printf("[WarehouseManager] Conversion failed for warehouse index %d: %v", i, err)
			return nil, fmt.Errorf("error in warehouse definition at index %d: %w", i, err)
		}

		// Validate the distance map
		if err := wm.ValidateDistanceMap(distMap, stores); err != nil {
			log.Printf("[WarehouseManager] Validation failed for warehouse index %d: %v", i, err)
			return nil, fmt.Errorf("invalid distances for warehouse at index %d: %w", i, err)
		}

		warehouse := models.WarehouseNode{
			ID:        fmt.Sprintf("%d", i+1), // Simple auto-increment ID
			Distances: distMap,
		}
		warehouses = append(warehouses, warehouse)
		log.Printf("[WarehouseManager] Created Warehouse %s with map: %v", warehouse.ID, distMap)
	}

	return warehouses, nil
}

// FindNearestWarehouse returns the warehouse closest to a specific store ID
func (wm *WarehouseManager) FindNearestWarehouse(warehouses []models.WarehouseNode, storeID string) (*models.WarehouseNode, int, error) {
	if len(warehouses) == 0 {
		return nil, 0, errors.New("no warehouses available")
	}

	var nearest *models.WarehouseNode
	minDist := math.MaxInt32

	for i := range warehouses {
		w := &warehouses[i]

		dist, ok := w.Distances[storeID]
		if !ok {
			// If a warehouse doesn't have a record for this store ID, skip it
			continue
		}

		if dist < minDist {
			minDist = dist
			nearest = w
		}
	}

	if nearest == nil {
		return nil, 0, fmt.Errorf("could not find a nearest warehouse for store %s", storeID)
	}

	return nearest, minDist, nil
}
