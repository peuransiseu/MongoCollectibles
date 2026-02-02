package data

import (
	"errors"
	"sync"

	"github.com/mongocollectibles/rental-system/models"
)

// InMemoryRepository handles in-memory data storage
type InMemoryRepository struct {
	collectibles map[string]*models.Collectible
	rentals      map[string]*models.Rental
	warehouses   map[string][]models.Warehouse // collectibleID -> warehouses
	mu           sync.RWMutex
}

// NewRepository creates a new in-memory repository instance
func NewRepository() *InMemoryRepository {
	return &InMemoryRepository{
		collectibles: make(map[string]*models.Collectible),
		rentals:      make(map[string]*models.Rental),
		warehouses:   make(map[string][]models.Warehouse),
	}
}

// GetAllCollectibles returns all collectibles
func (r *InMemoryRepository) GetAllCollectibles() ([]*models.Collectible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collectibles := make([]*models.Collectible, 0, len(r.collectibles))
	for _, c := range r.collectibles {
		collectibles = append(collectibles, c)
	}
	return collectibles, nil
}

// GetCollectibleByID returns a collectible by ID
func (r *InMemoryRepository) GetCollectibleByID(id string) (*models.Collectible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collectible, exists := r.collectibles[id]
	if !exists {
		return nil, errors.New("collectible not found")
	}
	return collectible, nil
}

// AddCollectible adds a new collectible
func (r *InMemoryRepository) AddCollectible(collectible *models.Collectible) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collectibles[collectible.ID] = collectible
	return nil
}

// GetWarehouses returns warehouses for a collectible
func (r *InMemoryRepository) GetWarehouses(collectibleID string) ([]models.Warehouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	warehouses, exists := r.warehouses[collectibleID]
	if !exists {
		return nil, errors.New("no warehouses found for collectible")
	}
	return warehouses, nil
}

// AddWarehouse adds a warehouse for a collectible
func (r *InMemoryRepository) AddWarehouse(collectibleID string, warehouse models.Warehouse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warehouses[collectibleID] = append(r.warehouses[collectibleID], warehouse)
	return nil
}

// GetAllWarehouses returns all warehouses (for allocation service)
func (r *InMemoryRepository) GetAllWarehouses() (map[string][]models.Warehouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.warehouses, nil
}

// CreateRental creates a new rental record
func (r *InMemoryRepository) CreateRental(rental *models.Rental) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rentals[rental.ID]; exists {
		return errors.New("rental already exists")
	}

	r.rentals[rental.ID] = rental
	return nil
}

// GetRentalByID returns a rental by ID
func (r *InMemoryRepository) GetRentalByID(id string) (*models.Rental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rental, exists := r.rentals[id]
	if !exists {
		return nil, errors.New("rental not found")
	}
	return rental, nil
}

// UpdateRental updates an existing rental
func (r *InMemoryRepository) UpdateRental(rental *models.Rental) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rentals[rental.ID]; !exists {
		return errors.New("rental not found")
	}

	r.rentals[rental.ID] = rental
	return nil
}

// GetAllRentals returns all rentals
func (r *InMemoryRepository) GetAllRentals() ([]*models.Rental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rentals := make([]*models.Rental, 0, len(r.rentals))
	for _, r := range r.rentals {
		rentals = append(rentals, r)
	}
	return rentals, nil
}

// GetRentalsByCustomerAndCollectible returns rentals for a specific customer and collectible
func (r *InMemoryRepository) GetRentalsByCustomerAndCollectible(email string, collectibleID string) ([]*models.Rental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*models.Rental
	for _, rental := range r.rentals {
		if rental.Customer.Email == email && rental.CollectibleID == collectibleID {
			matches = append(matches, rental)
		}
	}
	return matches, nil
}

// DeleteAllRentals clears all rental records
func (r *InMemoryRepository) DeleteAllRentals() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rentals = make(map[string]*models.Rental)
	return nil
}
