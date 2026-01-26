// Global state
let collectibles = [];
let stores = [];
let selectedCollectible = null;
let selectedStore = null;

// API Base URL
const API_BASE = '/api';

// Initialize app
document.addEventListener('DOMContentLoaded', async () => {
    await loadStores();
    await loadCollectibles();
    setupEventListeners();
});

// Load stores from config
async function loadStores() {
    try {
        // For now, hardcode stores (in production, fetch from API)
        stores = [
            { id: 'store-a', name: 'MongoCollectibles Store A - Manila' },
            { id: 'store-b', name: 'MongoCollectibles Store B - Quezon City' },
            { id: 'store-c', name: 'MongoCollectibles Store C - Makati' }
        ];

        const storeSelect = document.getElementById('storeSelect');
        storeSelect.innerHTML = '<option value="">Select a store...</option>';
        
        stores.forEach(store => {
            const option = document.createElement('option');
            option.value = store.id;
            option.textContent = store.name;
            storeSelect.appendChild(option);
        });

        // Set default store
        storeSelect.value = stores[0].id;
        selectedStore = stores[0].id;
    } catch (error) {
        console.error('Error loading stores:', error);
    }
}

// Load collectibles from API
async function loadCollectibles() {
    try {
        const response = await fetch(`${API_BASE}/collectibles`);
        const data = await response.json();
        
        if (data.success) {
            collectibles = data.data;
            renderCollectibles();
        }
    } catch (error) {
        console.error('Error loading collectibles:', error);
        document.getElementById('collectiblesGrid').innerHTML = 
            '<div class="loading">Failed to load collectibles. Please refresh the page.</div>';
    }
}

// Render collectibles grid
function renderCollectibles() {
    const grid = document.getElementById('collectiblesGrid');
    
    if (collectibles.length === 0) {
        grid.innerHTML = '<div class="loading">No collectibles available at this time.</div>';
        return;
    }

    grid.innerHTML = collectibles.map(collectible => `
        <div class="collectible-card" data-id="${collectible.id}">
            <img src="${collectible.image_url || '/images/placeholder.jpg'}" 
                 alt="${collectible.name}" 
                 class="collectible-image"
                 onerror="this.src='/images/placeholder.jpg'">
            <div class="collectible-content">
                <div class="collectible-header">
                    <h3 class="collectible-name">${collectible.name}</h3>
                    <span class="size-badge size-${collectible.size.toLowerCase()}">${collectible.size}</span>
                </div>
                <p class="collectible-description">${collectible.description}</p>
                <div class="collectible-footer">
                    <div class="price">
                        <span class="price-label">From</span>
                        ₱${formatPrice(getPriceForSize(collectible.size))}
                        <span class="price-label">/day</span>
                    </div>
                    <button class="btn btn-primary rent-btn" data-id="${collectible.id}">
                        Rent Now
                    </button>
                </div>
            </div>
        </div>
    `).join('');

    // Add click handlers to rent buttons
    document.querySelectorAll('.rent-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const id = btn.getAttribute('data-id');
            openRentalModal(id);
        });
    });

    // Add click handlers to cards
    document.querySelectorAll('.collectible-card').forEach(card => {
        card.addEventListener('click', () => {
            const id = card.getAttribute('data-id');
            openRentalModal(id);
        });
    });
}

// Get price for size
function getPriceForSize(size) {
    const prices = {
        'S': 1000,
        'M': 5000,
        'L': 10000
    };
    return prices[size] || 0;
}

