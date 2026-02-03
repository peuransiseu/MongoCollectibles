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

	amountCentavos := int(amount * 100)

	requestData := PayMongoSessionRequest{
		Data: PayMongoSessionData{
			Attributes: PayMongoSessionAttributes{
				LineItems: []PayMongoLineItem{
					{
						Amount:   amountCentavos,
						Currency: "PHP",
						Name:     fmt.Sprintf("%s (%d Days Rental)", collectibleName, duration),
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

	var firstName, lastName string

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
