package config

import (
	"os"
	"paylink/internal/models"
	"strconv"
)

// Config holds runtime server and service settings
type Config struct {
	Port     string
	BaseURL  string
	DemoMode bool
	School   models.School

	// Daraja M-Pesa Settings (Optional, used if DemoMode is false)
	DarajaConsumerKey    string
	DarajaConsumerSecret string
	DarajaPasskey        string
	DarajaShortcode      string
	DarajaCallbackURL    string

	// Africa's Talking SMS Settings
	ATApiKey   string
	ATUsername string
	ATSenderID string
}

// Load reads settings from environment variables or applies sensible defaults
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	demoMode := true
	if val := os.Getenv("DEMO_MODE"); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			demoMode = parsed
		}
	}

	schoolName := os.Getenv("SCHOOL_NAME")
	if schoolName == "" {
		schoolName = "Hillcrest Premier Academy"
	}

	motto := os.Getenv("SCHOOL_MOTTO")
	if motto == "" {
		motto = "Excellence in Knowledge & Character"
	}

	paybill := os.Getenv("SCHOOL_PAYBILL")
	if paybill == "" {
		paybill = "400200" // Standard demo cooperative/safaricom paybill
	}

	term := os.Getenv("SCHOOL_CURRENT_TERM")
	if term == "" {
		term = "Term 1, 2026"
	}

	return &Config{
		Port:     port,
		BaseURL:  baseURL,
		DemoMode: demoMode,
		School: models.School{
			Name:          schoolName,
			Motto:         motto,
			Paybill:       paybill,
			AccountPrefix: "ADM",
			ContactPhone:  "+254 712 345 678",
			Email:         "bursar@hillcrest.ac.ke",
			Currency:      "KES",
			CurrentTerm:   term,
		},
		DarajaConsumerKey:    os.Getenv("DARAJA_CONSUMER_KEY"),
		DarajaConsumerSecret: os.Getenv("DARAJA_CONSUMER_SECRET"),
		DarajaPasskey:        os.Getenv("DARAJA_PASSKEY"),
		DarajaShortcode:      os.Getenv("DARAJA_SHORTCODE"),
		DarajaCallbackURL:    os.Getenv("DARAJA_CALLBACK_URL"),
		ATApiKey:             os.Getenv("AT_API_KEY"),
		ATUsername:           os.Getenv("AT_USERNAME"),
		ATSenderID:           os.Getenv("AT_SENDER_ID"),
	}
}
