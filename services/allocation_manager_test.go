package services

import (
	"testing"

	"github.com/mongocollectibles/rental-system/models"
)

func TestAllocationManager_Allocate(t *testing.T) {
	// Setup Warehouses
	// W1: S1:10, S2:20
	// W2: S1:5, S2:25
	warehouses := []models.WarehouseNode{
		{ID: "1", Distances: map[string]int{"S1": 10, "S2": 20}},
		{ID: "2", Distances: map[string]int{"S1": 5, "S2": 25}},
	}

	// Setup Inventory
	units := []*models.CollectibleUnit{
		{ID: "U1", CollectibleID: "C1", WarehouseID: "1", IsAvailable: true},
		{ID: "U2", CollectibleID: "C1", WarehouseID: "2", IsAvailable: true},
		{ID: "U3", CollectibleID: "C2", WarehouseID: "1", IsAvailable: true},  // Diff collectible
		{ID: "U4", CollectibleID: "C1", WarehouseID: "1", IsAvailable: false}, // Unavailable
	}

	am := NewAllocationManager(units, warehouses)

	t.Run("Allocate nearest available", func(t *testing.T) {
		// Store S1: W2 (dist 5) is closer than W1 (dist 10)
		// Should pick U2
		got, _, err := am.Allocate("C1", "S1")
		if err != nil {
			t.Fatalf("Allocate failed: %v", err)
		}
		if got.ID != "U2" {
			t.Errorf("Expected unit U2 (dist 5), got %s", got.ID)
		}
		if got.IsAvailable {
			t.Error("Unit should be marked unavailable after allocation")
		}
	})

	t.Run("No available units (all booked or wrong type)", func(t *testing.T) {
		// Try allocating C2. Only U3 exists.
		// Let's allocate U3 first to make it unavailable.
		_, _, _ = am.Allocate("C2", "S1")

		// Now try allocating C2 again
		_, _, err := am.Allocate("C2", "S1")
		if err == nil {
			t.Error("Expected error for no available units, got nil")
		}
	})

	t.Run("Unknown store ID", func(t *testing.T) {
		_, _, err := am.Allocate("C1", "UNKNOWN")
		if err == nil {
			t.Error("Expected error for unknown store ID, got nil")
		}
	})
}

func TestAllocationManager_StockAndETA(t *testing.T) {
	// Setup Warehouses
	// W1: S1:1, S2:10
	// W2: S1:5, S2:20
	warehouses := []models.WarehouseNode{
		{ID: "1", Distances: map[string]int{"S1": 1, "S2": 10}},
		{ID: "2", Distances: map[string]int{"S1": 5, "S2": 20}},
	}

	// Setup Inventory
	units := []*models.CollectibleUnit{
		{ID: "U1", CollectibleID: "C1", WarehouseID: "1", IsAvailable: true},
		{ID: "U2", CollectibleID: "C1", WarehouseID: "2", IsAvailable: true},
		{ID: "U3", CollectibleID: "C1", WarehouseID: "1", IsAvailable: false}, // Unavailable
		{ID: "U4", CollectibleID: "C2", WarehouseID: "1", IsAvailable: true},  // Diff collectible
	}

	am := NewAllocationManager(units, warehouses)

	t.Run("GetTotalStock", func(t *testing.T) {
		// C1: U1, U2 available (2)
		if count := am.GetTotalStock("C1"); count != 2 {
			t.Errorf("Expected stock 2 for C1, got %d", count)
		}
		// C2: U4 available (1)
		if count := am.GetTotalStock("C2"); count != 1 {
			t.Errorf("Expected stock 1 for C2, got %d", count)
		}
		// C3: None (0)
		if count := am.GetTotalStock("C3"); count != 0 {
			t.Errorf("Expected stock 0 for C3, got %d", count)
		}
	})

	t.Run("GetETA", func(t *testing.T) {
		// C1 at Store S1
		// Available: U1(W1, dist 1), U2(W2, dist 5) -> Min 1
		dist, err := am.GetETA("C1", "S1")
		if err != nil {
			t.Fatalf("GetETA failed: %v", err)
		}
		if dist != 1 {
			t.Errorf("Expected ETA 1, got %d", dist)
		}

		// C1 at Store S2
		// Available: U1(W1, dist 10), U2(W2, dist 20) -> Min 10
		dist, err = am.GetETA("C1", "S2")
		if err != nil {
			t.Fatalf("GetETA failed: %v", err)
		}
		if dist != 10 {
			t.Errorf("Expected ETA 10, got %d", dist)
		}

		// Invalid Collectible
		_, err = am.GetETA("C3", "S1")
		if err == nil {
			t.Error("Expected error for ETA on missing collectible")
		}
	})
}
