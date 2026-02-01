package services

import (
	"testing"

	"github.com/mongocollectibles/rental-system/models"
)

func TestWarehouseManager(t *testing.T) {
	wm := NewWarehouseManager()

	t.Run("ValidateStoreCount", func(t *testing.T) {
		tests := []struct {
			name    string
			stores  []models.StoreNode
			wantErr bool
		}{
			{
				name:    "Valid number of stores (3)",
				stores:  make([]models.StoreNode, 3),
				wantErr: false,
			},
			{
				name:    "Valid number of stores (5)",
				stores:  make([]models.StoreNode, 5),
				wantErr: false,
			},
			{
				name:    "Insufficient stores (2)",
				stores:  make([]models.StoreNode, 2),
				wantErr: true,
			},
			{
				name:    "No stores",
				stores:  []models.StoreNode{},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := wm.ValidateStoreCount(tt.stores); (err != nil) != tt.wantErr {
					t.Errorf("ValidateStoreCount() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateDistanceMap", func(t *testing.T) {
		stores := []models.Store{
			{ID: "store-a"},
			{ID: "store-b"},
			{ID: "store-c"},
		}

		tests := []struct {
			name      string
			distances map[string]int
			wantErr   bool
		}{
			{
				name: "Valid distance map",
				distances: map[string]int{
					"store-a": 1,
					"store-b": 5,
					"store-c": 2,
				},
				wantErr: false,
			},
			{
				name: "Missing store distance",
				distances: map[string]int{
					"store-a": 1,
					"store-b": 5,
				},
				wantErr: true,
			},
			{
				name: "Negative distance",
				distances: map[string]int{
					"store-a": 1,
					"store-b": -5,
					"store-c": 2,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := wm.ValidateDistanceMap(tt.distances, stores); (err != nil) != tt.wantErr {
					t.Errorf("ValidateDistanceMap() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ConvertDistances", func(t *testing.T) {
		stores := []models.Store{
			{ID: "A"}, {ID: "B"}, {ID: "C"},
		}
		tuple := []int{1, 4, 5}

		got, err := wm.ConvertDistances(tuple, stores)
		if err != nil {
			t.Fatalf("ConvertDistances failed: %v", err)
		}

		expected := map[string]int{"A": 1, "B": 4, "C": 5}
		if len(got) != len(expected) {
			t.Errorf("got map len %d, want %d", len(got), len(expected))
		}
		for k, v := range expected {
			if got[k] != v {
				kStr := string(k)
				t.Errorf("key %s: got %d, want %d", kStr, got[kStr], v)
			}
		}

		// Length mismatch
		_, err = wm.ConvertDistances([]int{1, 2}, stores)
		if err == nil {
			t.Error("Expected error for length mismatch, got nil")
		}
	})

	t.Run("FindNearestWarehouse", func(t *testing.T) {
		// Setup: 3 stores, 2 warehouses
		// W1: A:1, B:10, C:5
		// W2: A:8, B:2, C:8
		warehouses := []models.WarehouseNode{
			{ID: "1", Distances: map[string]int{"A": 1, "B": 10, "C": 5}},
			{ID: "2", Distances: map[string]int{"A": 8, "B": 2, "C": 8}},
		}

		tests := []struct {
			name     string
			storeID  string
			wantID   string
			wantDist int
			wantErr  bool
		}{
			{
				name:     "Store A (W1 closest)",
				storeID:  "A",
				wantID:   "1",
				wantDist: 1,
				wantErr:  false,
			},
			{
				name:     "Store B (W2 closest)",
				storeID:  "B",
				wantID:   "2",
				wantDist: 2,
				wantErr:  false,
			},
			{
				name:     "Store C (W1 closest)",
				storeID:  "C",
				wantID:   "1",
				wantDist: 5,
				wantErr:  false,
			},
			{
				name:     "Unknown Store ID",
				storeID:  "D",
				wantID:   "",
				wantDist: 0,
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, dist, err := wm.FindNearestWarehouse(warehouses, tt.storeID)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindNearestWarehouse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if got.ID != tt.wantID {
						t.Errorf("FindNearestWarehouse() got ID = %v, want %v", got.ID, tt.wantID)
					}
					if dist != tt.wantDist {
						t.Errorf("FindNearestWarehouse() got dist = %v, want %v", dist, tt.wantDist)
					}
				}
			})
		}
	})

	t.Run("BuildWarehouses", func(t *testing.T) {
		stores := []models.Store{
			{ID: "A"}, {ID: "B"}, {ID: "C"},
		}
		tuples := [][]int{
			{1, 2, 3},
			{4, 5, 6},
		}

		warehouses, err := wm.BuildWarehouses(tuples, stores)
		if err != nil {
			t.Fatalf("BuildWarehouses failed: %v", err)
		}

		if len(warehouses) != 2 {
			t.Errorf("Expected 2 warehouses, got %d", len(warehouses))
		}

		// Verify map population
		if warehouses[0].Distances["A"] != 1 || warehouses[1].Distances["C"] != 6 {
			t.Error("Warehouse distances map not populated correctly in BuildWarehouses")
		}
	})
}
