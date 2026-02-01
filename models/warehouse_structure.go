package models

// StoreNode represents a brick-and-mortar store location for the warehouse logic
// Note: This is separate from the main Store struct to keep the new design clean as requested,
// though in a full refactor they might merge.
type StoreNode struct {
	ID   int
	Name string
}

// WarehouseNode represents a warehouse that stores one collectible
type WarehouseNode struct {
	ID        string
	Distances map[string]int // Map of StoreID -> distance (km) to eliminate index-based lookup errors
}
