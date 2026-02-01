package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mongocollectibles/rental-system/config"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// RentalsHandler handles rental-related endpoints
type RentalsHandler struct {
	repo              *data.Repository
	pricingService    *services.PricingService
	allocationManager *services.AllocationManager
	paymentService    *services.PaymentService
	config            *config.Config
	authHandler       *AuthHandler // For user authentication
}

// NewRentalsHandler creates a new rentals handler
func NewRentalsHandler(
	repo *data.Repository,
	pricingService *services.PricingService,
	allocationManager *services.AllocationManager,
	paymentService *services.PaymentService,
	cfg *config.Config,
	authHandler *AuthHandler,
) *RentalsHandler {
	return &RentalsHandler{
		repo:              repo,
		pricingService:    pricingService,
		allocationManager: allocationManager,
		paymentService:    paymentService,
		config:            cfg,
		authHandler:       authHandler,
	}
}

// GetQuote calculates a rental quote
func (h *RentalsHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	var req models.RentalQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate store first
	if req.StoreID == "" {
		// Optional: If no store selected, return stock but 0 ETA, or error.
		// For now, let's assume store is required for a valid quote including ETA.
		// However, the prompt implies "EU1 rents from Store1", so store is known context.
		// If the frontend calls getQuote before store selection, we might handle that.
		// Let's implement strict validation for now.
	}

	// Get collectible
	collectible, err := h.repo.GetCollectibleByID(req.CollectibleID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Collectible not found",
		})
		return
	}

	// Calculate quote
	quote := h.pricingService.CalculateQuote(collectible, req.Duration)

	// Get Stock
	stock := h.allocationManager.GetTotalStock(collectible.ID)
	quote.Stock = stock

	// Get ETA
	if req.StoreID != "" {
		eta, err := h.allocationManager.GetETA(collectible.ID, req.StoreID)
		if err == nil {
			quote.ETA = eta
		} else {
			// If error (e.g. no stock), ETA remains 0 or we could set a flag
			quote.ETA = 0
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    quote,
	})
}

