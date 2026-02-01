package data

import (
	"errors"
	"sync"
	"time"

	"github.com/mongocollectibles/rental-system/models"
)

// Repository handles in-memory data storage
type Repository struct {
	collectibles map[string]*models.Collectible
	rentals      map[string]*models.Rental
	warehouses   map[string][]models.Warehouse // collectibleID -> warehouses
	users        map[string]*models.User       // userID -> user
	usersByEmail map[string]*models.User       // email -> user (for login)
	sessions     map[string]string             // token -> userID
	carts        map[string]*models.Cart       // cartID -> cart
	cartsByUser  map[string]*models.Cart       // userID -> active cart
	orders       map[string]*models.Order      // orderID -> order
	ordersByUser map[string][]*models.Order    // userID -> orders
	refunds      map[string]*models.Refund     // refundID -> refund
	mu           sync.RWMutex
}

// NewRepository creates a new repository instance
func NewRepository() *Repository {
	return &Repository{
		collectibles: make(map[string]*models.Collectible),
		rentals:      make(map[string]*models.Rental),
		warehouses:   make(map[string][]models.Warehouse),
		users:        make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		sessions:     make(map[string]string),
		carts:        make(map[string]*models.Cart),
		cartsByUser:  make(map[string]*models.Cart),
		orders:       make(map[string]*models.Order),
		ordersByUser: make(map[string][]*models.Order),
		refunds:      make(map[string]*models.Refund),
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

// GetPendingRentalByCustomerAndCollectible finds a pending rental for a customer and collectible
func (r *Repository) GetPendingRentalByCustomerAndCollectible(customerEmail string, collectibleID string) (*models.Rental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rental := range r.rentals {
		if rental.Customer.Email == customerEmail &&
			rental.CollectibleID == collectibleID &&
			rental.PaymentStatus == models.PaymentPending {
			return rental, nil
		}
	}

	return nil, errors.New("no pending rental found")
}

// CreateUser creates a new user
func (r *Repository) CreateUser(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.usersByEmail[user.Email]; exists {
		return errors.New("email already registered")
	}

	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

// GetUserByID returns a user by ID
func (r *Repository) GetUserByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByEmail returns a user by email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// CreateSession creates a new session token for a user
func (r *Repository) CreateSession(userID, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[token] = userID
	return nil
}

// GetUserByToken returns the user ID associated with a session token
func (r *Repository) GetUserByToken(token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, exists := r.sessions[token]
	if !exists {
		return "", errors.New("invalid or expired token")
	}
	return userID, nil
}

// DeleteSession deletes a session token
func (r *Repository) DeleteSession(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, token)
	return nil
}

// GetActiveCartByUserID returns the active cart for a user
func (r *Repository) GetActiveCartByUserID(userID string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.cartsByUser[userID]
	if !exists || cart.Status != models.CartActive {
		return nil, errors.New("no active cart found")
	}
	return cart, nil
}

// CreateCart creates a new cart
func (r *Repository) CreateCart(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.ID] = cart
	r.cartsByUser[cart.UserID] = cart
	return nil
}

// UpdateCart updates an existing cart
func (r *Repository) UpdateCart(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.ID] = cart
	// Update user mapping if cart is active
	if cart.Status == models.CartActive {
		r.cartsByUser[cart.UserID] = cart
	} else {
		// Remove from active cart mapping if no longer active
		delete(r.cartsByUser, cart.UserID)
	}
	return nil
}

// AddCartItem adds an item to a cart
func (r *Repository) AddCartItem(cartID string, item models.CartItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cart, exists := r.carts[cartID]
	if !exists {
		return errors.New("cart not found")
	}

	// Check if item already exists, update quantity if so
	for i, existingItem := range cart.Items {
		if existingItem.CollectibleID == item.CollectibleID && existingItem.StoreID == item.StoreID {
			cart.Items[i].Quantity += item.Quantity
			cart.Items[i].RentalDays = item.RentalDays // Update rental days
			return nil
		}
	}

	// Add new item
	cart.Items = append(cart.Items, item)
	return nil
}

// UpdateCartItem updates a specific cart item
func (r *Repository) UpdateCartItem(cartID string, collectibleID string, item models.CartItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cart, exists := r.carts[cartID]
	if !exists {
		return errors.New("cart not found")
	}

	for i, existingItem := range cart.Items {
		if existingItem.CollectibleID == collectibleID {
			cart.Items[i] = item
			return nil
		}
	}

	return errors.New("item not found in cart")
}

// RemoveCartItem removes an item from a cart
func (r *Repository) RemoveCartItem(cartID string, collectibleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cart, exists := r.carts[cartID]
	if !exists {
		return errors.New("cart not found")
	}

	for i, item := range cart.Items {
		if item.CollectibleID == collectibleID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			return nil
		}
	}

	return errors.New("item not found in cart")
}

// CreateOrder creates a new order
func (r *Repository) CreateOrder(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	r.ordersByUser[order.UserID] = append(r.ordersByUser[order.UserID], order)
	return nil
}

// GetOrderByID returns an order by ID
func (r *Repository) GetOrderByID(orderID string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// GetOrdersByUserID returns all orders for a user
func (r *Repository) GetOrdersByUserID(userID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, exists := r.ordersByUser[userID]
	if !exists {
		return []*models.Order{}, nil
	}
	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (r *Repository) UpdateOrderStatus(orderID string, status models.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[orderID]
	if !exists {
		return errors.New("order not found")
	}

	order.Status = status
	order.UpdatedAt = time.Now()
	return nil
}

// UpdateOrder updates an entire order
func (r *Repository) UpdateOrder(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	return nil
}

// CreateRefund creates a new refund
func (r *Repository) CreateRefund(refund *models.Refund) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refunds[refund.ID] = refund
	return nil
}

// GetRefundByOrderID returns a refund by order ID
func (r *Repository) GetRefundByOrderID(orderID string) (*models.Refund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, refund := range r.refunds {
		if refund.OrderID == orderID {
			return refund, nil
		}
	}

	return nil, errors.New("refund not found")
}

// UpdateRefund updates a refund
func (r *Repository) UpdateRefund(refund *models.Refund) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refunds[refund.ID] = refund
	return nil
}
