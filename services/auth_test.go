package services

import (
	"testing"

	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
)

func TestAuthService_HashPassword(t *testing.T) {
	authService := NewAuthService()

	password := "testpassword123"
	hash, err := authService.HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not equal plain password")
	}
}

func TestAuthService_CheckPassword(t *testing.T) {
	authService := NewAuthService()

	password := "testpassword123"
	hash, _ := authService.HashPassword(password)

	// Test correct password
	if !authService.CheckPassword(password, hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}

	// Test incorrect password
	if authService.CheckPassword("wrongpassword", hash) {
		t.Fatal("CheckPassword should return false for incorrect password")
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService()

	token1, err := authService.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token1 == "" {
		t.Fatal("Token should not be empty")
	}

	// Generate another token to ensure uniqueness
	token2, _ := authService.GenerateToken()
	if token1 == token2 {
		t.Fatal("Tokens should be unique")
	}
}

func TestOrderService_CheckCancellationEligibility(t *testing.T) {
	allocationManager := NewAllocationManager(nil, nil)
	orderService := NewOrderService(allocationManager)

	tests := []struct {
		name           string
		status         models.OrderStatus
		totalAmount    float64
		expectedCancel bool
		expectedRefund float64
	}{
		{
			name:           "PENDING_PAYMENT - can cancel, no refund",
			status:         models.OrderPendingPayment,
			totalAmount:    1000.0,
			expectedCancel: true,
			expectedRefund: 0,
		},
		{
			name:           "PAID - can cancel, full refund",
			status:         models.OrderPaid,
			totalAmount:    1000.0,
			expectedCancel: true,
			expectedRefund: 1000.0,
		},
		{
			name:           "IN_TRANSIT - can cancel, 50% refund",
			status:         models.OrderInTransit,
			totalAmount:    1000.0,
			expectedCancel: true,
			expectedRefund: 500.0,
		},
		{
			name:           "COMPLETED - cannot cancel",
			status:         models.OrderCompleted,
			totalAmount:    1000.0,
			expectedCancel: false,
			expectedRefund: 0,
		},
		{
			name:           "CANCELLED - cannot cancel",
			status:         models.OrderCancelled,
			totalAmount:    1000.0,
			expectedCancel: false,
			expectedRefund: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &models.Order{
				Status:      tt.status,
				TotalAmount: tt.totalAmount,
			}

			result := orderService.CheckCancellationEligibility(order)

			if result.CanCancel != tt.expectedCancel {
				t.Errorf("Expected CanCancel=%v, got %v", tt.expectedCancel, result.CanCancel)
			}

			if result.RefundAmount != tt.expectedRefund {
				t.Errorf("Expected RefundAmount=%.2f, got %.2f", tt.expectedRefund, result.RefundAmount)
			}
		})
	}
}

func TestRefundService_CreateRefund_Idempotency(t *testing.T) {
	repo := data.NewRepository()
	refundService := NewRefundService(repo)

	// Create test order
	order := &models.Order{
		ID:          "test-order-1",
		UserID:      "test-user-1",
		TotalAmount: 1000.0,
		Status:      models.OrderPaid,
	}
	repo.CreateOrder(order)

	// Create refund first time
	refund1, err := refundService.CreateRefund("test-order-1", 1000.0, "Test refund")
	if err != nil {
		t.Fatalf("CreateRefund failed: %v", err)
	}

	if refund1.Amount != 1000.0 {
		t.Errorf("Expected refund amount 1000.0, got %.2f", refund1.Amount)
	}

	// Create refund second time (should return same refund - idempotency)
	refund2, err := refundService.CreateRefund("test-order-1", 1000.0, "Test refund")
	if err != nil {
		t.Fatalf("CreateRefund failed on second call: %v", err)
	}

	if refund1.ID != refund2.ID {
		t.Errorf("Expected same refund ID (idempotency), got different IDs: %s vs %s", refund1.ID, refund2.ID)
	}
}

func TestRefundService_GetRefundByOrderID(t *testing.T) {
	repo := data.NewRepository()
	refundService := NewRefundService(repo)

	// Create test order
	order := &models.Order{
		ID:          "test-order-2",
		UserID:      "test-user-2",
		TotalAmount: 500.0,
		Status:      models.OrderPaid,
	}
	repo.CreateOrder(order)

	// Create refund
	refund, _ := refundService.CreateRefund("test-order-2", 500.0, "Test refund")

	// Get refund by order ID
	retrieved, err := refundService.GetRefundByOrderID("test-order-2")
	if err != nil {
		t.Fatalf("GetRefundByOrderID failed: %v", err)
	}

	if retrieved.ID != refund.ID {
		t.Errorf("Expected refund ID %s, got %s", refund.ID, retrieved.ID)
	}

	// Try to get non-existent refund
	_, err = refundService.GetRefundByOrderID("non-existent-order")
	if err == nil {
		t.Fatal("Expected error for non-existent order, got nil")
	}
}
