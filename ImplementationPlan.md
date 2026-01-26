MongoCollectibles Rental System - Implementation Plan
A comprehensive rental management system built with Go backend and modern web frontend, featuring intelligent warehouse allocation, dynamic pricing, and PayMongo payment integration.

User Review Required
IMPORTANT

PayMongo API Credentials: You'll need to provide PayMongo API keys (test/production) for payment integration. The system will be designed to accept these via environment variables.

IMPORTANT

Warehouse Distance Input Format: The system will accept warehouse distances as a list of tuples per collectible, e.g., [(1,4,5), (3,2,3)] where each tuple represents distances from warehouses to stores A, B, C respectively. Please confirm this format works for your data structure.

NOTE

Payment Methods: The system will support Cards, GCash, GrabPay, and BPI/UBP Direct Online Banking through PayMongo's payment method APIs.

Proposed Changes
Backend - Go Application
The backend will be structured using a clean architecture pattern with clear separation of concerns:

[NEW] 
main.go
Entry point for the application. Initializes HTTP server, routes, and dependencies.

[NEW] 
go.mod
Go module definition with dependencies:

github.com/gorilla/mux - HTTP routing
github.com/joho/godotenv - Environment variable management
PayMongo SDK or HTTP client for API integration
Domain Models (models/)
[NEW] 
models/collectible.go
Defines core domain entities:

Collectible struct with ID, Name, Size (S/M/L), WarehouseDistances
Size enum type with pricing constants
Store struct representing brick-and-mortar locations
Warehouse struct with location and inventory
[NEW] 
models/rental.go
Rental transaction models:

Rental struct with CollectibleID, StoreID, Duration, TotalFee
Customer struct with billing details
PaymentMethod enum (Card, GCash, GrabPay, BankTransfer)
Business Logic (services/)
[NEW] 
services/allocation.go
Warehouse allocation algorithm:

AllocateWarehouse(collectibleID, storeID) - Finds nearest warehouse with available unit
Uses distance tuples to calculate minimum distance
Returns allocated warehouse ID or error if unavailable
[NEW] 
services/pricing.go
Rental fee calculation:

CalculateRentalFee(size, duration) - Applies pricing rules
Base rates: S=1000, M=5000, L=10000 PHP/day
Minimum 7-day duration for normal rate
Double rate for <7 days
Returns total fee in PHP
[NEW] 
services/payment.go
PayMongo API integration:

CreatePaymentIntent(amount, method, billingDetails) - Initiates payment
HandlePaymentCallback(paymentID) - Processes webhooks
Supports Cards, E-wallets (GCash/GrabPay), Direct Banking (BPI/UBP)
Returns payment URL for customer checkout
API Handlers (handlers/)
[NEW] 
handlers/collectibles.go
REST endpoints:

GET /api/collectibles - List available collectibles
GET /api/collectibles/:id - Get collectible details with warehouse info
[NEW] 
handlers/rentals.go
Rental management endpoints:

POST /api/rentals/quote - Calculate rental fee quote
POST /api/rentals/checkout - Create rental and initiate payment
Request body includes: collectibleID, storeID, duration, paymentMethod, billingDetails
[NEW] 
handlers/payments.go
Payment webhook handling:

POST /api/webhooks/paymongo - Receives payment status updates
Updates rental status based on payment confirmation
Configuration (config/)
[NEW] 
config/config.go
Application configuration:

Loads environment variables (PayMongo API keys, server port)
Defines store and warehouse data structures
Initializes default stores (minimum 3, expandable)
[NEW] 
.env.example
Environment variable template:

PAYMONGO_SECRET_KEY=sk_test_...
PAYMONGO_PUBLIC_KEY=pk_test_...
SERVER_PORT=8080
Frontend - Web Application
Modern, responsive web interface for customer rental experience.

[NEW] 
static/index.html
Main application page with:

Collectible browsing grid with size/pricing display
Store selection dropdown
Rental duration input with dynamic fee calculation
Checkout button
[NEW] 
static/css/styles.css
Premium design system:

Modern color palette with gradients
Responsive grid layouts
Smooth animations and transitions
Mobile-first approach
[NEW] 
static/js/app.js
Frontend application logic:

Fetches collectibles from API
Real-time rental fee calculation
Handles checkout flow
Redirects to PayMongo payment page
[NEW] 
static/js/checkout.js
Checkout process:

Payment method selection UI (Cards, GCash, GrabPay, BPI/UBP)
Billing details form with validation
API integration for rental creation
Payment redirect handling
Data Layer (data/)
[NEW] 
data/repository.go
In-memory data store (can be replaced with MongoDB later):

Stores collectibles with warehouse distance data
Manages rental records
Tracks warehouse inventory availability
[NEW] 
data/seed.go
Sample data initialization:

Creates 3 default stores (A, B, C)
Seeds collectibles with varied sizes
Example warehouse distances: [(1,4,5), (3,2,3), (2,1,4)]
Verification Plan
Automated Tests
Unit Tests:

go test ./services/... -v
Test allocation algorithm with various distance scenarios
Verify pricing calculations for all size/duration combinations
Mock PayMongo API responses
Integration Tests:

go test ./handlers/... -v
Test API endpoints with sample requests
Verify end-to-end rental creation flow
Manual Verification
Local Development Server:

go run main.go
Access web interface at http://localhost:8080
Test collectible browsing and selection
Verify real-time fee calculation
Payment Flow Testing:

Use PayMongo test credentials
Complete checkout with test card numbers
Verify payment redirect and webhook handling
Warehouse Allocation Testing:

Create rentals from different stores
Verify system selects nearest warehouse
Test behavior when warehouses are unavailable
Browser Testing:

Test responsive design on mobile/tablet/desktop
Verify all payment methods display correctly
Ensure billing form validation works