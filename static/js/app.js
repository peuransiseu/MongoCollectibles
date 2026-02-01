// Global state
let collectibles = [];
let stores = [];
let selectedCollectible = null;
let selectedStore = null;
let authToken = localStorage.getItem('authToken') || null;
let currentUser = null;
let cart = null;

// API Base URL
const API_BASE = '/api';

// Initialize app
document.addEventListener('DOMContentLoaded', async () => {
    await loadStores();
    await loadCollectibles();
    setupEventListeners();
    checkAuthStatus();

    // Load cart if logged in
    if (authToken) {
        await loadCart();
    }
});

// Check auth status and update UI
function checkAuthStatus() {
    const authButtons = document.getElementById('authButtons');
    const userSection = document.getElementById('userSection');

    if (authToken) {
        authButtons.style.display = 'none';
        userSection.style.display = 'flex';
    } else {
        authButtons.style.display = 'flex';
        userSection.style.display = 'none';
    }
}

// Auth Functions
async function register(email, password) {
    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (data.success) {
            authToken = data.data.token;
            currentUser = data.data;
            localStorage.setItem('authToken', authToken);
            checkAuthStatus();
            await loadCart();
            showNotification('Registration successful!', 'success');
            return true;
        } else {
            showNotification(data.error || 'Registration failed', 'error');
            return false;
        }
    } catch (error) {
        console.error('Registration error:', error);
        showNotification('Registration failed', 'error');
        return false;
    }
}

async function login(email, password) {
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (data.success) {
            authToken = data.data.token;
            currentUser = data.data;
            localStorage.setItem('authToken', authToken);
            checkAuthStatus();
            await loadCart();
            showNotification('Login successful!', 'success');
            return true;
        } else {
            showNotification(data.error || 'Login failed', 'error');
            return false;
        }
    } catch (error) {
        console.error('Login error:', error);
        showNotification('Login failed', 'error');
        return false;
    }
}

async function logout() {
    try {
        if (authToken) {
            await fetch(`${API_BASE}/auth/logout`, {
                method: 'POST',
                headers: { 'Authorization': authToken }
            });
        }

        authToken = null;
        currentUser = null;
        cart = null;
        localStorage.removeItem('authToken');
        checkAuthStatus();
        updateCartCount(0);
        showNotification('Logged out successfully', 'success');
    } catch (error) {
        console.error('Logout error:', error);
    }
}

// Cart Functions
async function loadCart() {
    if (!authToken) return;

    try {
        const response = await fetch(`${API_BASE}/cart`, {
            headers: { 'Authorization': authToken }
        });

        const data = await response.json();

        if (data.success) {
            cart = data.data;
            updateCartCount(cart.items ? cart.items.length : 0);
        }
    } catch (error) {
        console.error('Load cart error:', error);
    }
}

async function addToCart(collectibleId, rentalDays = 7, quantity = 1) {
    if (!authToken) {
        showNotification('Please login to add items to cart', 'error');
        document.getElementById('loginModal').style.display = 'flex';
        return false;
    }

    if (!selectedStore) {
        showNotification('Please select a store first', 'error');
        return false;
    }

    try {
        const response = await fetch(`${API_BASE}/cart/items`, {
            method: 'POST',
            headers: {
                'Authorization': authToken,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                collectible_id: collectibleId,
                store_id: selectedStore,
                rental_days: rentalDays,
                quantity: quantity
            })
        });

        const data = await response.json();

        if (data.success) {
            cart = data.data;
            updateCartCount(cart.items.length);
            showNotification('Added to cart!', 'success');
            return true;
        } else {
            showNotification(data.error || 'Failed to add to cart', 'error');
            return false;
        }
    } catch (error) {
        console.error('Add to cart error:', error);
        showNotification('Failed to add to cart', 'error');
        return false;
    }
}

async function removeFromCart(collectibleId) {
    if (!authToken) return;

    try {
        const response = await fetch(`${API_BASE}/cart/items/${collectibleId}`, {
            method: 'DELETE',
            headers: { 'Authorization': authToken }
        });

        const data = await response.json();

        if (data.success) {
            cart = data.data;
            updateCartCount(cart.items.length);
            showNotification('Removed from cart', 'success');
            displayCart();
        }
    } catch (error) {
        console.error('Remove from cart error:', error);
    }
}

