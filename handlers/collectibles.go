package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/services"
)

// CollectiblesHandler handles collectible-related endpoints
type CollectiblesHandler struct {
	repo              *data.Repository
	allocationManager *services.AllocationManager
}

// NewCollectiblesHandler creates a new collectibles handler
func NewCollectiblesHandler(repo *data.Repository, allocationManager *services.AllocationManager) *CollectiblesHandler {
	return &CollectiblesHandler{
		repo:              repo,
		allocationManager: allocationManager,
	}
}

// GetAllCollectibles returns all available collectibles
func (h *CollectiblesHandler) GetAllCollectibles(w http.ResponseWriter, r *http.Request) {
	collectibles := h.repo.GetAllCollectibles()

	// Populate dynamic fields (stock and ETA)
	// Use first store as default for ETA calculation
	defaultStoreID := "store-a"

	for _, c := range collectibles {
		c.Stock = h.allocationManager.GetTotalStock(c.ID)
		eta, err := h.allocationManager.GetETA(c.ID, defaultStoreID)
		if err == nil {
			c.ETADays = eta
		} else {
			c.ETADays = 0 // No units available
		}
		// Set daily rate based on size
		c.DailyRate = c.Size.GetDailyRate()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    collectibles,
	})
}

// GetCollectibleByID returns a specific collectible with warehouse information
func (h *CollectiblesHandler) GetCollectibleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	collectible, err := h.repo.GetCollectibleByID(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Collectible not found",
		})
		return
	}

	warehouses, _ := h.repo.GetWarehouses(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"data":       collectible,
		"warehouses": warehouses,
	})
}
