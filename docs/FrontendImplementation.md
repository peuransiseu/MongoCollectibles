# Frontend Implementation Guide

## Overview
The MongoCollectibles frontend is a modern, premium web application built with vanilla HTML, CSS, and JavaScript. It features a responsive design with smooth animations, real-time quote calculation, and seamless PayMongo payment integration.

---

## Architecture

### Technology Stack
- **HTML5** - Semantic markup with SEO optimization
- **CSS3** - Modern design with gradients, glassmorphism, and animations
- **Vanilla JavaScript** - No frameworks, pure ES6+ code
- **Google Fonts** - Inter font family for premium typography

### File Structure
```
static/
├── index.html          # Main application page
├── success.html        # Payment success page
├── failed.html         # Payment failure page
├── css/
│   └── styles.css      # Complete design system
├── js/
│   ├── app.js          # Main application logic
│   └── checkout.js     # Checkout and payment handling
└── images/
    ├── batman.jpg
    ├── falcon.jpg
    ├── ironman.jpg
    ├── pokemon.jpg
    ├── gundam.jpg
    ├── arcade.jpg
    └── placeholder.jpg
```

---

## Design System

### Color Palette
```css
--primary: hsl(260, 80%, 60%)        /* Purple */
--primary-dark: hsl(260, 80%, 50%)   /* Dark purple */
--primary-light: hsl(260, 80%, 70%)  /* Light purple */
--secondary: hsl(200, 80%, 55%)      /* Blue */
--accent: hsl(320, 70%, 60%)         /* Pink */

--bg-dark: hsl(240, 20%, 10%)        /* Dark background */
--bg-card: hsl(240, 15%, 15%)        /* Card background */
--bg-card-hover: hsl(240, 15%, 18%)  /* Card hover state */

--text-primary: hsl(0, 0%, 95%)      /* White text */
--text-secondary: hsl(0, 0%, 70%)    /* Gray text */
--text-muted: hsl(0, 0%, 50%)        /* Muted text */
```

### Typography
- **Font Family**: Inter (Google Fonts)
- **Weights**: 300, 400, 500, 600, 700, 800
- **Base Size**: 16px
- **Line Height**: 1.6

### Spacing Scale
```css
--spacing-xs: 0.5rem   /* 8px */
--spacing-sm: 1rem     /* 16px */
--spacing-md: 1.5rem   /* 24px */
--spacing-lg: 2rem     /* 32px */
--spacing-xl: 3rem     /* 48px */
```

### Border Radius
```css
--radius-sm: 0.5rem    /* 8px */
--radius-md: 1rem      /* 16px */
--radius-lg: 1.5rem    /* 24px */
```

### Shadows & Effects
```css
--shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.3)
--shadow-md: 0 4px 16px rgba(0, 0, 0, 0.4)
--shadow-lg: 0 8px 32px rgba(0, 0, 0, 0.5)
--shadow-glow: 0 0 20px rgba(160, 100, 255, 0.3)
```

---

## Components

### 1. Header
**Location**: Top of page, sticky  
**Features**:
- Logo with gradient text effect
- Sticky positioning with backdrop blur
- Navigation buttons

**CSS Classes**:
- `.header` - Main container
- `.header-content` - Inner wrapper
- `.logo` - Gradient text logo

### 2. Hero Section
**Purpose**: Eye-catching banner with tagline  
**Features**:
- Large typography
- Gradient text effects
- Centered content

**CSS Classes**:
- `.hero` - Container
- `.hero-content` - Content wrapper
- `.hero-title` - Main heading
- `.hero-subtitle` - Subheading

### 3. Store Selection
**Purpose**: Choose pickup location  
**Features**:
- Dropdown with 3 stores
- Auto-selects first store
- Updates on change

**HTML**:
```html
<select id="storeSelect" class="form-select">
  <option value="store-a">Store A - Manila</option>
  <option value="store-b">Store B - Quezon City</option>
  <option value="store-c">Store C - Makati</option>
</select>
```

### 4. Collectibles Grid
**Purpose**: Display available items  
**Features**:
- Responsive grid (auto-fill, min 320px)
- Hover effects with elevation
- Click to open rental modal

**CSS Classes**:
- `.collectibles-grid` - Grid container
- `.collectible-card` - Individual card
- `.collectible-image` - Product image
- `.collectible-content` - Card content
- `.size-badge` - Size indicator (S/M/L)
- `.price` - Pricing display

**Card Structure**:
```html
<div class="collectible-card">
  <img src="/images/batman.jpg" class="collectible-image">
  <div class="collectible-content">
    <div class="collectible-header">
      <h3 class="collectible-name">Batman Figure</h3>
      <span class="size-badge size-s">S</span>
    </div>
    <p class="collectible-description">...</p>
    <div class="collectible-footer">
      <div class="price">₱1,000.00/day</div>
      <button class="btn btn-primary">Rent Now</button>
    </div>
  </div>
</div>
```

### 5. Rental Modal
**Purpose**: Configure rental and checkout  
**Features**:
- Overlay with backdrop blur
- Rental duration input
- Real-time quote calculation
- Payment method selection
- Billing details form

