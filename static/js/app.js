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

// UI Helper: Show Notification
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 2rem;
        right: 2rem;
        padding: 1rem 1.5rem;
        background: ${type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6'};
        color: white;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        z-index: 10000;
        animation: slideIn 0.3s ease;
    `;
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Load stores from config
async function loadStores() {
    try {
        stores = [
            { id: 'store-a', name: 'MongoCollectibles Store A - Manila' },
            { id: 'store-b', name: 'MongoCollectibles Store B - Quezon City' },
            { id: 'store-c', name: 'MongoCollectibles Store C - Makati' }
        ];

        const storeOptions = document.getElementById('storeOptions');

        storeOptions.innerHTML = '';

        stores.forEach(store => {
            const option = document.createElement('div');
            option.className = 'select-option';
            option.dataset.value = store.id;
            option.innerHTML = `
                <span class="select-icon">📍</span>
                <span>${store.name}</span>
            `;

            option.addEventListener('click', (e) => {
                e.stopPropagation();
                selectStore(store.id);
            });

            storeOptions.appendChild(option);
        });

        if (stores.length > 0) {
            selectStore(stores[0].id);
        }
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
            displayCollectibles();
        }
    } catch (error) {
        console.error('Error loading collectibles:', error);
    }
}

// Display collectibles
function displayCollectibles() {
    const grid = document.getElementById('collectiblesGrid');
    grid.innerHTML = '';

    collectibles.forEach(collectible => {
        const card = createCollectibleCard(collectible);
        grid.appendChild(card);
    });
}

// Create collectible card
function createCollectibleCard(collectible) {
    const card = document.createElement('div');
    card.className = 'collectible-card';
    card.innerHTML = `
        <div class="collectible-image" style="background-image: url('${collectible.image_url}'); background-size: cover; background-position: center;">
            <div class="size-badge">${collectible.size}</div>
            <div class="stock-badge ${collectible.stock > 0 ? 'in-stock' : 'out-of-stock'}">
                ${collectible.stock > 0 ? `${collectible.stock} in stock` : 'Out of stock'}
            </div>
        </div>
        <div class="collectible-info">
            <h3 class="collectible-name">${collectible.name}</h3>
            <div class="collectible-meta">
                <span class="meta-item">📦 ${collectible.size}</span>
                <span class="meta-item">⏱️ ${collectible.eta_days} days</span>
            </div>
            <div class="collectible-footer">
                <div class="price">₱${collectible.daily_rate}/day</div>
                <button class="btn btn-primary btn-sm rental-btn" 
                    data-id="${collectible.id}" 
                    ${collectible.stock === 0 ? 'disabled' : ''}>
                    Rent Now
                </button>
            </div>
        </div>
    `;

    // Add event listener directly to the button
    const btn = card.querySelector('.rental-btn');
    if (btn && !btn.disabled) {
        btn.addEventListener('click', () => openRentalModal(collectible.id));
    }

    return card;
}

// Select store
function selectStore(storeId) {
    selectedStore = storeId;
    const store = stores.find(s => s.id === storeId);
    const selectedText = document.querySelector('.selected-text');
    const options = document.querySelectorAll('.select-option');

    if (store) {
        selectedText.textContent = store.name;
    }

    options.forEach(opt => {
        opt.classList.remove('active');
        if (opt.dataset.value === storeId) {
            opt.classList.add('active');
        }
    });

    const storeSelect = document.getElementById('storeSelect');
    storeSelect.classList.remove('active');

    loadCollectibles();
}

// Open Rental Modal
function openRentalModal(collectibleId) {
    if (!selectedStore) {
        showNotification('Please select a store first', 'error');
        return;
    }

    const collectible = collectibles.find(c => c.id === collectibleId);
    if (!collectible) return;

    selectedCollectible = collectible;

    // Populate modal
    document.getElementById('modalTitle').textContent = `Rent ${collectible.name}`;

    // Reset inputs
    document.getElementById('rentalDuration').value = 7;
    if (window.fillDemoData) window.fillDemoData(); // Auto-fill for ease

    // Reset Quote UI
    updateQuote(collectible.daily_rate, 7);

    // Show modal
    document.getElementById('rentalModal').style.display = 'flex';
}

// Update Quote Calculation
function updateQuote(dailyRate, duration) {
    let finalRate = dailyRate;
    let isSpecial = false;

    if (duration < 7) {
        finalRate *= 2;
        isSpecial = true;
    }

    const total = finalRate * duration;

    // DOM Elements
    const quoteDaily = document.getElementById('quoteDaily');
    const quoteDuration = document.getElementById('quoteDuration');
    const quoteTotal = document.getElementById('quoteTotal');
    const quoteETA = document.getElementById('quoteETA');
    const specialNotice = document.getElementById('specialRateNotice');
    const quoteSummary = document.getElementById('quoteSummary');

    if (quoteDaily) quoteDaily.textContent = `₱${finalRate.toFixed(2)}`;
    if (quoteDuration) quoteDuration.textContent = `${duration} days`;
    if (quoteTotal) quoteTotal.textContent = `₱${total.toFixed(2)}`;

    // Estimate ETA (Simplistic: +1 day from now, or use collectible data if available)
    // In a real app, we might call the API for precise ETA based on store selection
    const etaDays = selectedCollectible ? selectedCollectible.eta_days : 1;
    const etaDate = new Date();
    etaDate.setDate(etaDate.getDate() + etaDays);
    if (quoteETA) quoteETA.textContent = etaDate.toDateString();

    if (specialNotice) {
        specialNotice.style.display = isSpecial ? 'block' : 'none';
    }

    if (quoteSummary) quoteSummary.style.display = 'block';
}

// Setup event listeners
function setupEventListeners() {
    // Store dropdown
    const trigger = document.getElementById('storeSelectTrigger');
    const storeSelect = document.getElementById('storeSelect');

    trigger.addEventListener('click', (e) => {
        e.stopPropagation();
        storeSelect.classList.toggle('active');
    });

    document.addEventListener('click', () => {
        storeSelect.classList.remove('active');
    });

    // Close Modal
    const closeModal = document.getElementById('closeModal');
    if (closeModal) {
        closeModal.addEventListener('click', () => {
            document.getElementById('rentalModal').style.display = 'none';
        });
    }

    // Close modal on outside click
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('rentalModal');
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });

    // Duration Change -> Update Quote
    const durationInput = document.getElementById('rentalDuration');
    if (durationInput) {
        durationInput.addEventListener('input', (e) => {
            const duration = parseInt(e.target.value) || 0;
            if (selectedCollectible && duration > 0) {
                updateQuote(selectedCollectible.daily_rate, duration);
            }
        });
    }

    // Cancel Button in Modal
    const cancelBtn = document.getElementById('cancelBtn');
    if (cancelBtn) {
        cancelBtn.addEventListener('click', () => {
            document.getElementById('rentalModal').style.display = 'none';
        });
    }
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
    
    .modal {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }
    
    .modal-content {
        background: var(--card-bg);
        padding: 2rem;
        border-radius: 12px;
        max-width: 500px;
        width: 90%;
        max-height: 90vh;
        overflow-y: auto;
    }
    
    .btn-outline {
        background: transparent;
        border: 2px solid var(--primary);
        color: var(--primary);
    }
    
    .btn-outline:hover {
        background: var(--primary);
        color: white;
    }
`;
document.head.appendChild(style);
