package shkeeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client for SHKeeper API
type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// InvoiceRequest for creating payment invoices
type InvoiceRequest struct {
	OrderID   string  `json:"order_id"`
	Amount    string  `json:"amount"`
	Currency  string  `json:"currency"`
	Callback  string  `json:"callback,omitempty"`
}

// InvoiceResponse from SHKeeper
type InvoiceResponse struct {
	OrderID     string `json:"order_id"`
	PaymentURL  string `json:"payment_url"`
	Address     string `json:"address"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PaymentStatus for checking payments
type PaymentStatus struct {
	OrderID      string    `json:"order_id"`
	Status       string    `json:"status"` // pending, confirmed, expired
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	Received     string    `json:"received"`
	ConfirmedAt  time.Time `json:"confirmed_at,omitempty"`
}

// Balance represents wallet balance
type Balance struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// New creates a new SHKeeper client
func New(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateInvoice creates a new payment invoice
func (c *Client) CreateInvoice(ctx context.Context, req InvoiceRequest) (*InvoiceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/invoice", c.BaseURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shkeeper returned status %d", resp.StatusCode)
	}

	var invoice InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// CheckPayment checks payment status
func (c *Client) CheckPayment(ctx context.Context, orderID string) (*PaymentStatus, error) {
	url := fmt.Sprintf("%s/api/v1/payment/%s", c.BaseURL, orderID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shkeeper returned status %d", resp.StatusCode)
	}

	var status PaymentStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// GetBalances returns all wallet balances
func (c *Client) GetBalances(ctx context.Context) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/v1/balances", c.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shkeeper returned status %d", resp.StatusCode)
	}

	var balances map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		return nil, err
	}

	return balances, nil
}

// SendPayment sends payment from SHKeeper wallet
func (c *Client) SendPayment(ctx context.Context, currency, toAddress, amount string) error {
	url := fmt.Sprintf("%s/api/v1/send", c.BaseURL)

	payload := map[string]string{
		"currency": currency,
		"to":       toAddress,
		"amount":   amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shkeeper returned status %d", resp.StatusCode)
	}

	return nil
}
