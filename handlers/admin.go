package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/services"
)

// AdminHandler handles admin dashboard requests
type AdminHandler struct {
	repo              data.Repository
	allocationManager *services.AllocationManager
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo data.Repository, allocationManager *services.AllocationManager) *AdminHandler {
	return &AdminHandler{
		repo:              repo,
		allocationManager: allocationManager,
	}
}

// GetDashboardData aggregates all data for the dashboard
func (h *AdminHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	// 1. Get Inventory State
	inventory := h.allocationManager.GetAllInventory()

	// 2. Get Orders (Rentals)
	rentals, err := h.repo.GetAllRentals()
	if err != nil {
		http.Error(w, "Failed to fetch rentals", http.StatusInternalServerError)
		return
	}

	// 3. Get Collectibles (for names)
	collectibles, _ := h.repo.GetAllCollectibles()
	// Map for faster lookup
	collectibleNames := make(map[string]string)
	for _, c := range collectibles {
		collectibleNames[c.ID] = c.Name
	}

	response := map[string]interface{}{
		"inventory":    inventory,
		"rentals":      rentals,
		"collectibles": collectibleNames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
