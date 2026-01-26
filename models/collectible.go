package models

// Size represents the size category of a collectible
type Size string

const (
	SizeSmall  Size = "S"
	SizeMedium Size = "M"
	SizeLarge  Size = "L"
)

// GetDailyRate returns the daily rental rate for a given size
func (s Size) GetDailyRate() float64 {
	switch s {
	case SizeSmall:
		return 1000.00
	case SizeMedium:
		return 5000.00
	case SizeLarge:
		return 10000.00
	default:
		return 0.00
	}
}

// Collectible represents a rentable collectible item
type Collectible struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	Size               Size       `json:"size"`
	ImageURL           string     `json:"image_url"`
	WarehouseDistances [][]int    `json:"warehouse_distances"` // List of tuples: [(w1_to_s1, w1_to_s2, w1_to_s3), ...]
	Available          bool       `json:"available"`
}

// Store represents a brick-and-mortar store location
type Store struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Warehouse represents a storage location for collectibles
type Warehouse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CollectibleID   string `json:"collectible_id"`
	Available       bool   `json:"available"`
	DistancesToStores []int `json:"distances_to_stores"` // Distance to each store in order
}
