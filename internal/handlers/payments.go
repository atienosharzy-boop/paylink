package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"paylink/internal/config"
	"paylink/internal/models"
	"paylink/internal/mpesa"
	"paylink/internal/sms"
	"paylink/internal/store"
)

type PaymentHandler struct {
	cfg   *config.Config
	store *store.Store
	mpesa *mpesa.Client
	sms   *sms.Service
	tmpl  *template.Template
}

func NewPaymentHandler(cfg *config.Config, s *store.Store, m *mpesa.Client, smsSvc *sms.Service, tmpl *template.Template) *PaymentHandler {
	return &PaymentHandler{
		cfg:   cfg,
		store: s,
		mpesa: m,
		sms:   smsSvc,
		tmpl:  tmpl,
	}
}

type PayViewData struct {
	School       models.School
	Student      models.Student
	HalfBalance  int64
	ThirdBalance int64
	BaseURL      string
	DemoMode     bool
}

type ReceiptViewData struct {
	School   models.School
	Payment  models.Payment
	Student  models.Student
	BaseURL  string
	DemoMode bool
}

// ServeLandingPage renders the school-focused PayLink marketing & intro page
func (h *PaymentHandler) ServeLandingPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"School":   h.store.School,
		"Stats":    h.store.GetStats(),
		"BaseURL":  h.cfg.BaseURL,
		"DemoMode": h.cfg.DemoMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "landing.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// ServePayPage renders the parent fee payment link
func (h *PaymentHandler) ServePayPage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	std, err := h.store.GetStudentBySlug(slug)
	if err != nil {
		http.Error(w, "Payment link not found or expired", http.StatusNotFound)
		return
	}

	half := std.Balance / 2
	if half < 1 {
		half = std.Balance
	}

	third := std.Balance / 3
	if third < 1 {
		third = half
	}

	data := PayViewData{
		School:       h.store.School,
		Student:      *std,
		HalfBalance:  half,
		ThirdBalance: third,
		BaseURL:      h.cfg.BaseURL,
		DemoMode:     h.cfg.DemoMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "pay.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// InitiateSTKPush starts an M-Pesa prompt to the parent
func (h *PaymentHandler) InitiateSTKPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StudentID   string `json:"student_id"`
		Amount      int64  `json:"amount"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON", http.StatusBadRequest)
		return
	}

	std, err := h.store.GetStudentByID(req.StudentID)
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	if req.Amount <= 0 || req.Amount > std.Balance {
		http.Error(w, fmt.Sprintf("Amount must be between KES 1 and remaining balance of %s", models.FormatMoney(std.Balance)), http.StatusBadRequest)
		return
	}

	pushResp, err := h.mpesa.InitiateSTKPush(mpesa.STKPushRequest{
		PhoneNumber: models.NormalizeKenyanPhone(req.PhoneNumber),
		Amount:      req.Amount,
		AccountRef:  std.AdmNo,
		Description: fmt.Sprintf("%s Fee", h.store.School.Name),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pushResp)
}

// ConfirmPayment finalizes the transaction, records the payment, and issues SMS/receipt
func (h *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StudentID     string `json:"student_id"`
		Amount        int64  `json:"amount"`
		PhoneNumber   string `json:"phone_number"`
		ReceiptNumber string `json:"receipt_number"`
		Notes         string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON", http.StatusBadRequest)
		return
	}

	if req.ReceiptNumber == "" {
		req.ReceiptNumber = mpesa.GenerateMpesaReceipt()
	}

	payment, std, err := h.store.RecordPayment(req.StudentID, req.PhoneNumber, req.Notes, req.Amount, req.ReceiptNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	receiptURL := fmt.Sprintf("%s/receipt/%s", h.cfg.BaseURL, payment.MpesaReceipt)
	smsMsg := h.sms.BuildReceiptMessage(h.store.School, *std, *payment, receiptURL)
	_ = h.sms.SendSMS(payment.PhoneNumber, smsMsg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"payment_id":     payment.ID,
		"receipt_number": payment.MpesaReceipt,
		"amount":         payment.Amount,
		"balance_after":  payment.BalanceAfter,
		"receipt_url":    receiptURL,
	})
}

// ServeReceiptPage renders the official printable digital fee receipt
func (h *PaymentHandler) ServeReceiptPage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	payment, err := h.store.GetPaymentByID(id)
	if err != nil {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	std, _ := h.store.GetStudentByID(payment.StudentID)
	var studentData models.Student
	if std != nil {
		studentData = *std
	} else {
		studentData = models.Student{
			FullName:   payment.StudentName,
			AdmNo:      payment.AdmNo,
			ClassGrade: payment.ClassGrade,
		}
	}

	data := ReceiptViewData{
		School:   h.store.School,
		Payment:  *payment,
		Student:  studentData,
		BaseURL:  h.cfg.BaseURL,
		DemoMode: h.cfg.DemoMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "receipt.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}
