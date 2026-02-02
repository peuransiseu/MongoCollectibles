package data

import (
	"github.com/mongocollectibles/rental-system/models"
)

// SeedData populates the repository with sample data
func SeedData(repo Repository) {
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
			ImageURL:    "/images/street-fighter.jpg",
			Available:   true,
		},
	}

	for _, c := range collectibles {
		repo.AddCollectible(c)
	}

	// Standard definition for 4 regional warehouses
	regions := []struct {
		IDSuffix  string
		Name      string
		Distances map[string]int
	}{
		{
			IDSuffix: "-north",
			Name:     "Warehouse North (QC)",
			Distances: map[string]int{
				"store-a": 5,  // Manila
				"store-b": 1,  // QC (Close)
				"store-c": 10, // Makati (Far)
			},
		},
		{
			IDSuffix: "-south",
			Name:     "Warehouse South (Alabang)",
			Distances: map[string]int{
				"store-a": 8,  // Manila
				"store-b": 12, // QC (Far)
				"store-c": 1,  // Makati (Close)
			},
		},
		{
			IDSuffix: "-east",
			Name:     "Warehouse East (Pasig)",
			Distances: map[string]int{
				"store-a": 6, // Manila
				"store-b": 4, // QC
				"store-c": 3, // Makati (Med)
			},
		},
		{
			IDSuffix: "-west",
			Name:     "Warehouse West (Port Area)",
			Distances: map[string]int{
				"store-a": 1, // Manila (Close)
				"store-b": 6, // QC
				"store-c": 6, // Makati
			},
		},
	}

	// Loop through all collectibles and assign one unit in each warehouse
	for _, c := range collectibles {
		for _, r := range regions {
			repo.AddWarehouse(c.ID, models.Warehouse{
				ID:            c.ID + r.IDSuffix,
				Name:          r.Name,
				CollectibleID: c.ID,
				Available:     true,
				Distances:     r.Distances,
			})
		}
	}
}
