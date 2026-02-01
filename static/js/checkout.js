// Checkout functionality

// Handle checkout button click (only if it exists)
const proceedBtn = document.getElementById('proceedToPaymentBtn');
if (proceedBtn) {
    proceedBtn.addEventListener('click', handleCheckout);
}

async function handleCheckout() {
    console.log("Handle Checkout Started - v2");

    // 1. Validate form fields first
    if (!validateCheckoutForm()) {
        console.log("Validation failed");
        return;
    }
    console.log("Validation passed");

    // 2. Collect only what we need
    const durationInput = document.getElementById('rentalDuration');
    const duration = parseInt(durationInput.value);

    const customer = {
        name: document.getElementById('customerName').value.trim(),
        email: document.getElementById('customerEmail').value.trim(),
        phone: document.getElementById('customerPhone').value.trim(),
        address: document.getElementById('customerAddress').value.trim(),
        city: document.getElementById('customerCity').value.trim(),
        postal_code: document.getElementById('customerPostal').value.trim()
    };

    // 3. Prepare data - we MUST include payment_method for the API
    const checkoutData = {
        collectible_id: selectedCollectible.id,
        store_id: selectedStore,
        duration: duration,
        payment_method: "external",
        customer: customer
    };

    // 4. Update UI to show processing
    const proceedBtn = document.getElementById('proceedToPaymentBtn');
    const originalText = proceedBtn.textContent;
    proceedBtn.disabled = true;
    proceedBtn.textContent = 'Preparing Payment...';

    try {
        const response = await fetch(`${API_BASE}/rentals/checkout`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(checkoutData)
        });

        const data = await response.json();

        if (data.success && data.data.payment_url) {
            // Successfully got the PayMongo session URL
            window.location.href = data.data.payment_url;
        } else {
            console.error('API Error:', data);
            alert(`Checkout Error: ${data.error || 'The server encountered an issue. Please try again.'}`);
            proceedBtn.disabled = false;
            proceedBtn.textContent = originalText;
        }
    } catch (error) {
        console.error('Network/System Error:', error);
        alert('Could not connect to the payment server. Please check your connection.');
        proceedBtn.disabled = false;
        proceedBtn.textContent = originalText;
    }
}

// Validate checkout form
function validateCheckoutForm() {
    // Check duration
    const durationInput = document.getElementById('rentalDuration');
    const duration = parseInt(durationInput.value);
    if (!duration || duration < 1) {
        alert('Please enter a valid rental duration (minimum 1 day)');
        durationInput.focus();
        return false;
    }

    // Check customer details
    const nameInput = document.getElementById('customerName');
    const name = nameInput.value.trim();
    if (!name) {
        alert('Please enter your full name');
        nameInput.focus();
        return false;
    }

    const emailInput = document.getElementById('customerEmail');
    const email = emailInput.value.trim();
    if (!email || !isValidEmail(email)) {
        alert('Please enter a valid email address');
        emailInput.focus();
        return false;
    }

    const phoneInput = document.getElementById('customerPhone');
    const phone = phoneInput.value.trim();
    if (!phone) {
        alert('Please enter your phone number');
        phoneInput.focus();
        return false;
    }

    const addressInput = document.getElementById('customerAddress');
    const address = addressInput.value.trim();
    if (!address) {
        alert('Please enter your address');
        addressInput.focus();
        return false;
    }

    const cityInput = document.getElementById('customerCity');
    const city = cityInput.value.trim();
    if (!city) {
        alert('Please enter your city');
        cityInput.focus();
        return false;
    }

    const postalInput = document.getElementById('customerPostal');
    const postal = postalInput.value.trim();
    if (!postal) {
        alert('Please enter your postal code');
        postalInput.focus();
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
