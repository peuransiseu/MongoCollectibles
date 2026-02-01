package services

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
)

// RefundService handles refund operations
type RefundService struct {
	repo *data.Repository
}

// NewRefundService creates a new refund service
func NewRefundService(repo *data.Repository) *RefundService {
	return &RefundService{
		repo: repo,
	}
}

// CreateRefund creates a refund for an order (idempotent)
func (s *RefundService) CreateRefund(orderID string, amount float64, reason string) (*models.Refund, error) {
	// 1. Check if refund already exists (idempotency)
	existing, err := s.repo.GetRefundByOrderID(orderID)
	if err == nil {
		log.Printf("[RefundService] Refund already exists for order %s: %s", orderID, existing.ID)
		return existing, nil // Return existing refund
	}

	// 2. Get order to validate
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// 3. Create refund record
	refund := &models.Refund{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		UserID:    order.UserID,
		Amount:    amount,
		Reason:    reason,
		Status:    models.RefundPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. Save refund
	if err := s.repo.CreateRefund(refund); err != nil {
		return nil, err
	}

	log.Printf("[RefundService] Created refund %s for order %s: Amount %.2f", refund.ID, orderID, amount)

	// 5. Process refund (integrate with payment provider)
	// For now, mark as processed immediately (mock implementation)
	if err := s.ProcessRefund(refund); err != nil {
		log.Printf("[RefundService] Failed to process refund %s: %v", refund.ID, err)
		refund.Status = models.RefundFailed
		s.repo.UpdateRefund(refund)
		return refund, err
	}

	return refund, nil
}

// ProcessRefund processes a refund with the payment provider
func (s *RefundService) ProcessRefund(refund *models.Refund) error {
	// TODO: Integrate with PayMongo refund API
	// For now, this is a mock implementation

	log.Printf("[RefundService] Processing refund %s: Amount %.2f", refund.ID, refund.Amount)

	// Simulate successful refund processing
	refund.Status = models.RefundProcessed
	refund.UpdatedAt = time.Now()
	s.repo.UpdateRefund(refund)

	log.Printf("[RefundService] Refund %s processed successfully", refund.ID)

	return nil
}

// GetRefundByOrderID retrieves a refund by order ID
func (s *RefundService) GetRefundByOrderID(orderID string) (*models.Refund, error) {
	return s.repo.GetRefundByOrderID(orderID)
}
