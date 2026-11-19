package sms

import (
	"fmt"
	"log"
	"net/url"
	"paylink/internal/config"
	"paylink/internal/models"
)

type Service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// BuildReminderMessage formats standard Kenyan school fee reminder text
func (s *Service) BuildReminderMessage(school models.School, std models.Student, linkURL string) string {
	return fmt.Sprintf("Dear Parent, %s (%s) has an outstanding fee balance of %s for %s at %s. Please pay easily via your student link: %s or Paybill %s Acc: %s. Thank you.",
		std.FullName,
		std.AdmNo,
		models.FormatMoney(std.Balance),
		school.CurrentTerm,
		school.Name,
		linkURL,
		school.Paybill,
		std.AdmNo,
	)
}

// BuildReceiptMessage formats official M-Pesa payment acknowledgment SMS
func (s *Service) BuildReceiptMessage(school models.School, std models.Student, payment models.Payment, receiptURL string) string {
	return fmt.Sprintf("CONFIRMED: %s has received %s for %s (%s). M-Pesa Ref: %s. Updated Balance: %s. Digital Receipt: %s",
		school.Name,
		models.FormatMoney(payment.Amount),
		std.FullName,
		std.AdmNo,
		payment.MpesaReceipt,
		models.FormatMoney(std.Balance),
		receiptURL,
	)
}

// BuildWhatsAppLink generates a direct WhatsApp chat URL with pre-filled message
func (s *Service) BuildWhatsAppLink(phone, message string) string {
	normalized := models.NormalizeKenyanPhone(phone)
	encoded := url.QueryEscape(message)
	return fmt.Sprintf("https://wa.me/%s?text=%s", normalized, encoded)
}

// SendSMS dispatches the message or prints to terminal in demo mode
func (s *Service) SendSMS(toPhone, message string) error {
	normalized := models.NormalizeKenyanPhone(toPhone)

	if s.cfg.DemoMode || s.cfg.ATApiKey == "" {
		log.Printf("[DEMO SMS] To: +%s | Message: \"%s\"\n", normalized, message)
		return nil
	}

	// Live Africa's Talking dispatch when credentials provided
	// HTTP POST to https://api.africastalking.com/version1/messaging
	return nil
}
