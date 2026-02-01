package services

import (
	"errors"
	"log"

	"github.com/mongocollectibles/rental-system/models"
)

// OrderService handles order-related business logic
type OrderService struct {
	allocationManager *AllocationManager
}

// NewOrderService creates a new order service
func NewOrderService(allocationManager *AllocationManager) *OrderService {
	return &OrderService{
		allocationManager: allocationManager,
	}
}

// CancellationResult represents the result of a cancellation eligibility check
type CancellationResult struct {
	CanCancel    bool
	RefundAmount float64
	RefundReason string
}

// CheckCancellationEligibility determines if an order can be cancelled and calculates refund
func (s *OrderService) CheckCancellationEligibility(order *models.Order) CancellationResult {
	switch order.Status {
	case models.OrderPendingPayment:
		return CancellationResult{
			CanCancel:    true,
			RefundAmount: 0,
			RefundReason: "Payment not completed",
		}

	case models.OrderPaid, models.OrderAllocated:
		// Full refund if not shipped yet
		return CancellationResult{
			CanCancel:    true,
			RefundAmount: order.TotalAmount,
			RefundReason: "Full refund - order not shipped",
		}

	case models.OrderInTransit:
		// Partial refund - 50% of total amount
		return CancellationResult{
			CanCancel:    true,
			RefundAmount: order.TotalAmount * 0.5,
			RefundReason: "Partial refund (50%) - order in transit",
		}

	case models.OrderReadyForPickup, models.OrderCompleted:
		// Cannot cancel delivered orders
		return CancellationResult{
			CanCancel:    false,
			RefundAmount: 0,
			RefundReason: "Cannot cancel delivered or completed orders",
		}

	case models.OrderCancelled, models.OrderRefunded:
		// Already cancelled
		return CancellationResult{
			CanCancel:    false,
			RefundAmount: 0,
			RefundReason: "Order already cancelled or refunded",
		}

	default:
		return CancellationResult{
			CanCancel:    false,
			RefundAmount: 0,
			RefundReason: "Invalid order status",
		}
	}
}

// CancelOrder cancels an order and releases allocated units if applicable
func (s *OrderService) CancelOrder(order *models.Order) (*CancellationResult, error) {
	// 1. Check eligibility
	result := s.CheckCancellationEligibility(order)
	if !result.CanCancel {
		return &result, errors.New(result.RefundReason)
	}

	// 2. Release allocated units (if not shipped)
	if order.Status == models.OrderPaid || order.Status == models.OrderAllocated {
		for _, item := range order.Items {
			if err := s.allocationManager.ReleaseUnit(item.CollectibleID, item.WarehouseID); err != nil {
				log.Printf("[OrderService] Warning: Failed to release unit %s: %v", item.UnitID, err)
				// Continue with cancellation even if release fails
			} else {
				log.Printf("[OrderService] Released unit %s for cancelled order %s", item.UnitID, order.ID)
			}
		}
	}

	// 3. Update order status
	order.Status = models.OrderCancelled

	log.Printf("[OrderService] Order %s cancelled. Refund amount: %.2f", order.ID, result.RefundAmount)

	return &result, nil
}
