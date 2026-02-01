# MongoCollectibles API Documentation

## Overview

This document provides comprehensive API documentation for the MongoCollectibles rental system, including the newly implemented user authentication, cart system, checkout flow, order management, cancellation, and refund handling.

**Base URL:** `http://localhost:8080`

---

## Table of Contents

1. [Authentication](#authentication)
2. [Cart Management](#cart-management)
3. [Checkout](#checkout)
4. [Order Management](#order-management)
5. [Collectibles](#collectibles)
6. [Rentals (Legacy)](#rentals-legacy)
7. [Payment Webhooks](#payment-webhooks)
8. [Error Handling](#error-handling)
9. [Order Status Lifecycle](#order-status-lifecycle)
10. [Cancellation & Refund Rules](#cancellation--refund-rules)

---

## Authentication

### Register User

**Endpoint:** `POST /api/auth/register`

**Description:** Register a new user account with email and password.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "token": "dGVzdC1zZXNzaW9uLXRva2Vu..."
  }
}
```

**Response (Error - 409 Conflict):**
```json
{
  "success": false,
  "error": "email already registered"
}
```

---

### Login

**Endpoint:** `POST /api/auth/login`

**Description:** Login with email and password to receive a session token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "token": "dGVzdC1zZXNzaW9uLXRva2Vu..."
  }
}
```

**Response (Error - 401 Unauthorized):**
```json
{
  "success": false,
  "error": "Invalid email or password"
}
```

---

### Logout

**Endpoint:** `POST /api/auth/logout`

**Description:** Logout and invalidate the current session token.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

## Cart Management

**Note:** All cart endpoints require authentication via the `Authorization` header.

### Get Active Cart

**Endpoint:** `GET /api/cart`

**Description:** Retrieve the user's active shopping cart. Creates a new empty cart if none exists.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "id": "cart-550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "ACTIVE",
    "items": [
      {
        "collectible_id": "collectible-1",
        "store_id": "store-1",
        "rental_days": 7,
        "quantity": 2
      }
    ],
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-01T10:05:00Z"
  }
}
```

**Important:** Cart is **intent only** - no stock checks or allocation happens at this stage.

---

### Add Item to Cart

**Endpoint:** `POST /api/cart/items`

**Description:** Add an item to the cart. If the item already exists, quantity is incremented.

**Headers:**
```
Authorization: <session-token>
```

**Request Body:**
```json
{
  "collectible_id": "collectible-1",
  "store_id": "store-1",
  "rental_days": 7,
  "quantity": 2
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Item added to cart",
  "data": {
    "id": "cart-550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "ACTIVE",
    "items": [
      {
        "collectible_id": "collectible-1",
        "store_id": "store-1",
        "rental_days": 7,
        "quantity": 2
      }
    ],
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-01T10:05:00Z"
  }
}
```

---

### Update Cart Item

**Endpoint:** `PUT /api/cart/items/{collectible_id}`

**Description:** Update the quantity or rental days for a specific cart item.

**Headers:**
```
Authorization: <session-token>
```

**Request Body:**
```json
{
  "rental_days": 14,
  "quantity": 3
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Cart item updated",
  "data": { /* updated cart */ }
}
```

---

### Remove Item from Cart

**Endpoint:** `DELETE /api/cart/items/{collectible_id}`

**Description:** Remove a specific item from the cart.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Item removed from cart",
  "data": { /* updated cart */ }
}
```

---

### Clear Cart

**Endpoint:** `DELETE /api/cart`

**Description:** Remove all items from the cart.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Cart cleared",
  "data": { /* empty cart */ }
}
```

---

## Checkout

### Checkout from Cart

**Endpoint:** `POST /api/checkout`

**Description:** Process checkout from the user's cart. **THIS IS WHERE ALLOCATION HAPPENS** - units are allocated during checkout, not when adding to cart.

**Headers:**
```
Authorization: <session-token>
```

**Request Body:** None (uses active cart)

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "order_id": "order-550e8400-e29b-41d4-a716-446655440000",
    "total": 1500.00,
    "payment_url": "https://checkout.paymongo.com/...",
    "status": "PENDING_PAYMENT"
  }
}
```

**Response (Error - 404 Not Found):**
```json
{
  "success": false,
  "error": "No active cart found"
}
```

**Response (Error - 409 Conflict):**
```json
{
  "success": false,
  "error": "Some items could not be allocated",
  "failed": ["collectible-1", "collectible-2"]
}
```

**Checkout Flow:**
1. Validates cart is not empty
2. **Allocates units for each cart item** (critical allocation point)
3. Creates payment session
4. Creates order with `PENDING_PAYMENT` status
5. Marks cart as `CHECKED_OUT`
6. Returns payment URL for user to complete payment

**Error Handling:**
- If allocation fails for any item → releases all allocated units
- If payment session creation fails → releases all allocated units
- If order creation fails → releases all allocated units

---

## Order Management

### Get All Orders

**Endpoint:** `GET /api/orders`

**Description:** Retrieve all orders for the authenticated user.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "order-550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "store_id": "store-1",
      "status": "PAID",
      "total_amount": 1500.00,
      "items": [
        {
          "collectible_id": "collectible-1",
          "collectible_name": "Funko Pop - Iron Man",
          "rental_days": 7,
          "eta_days": 3,
          "price": 750.00
        }
      ],
      "payment_id": "payment-123",
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-01T10:05:00Z"
    }
  ]
}
```

**Note:** `unit_id` and `warehouse_id` are **NOT** exposed to the frontend (marked with `json:"-"`).

---

### Get Order by ID

**Endpoint:** `GET /api/orders/{id}`

**Description:** Retrieve details of a specific order.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "id": "order-550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "store_id": "store-1",
    "status": "PAID",
    "total_amount": 1500.00,
    "items": [ /* order items */ ],
    "payment_id": "payment-123",
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-01T10:05:00Z"
  }
}
```

**Response (Error - 403 Forbidden):**
```json
{
  "success": false,
  "error": "Access denied"
}
```

---

### Cancel Order

**Endpoint:** `POST /api/orders/{id}/cancel`

**Description:** Cancel an order and process refund if applicable. Automatically releases allocated units for eligible orders.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "order_id": "order-550e8400-e29b-41d4-a716-446655440000",
    "status": "REFUNDED",
    "refund_amount": 1500.00,
    "refund_reason": "Full refund - order not shipped",
    "refund_id": "refund-550e8400-e29b-41d4-a716-446655440000",
    "refund_status": "PROCESSED"
  }
}
```

**Response (Error - 400 Bad Request):**
```json
{
  "success": false,
  "error": "Cannot cancel delivered or completed orders"
}
```

**Cancellation Rules:** See [Cancellation & Refund Rules](#cancellation--refund-rules)

---

### Get Refund Status

**Endpoint:** `GET /api/orders/{id}/refund`

**Description:** Retrieve refund status for a cancelled order.

**Headers:**
```
Authorization: <session-token>
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "id": "refund-550e8400-e29b-41d4-a716-446655440000",
    "order_id": "order-550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 1500.00,
    "reason": "Full refund - order not shipped",
    "status": "PROCESSED",
    "created_at": "2026-02-01T10:10:00Z",
    "updated_at": "2026-02-01T10:10:05Z"
  }
}
```

**Response (Error - 404 Not Found):**
```json
{
  "success": false,
  "error": "No refund found for this order"
}
```

---

## Collectibles

### Get All Collectibles

**Endpoint:** `GET /api/collectibles`

**Description:** Retrieve all available collectibles.

**Response (Success - 200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "collectible-1",
      "name": "Funko Pop - Iron Man",
      "size": "MEDIUM",
      "stock": 5,
      "eta_days": 3
    }
  ]
}
```

