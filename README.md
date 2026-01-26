# MongoCollectibles - Quick Start Guide

## ✅ System Status

**Server Status:** ✅ Running on http://localhost:8080  
**Dependencies:** ✅ All Go modules downloaded  
**API Status:** ✅ Verified working

## 🚀 Access the Application

Open your web browser and navigate to:
```
http://localhost:8080
```

## 🧪 Verified Functionality

### API Endpoints Tested

✅ **GET /api/collectibles** - Returns all 6 collectibles:
- Vintage Batman Action Figure (Small - ₱1,000/day)
- Star Wars Millennium Falcon Model (Medium - ₱5,000/day)
- Life-Size Iron Man Suit (Large - ₱10,000/day)
- Pokemon Card Collection Set (Small - ₱1,000/day)
- Gundam Perfect Grade Model (Medium - ₱5,000/day)
- Arcade Machine - Street Fighter II (Large - ₱10,000/day)

### Features Available

1. **Browse Collectibles** - View all available items with images and descriptions
2. **Select Store** - Choose from 3 pickup locations (Manila, Quezon City, Makati)
3. **Calculate Quote** - Real-time rental fee calculation based on duration
4. **Special Rates** - Automatic 2x rate for rentals under 7 days
5. **Payment Methods** - Cards, GCash, GrabPay, BPI/UBP Online Banking
6. **Checkout Flow** - Complete billing details form and payment processing

## 📝 How to Use

1. **Select a Store** - Use the dropdown at the top to choose your pickup location
2. **Browse Collectibles** - Click on any collectible card to view details
3. **Configure Rental** - Set rental duration (minimum 1 day)
4. **View Quote** - See real-time pricing with special rate warnings
5. **Choose Payment** - Select your preferred payment method
6. **Enter Details** - Fill in billing information
7. **Checkout** - Click "Proceed to Payment" to complete

## 🔧 Configuration

To enable PayMongo payments:

1. Create a `.env` file from the template:
   ```bash
   copy .env.example .env
   ```

2. Add your PayMongo credentials:
   ```
   PAYMONGO_SECRET_KEY=sk_test_your_key_here
   PAYMONGO_PUBLIC_KEY=pk_test_your_key_here
   ```

3. Restart the server:
   ```bash
   # Stop current server (Ctrl+C)
   go run main.go
   ```

## 🎯 Testing Scenarios

### Test Normal Rate (7+ days)
1. Select any collectible
2. Set duration to 7 days
3. Verify normal rate applies (₱1,000, ₱5,000, or ₱10,000/day)

### Test Special Rate (<7 days)
1. Select any collectible
2. Set duration to 3 days
3. Verify special rate warning appears
4. Verify rate is doubled (₱2,000, ₱10,000, or ₱20,000/day)

### Test Warehouse Allocation
1. Select different stores from dropdown
2. Rent the same collectible
3. System automatically allocates nearest warehouse

## 📊 Sample Calculations

**Small Collectible (7 days):**
- Daily Rate: ₱1,000
- Total: ₱7,000

**Medium Collectible (3 days - Special Rate):**
- Daily Rate: ₱10,000 (2x ₱5,000)
- Total: ₱30,000

**Large Collectible (14 days):**
- Daily Rate: ₱10,000
- Total: ₱140,000

## 🛑 Stop the Server

Press `Ctrl+C` in the terminal where the server is running.

## 📚 Additional Documentation

- **Implementation Plan:** See `implementation_plan.md` in artifacts
- **Walkthrough:** See `walkthrough.md` in artifacts
- **Full README:** See `README.md` in project root

## 🎉 You're All Set!

The MongoCollectibles rental system is fully operational and ready to use!