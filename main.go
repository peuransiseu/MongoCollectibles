package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mongocollectibles/rental-system/config"
	"github.com/mongocollectibles/rental-system/data"
	"github.com/mongocollectibles/rental-system/handlers"
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
	allocationService := services.NewAllocationService()
	allocationService.SetWarehouses(repo.GetAllWarehouses())
	paymentService := services.NewPaymentService(cfg.PayMongoSecretKey, cfg.PayMongoPublicKey)

	// Initialize handlers
	collectiblesHandler := handlers.NewCollectiblesHandler(repo)
	rentalsHandler := handlers.NewRentalsHandler(repo, pricingService, allocationService, paymentService, cfg)
	paymentsHandler := handlers.NewPaymentsHandler(repo, paymentService)

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
