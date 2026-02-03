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
	ID          string  `json:"id" dynamodbav:"id"`
	Name        string  `json:"name" dynamodbav:"name"`
	Description string  `json:"description" dynamodbav:"description"`
	Size        Size    `json:"size" dynamodbav:"size"`
	ImageURL    string  `json:"image_url" dynamodbav:"image_url"`
	Stock       int     `json:"stock" dynamodbav:"stock"` // Dynamic field for available units
	Available   bool    `json:"available" dynamodbav:"available"`
	DailyRate   float64 `json:"daily_rate" dynamodbav:"daily_rate"` // Daily rental rate
	ETADays     int     `json:"eta_days" dynamodbav:"-"`            // Estimated time of arrival in days, ignored in DB
}

// Store represents a brick-and-mortar store location
type Store struct {
	ID      string `json:"id" dynamodbav:"id"`
	Name    string `json:"name" dynamodbav:"name"`
	Address string `json:"address" dynamodbav:"address"`
	// Latitude  float64 `json:"latitude" dynamodbav:"latitude"`
	// Longitude float64 `json:"longitude" dynamodbav:"longitude"`
}

// Warehouse represents a storage location for collectibles
type Warehouse struct {
	ID            string         `json:"id" dynamodbav:"id"`
	Name          string         `json:"name" dynamodbav:"name"`
	CollectibleID string         `json:"collectible_id" dynamodbav:"collectible_id"`
	Available     bool           `json:"available" dynamodbav:"available"`
	Distances     map[string]int `json:"distances" dynamodbav:"distances"` // StoreID -> distance (km)
}