// Checkout creates a rental and initiates payment
func (h *RentalsHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// IDEMPOTENCY CHECK: Check for existing pending rental
	existingRental, err := h.repo.GetPendingRentalByCustomerAndCollectible(req.Customer.Email, req.CollectibleID)
	if err == nil {
		// Found existing pending rental - reuse it
		log.Printf("[Checkout] Reusing existing pending rental %s for customer %s", existingRental.ID, req.Customer.Email)

		response := models.CheckoutResponse{
			RentalID:   existingRental.ID,
			TotalFee:   existingRental.TotalFee,
			ETA:        existingRental.ETA,
			PaymentURL: existingRental.PaymentURL,
			Message:    "Existing rental found. Please complete payment.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    response,
		})
		return
	}

	// Get collectible
	collectible, err := h.repo.GetCollectibleByID(req.CollectibleID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Collectible not found",
		})
		return
	}

	// Create rental ID BEFORE allocation for tracking
	rentalID := uuid.New().String()

	// Allocate warehouse
	// Now using StoreID directly as the primary identifier for distance lookups
	unit, eta, err := h.allocationManager.Allocate(req.CollectibleID, req.StoreID, rentalID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No available warehouse for this collectible at the selected store",
		})
		return
	}
	warehouseID := unit.WarehouseID

	// Calculate pricing
	dailyRate, totalFee, _ := h.pricingService.CalculateRentalFee(collectible.Size, req.Duration)

	// Create rental record (rentalID already created above for allocation tracking)
	rental := &models.Rental{
		ID:              rentalID,
		CollectibleID:   req.CollectibleID,
		CollectibleName: collectible.Name,
		StoreID:         req.StoreID,
		WarehouseID:     warehouseID,
		Customer:        req.Customer,
		Duration:        req.Duration,
		DailyRate:       dailyRate,
		TotalFee:        totalFee,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   models.PaymentPending,
		ETA:             eta,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create payment session
	paymentID, paymentURL, err := h.paymentService.CreateCheckoutSession(
		totalFee,
		rentalID,
		collectible.Name,
		req.Duration,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create payment: " + err.Error(),
		})
		return
	}

	rental.PaymentID = paymentID
	rental.PaymentURL = paymentURL

	// Save rental
	if err := h.repo.CreateRental(rental); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create rental",
		})
		return
	}

	// Mark warehouse as unavailable
	// h.allocationManager.Allocate already marked it as unavailable

	// Return response
	response := models.CheckoutResponse{
		RentalID:   rentalID,
		TotalFee:   totalFee,
		ETA:        eta,
		PaymentURL: paymentURL,
		Message:    "Rental created successfully. Please complete payment.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// CheckoutFromCart processes checkout from the user's cart
// THIS IS WHERE ALLOCATION HAPPENS - not in the cart
func (h *RentalsHandler) CheckoutFromCart(w http.ResponseWriter, r *http.Request) {
	// 1. Get user from auth token
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

	// 2. Get active cart
	cart, err := h.repo.GetActiveCartByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No active cart found",
		})
		return
	}

	// 3. Validate cart is not empty
	if len(cart.Items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cart is empty",
		})
		return
	}

	// 4. Create order ID BEFORE allocation
	orderID := uuid.New().String()
	var orderItems []models.OrderItem
	var totalAmount float64
	var failedAllocations []string

	// 5. For each cart item, ALLOCATE units (THIS IS THE CRITICAL POINT)
	for _, cartItem := range cart.Items {
		// Get collectible details
		collectible, err := h.repo.GetCollectibleByID(cartItem.CollectibleID)
		if err != nil {
			failedAllocations = append(failedAllocations, cartItem.CollectibleID)
			continue
		}

		// ALLOCATION HAPPENS HERE - not in cart
		unit, eta, err := h.allocationManager.Allocate(
			cartItem.CollectibleID,
			cartItem.StoreID,
			orderID,
		)
		if err != nil {
			log.Printf("[Checkout] Allocation failed for %s: %v", cartItem.CollectibleID, err)
			failedAllocations = append(failedAllocations, cartItem.CollectibleID)
			continue
		}

		// Calculate pricing
		dailyRate, itemTotal, _ := h.pricingService.CalculateRentalFee(
			collectible.Size,
			cartItem.RentalDays,
		)
		_ = dailyRate // Unused for now

		// Create order item with ALLOCATED unit details
		orderItem := models.OrderItem{
			CollectibleID:   cartItem.CollectibleID,
			CollectibleName: collectible.Name,
			UnitID:          unit.ID,          // Allocated unit (internal only)
			WarehouseID:     unit.WarehouseID, // Allocated warehouse (internal only)
			RentalDays:      cartItem.RentalDays,
			ETADays:         eta,
			Price:           itemTotal,
		}

		orderItems = append(orderItems, orderItem)
		totalAmount += itemTotal
	}

	// 6. Check if any allocations failed
	if len(failedAllocations) > 0 {
		// Release any successfully allocated units
		for _, item := range orderItems {
			h.allocationManager.ReleaseUnit(item.CollectibleID, item.WarehouseID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Some items could not be allocated",
			"failed":  failedAllocations,
		})
		return
	}

	// 7. Create payment session
	paymentID, paymentURL, err := h.paymentService.CreateCheckoutSession(
		totalAmount,
		orderID,
		"Rental Order",
		7, // Default rental duration (can be improved)
	)
	if err != nil {
		// Release allocated units on payment session creation failure
		for _, item := range orderItems {
			h.allocationManager.ReleaseUnit(item.CollectibleID, item.WarehouseID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create payment session",
		})
		return
	}

	// 8. Create order with PENDING_PAYMENT status
	order := &models.Order{
		ID:          orderID,
		UserID:      userID,
		StoreID:     cart.Items[0].StoreID, // Use first item's store
		Status:      models.OrderPendingPayment,
		TotalAmount: totalAmount,
		Items:       orderItems,
		PaymentID:   paymentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.CreateOrder(order); err != nil {
		// Release allocated units on order creation failure
		for _, item := range orderItems {
			h.allocationManager.ReleaseUnit(item.CollectibleID, item.WarehouseID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create order",
		})
		return
	}

	// 9. Mark cart as CHECKED_OUT
	cart.Status = models.CartCheckedOut
	cart.UpdatedAt = time.Now()
	h.repo.UpdateCart(cart)

	log.Printf("[Checkout] Order created: %s for user %s, Total: %.2f", orderID, userID, totalAmount)

	// 10. Return payment URL to frontend
	response := map[string]interface{}{
		"order_id":    orderID,
		"total":       totalAmount,
		"payment_url": paymentURL,
		"status":      models.OrderPendingPayment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}
