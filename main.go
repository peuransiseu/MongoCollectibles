package main

import (
	"context"
	"log"
	"net/http"

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

	// Start background cleanup job for expired reservations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go allocationManager.StartCleanupJob(ctx)

	paymentService := services.NewPaymentService(cfg.PayMongoSecretKey, cfg.PayMongoPublicKey)

	// Initialize auth service and handlers
	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(repo, authService)
	cartHandler := handlers.NewCartHandler(repo, authHandler)

	// Initialize order and refund services
	orderService := services.NewOrderService(allocationManager)
	refundService := services.NewRefundService(repo)
	orderHandler := handlers.NewOrderHandler(repo, authHandler, orderService, refundService)

	// Initialize handlers
	collectiblesHandler := handlers.NewCollectiblesHandler(repo, allocationManager)
	rentalsHandler := handlers.NewRentalsHandler(repo, pricingService, allocationManager, paymentService, cfg, authHandler)
	paymentsHandler := handlers.NewPaymentsHandler(repo, paymentService, allocationManager)

	// Setup router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Auth endpoints
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	// Cart endpoints (require auth)
	api.HandleFunc("/cart", cartHandler.GetCart).Methods("GET")
	api.HandleFunc("/cart/items", cartHandler.AddToCart).Methods("POST")
	api.HandleFunc("/cart/items/{collectible_id}", cartHandler.UpdateCartItem).Methods("PUT")
	api.HandleFunc("/cart/items/{collectible_id}", cartHandler.RemoveFromCart).Methods("DELETE")
	api.HandleFunc("/cart", cartHandler.ClearCart).Methods("DELETE")

	// Collectibles endpoints
	api.HandleFunc("/collectibles", collectiblesHandler.GetAllCollectibles).Methods("GET")
	api.HandleFunc("/collectibles/{id}", collectiblesHandler.GetCollectibleByID).Methods("GET")

	// Rentals endpoints
	api.HandleFunc("/rentals/quote", rentalsHandler.GetQuote).Methods("POST")
	api.HandleFunc("/rentals/checkout", rentalsHandler.Checkout).Methods("POST")
	api.HandleFunc("/checkout", rentalsHandler.CheckoutFromCart).Methods("POST") // NEW: Checkout from cart

	// Order endpoints (require auth)
	api.HandleFunc("/orders", orderHandler.GetOrders).Methods("GET")
	api.HandleFunc("/orders/{id}", orderHandler.GetOrderByID).Methods("GET")
	api.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrder).Methods("POST")
	api.HandleFunc("/orders/{id}/refund", orderHandler.GetRefundStatus).Methods("GET")

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
