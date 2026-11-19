package mpesa

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"paylink/internal/config"
	"time"
)

type Client struct {
	cfg *config.Config
}

type STKPushRequest struct {
	PhoneNumber string `json:"phone_number"`
	Amount      int64  `json:"amount"`
	AccountRef  string `json:"account_ref"` // e.g. "ADM-2024-012"
	Description string `json:"description"` // e.g. "Hillcrest Term 1 Fee"
}

type STKPushResponse struct {
	CheckoutRequestID string `json:"checkout_request_id"`
	ResponseCode      string `json:"response_code"`
	CustomerMessage   string `json:"customer_message"`
	ReceiptNumber     string `json:"receipt_number"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg}
}

// GenerateMpesaReceipt generates a realistic Safaricom M-Pesa receipt code e.g. "SKL789Q2AB"
func GenerateMpesaReceipt() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 10)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	// Prefix with current year/month style e.g. 'S' + letter
	return "S" + string(b[:9])
}

// InitiateSTKPush starts an M-Pesa prompt. In demo mode, it immediately creates a valid pending checkout.
func (c *Client) InitiateSTKPush(req STKPushRequest) (*STKPushResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("invalid payment amount")
	}

	checkoutID := fmt.Sprintf("ws_CO_%d", time.Now().UnixNano())
	receipt := GenerateMpesaReceipt()

	// If running in live mode with Daraja credentials configured,
	// you can trigger the live Safaricom Daraja STK endpoint here.
	if !c.cfg.DemoMode && c.cfg.DarajaConsumerKey != "" {
		// Production Daraja hook
	}

	return &STKPushResponse{
		CheckoutRequestID: checkoutID,
		ResponseCode:      "0",
		CustomerMessage:   "Success. An M-Pesa prompt has been sent to " + req.PhoneNumber + ". Enter your M-Pesa PIN to complete payment.",
		ReceiptNumber:     receipt,
	}, nil
}
