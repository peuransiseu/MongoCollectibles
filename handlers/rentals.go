package handlers

import (
	"encoding/json"
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
}

// NewRentalsHandler creates a new rentals handler
func NewRentalsHandler(
	repo *data.Repository,
	pricingService *services.PricingService,
	allocationManager *services.AllocationManager,
	paymentService *services.PaymentService,
	cfg *config.Config,
) *RentalsHandler {
	return &RentalsHandler{
		repo:              repo,
		pricingService:    pricingService,
		allocationManager: allocationManager,
		paymentService:    paymentService,
		config:            cfg,
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

	// Allocate warehouse
	// Now using StoreID directly as the primary identifier for distance lookups
	unit, eta, err := h.allocationManager.Allocate(req.CollectibleID, req.StoreID)
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

	// Idempotency: Check if user already has a pending rental for this collectible
	// note: This simplistic check assumes 1 pending rental per user/collectible pair is allowed
	existingRentals := h.repo.GetRentalsByCustomerAndCollectible(req.Customer.Email, req.CollectibleID)
	for _, rent := range existingRentals {
		if rent.PaymentStatus == models.PaymentPending {
			// Found existing pending rental, reuse it
			// Could verify if storeID matches, but for simplicity we return the existing one
			// If we wanted to support multiple stores, we'd filter by store too.

			// We might need to update the payment session if amount changed, but keeping it simple:
			// just return existing rental info
			response := models.CheckoutResponse{
				RentalID:   rent.ID,
				TotalFee:   rent.TotalFee,
				ETA:        rent.ETA,
				PaymentURL: rent.PaymentURL,
				Message:    "Found pending rental. Please complete payment.",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    response,
			})
			return
		}
	}

	// Create rental record
	rentalID := uuid.New().String()
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
