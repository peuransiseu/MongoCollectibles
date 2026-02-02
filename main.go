package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/config"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/handlers"
	"github.com/mongocollectibles/rental-system/models"
	"github.com/mongocollectibles/rental-system/services"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize repository
	var repo data.Repository
	if os.Getenv("USE_DYNAMODB") == "true" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}
		repo = data.NewDynamoDBRepository(awsCfg)
		log.Println("Using DynamoDB Repository")

		// Auto-seed if empty
		collectibles, err := repo.GetAllCollectibles()
		if err != nil {
			log.Printf("Error checking database content: %v", err)
		} else {
			log.Printf("Found %d collectibles in database", len(collectibles))
			if len(collectibles) == 0 {
				log.Println("Database is empty. Seeding data...")
				data.SeedData(repo)
			}
		}
	} else {
		repo = data.NewRepository()
		log.Println("Using In-Memory Repository")
		data.SeedData(repo)
	}

	// Optional: Reset rentals if requested (Useful for demos/testing during restart)
	if os.Getenv("RESET_RENTALS") == "true" {
		log.Println("RESET_RENTALS=true detected. Clearing all rental records...")
		if err := repo.DeleteAllRentals(); err != nil {
			log.Printf("Error clearing rentals: %v", err)
		} else {
			log.Println("All rentals cleared successfully.")
		}
	}

	// Initialize services
	pricingService := services.NewPricingService()

	// Bridge: Transform legacy data for new AllocationManager
	log.Println("Initializing AllocationManager with warehouse data...")
	allWarehouses, _ := repo.GetAllWarehouses()
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

	// Sync with persistent storage (fix for inventory reset on restart)
	log.Println("Syncing inventory with persistent rentals...")
	if allRentals, err := repo.GetAllRentals(); err == nil {
		allocationManager.SyncInventory(allRentals)
	} else {
		log.Printf("Warning: Failed to fetch rentals for sync: %v", err)
	}

	// VALIDATION: Enforce Minimum 3 Stores Rule PER WAREHOUSE
	// We iterate through every physical warehouse node to ensure full connectivity.
	for _, wh := range newDistances {
		storeCount := len(wh.Distances)
		if storeCount < 3 {
			log.Fatalf("System Startup Failed: Constraint Violation. Warehouse '%s' only has %d stores connected (Minimum 3 required).", wh.ID, storeCount)
		}
	}
	log.Printf("System Validation Passed: All warehouses meet connectivity requirements.", len(newDistances))

	// Start reservation cleanup job (Run every 5 mins, expire after 15 mins)
	allocationManager.StartCleanupJob(5*time.Minute, 15*time.Minute)

	paymentService := services.NewPaymentService(cfg.PayMongoSecretKey, cfg.PayMongoPublicKey)

	// Initialize handlers
	collectiblesHandler := handlers.NewCollectiblesHandler(repo, allocationManager)
	rentalsHandler := handlers.NewRentalsHandler(repo, pricingService, allocationManager, paymentService, cfg)
	paymentsHandler := handlers.NewPaymentsHandler(repo, paymentService, allocationManager)
	adminHandler := handlers.NewAdminHandler(repo, allocationManager)

	// Setup router
	router := mux.NewRouter()

	// --- Admin Configuration ---
	// Use path prefix instead of host for simpler access
	adminRouter := router.PathPrefix("/admin").Subrouter()

	// API Route for admin data
	adminRouter.HandleFunc("/dashboard/api", adminHandler.GetDashboardData).Methods("GET")

	// Serve admin static files at /admin/
	// We need StripPrefix so the file server doesn't look for /admin/ inside static/admin/
	adminRouter.PathPrefix("/").Handler(http.StripPrefix("/admin/", http.FileServer(http.Dir("./static/admin"))))

	// --- Main App Routes ---
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

	// Serve static files (Catch-all for main app)
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