---

### Get Collectible by ID

**Endpoint:** `GET /api/collectibles/{id}`

**Description:** Retrieve details of a specific collectible.

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "id": "collectible-1",
    "name": "Funko Pop - Iron Man",
    "size": "MEDIUM",
    "stock": 5,
    "eta_days": 3
  }
}
```

---

## Rentals (Legacy)

### Get Quote

**Endpoint:** `POST /api/rentals/quote`

**Description:** Get a rental quote for a collectible.

**Request Body:**
```json
{
  "collectible_id": "collectible-1",
  "store_id": "store-1",
  "rental_days": 7
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "collectible_id": "collectible-1",
    "collectible_name": "Funko Pop - Iron Man",
    "rental_days": 7,
    "daily_rate": 100.00,
    "total_fee": 700.00,
    "eta_days": 3,
    "stock_available": 5
  }
}
```

---

### Checkout (Legacy - Single Item)

**Endpoint:** `POST /api/rentals/checkout`

**Description:** Legacy checkout endpoint for single item rental (without cart).

**Request Body:**
```json
{
  "collectible_id": "collectible-1",
  "store_id": "store-1",
  "rental_days": 7,
  "customer_email": "user@example.com"
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "data": {
    "rental_id": "rental-550e8400-e29b-41d4-a716-446655440000",
    "payment_url": "https://checkout.paymongo.com/...",
    "total_fee": 700.00
  }
}
```

**Note:** This endpoint is maintained for backward compatibility. New implementations should use `/api/checkout` with the cart system.

---

## Payment Webhooks

### PayMongo Webhook

**Endpoint:** `POST /api/webhooks/paymongo`

**Description:** Webhook endpoint for PayMongo payment events.

**Request Body:** PayMongo webhook payload

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Webhook processed"
}
```

---

### Payment Success Redirect

**Endpoint:** `GET /payment/success?order_id={order_id}`

**Description:** Redirect endpoint after successful payment.

---

