package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// OrderHandler handles order-related endpoints
type OrderHandler struct {
	repo          *data.Repository
	authHandler   *AuthHandler
	orderService  *services.OrderService
	refundService *services.RefundService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(
	repo *data.Repository,
	authHandler *AuthHandler,
	orderService *services.OrderService,
	refundService *services.RefundService,
) *OrderHandler {
	return &OrderHandler{
		repo:          repo,
		authHandler:   authHandler,
		orderService:  orderService,
		refundService: refundService,
	}
}

// GetOrders returns all orders for the authenticated user
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get orders
	orders, err := h.repo.GetOrdersByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to retrieve orders",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    orders,
	})
}

// GetOrderByID returns a specific order by ID
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID := vars["id"]

	// Get order
	order, err := h.repo.GetOrderByID(orderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// Verify order belongs to user
	if order.UserID != userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    order,
	})
}

// CancelOrder cancels an order and processes refund if applicable
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID := vars["id"]

	// Get order
	order, err := h.repo.GetOrderByID(orderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// Verify order belongs to user
	if order.UserID != userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Cancel order using order service
	result, err := h.orderService.CancelOrder(order)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update order in repository
	h.repo.UpdateOrder(order)

	// Create refund if applicable
	var refund *models.Refund
	if result.RefundAmount > 0 {
		refund, err = h.refundService.CreateRefund(orderID, result.RefundAmount, result.RefundReason)
		if err != nil {
			log.Printf("[OrderHandler] Failed to create refund for order %s: %v", orderID, err)
			// Continue with cancellation even if refund fails
		} else {
			// Update order status to REFUNDED
			order.Status = models.OrderRefunded
			h.repo.UpdateOrder(order)
		}
	}

	log.Printf("[OrderHandler] Order %s cancelled by user %s. Refund: %.2f", orderID, userID, result.RefundAmount)

	response := map[string]interface{}{
		"order_id":      orderID,
		"status":        order.Status,
		"refund_amount": result.RefundAmount,
		"refund_reason": result.RefundReason,
	}

	if refund != nil {
		response["refund_id"] = refund.ID
		response["refund_status"] = refund.Status
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// GetRefundStatus returns the refund status for an order
func (h *OrderHandler) GetRefundStatus(w http.ResponseWriter, r *http.Request) {
	// Get user from token
	userID, err := h.authHandler.GetUserFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID := vars["id"]

	// Get order to verify ownership
	order, err := h.repo.GetOrderByID(orderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// Verify order belongs to user
	if order.UserID != userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Get refund
	refund, err := h.refundService.GetRefundByOrderID(orderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No refund found for this order",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}