// Format price
function formatPrice(price) {
    return price.toLocaleString('en-PH', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

// Open rental modal
function openRentalModal(collectibleId) {
    selectedCollectible = collectibles.find(c => c.id === collectibleId);
    
    if (!selectedCollectible) {
        alert('Collectible not found');
        return;
    }

    // Check if store is selected
    const storeSelect = document.getElementById('storeSelect');
    selectedStore = storeSelect.value;
    
    if (!selectedStore) {
        alert('Please select a pickup store first');
        storeSelect.focus();
        return;
    }

    // Update modal title
    document.getElementById('modalTitle').textContent = `Rent ${selectedCollectible.name}`;

    // Reset form
    document.getElementById('rentalDuration').value = 7;
    document.querySelectorAll('.payment-method').forEach(el => el.classList.remove('selected'));
    
    // Show modal
    document.getElementById('rentalModal').classList.add('active');

    // Calculate initial quote
    calculateQuote();
}

// Setup event listeners
function setupEventListeners() {
    // Store selection
    document.getElementById('storeSelect').addEventListener('change', (e) => {
        selectedStore = e.target.value;
    });

    // Modal close
    document.getElementById('closeModal').addEventListener('click', closeModal);
    document.getElementById('cancelBtn').addEventListener('click', closeModal);

    // Click outside modal to close
    document.getElementById('rentalModal').addEventListener('click', (e) => {
        if (e.target.id === 'rentalModal') {
            closeModal();
        }
    });

    // Duration change
    document.getElementById('rentalDuration').addEventListener('input', calculateQuote);

    // Payment method selection
    document.querySelectorAll('.payment-method').forEach(method => {
        method.addEventListener('click', () => {
            document.querySelectorAll('.payment-method').forEach(m => m.classList.remove('selected'));
            method.classList.add('selected');
        });
    });
}

// Close modal
function closeModal() {
    document.getElementById('rentalModal').classList.remove('active');
    selectedCollectible = null;
}

// Calculate quote
async function calculateQuote() {
    if (!selectedCollectible) return;

    const duration = parseInt(document.getElementById('rentalDuration').value) || 0;
    
    if (duration < 1) {
        document.getElementById('quoteSummary').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/rentals/quote`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                collectible_id: selectedCollectible.id,
                duration: duration
            })
        });

        const data = await response.json();
        
        if (data.success) {
            const quote = data.data;
            
            // Update quote display
            document.getElementById('quoteDaily').textContent = `₱${formatPrice(quote.daily_rate)}`;
            document.getElementById('quoteDuration').textContent = `${quote.duration} days`;
            document.getElementById('quoteTotal').textContent = `₱${formatPrice(quote.total_fee)}`;
            document.getElementById('quoteSummary').style.display = 'block';

            // Show special rate notice if applicable
            const specialNotice = document.getElementById('specialRateNotice');
            if (quote.is_special_rate) {
                specialNotice.style.display = 'block';
            } else {
                specialNotice.style.display = 'none';
            }
        }
    } catch (error) {
        console.error('Error calculating quote:', error);
    }
}

// Check URL for success/failure
window.addEventListener('load', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const rentalId = urlParams.get('rental_id');
    
    if (window.location.pathname === '/success.html' && rentalId) {
        document.body.innerHTML = `
            <div class="container" style="text-align: center; padding: 4rem 2rem;">
                <h1 style="font-size: 3rem; margin-bottom: 1rem;">✅ Payment Successful!</h1>
                <p style="font-size: 1.2rem; color: var(--text-secondary); margin-bottom: 2rem;">
                    Your rental (ID: ${rentalId}) has been confirmed. You'll receive a confirmation email shortly.
                </p>
                <button class="btn btn-primary" onclick="window.location.href='/'">Browse More Collectibles</button>
            </div>
        `;
    } else if (window.location.pathname === '/failed.html' && rentalId) {
        document.body.innerHTML = `
            <div class="container" style="text-align: center; padding: 4rem 2rem;">
                <h1 style="font-size: 3rem; margin-bottom: 1rem;">❌ Payment Failed</h1>
                <p style="font-size: 1.2rem; color: var(--text-secondary); margin-bottom: 2rem;">
                    Unfortunately, your payment could not be processed. Please try again.
                </p>
                <button class="btn btn-primary" onclick="window.location.href='/'">Return to Home</button>
            </div>
        `;
    }
});
