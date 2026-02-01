package data

import (
	"errors"
	"sync"

	"github.com/mongocollectibles/rental-system/models"
)

// Repository handles in-memory data storage
type Repository struct {
	collectibles map[string]*models.Collectible
	rentals      map[string]*models.Rental
	warehouses   map[string][]models.Warehouse // collectibleID -> warehouses
	mu           sync.RWMutex
}

// NewRepository creates a new repository instance
func NewRepository() *Repository {
	return &Repository{
		collectibles: make(map[string]*models.Collectible),
		rentals:      make(map[string]*models.Rental),
		warehouses:   make(map[string][]models.Warehouse),
	}
}

// GetAllCollectibles returns all collectibles
func (r *Repository) GetAllCollectibles() []*models.Collectible {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collectibles := make([]*models.Collectible, 0, len(r.collectibles))
	for _, c := range r.collectibles {
		collectibles = append(collectibles, c)
	}
	return collectibles
}

// GetCollectibleByID returns a collectible by ID
func (r *Repository) GetCollectibleByID(id string) (*models.Collectible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collectible, exists := r.collectibles[id]
	if !exists {
		return nil, errors.New("collectible not found")
	}
	return collectible, nil
}

// AddCollectible adds a new collectible
func (r *Repository) AddCollectible(collectible *models.Collectible) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collectibles[collectible.ID] = collectible
}

// GetWarehouses returns warehouses for a collectible
func (r *Repository) GetWarehouses(collectibleID string) ([]models.Warehouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	warehouses, exists := r.warehouses[collectibleID]
	if !exists {
		return nil, errors.New("no warehouses found for collectible")
	}
	return warehouses, nil
}

// AddWarehouse adds a warehouse for a collectible
func (r *Repository) AddWarehouse(collectibleID string, warehouse models.Warehouse) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warehouses[collectibleID] = append(r.warehouses[collectibleID], warehouse)
}

// GetAllWarehouses returns all warehouses (for allocation service)
func (r *Repository) GetAllWarehouses() map[string][]models.Warehouse {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.warehouses
}

// CreateRental creates a new rental record
func (r *Repository) CreateRental(rental *models.Rental) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rentals[rental.ID]; exists {
		return errors.New("rental already exists")
	}

	r.rentals[rental.ID] = rental
	return nil
}

// GetRentalByID returns a rental by ID
func (r *Repository) GetRentalByID(id string) (*models.Rental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rental, exists := r.rentals[id]
	if !exists {
		return nil, errors.New("rental not found")
	}
	return rental, nil
}

// UpdateRental updates an existing rental
func (r *Repository) UpdateRental(rental *models.Rental) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rentals[rental.ID]; !exists {
		return errors.New("rental not found")
	}

	r.rentals[rental.ID] = rental
	return nil
}

// GetAllRentals returns all rentals
func (r *Repository) GetAllRentals() []*models.Rental {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rentals := make([]*models.Rental, 0, len(r.rentals))
	for _, r := range r.rentals {
		rentals = append(rentals, r)
	}
	return rentals
}

// GetRentalsByCustomerAndCollectible returns rentals for a specific customer and collectible
func (r *Repository) GetRentalsByCustomerAndCollectible(email string, collectibleID string) []*models.Rental {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*models.Rental
	for _, rental := range r.rentals {
		if rental.Customer.Email == email && rental.CollectibleID == collectibleID {
			matches = append(matches, rental)
		}
	}
	return matches
}
