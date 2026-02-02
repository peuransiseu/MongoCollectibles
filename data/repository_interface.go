package data

import (
	"github.com/mongocollectibles/rental-system/models"
)

// Repository defines the interface for data access
type Repository interface {
	GetAllCollectibles() ([]*models.Collectible, error)
	GetCollectibleByID(id string) (*models.Collectible, error)
	AddCollectible(collectible *models.Collectible) error
	GetWarehouses(collectibleID string) ([]models.Warehouse, error)
	AddWarehouse(collectibleID string, warehouse models.Warehouse) error
	GetAllWarehouses() (map[string][]models.Warehouse, error)
	CreateRental(rental *models.Rental) error
	GetRentalByID(id string) (*models.Rental, error)
	UpdateRental(rental *models.Rental) error
	GetAllRentals() ([]*models.Rental, error)
	GetRentalsByCustomerAndCollectible(email string, collectibleID string) ([]*models.Rental, error)
	DeleteAllRentals() error
}