### Payment Failed Redirect

**Endpoint:** `GET /payment/failed?order_id={order_id}`

**Description:** Redirect endpoint after failed payment.

---

## Error Handling

All API endpoints return errors in the following format:

```json
{
  "success": false,
  "error": "Error message description"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (access denied)
- `404` - Not Found (resource not found)
- `409` - Conflict (duplicate resource or allocation failure)
- `500` - Internal Server Error

---

## Order Status Lifecycle

```
PENDING_PAYMENT → PAID → ALLOCATED → IN_TRANSIT → READY_FOR_PICKUP → COMPLETED
                                                                     ↓
                                                               CANCELLED → REFUNDED
```

**Status Definitions:**

| Status | Description |
|---|---|
| `PENDING_PAYMENT` | Order created, awaiting payment |
| `PAID` | Payment completed, units allocated |
| `ALLOCATED` | Units confirmed allocated |
| `IN_TRANSIT` | Shipment in progress |
| `READY_FOR_PICKUP` | Available at store |
| `COMPLETED` | Rental completed |
| `CANCELLED` | Order cancelled |
| `REFUNDED` | Refund processed |

---

## Cancellation & Refund Rules

| Order Status | Can Cancel | Refund Amount | Unit Release |
|---|---|---|---|
| `PENDING_PAYMENT` | ✅ Yes | None | N/A (not allocated) |
| `PAID` | ✅ Yes | 100% | ✅ Released |
| `ALLOCATED` | ✅ Yes | 100% | ✅ Released |
| `IN_TRANSIT` | ⚠️ Conditional | 50% | ❌ Not released |
| `READY_FOR_PICKUP` | ❌ No | None | ❌ Not released |
| `COMPLETED` | ❌ No | None | ❌ Not released |
| `CANCELLED` | ❌ No | None | N/A |
| `REFUNDED` | ❌ No | None | N/A |

**Refund Processing:**
- Refunds are **idempotent** - multiple cancellation requests return the same refund
- Refund status: `PENDING` → `PROCESSED` → `FAILED`
- Refunds are automatically created when eligible orders are cancelled

---

## Authentication Flow

1. **Register:** `POST /api/auth/register` → Receive session token
2. **Login:** `POST /api/auth/login` → Receive session token
3. **Use Token:** Include token in `Authorization` header for protected endpoints
4. **Logout:** `POST /api/auth/logout` → Invalidate token

**Example:**
```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123"}'

# Use token for protected endpoints
curl -X GET http://localhost:8080/api/cart \
  -H "Authorization: <token-from-register>"
```

---

## Complete User Flow Example

### 1. Register & Login
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123"}'
```

### 2. Add Items to Cart
```bash
curl -X POST http://localhost:8080/api/cart/items \
  -H "Authorization: <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "collectible_id":"collectible-1",
    "store_id":"store-1",
    "rental_days":7,
    "quantity":2
  }'
```

### 3. Checkout (Allocation Happens)
```bash
curl -X POST http://localhost:8080/api/checkout \
  -H "Authorization: <token>"
```

### 4. Track Order
```bash
curl -X GET http://localhost:8080/api/orders \
  -H "Authorization: <token>"
```

### 5. Cancel Order
```bash
curl -X POST http://localhost:8080/api/orders/{order-id}/cancel \
  -H "Authorization: <token>"
```

### 6. Check Refund Status
```bash
curl -X GET http://localhost:8080/api/orders/{order-id}/refund \
  -H "Authorization: <token>"
```

---

## Key Design Principles

1. **No Allocation Until Checkout**
   - Cart is **intent only** - no stock checks or allocation
   - Allocation happens **only during checkout** (`POST /api/checkout`)
   - Orders contain allocated unit details (internal only)

2. **Frontend Never Sees Warehouse/Unit Details**
   - `unit_id` and `warehouse_id` are marked with `json:"-"`
   - Frontend only sees `collectible_id`, `eta_days`, and `price`

3. **Automatic Resource Cleanup**
   - Allocation failure → Release all allocated units
   - Payment failure → Release all allocated units
   - Order cancellation → Release allocated units (if eligible)

4. **Idempotency**
   - Refund creation is idempotent
   - Multiple calls with same `order_id` return same refund

---

## Rate Limiting

**Note:** Currently not implemented. Future versions will include rate limiting to prevent abuse.

---

## Security Considerations

**Implemented:**
- ✅ Password hashing with bcrypt
- ✅ Session token authentication
- ✅ User ID validation from token
- ✅ Order ownership verification

**Not Implemented (Future):**
- ❌ Auth middleware (routes currently unprotected)
- ❌ JWT tokens
- ❌ Token expiration
- ❌ Rate limiting
- ❌ HTTPS enforcement

---

## Support

For issues or questions, please contact the development team or refer to the project repository.

---

**Last Updated:** 2026-02-01  
**API Version:** 1.0.0