async function checkout() {
    if (!authToken) {
        showNotification('Please login to checkout', 'error');
        return;
    }

    if (!cart || cart.items.length === 0) {
        showNotification('Cart is empty', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/checkout`, {
            method: 'POST',
            headers: { 'Authorization': authToken }
        });

        const data = await response.json();

        if (data.success) {
            showNotification('Order created! Redirecting to payment...', 'success');
            // Redirect to payment URL
            if (data.data.payment_url) {
                window.location.href = data.data.payment_url;
            }
        } else {
            showNotification(data.error || 'Checkout failed', 'error');
        }
    } catch (error) {
        console.error('Checkout error:', error);
        showNotification('Checkout failed', 'error');
    }
}

// Orders Functions
async function loadOrders() {
    if (!authToken) return [];

    try {
        const response = await fetch(`${API_BASE}/orders`, {
            headers: { 'Authorization': authToken }
        });

        const data = await response.json();

        if (data.success) {
            return data.data || [];
        }
        return [];
    } catch (error) {
        console.error('Load orders error:', error);
        return [];
    }
}

async function cancelOrder(orderId) {
    if (!authToken) return;

    if (!confirm('Are you sure you want to cancel this order?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/orders/${orderId}/cancel`, {
            method: 'POST',
            headers: { 'Authorization': authToken }
        });

        const data = await response.json();

        if (data.success) {
            showNotification(`Order cancelled. Refund: ₱${data.data.refund_amount.toFixed(2)}`, 'success');
            displayOrders();
        } else {
            showNotification(data.error || 'Cancellation failed', 'error');
        }
    } catch (error) {
        console.error('Cancel order error:', error);
        showNotification('Cancellation failed', 'error');
    }
}

// UI Functions
function updateCartCount(count) {
    const cartCount = document.getElementById('cartCount');
    if (cartCount) {
        cartCount.textContent = count;
    }
}

