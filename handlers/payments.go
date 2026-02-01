package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// PaymentsHandler handles payment webhooks and callbacks
type PaymentsHandler struct {
	repo              *data.Repository
	paymentService    *services.PaymentService
	allocationManager *services.AllocationManager
}

// NewPaymentsHandler creates a new payments handler
func NewPaymentsHandler(repo *data.Repository, paymentService *services.PaymentService, allocationManager *services.AllocationManager) *PaymentsHandler {
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

	// Extract event type
	data, ok := webhookData["data"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check event type
	eventType, _ := attributes["type"].(string)

	switch eventType {
	case "checkout_session.payment.paid":
		// Handle successful payment
		h.handlePaymentSuccess(attributes)

	case "checkout_session.expired":
		// Handle session expiry - release the unit
		h.handleSessionExpiry(attributes)

	default:
		// For backward compatibility, try to extract payment ID
		paymentID, ok := attributes["id"].(string)
		if ok {
			// Verify payment status
			status, err := h.paymentService.VerifyPayment(paymentID)
			if err != nil {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Update rental status based on payment
			rentals := h.repo.GetAllRentals()
			for _, rental := range rentals {
				if rental.PaymentID == paymentID {
					rental.PaymentStatus = status
					h.repo.UpdateRental(rental)
					break
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// handlePaymentSuccess processes successful payment webhook events
func (h *PaymentsHandler) handlePaymentSuccess(attributes map[string]interface{}) {
	paymentID, ok := attributes["id"].(string)
	if !ok {
		return
	}

	rentals := h.repo.GetAllRentals()
	for _, rental := range rentals {
		if rental.PaymentID == paymentID {
			rental.PaymentStatus = models.PaymentCompleted
			h.repo.UpdateRental(rental)
			log.Printf("[Webhook] Payment completed for rental %s", rental.ID)
			break
		}
	}
}

// handleSessionExpiry releases the unit when payment session expires
func (h *PaymentsHandler) handleSessionExpiry(attributes map[string]interface{}) {
	// Extract session ID or payment ID from attributes
	// The exact field depends on PayMongo's webhook structure
	sessionID, ok := attributes["id"].(string)
	if !ok {
		return
	}

	// Find rental by payment/session ID
	rentals := h.repo.GetAllRentals()
	for _, rental := range rentals {
		if rental.PaymentID == sessionID && rental.PaymentStatus == models.PaymentPending {
			// Release the unit
			if err := h.allocationManager.ReleaseUnit(rental.CollectibleID, rental.WarehouseID); err != nil {
				log.Printf("[Webhook] Failed to release unit for expired session: %v", err)
			}

			// Update rental status
			rental.PaymentStatus = models.PaymentFailed
			h.repo.UpdateRental(rental)

			log.Printf("[Webhook] Released unit for expired session: Rental %s", rental.ID)
			break
		}
	}
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
