package data

import (
	"github.com/mongocollectibles/rental-system/models"
)

// SeedData populates the repository with sample data
func SeedData(repo *Repository) {
	// Seed collectibles
	collectibles := []*models.Collectible{
		{
			ID:          "col-001",
			Name:        "Vintage Batman Action Figure",
			Description: "Rare 1989 Batman action figure in mint condition",
			Size:        models.SizeSmall,
			ImageURL:    "/images/batman.jpg",
			Available:   true,
		},
		{
			ID:          "col-002",
			Name:        "Star Wars Millennium Falcon Model",
			Description: "Detailed replica of the iconic spaceship",
			Size:        models.SizeMedium,
			ImageURL:    "/images/falcon.jpg",
			Available:   true,
		},
		{
			ID:          "col-003",
			Name:        "Life-Size Iron Man Suit",
			Description: "Full-scale Mark 42 armor replica",
			Size:        models.SizeLarge,
			ImageURL:    "/images/ironman.jpg",
			Available:   true,
		},
		{
			ID:          "col-004",
			Name:        "Pokemon Card Collection Set",
			Description: "Complete first edition holographic set",
			Size:        models.SizeSmall,
			ImageURL:    "/images/pokemon.jpg",
			Available:   true,
		},
		{
			ID:          "col-005",
			Name:        "Gundam Perfect Grade Model",
			Description: "RX-78-2 Gundam 1/60 scale model kit",
			Size:        models.SizeMedium,
			ImageURL:    "/images/gundam.jpg",
			Available:   true,
		},
		{
			ID:          "col-006",
			Name:        "Arcade Machine - Street Fighter II",
			Description: "Original 1991 arcade cabinet, fully functional",
			Size:        models.SizeLarge,
			ImageURL:    "/images/arcade.jpg",
			Available:   true,
		},
	}

	for _, c := range collectibles {
		repo.AddCollectible(c)
	}

	// Seed warehouses with distance tuples
	// Format: [(distance_to_store_A, distance_to_store_B, distance_to_store_C), ...]
	
	// Batman - 2 warehouses
	repo.AddWarehouse("col-001", models.Warehouse{
		ID:                "wh-001-1",
		Name:              "Warehouse North - Batman",
		CollectibleID:     "col-001",
		Available:         true,
		DistancesToStores: []int{1, 4, 5}, // Distances to stores A, B, C
	})
	repo.AddWarehouse("col-001", models.Warehouse{
		ID:                "wh-001-2",
		Name:              "Warehouse South - Batman",
		CollectibleID:     "col-001",
		Available:         true,
		DistancesToStores: []int{3, 2, 3},
	})

	// Millennium Falcon - 3 warehouses
	repo.AddWarehouse("col-002", models.Warehouse{
		ID:                "wh-002-1",
		Name:              "Warehouse East - Falcon",
		CollectibleID:     "col-002",
		Available:         true,
		DistancesToStores: []int{2, 1, 4},
	})
	repo.AddWarehouse("col-002", models.Warehouse{
		ID:                "wh-002-2",
		Name:              "Warehouse West - Falcon",
		CollectibleID:     "col-002",
		Available:         true,
		DistancesToStores: []int{5, 3, 2},
	})
	repo.AddWarehouse("col-002", models.Warehouse{
		ID:                "wh-002-3",
		Name:              "Warehouse Central - Falcon",
		CollectibleID:     "col-002",
		Available:         true,
		DistancesToStores: []int{3, 3, 3},
	})

	// Iron Man Suit - 2 warehouses
	repo.AddWarehouse("col-003", models.Warehouse{
		ID:                "wh-003-1",
		Name:              "Warehouse Premium - Iron Man",
		CollectibleID:     "col-003",
		Available:         true,
		DistancesToStores: []int{4, 2, 1},
	})
	repo.AddWarehouse("col-003", models.Warehouse{
		ID:                "wh-003-2",
		Name:              "Warehouse Secure - Iron Man",
		CollectibleID:     "col-003",
		Available:         true,
		DistancesToStores: []int{2, 5, 4},
	})

	// Pokemon Cards - 3 warehouses
	repo.AddWarehouse("col-004", models.Warehouse{
		ID:                "wh-004-1",
		Name:              "Warehouse A - Pokemon",
		CollectibleID:     "col-004",
		Available:         true,
		DistancesToStores: []int{1, 3, 6},
	})
	repo.AddWarehouse("col-004", models.Warehouse{
		ID:                "wh-004-2",
		Name:              "Warehouse B - Pokemon",
		CollectibleID:     "col-004",
		Available:         true,
		DistancesToStores: []int{4, 1, 5},
	})
	repo.AddWarehouse("col-004", models.Warehouse{
		ID:                "wh-004-3",
		Name:              "Warehouse C - Pokemon",
		CollectibleID:     "col-004",
		Available:         true,
		DistancesToStores: []int{6, 5, 1},
	})

	// Gundam - 2 warehouses
	repo.AddWarehouse("col-005", models.Warehouse{
		ID:                "wh-005-1",
		Name:              "Warehouse Tech - Gundam",
		CollectibleID:     "col-005",
		Available:         true,
		DistancesToStores: []int{3, 4, 2},
	})
	repo.AddWarehouse("col-005", models.Warehouse{
		ID:                "wh-005-2",
		Name:              "Warehouse Main - Gundam",
		CollectibleID:     "col-005",
		Available:         true,
		DistancesToStores: []int{2, 2, 5},
	})

	// Arcade Machine - 2 warehouses
	repo.AddWarehouse("col-006", models.Warehouse{
		ID:                "wh-006-1",
		Name:              "Warehouse Retro - Arcade",
		CollectibleID:     "col-006",
		Available:         true,
		DistancesToStores: []int{5, 1, 3},
	})
	repo.AddWarehouse("col-006", models.Warehouse{
		ID:                "wh-006-2",
		Name:              "Warehouse Gaming - Arcade",
		CollectibleID:     "col-006",
		Available:         true,
		DistancesToStores: []int{1, 6, 2},
	})
}
