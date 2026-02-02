package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// PaymentsHandler handles payment webhooks and callbacks
type PaymentsHandler struct {
	repo              data.Repository
	paymentService    *services.PaymentService
	allocationManager *services.AllocationManager
}

// NewPaymentsHandler creates a new payments handler
func NewPaymentsHandler(repo data.Repository, paymentService *services.PaymentService, allocationManager *services.AllocationManager) *PaymentsHandler {
	return &PaymentsHandler{
		repo:              repo,
		paymentService:    paymentService,
		allocationManager: allocationManager,
	}
}

// WebhookPayMongo handles PayMongo webhook events
func (h *PaymentsHandler) WebhookPayMongo(w http.ResponseWriter, r *http.Request) {
	var webhookData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extract payment ID from webhook
	// Note: Actual webhook structure may vary, this is a simplified version
	data, ok := webhookData["data"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract event type
	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	eventType, _ := attributes["type"].(string)

	// Use the data structure we need
	dataResource, ok := attributes["data"].(map[string]interface{})
	if !ok {
		// Some events might be structured differently, safely ignore or log
		w.WriteHeader(http.StatusOK)
		return
	}
	resourceAttr, ok := dataResource["attributes"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}
	paymentID, _ := resourceAttr["id"].(string)

	if eventType == "checkout_session.expired" {
		// Find rental by payment ID (which is the session ID in this context)
		// Or if we store session ID separately
		rentals, _ := h.repo.GetAllRentals()
		for _, rental := range rentals {
			if rental.PaymentID == paymentID {
				// Release unit
				h.allocationManager.ReleaseUnit(rental.CollectibleID, rental.WarehouseID)
				rental.PaymentStatus = models.PaymentFailed
				h.repo.UpdateRental(rental)
				break
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify payment status for strictness, or trust the webhook
	status, err := h.paymentService.VerifyPayment(paymentID)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Update rental status based on payment
	rentals, _ := h.repo.GetAllRentals()
	for _, rental := range rentals {
		if rental.PaymentID == paymentID {
			rental.PaymentStatus = status
			h.repo.UpdateRental(rental)
			break
		}
	}

	w.WriteHeader(http.StatusOK)
}

// PaymentSuccess handles successful payment redirects
func (h *PaymentsHandler) PaymentSuccess(w http.ResponseWriter, r *http.Request) {
	rentalID := r.URL.Query().Get("rental_id")

	rental, err := h.repo.GetRentalByID(rentalID)
	if err != nil {
		http.Error(w, "Rental not found", http.StatusNotFound)
		return
	}

	rental.PaymentStatus = models.PaymentCompleted
	h.repo.UpdateRental(rental)

	// Confirm reservation in allocation manager to prevent auto-cleanup
	if err := h.allocationManager.ConfirmReservation(rental.CollectibleID, rental.WarehouseID); err != nil {
		// Log error but assume valid since we are in success flow
		// log.Printf("Warning: Failed to confirm reservation: %v", err)
	}

	// Redirect to success page
	http.Redirect(w, r, "/success.html?rental_id="+rentalID, http.StatusSeeOther)
}

// PaymentFailed handles failed payment redirects
func (h *PaymentsHandler) PaymentFailed(w http.ResponseWriter, r *http.Request) {
	rentalID := r.URL.Query().Get("rental_id")

	rental, err := h.repo.GetRentalByID(rentalID)
	if err != nil {
		http.Error(w, "Rental not found", http.StatusNotFound)
		return
	}

	// Release the allocated unit back to inventory
	if err := h.allocationManager.ReleaseUnit(rental.CollectibleID, rental.WarehouseID); err != nil {
		// Log the error but continue - we still want to update the rental status
		// The unit might have already been released or not found
	}

	rental.PaymentStatus = models.PaymentFailed
	h.repo.UpdateRental(rental)

	// Redirect to failure page
	http.Redirect(w, r, "/failed.html?rental_id="+rentalID, http.StatusSeeOther)
}