function displayCart() {
    const cartItems = document.getElementById('cartItems');
    const cartTotal = document.getElementById('cartTotal');

    if (!cart || !cart.items || cart.items.length === 0) {
        cartItems.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: 2rem;">Your cart is empty</p>';
        cartTotal.innerHTML = '';
        return;
    }

    let html = '<div style="display: flex; flex-direction: column; gap: 1rem;">';
    let total = 0;

    cart.items.forEach(item => {
        const collectible = collectibles.find(c => c.id === item.collectible_id);

        let dailyRate = collectible ? collectible.daily_rate : 0;
        let isSpecialRate = false;

        // Apply special rate (double) for rentals < 7 days
        if (item.rental_days < 7) {
            dailyRate *= 2;
            isSpecialRate = true;
        }

        const itemTotal = dailyRate * item.rental_days * item.quantity;
        total += itemTotal;

        html += `
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 1rem; background: var(--card-bg); border-radius: 8px;">
                <div>
                    <h3 style="margin: 0 0 0.5rem 0;">${collectible ? collectible.name : item.collectible_id}</h3>
                    <p style="margin: 0; color: var(--text-secondary);">
                        ${item.rental_days} days @ ₱${dailyRate.toFixed(2)}/day ${isSpecialRate ? '(Special Rate)' : ''}
                    </p>
                    <p style="margin: 0; color: var(--text-secondary); font-weight: bold;">
                        Total: ₱${itemTotal.toFixed(2)}
                    </p>
                </div>
                <button class="btn btn-outline" onclick="removeFromCart('${item.collectible_id}')" style="padding: 0.5rem 1rem;">Remove</button>
            </div>
        `;
    });

    html += '</div>';
    cartItems.innerHTML = html;

    cartTotal.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: center;">
            <h3>Total:</h3>
            <h2 style="color: var(--primary);">₱${total.toFixed(2)}</h2>
        </div>
    `;
}

async function displayOrders() {
    const ordersList = document.getElementById('ordersList');
    const orders = await loadOrders();

    if (orders.length === 0) {
        ordersList.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: 2rem;">No orders yet</p>';
        return;
    }

    let html = '<div style="display: flex; flex-direction: column; gap: 1rem;">';

    orders.forEach(order => {
        const statusColors = {
            'PENDING_PAYMENT': '#f59e0b',
            'PAID': '#10b981',
            'ALLOCATED': '#3b82f6',
            'IN_TRANSIT': '#8b5cf6',
            'READY_FOR_PICKUP': '#06b6d4',
            'COMPLETED': '#22c55e',
            'CANCELLED': '#ef4444',
            'REFUNDED': '#6b7280'
        };

        const canCancel = ['PENDING_PAYMENT', 'PAID', 'ALLOCATED', 'IN_TRANSIT'].includes(order.status);

        html += `
            <div style="padding: 1.5rem; background: var(--card-bg); border-radius: 8px; border-left: 4px solid ${statusColors[order.status] || '#6b7280'};">
                <div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 1rem;">
                    <div>
                        <h3 style="margin: 0 0 0.5rem 0;">Order #${order.id.substring(0, 8)}</h3>
                        <p style="margin: 0; color: var(--text-secondary); font-size: 0.9rem;">
                            ${new Date(order.created_at).toLocaleDateString()}
                        </p>
                    </div>
                    <span style="padding: 0.25rem 0.75rem; background: ${statusColors[order.status]}; color: white; border-radius: 4px; font-size: 0.85rem; font-weight: 600;">
                        ${order.status.replace(/_/g, ' ')}
                    </span>
                </div>
                <div style="margin-bottom: 1rem;">
                    ${order.items.map(item => `
                        <p style="margin: 0.25rem 0; color: var(--text-secondary);">
                            ${item.collectible_name} - ${item.rental_days} days - ₱${item.price.toFixed(2)}
                        </p>
                    `).join('')}
                </div>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <h3 style="margin: 0;">Total: ₱${order.total_amount.toFixed(2)}</h3>
                    ${canCancel ? `<button class="btn btn-outline" onclick="cancelOrder('${order.id}')" style="padding: 0.5rem 1rem;">Cancel Order</button>` : ''}
                </div>
            </div>
        `;
    });

    html += '</div>';
    ordersList.innerHTML = html;
}

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
        const selectedText = document.querySelector('.selected-text');

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
                <button class="btn btn-primary btn-sm" onclick="addToCart('${collectible.id}')" ${collectible.stock === 0 ? 'disabled' : ''}>
                    Add to Cart
                </button>
            </div>
        </div>
    `;

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

    // Auth buttons
    document.getElementById('loginBtn')?.addEventListener('click', () => {
        document.getElementById('loginModal').style.display = 'flex';
    });

    document.getElementById('registerBtn')?.addEventListener('click', () => {
        document.getElementById('registerModal').style.display = 'flex';
    });

    document.getElementById('logoutBtn')?.addEventListener('click', logout);

    // Modal close buttons
    document.getElementById('closeLoginModal')?.addEventListener('click', () => {
        document.getElementById('loginModal').style.display = 'none';
    });

    document.getElementById('closeRegisterModal')?.addEventListener('click', () => {
        document.getElementById('registerModal').style.display = 'none';
    });

    document.getElementById('closeCartModal')?.addEventListener('click', () => {
        document.getElementById('cartModal').style.display = 'none';
    });

    document.getElementById('closeOrdersModal')?.addEventListener('click', () => {
        document.getElementById('ordersModal').style.display = 'none';
    });

    // Auth forms
    document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;

        if (await login(email, password)) {
            document.getElementById('loginModal').style.display = 'none';
            document.getElementById('loginForm').reset();
        }
    });

    document.getElementById('registerForm')?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('registerEmail').value;
        const password = document.getElementById('registerPassword').value;
        const confirmPassword = document.getElementById('registerPasswordConfirm').value;

        if (password !== confirmPassword) {
            showNotification('Passwords do not match', 'error');
            return;
        }

        if (await register(email, password)) {
            document.getElementById('registerModal').style.display = 'none';
            document.getElementById('registerForm').reset();
        }
    });

    // Cart button
    document.getElementById('viewCartBtn')?.addEventListener('click', () => {
        displayCart();
        document.getElementById('cartModal').style.display = 'flex';
    });

    // Orders button
    document.getElementById('viewOrdersBtn')?.addEventListener('click', () => {
        displayOrders();
        document.getElementById('ordersModal').style.display = 'flex';
    });

    // Checkout button
    document.getElementById('checkoutBtn')?.addEventListener('click', checkout);
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
