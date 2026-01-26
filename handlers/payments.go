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
	repo           *data.Repository
	paymentService *services.PaymentService
}

// NewPaymentsHandler creates a new payments handler
func NewPaymentsHandler(repo *data.Repository, paymentService *services.PaymentService) *PaymentsHandler {
	return &PaymentsHandler{
		repo:           repo,
		paymentService: paymentService,
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

	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	paymentID, ok := attributes["id"].(string)
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	rental.PaymentStatus = models.PaymentFailed
	h.repo.UpdateRental(rental)

	// Redirect to failure page
	http.Redirect(w, r, "/failed.html?rental_id="+rentalID, http.StatusSeeOther)
}