**CSS Classes**:
- `.modal` - Overlay container
- `.modal-content` - Modal box
- `.modal-close` - Close button
- `.form-group` - Form field wrapper
- `.form-input` - Text inputs
- `.form-select` - Dropdowns
- `.payment-method` - Payment option card
- `.quote-summary` - Price breakdown

**Modal Sections**:
1. Duration input
2. Quote summary (dynamic)
3. Special rate notice (conditional)
4. Payment method selection
5. Billing details form
6. Action buttons

---

## JavaScript Architecture

### app.js - Main Application

#### Global State
```javascript
let collectibles = [];      // All collectibles from API
let stores = [];            // Available stores
let selectedCollectible = null;  // Currently selected item
let selectedStore = null;   // Selected pickup location
```

#### Key Functions

**`loadStores()`**
- Initializes store data
- Populates dropdown
- Sets default selection

**`loadCollectibles()`**
- Fetches from `/api/collectibles`
- Stores in global state
- Triggers render

**`renderCollectibles()`**
- Generates HTML for each collectible
- Attaches event listeners
- Handles image fallbacks

**`openRentalModal(collectibleId)`**
- Finds collectible by ID
- Validates store selection
- Shows modal
- Calculates initial quote

**`calculateQuote()`**
- Reads duration input
- Calls `/api/rentals/quote`
- Updates price display
- Shows special rate warning if applicable

**`setupEventListeners()`**
- Store selection change
- Modal open/close
- Duration input
- Payment method selection

### checkout.js - Payment Processing

#### Key Functions

**`handleCheckout()`**
- Validates all form fields
- Collects customer data
- Calls `/api/rentals/checkout`
- Redirects to PayMongo

**`validateCheckoutForm()`**
- Checks duration (min 1 day)
- Validates payment method selected
- Validates customer details
- Email format validation

**`isValidEmail(email)`**
- Regex validation for email

**`fillDemoData()`** (dev only)
- Auto-fills form for testing
- Available in console: `window.fillDemoData()`

---

## User Flows

### 1. Browse Collectibles
1. Page loads
2. API fetches collectibles
3. Grid renders with images
4. User can scroll and view items

### 2. Get Quote
1. User selects store from dropdown
2. User clicks collectible card
3. Modal opens
4. User enters rental duration
5. Quote calculates in real-time
6. Special rate warning shows if <7 days

### 3. Checkout
1. User selects payment method
2. User fills billing details
3. User clicks "Proceed to Payment"
4. Form validates
5. API creates rental
6. User redirects to PayMongo
7. After payment, redirects to success/failed page

---

## API Integration

### Endpoints Used

**GET /api/collectibles**
```javascript
const response = await fetch('/api/collectibles');
const data = await response.json();
// Returns: { success: true, data: [...collectibles] }
```

**POST /api/rentals/quote**
```javascript
const response = await fetch('/api/rentals/quote', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    collectible_id: 'col-001',
    duration: 7
  })
});
// Returns: { success: true, data: { daily_rate, total_fee, is_special_rate } }
```

**POST /api/rentals/checkout**
```javascript
const response = await fetch('/api/rentals/checkout', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    collectible_id: 'col-001',
    store_id: 'store-a',
    duration: 7,
    payment_method: 'card',
    customer: { name, email, phone, address, city, postal_code }
  })
});
// Returns: { success: true, data: { rental_id, payment_url } }
```

---

## Responsive Design

### Breakpoints
- **Mobile**: < 768px
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

### Mobile Optimizations
```css
@media (max-width: 768px) {
  .hero-title { font-size: 2.5rem; }
  .collectibles-grid { grid-template-columns: 1fr; }
  .payment-methods { grid-template-columns: 1fr; }
}
```

---

## Performance Optimizations

### Image Handling
- Lazy loading with `onerror` fallback
- Placeholder for missing images
- Optimized file sizes

### CSS
- CSS variables for consistency
- Hardware-accelerated animations
- Minimal repaints

### JavaScript
- Event delegation where possible
- Debounced API calls
- Minimal DOM manipulation

---

## Accessibility

### Semantic HTML
- Proper heading hierarchy (h1 → h2 → h3)
- Form labels for all inputs
- Alt text for images

### Keyboard Navigation
- Modal closes with Escape key
- Tab order follows logical flow
- Focus states visible

### Color Contrast
- Text meets WCAG AA standards
- Interactive elements clearly visible

---

## Browser Compatibility
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

---

## Development Tips

### Testing Locally
1. Start server: `go run main.go`
2. Open: `http://localhost:8080`
3. Open DevTools (F12)
4. Check Console for errors

### Debugging
- Use `console.log()` for state inspection
- Network tab for API calls
- Elements tab for CSS debugging

### Common Issues
- **Images not loading**: Hard refresh (Ctrl+Shift+R)
- **API errors**: Check server is running
- **Modal not opening**: Check store is selected

---

## Future Enhancements
- [ ] Add loading spinners
- [ ] Implement search/filter
- [ ] Add favorites/wishlist
- [ ] Progressive Web App (PWA)
- [ ] Dark/light mode toggle
- [ ] Accessibility improvements
- [ ] Animation performance optimization
