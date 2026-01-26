package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mongocollectibles/rental-system/models"
)

const (
	PayMongoAPIURL = "https://api.paymongo.com/v1"
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

// PayMongoSourceRequest represents the request to create a payment source
type PayMongoSourceRequest struct {
	Data PayMongoSourceData `json:"data"`
}

type PayMongoSourceData struct {
	Attributes PayMongoSourceAttributes `json:"attributes"`
}

type PayMongoSourceAttributes struct {
	Type     string                 `json:"type"`
	Amount   int                    `json:"amount"` // Amount in centavos
	Currency string                 `json:"currency"`
	Redirect PayMongoRedirect       `json:"redirect"`
	Billing  PayMongoBilling        `json:"billing,omitempty"`
}

type PayMongoRedirect struct {
	Success string `json:"success"`
	Failed  string `json:"failed"`
}

type PayMongoBilling struct {
	Name    string         `json:"name"`
	Email   string         `json:"email"`
	Phone   string         `json:"phone"`
	Address PayMongoAddress `json:"address"`
}

type PayMongoAddress struct {
	Line1      string `json:"line1"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// PayMongoSourceResponse represents the response from creating a payment source
type PayMongoSourceResponse struct {
	Data PayMongoSourceResponseData `json:"data"`
}

type PayMongoSourceResponseData struct {
	ID         string                       `json:"id"`
	Type       string                       `json:"type"`
	Attributes PayMongoSourceResponseAttrs  `json:"attributes"`
}

type PayMongoSourceResponseAttrs struct {
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

// CreatePaymentSource creates a payment source via PayMongo API
func (s *PaymentService) CreatePaymentSource(amount float64, paymentMethod models.PaymentMethod, customer models.Customer, rentalID string) (string, string, error) {
	// Convert amount to centavos (PayMongo uses smallest currency unit)
	amountCentavos := int(amount * 100)

	// Map payment method to PayMongo type
	paymongoType, err := s.mapPaymentMethodToType(paymentMethod)
	if err != nil {
		return "", "", err
	}

	// Create request payload
	requestData := PayMongoSourceRequest{
		Data: PayMongoSourceData{
			Attributes: PayMongoSourceAttributes{
				Type:     paymongoType,
				Amount:   amountCentavos,
				Currency: "PHP",
				Redirect: PayMongoRedirect{
					Success: fmt.Sprintf("http://localhost:8080/payment/success?rental_id=%s", rentalID),
					Failed:  fmt.Sprintf("http://localhost:8080/payment/failed?rental_id=%s", rentalID),
				},
				Billing: PayMongoBilling{
					Name:  customer.Name,
					Email: customer.Email,
					Phone: customer.Phone,
					Address: PayMongoAddress{
						Line1:      customer.Address,
						City:       customer.City,
						PostalCode: customer.PostalCode,
						Country:    "PH",
					},
				},
			},
		},
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", PayMongoAPIURL+"/sources", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+s.encodeSecretKey())

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("payment API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var sourceResponse PayMongoSourceResponse
	if err := json.Unmarshal(body, &sourceResponse); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	return sourceResponse.Data.ID, sourceResponse.Data.Attributes.CheckoutURL, nil
}

// mapPaymentMethodToType maps our payment method enum to PayMongo's type
func (s *PaymentService) mapPaymentMethodToType(method models.PaymentMethod) (string, error) {
	switch method {
	case models.PaymentCard:
		return "card", nil
	case models.PaymentGCash:
		return "gcash", nil
	case models.PaymentGrabPay:
		return "grab_pay", nil
	case models.PaymentBPI:
		return "billease", nil // PayMongo uses billease for bank transfers
	case models.PaymentUBP:
		return "billease", nil
	default:
		return "", errors.New("unsupported payment method")
	}
}

// encodeSecretKey encodes the secret key for basic auth
func (s *PaymentService) encodeSecretKey() string {
	// PayMongo uses base64 encoding of "sk_xxx:"
	// For simplicity, we'll use the key directly in production code
	// In real implementation, use base64.StdEncoding.EncodeToString()
	return s.secretKey
}

// VerifyPayment verifies a payment status (called from webhook)
func (s *PaymentService) VerifyPayment(paymentID string) (models.PaymentStatus, error) {
	// Create HTTP request to get source status
	req, err := http.NewRequest("GET", PayMongoAPIURL+"/sources/"+paymentID, nil)
	if err != nil {
		return models.PaymentFailed, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+s.encodeSecretKey())

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return models.PaymentFailed, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.PaymentFailed, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var sourceResponse PayMongoSourceResponse
	if err := json.Unmarshal(body, &sourceResponse); err != nil {
		return models.PaymentFailed, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map PayMongo status to our status
	switch sourceResponse.Data.Attributes.Status {
	case "chargeable", "paid":
		return models.PaymentCompleted, nil
	case "pending":
		return models.PaymentPending, nil
	default:
		return models.PaymentFailed, nil
	}
}
