// Checkout functionality

// Handle checkout button click
document.getElementById('checkoutBtn').addEventListener('click', handleCheckout);

async function handleCheckout() {
    // Validate form
    if (!validateCheckoutForm()) {
        return;
    }

    // Get form data
    const duration = parseInt(document.getElementById('rentalDuration').value);
    const paymentMethod = document.querySelector('.payment-method.selected')?.getAttribute('data-method');

    const customer = {
        name: document.getElementById('customerName').value.trim(),
        email: document.getElementById('customerEmail').value.trim(),
        phone: document.getElementById('customerPhone').value.trim(),
        address: document.getElementById('customerAddress').value.trim(),
        city: document.getElementById('customerCity').value.trim(),
        postal_code: document.getElementById('customerPostal').value.trim()
    };

    // Prepare checkout request
    const checkoutData = {
        collectible_id: selectedCollectible.id,
        store_id: selectedStore,
        duration: duration,
        payment_method: paymentMethod,
        customer: customer
    };

    // Disable button and show loading
    const checkoutBtn = document.getElementById('checkoutBtn');
    const originalText = checkoutBtn.textContent;
    checkoutBtn.disabled = true;
    checkoutBtn.textContent = 'Processing...';

    try {
        const response = await fetch(`${API_BASE}/rentals/checkout`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(checkoutData)
        });

        const data = await response.json();

        if (data.success) {
            // Redirect to PayMongo payment page
            window.location.href = data.data.payment_url;
        } else {
            alert(`Checkout failed: ${data.error || 'Unknown error'}`);
            checkoutBtn.disabled = false;
            checkoutBtn.textContent = originalText;
        }
    } catch (error) {
        console.error('Checkout error:', error);
        alert('An error occurred during checkout. Please try again.');
        checkoutBtn.disabled = false;
        checkoutBtn.textContent = originalText;
    }
}

// Validate checkout form
function validateCheckoutForm() {
    // Check duration
    const duration = parseInt(document.getElementById('rentalDuration').value);
    if (!duration || duration < 1) {
        alert('Please enter a valid rental duration (minimum 1 day)');
        document.getElementById('rentalDuration').focus();
        return false;
    }

    // Check payment method
    const paymentMethod = document.querySelector('.payment-method.selected');
    if (!paymentMethod) {
        alert('Please select a payment method');
        return false;
    }

    // Check customer details
    const name = document.getElementById('customerName').value.trim();
    if (!name) {
        alert('Please enter your full name');
        document.getElementById('customerName').focus();
        return false;
    }

    const email = document.getElementById('customerEmail').value.trim();
    if (!email || !isValidEmail(email)) {
        alert('Please enter a valid email address');
        document.getElementById('customerEmail').focus();
        return false;
    }

    const phone = document.getElementById('customerPhone').value.trim();
    if (!phone) {
        alert('Please enter your phone number');
        document.getElementById('customerPhone').focus();
        return false;
    }

    const address = document.getElementById('customerAddress').value.trim();
    if (!address) {
        alert('Please enter your address');
        document.getElementById('customerAddress').focus();
        return false;
    }

    const city = document.getElementById('customerCity').value.trim();
    if (!city) {
        alert('Please enter your city');
        document.getElementById('customerCity').focus();
        return false;
    }

    const postal = document.getElementById('customerPostal').value.trim();
    if (!postal) {
        alert('Please enter your postal code');
        document.getElementById('customerPostal').focus();
        return false;
    }

    return true;
}

// Email validation
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

// Auto-fill demo data (for testing)
function fillDemoData() {
    document.getElementById('customerName').value = 'Juan Dela Cruz';
    document.getElementById('customerEmail').value = 'juan@example.com';
    document.getElementById('customerPhone').value = '+63 912 345 6789';
    document.getElementById('customerAddress').value = '123 Main Street, Barangay San Juan';
    document.getElementById('customerCity').value = 'Manila';
    document.getElementById('customerPostal').value = '1000';
}

// Expose for debugging
if (window.location.hostname === 'localhost') {
    window.fillDemoData = fillDemoData;
}
