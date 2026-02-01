package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/config"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/handlers"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize repository and seed data
	repo := data.NewRepository()
	data.SeedData(repo)

	// Initialize services
	pricingService := services.NewPricingService()

	// Bridge: Transform legacy data for new AllocationManager
	log.Println("Initializing AllocationManager with warehouse data...")
	allWarehouses := repo.GetAllWarehouses()
	var newInventory []*models.CollectibleUnit
	var newDistances []models.WarehouseNode
	seenWarehouses := make(map[string]bool)

	for collectibleID, warehouseList := range allWarehouses {
		for _, wh := range warehouseList {
			// Create Unit
			unit := &models.CollectibleUnit{
				ID:            wh.ID, // Assuming Unit ID = Warehouse ID for legacy
				CollectibleID: collectibleID,
				WarehouseID:   wh.ID,
				IsAvailable:   wh.Available,
			}
			newInventory = append(newInventory, unit)

			// Create Warehouse Node (Physical)
			if !seenWarehouses[wh.ID] {
				// Use the Distances map directly from the warehouse data
				node := models.WarehouseNode{
					ID:        wh.ID,
					Distances: wh.Distances,
				}
				newDistances = append(newDistances, node)
				seenWarehouses[wh.ID] = true
			}
		}
	}

	allocationManager := services.NewAllocationManager(newInventory, newDistances)

	// Start reservation cleanup job (Run every 5 mins, expire after 15 mins)
	allocationManager.StartCleanupJob(5*time.Minute, 15*time.Minute)

	paymentService := services.NewPaymentService(cfg.PayMongoSecretKey, cfg.PayMongoPublicKey)

	// Initialize handlers
	collectiblesHandler := handlers.NewCollectiblesHandler(repo, allocationManager)
	rentalsHandler := handlers.NewRentalsHandler(repo, pricingService, allocationManager, paymentService, cfg)
	paymentsHandler := handlers.NewPaymentsHandler(repo, paymentService, allocationManager)

	// Setup router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Collectibles endpoints
	api.HandleFunc("/collectibles", collectiblesHandler.GetAllCollectibles).Methods("GET")
	api.HandleFunc("/collectibles/{id}", collectiblesHandler.GetCollectibleByID).Methods("GET")

	// Rentals endpoints
	api.HandleFunc("/rentals/quote", rentalsHandler.GetQuote).Methods("POST")
	api.HandleFunc("/rentals/checkout", rentalsHandler.Checkout).Methods("POST")

	// Payment endpoints
	api.HandleFunc("/webhooks/paymongo", paymentsHandler.WebhookPayMongo).Methods("POST")
	router.HandleFunc("/payment/success", paymentsHandler.PaymentSuccess).Methods("GET")
	router.HandleFunc("/payment/failed", paymentsHandler.PaymentFailed).Methods("GET")

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Enable CORS for development
	corsRouter := enableCORS(router)

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Environment: %s", cfg.Environment)
	log.Fatal(http.ListenAndServe(addr, corsRouter))
}

// enableCORS adds CORS headers to responses
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
