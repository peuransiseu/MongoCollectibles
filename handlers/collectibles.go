package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

// CollectiblesHandler handles collectible-related endpoints
type CollectiblesHandler struct {
	repo              data.Repository
	allocationManager *services.AllocationManager
}

// NewCollectiblesHandler creates a new collectibles handler
func NewCollectiblesHandler(repo data.Repository, allocationManager *services.AllocationManager) *CollectiblesHandler {
	return &CollectiblesHandler{
		repo:              repo,
		allocationManager: allocationManager,
	}
}

// GetAllCollectibles returns all available collectibles
func (h *CollectiblesHandler) GetAllCollectibles(w http.ResponseWriter, r *http.Request) {
	collectibles, err := h.repo.GetAllCollectibles()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch collectibles",
		})
		return
	}

	// Ensure empty slice instead of nil for JSON
	if collectibles == nil {
		collectibles = []*models.Collectible{}
	}

	// Sort collectibles by Name to ensure consistent order
	sort.Slice(collectibles, func(i, j int) bool {
		return collectibles[i].Name < collectibles[j].Name
	})

	// Get store_id from query params, default to "store-a"
	targetStore := r.URL.Query().Get("store_id")
	if targetStore == "" {
		targetStore = "store-a"
	}

	for _, c := range collectibles {
		c.Stock = h.allocationManager.GetTotalStock(c.ID)

		// Calculate ETA logic based on target store
		// If stock is available locally (distance 0 or handled by logic), ETA is 0 or 1
		eta, err := h.allocationManager.GetETA(c.ID, targetStore)
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
