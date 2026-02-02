package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mongocollectibles/rental-system/models"
)

// PaymentService handles PayMongo API integration
type PaymentService struct {
	secretKey string
	publicKey string
	client    *http.Client
}

// NewPaymentService creates a new payment service
func NewPaymentService(secretKey, publicKey string) *PaymentService {
	return &PaymentService{
		secretKey: secretKey,
		publicKey: publicKey,
		client:    &http.Client{},
	}
}

// PayMongoSessionRequest represents the request to create a checkout session
type PayMongoSessionRequest struct {
	Data PayMongoSessionData `json:"data"`
}

type PayMongoSessionData struct {
	Attributes PayMongoSessionAttributes `json:"attributes"`
}

type PayMongoSessionAttributes struct {
	LineItems          []PayMongoLineItem `json:"line_items"`
	PaymentMethodTypes []string           `json:"payment_method_types"`
	Description        string             `json:"description"`
	SuccessUrl         string             `json:"success_url,omitempty"`
	CancelUrl          string             `json:"cancel_url,omitempty"`
	SendEmailReceipt   bool               `json:"send_email_receipt"`
	ShowDescription    bool               `json:"show_description"`
	ShowLineItems      bool               `json:"show_line_items"`
}

type PayMongoLineItem struct {
	Amount   int    `json:"amount"` // Amount in centavos
	Currency string `json:"currency"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// PayMongoSessionResponse represents the response from creating a checkout session
type PayMongoSessionResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			CheckoutURL string `json:"checkout_url"`
			Status      string `json:"status"`
		} `json:"attributes"`
	} `json:"data"`
}

// CreateCheckoutSession creates a checkout session via PayMongo API
func (s *PaymentService) CreateCheckoutSession(baseURL string, amount float64, rentalID string, collectibleName string, duration int) (string, string, error) {
	// Convert amount to centavos
	// Note: Total amount for the session.
	// The user's snippet showed amount 10000 and quantity 7.
	// We'll pass the total fee as one item with quantity 1 for simplicity,
	// or match the rental details.
	amountCentavos := int(amount * 100)

	requestData := PayMongoSessionRequest{
		Data: PayMongoSessionData{
			Attributes: PayMongoSessionAttributes{
				LineItems: []PayMongoLineItem{
					{
						Amount:   amountCentavos,
						Currency: "PHP",
						Name:     collectibleName,
						Quantity: 1,
					},
				},
				PaymentMethodTypes: []string{"qrph", "gcash", "paymaya", "card", "grab_pay", "dob", "dob_ubp"},
				Description:        fmt.Sprintf("Rental for %s (%d days)", collectibleName, duration),
				SendEmailReceipt:   true,
				ShowDescription:    true,
				ShowLineItems:      true,
				SuccessUrl:         fmt.Sprintf("%s/payment/success?rental_id=%s", baseURL, rentalID),
				CancelUrl:          fmt.Sprintf("%s/payment/failed?rental_id=%s", baseURL, rentalID),
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.paymongo.com/v1/checkout_sessions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// Use secret key from config
	authKey := s.secretKey

	// PayMongo requires Basic Auth with Secret Key as username and empty password
	encodedKey := base64.StdEncoding.EncodeToString([]byte(authKey + ":"))
	req.Header.Add("authorization", "Basic "+encodedKey)

	res, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("paymongo api error (%d): %s", res.StatusCode, string(body))
	}

	var sessionResponse PayMongoSessionResponse
	if err := json.Unmarshal(body, &sessionResponse); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	return sessionResponse.Data.ID, sessionResponse.Data.Attributes.CheckoutURL, nil
}

// PayMongoCustomerRequest represents request to create a customer
type PayMongoCustomerRequest struct {
	Data PayMongoCustomerData `json:"data"`
}

type PayMongoCustomerData struct {
	Attributes PayMongoCustomerAttributes `json:"attributes"`
}

type PayMongoCustomerAttributes struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// PayMongoCustomerResponse represents response from creating a customer
type PayMongoCustomerResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// CreateCustomer creates a customer in PayMongo
func (s *PaymentService) CreateCustomer(customer models.Customer) (string, error) {
	// Split name into first and last name (simple heuristic)
	// In a real app, you might want to store them separately in models.Customer
	var firstName, lastName string
	// Split by space
	// This is just a basic implementation
	// If name has no spaces, everything goes to firstName
	// If multiple spaces, last word is lastName, rest is firstName
	// ...
	// For simplicity, let's just pass the full name as first name if needed, or split.
	// PayMongo attributes: first_name, last_name.
	// Let's do a simple split.
	// actually standard PayMongo API might allow simple name?
	// Checking docs (simulation): create customer takes first_name, last_name, email, phone.

	// Simple split:
	// If Name is "John Doe" -> First: John, Last: Doe
	// If Name is "John" -> First: John, Last: .

	nameParts := SplitName(customer.Name)
	firstName = nameParts[0]
	if len(nameParts) > 1 {
		lastName = nameParts[1]
	}

	reqData := PayMongoCustomerRequest{
		Data: PayMongoCustomerData{
			Attributes: PayMongoCustomerAttributes{
				FirstName: firstName,
				LastName:  lastName,
				Email:     customer.Email,
				Phone:     customer.Phone,
			},
		},
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal customer request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.paymongo.com/v1/customers", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	authKey := s.secretKey
	encodedKey := base64.StdEncoding.EncodeToString([]byte(authKey + ":"))
	req.Header.Add("authorization", "Basic "+encodedKey)

	res, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("paymongo api error (%d): %s", res.StatusCode, string(body))
	}

	var custResponse PayMongoCustomerResponse
	if err := json.Unmarshal(body, &custResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return custResponse.Data.ID, nil
}

// SplitName helper
func SplitName(name string) []string {
	// This is a placeholder, implementation can be improved
	// For now, assuming simple space separation
	// Return [First, Last]
	// If no space, return [Name, ""]
	// Imports "strings" needed? Yes.
	// I will add strings import.

	// Actually, I can avoid helper and do logic inline or simple substring.
	// Let's keep it inline-ish with basic logic to avoid extra function/import for now if possible,
	// OR add "strings" to imports. "strings" is standard.
	// I'll add "strings" to imports in a separate Edit or use basic iteration.
	// Wait, I can't add imports with replace_file_content in the middle easily.
	// I will implement a loop or just assume First Name = Name for now to keep it simple,
	// OR just use a simple space check.

	var first, last string
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ' ' {
			first = name[:i]
			last = name[i+1:]
			return []string{first, last}
		}
	}
	return []string{name, ""}
}

// VerifyPayment verifies a payment status
func (s *PaymentService) VerifyPayment(sessionID string) (models.PaymentStatus, error) {
	req, err := http.NewRequest("GET", "https://api.paymongo.com/v1/checkout_sessions/"+sessionID, nil)
	if err != nil {
		return models.PaymentFailed, err
	}

	authKey := s.secretKey
	encodedKey := base64.StdEncoding.EncodeToString([]byte(authKey + ":"))
	req.Header.Add("authorization", "Basic "+encodedKey)

	res, err := s.client.Do(req)
	if err != nil {
		return models.PaymentFailed, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.PaymentFailed, err
	}

	var sessionResponse PayMongoSessionResponse
	if err := json.Unmarshal(body, &sessionResponse); err != nil {
		return models.PaymentFailed, err
	}

	if sessionResponse.Data.Attributes.Status == "paid" {
		return models.PaymentCompleted, nil
	}

	return models.PaymentPending, nil
}
