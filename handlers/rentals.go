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
	allocationService *services.AllocationService
	paymentService    *services.PaymentService
	config            *config.Config
}

// NewRentalsHandler creates a new rentals handler
func NewRentalsHandler(
	repo *data.Repository,
	pricingService *services.PricingService,
	allocationService *services.AllocationService,
	paymentService *services.PaymentService,
	cfg *config.Config,
) *RentalsHandler {
	return &RentalsHandler{
		repo:              repo,
		pricingService:    pricingService,
		allocationService: allocationService,
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

	// Validate store
	storeIndex := h.config.GetStoreIndex(req.StoreID)
	if storeIndex == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid store ID",
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
	warehouseID, _, err := h.allocationService.AllocateWarehouse(req.CollectibleID, storeIndex)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No available warehouse for this collectible at the selected store",
		})
		return
	}

	// Calculate pricing
	dailyRate, totalFee, _ := h.pricingService.CalculateRentalFee(collectible.Size, req.Duration)

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
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create payment source
	paymentID, paymentURL, err := h.paymentService.CreatePaymentSource(
		totalFee,
		req.PaymentMethod,
		req.Customer,
		rentalID,
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
	h.allocationService.MarkWarehouseUnavailable(req.CollectibleID, warehouseID)

	// Return response
	response := models.CheckoutResponse{
		RentalID:   rentalID,
		TotalFee:   totalFee,
		PaymentURL: paymentURL,
		Message:    "Rental created successfully. Please complete payment.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}
