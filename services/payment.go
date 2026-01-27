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
func (s *PaymentService) CreateCheckoutSession(amount float64, rentalID string, collectibleName string, duration int) (string, string, error) {
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
				PaymentMethodTypes: []string{"qrph", "gcash", "paymaya", "card", "grab_pay"},
				Description:        fmt.Sprintf("Rental for %s (%d days)", collectibleName, duration),
				SendEmailReceipt:   false,
				ShowDescription:    true,
				ShowLineItems:      true,
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
