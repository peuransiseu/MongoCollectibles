package data

import (
	"testing"
	"time"

	"github.com/mongocollectibles/rental-system/models"
)

func TestRepository_CreateUser(t *testing.T) {
	repo := NewRepository()

	user := &models.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Try to create duplicate email
	user2 := &models.User{
		ID:           "user-2",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.CreateUser(user2)
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	repo := NewRepository()

	user := &models.User{
		ID:           "user-3",
		Email:        "user3@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	repo.CreateUser(user)

	retrieved, err := repo.GetUserByEmail("user3@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrieved.ID)
	}

	// Try to get non-existent user
	_, err = repo.GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

func TestRepository_SessionManagement(t *testing.T) {
	repo := NewRepository()

	userID := "user-4"
	token := "test-session-token"

	// Create session
	err := repo.CreateSession(userID, token)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get user by token
	retrievedUserID, err := repo.GetUserByToken(token)
	if err != nil {
		t.Fatalf("GetUserByToken failed: %v", err)
	}

	if retrievedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrievedUserID)
	}

	// Delete session
	err = repo.DeleteSession(token)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Try to get user by deleted token
	_, err = repo.GetUserByToken(token)
	if err == nil {
		t.Fatal("Expected error for deleted token, got nil")
	}
}

func TestRepository_CartManagement(t *testing.T) {
	repo := NewRepository()

	cart := &models.Cart{
		ID:        "cart-1",
		UserID:    "user-5",
		Status:    models.CartActive,
		Items:     []models.CartItem{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create cart
	err := repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("CreateCart failed: %v", err)
	}

	// Get active cart by user ID
	retrieved, err := repo.GetActiveCartByUserID("user-5")
	if err != nil {
		t.Fatalf("GetActiveCartByUserID failed: %v", err)
	}

	if retrieved.ID != cart.ID {
		t.Errorf("Expected cart ID %s, got %s", cart.ID, retrieved.ID)
	}

	// Add item to cart
	item := models.CartItem{
		CollectibleID: "collectible-1",
		StoreID:       "store-1",
		RentalDays:    7,
		Quantity:      2,
	}

	err = repo.AddCartItem(cart.ID, item)
	if err != nil {
		t.Fatalf("AddCartItem failed: %v", err)
	}

	// Verify item was added
	retrieved, _ = repo.GetActiveCartByUserID("user-5")
	if len(retrieved.Items) != 1 {
		t.Errorf("Expected 1 item in cart, got %d", len(retrieved.Items))
	}

	// Remove item from cart
	err = repo.RemoveCartItem(cart.ID, "collectible-1")
	if err != nil {
		t.Fatalf("RemoveCartItem failed: %v", err)
	}

	// Verify item was removed
	retrieved, _ = repo.GetActiveCartByUserID("user-5")
	if len(retrieved.Items) != 0 {
		t.Errorf("Expected 0 items in cart, got %d", len(retrieved.Items))
	}
}

func TestRepository_OrderManagement(t *testing.T) {
	repo := NewRepository()

	order := &models.Order{
		ID:          "order-1",
		UserID:      "user-6",
		Status:      models.OrderPendingPayment,
		TotalAmount: 1500.0,
		Items:       []models.OrderItem{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create order
	err := repo.CreateOrder(order)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Get order by ID
	retrieved, err := repo.GetOrderByID("order-1")
	if err != nil {
		t.Fatalf("GetOrderByID failed: %v", err)
	}

	if retrieved.Status != models.OrderPendingPayment {
		t.Errorf("Expected status PENDING_PAYMENT, got %s", retrieved.Status)
	}

	// Update order status
	err = repo.UpdateOrderStatus("order-1", models.OrderPaid)
	if err != nil {
		t.Fatalf("UpdateOrderStatus failed: %v", err)
	}

	// Verify status was updated
	retrieved, _ = repo.GetOrderByID("order-1")
	if retrieved.Status != models.OrderPaid {
		t.Errorf("Expected status PAID, got %s", retrieved.Status)
	}

	// Get orders by user ID
	orders, err := repo.GetOrdersByUserID("user-6")
	if err != nil {
		t.Fatalf("GetOrdersByUserID failed: %v", err)
	}

	if len(orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders))
	}
}

func TestRepository_RefundManagement(t *testing.T) {
	repo := NewRepository()

	refund := &models.Refund{
		ID:        "refund-1",
		OrderID:   "order-2",
		UserID:    "user-7",
		Amount:    1000.0,
		Reason:    "Test refund",
		Status:    models.RefundPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create refund
	err := repo.CreateRefund(refund)
	if err != nil {
		t.Fatalf("CreateRefund failed: %v", err)
	}

	// Get refund by order ID
	retrieved, err := repo.GetRefundByOrderID("order-2")
	if err != nil {
		t.Fatalf("GetRefundByOrderID failed: %v", err)
	}

	if retrieved.Amount != 1000.0 {
		t.Errorf("Expected refund amount 1000.0, got %.2f", retrieved.Amount)
	}

	// Update refund status
	refund.Status = models.RefundProcessed
	err = repo.UpdateRefund(refund)
	if err != nil {
		t.Fatalf("UpdateRefund failed: %v", err)
	}

	// Verify status was updated
	retrieved, _ = repo.GetRefundByOrderID("order-2")
	if retrieved.Status != models.RefundProcessed {
		t.Errorf("Expected status PROCESSED, got %s", retrieved.Status)
	}
}
