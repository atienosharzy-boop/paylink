package models

import (
	"fmt"
	"strings"
	"time"
)

// School holds configuration and profile for the school or academy
type School struct {
	Name         string `json:"name"`
	Motto        string `json:"motto"`
	Paybill      string `json:"paybill"`
	AccountPrefix string `json:"account_prefix"`
	ContactPhone string `json:"contact_phone"`
	Email        string `json:"email"`
	Currency     string `json:"currency"`
	CurrentTerm  string `json:"current_term"`
}

// Student represents a registered learner with fee tracking
type Student struct {
	ID          string    `json:"id"`
	AdmNo       string    `json:"adm_no"`
	FullName    string    `json:"full_name"`
	ClassGrade  string    `json:"class_grade"`
	ParentName  string    `json:"parent_name"`
	ParentPhone string    `json:"parent_phone"`
	TermFee     int64     `json:"term_fee"`
	PaidAmount  int64     `json:"paid_amount"`
	Balance     int64     `json:"balance"`
	Status      string    `json:"status"` // "cleared", "partial", "owing"
	Slug        string    `json:"slug"`   // Unique slug for payment link e.g. "amani-kimani-042"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Recalculate recalculates Balance and Status based on TermFee and PaidAmount
func (s *Student) Recalculate() {
	s.Balance = s.TermFee - s.PaidAmount
	if s.Balance <= 0 {
		s.Balance = 0
		s.Status = "cleared"
	} else if s.PaidAmount > 0 {
		s.Status = "partial"
	} else {
		s.Status = "owing"
	}
	s.UpdatedAt = time.Now()
}

// Payment records a specific fee payment via M-Pesa
type Payment struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	StudentName  string    `json:"student_name"`
	AdmNo        string    `json:"adm_no"`
	ClassGrade   string    `json:"class_grade"`
	Amount       int64     `json:"amount"`
	MpesaReceipt string    `json:"mpesa_receipt"`
	PhoneNumber  string    `json:"phone_number"`
	Status       string    `json:"status"` // "completed", "pending", "failed"
	PaidAt       time.Time `json:"paid_at"`
	BalanceAfter int64     `json:"balance_after"`
	Notes        string    `json:"notes"`
}

// Reminder logs an SMS/WhatsApp reminder sent to a parent
type Reminder struct {
	ID          string    `json:"id"`
	StudentID   string    `json:"student_id"`
	StudentName string    `json:"student_name"`
	ParentPhone string    `json:"parent_phone"`
	AmountOwed  int64     `json:"amount_owed"`
	Message     string    `json:"message"`
	SentAt      time.Time `json:"sent_at"`
	Status      string    `json:"status"` // "sent", "simulated"
}

// DashboardStats provides real-time financial and clearance aggregates
type DashboardStats struct {
	TotalStudents      int       `json:"total_students"`
	TotalExpectedFees  int64     `json:"total_expected_fees"`
	TotalCollected     int64     `json:"total_collected"`
	TotalOutstanding   int64     `json:"total_outstanding"`
	ClearedCount       int       `json:"cleared_count"`
	PartialCount       int       `json:"partial_count"`
	OwingCount         int       `json:"owing_count"`
	ClearancePercentage float64   `json:"clearance_percentage"`
	RecentPayments     []Payment `json:"recent_payments"`
}

// NormalizeKenyanPhone standardizes local numbers (0712..., +254712..., 254712...) to 2547XXXXXXXX
func NormalizeKenyanPhone(phone string) string {
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.TrimPrefix(cleaned, "+")

	if strings.HasPrefix(cleaned, "0") && len(cleaned) == 10 {
		return "254" + cleaned[1:]
	}
	if strings.HasPrefix(cleaned, "7") && len(cleaned) == 9 {
		return "254" + cleaned
	}
	if strings.HasPrefix(cleaned, "1") && len(cleaned) == 9 {
		return "254" + cleaned
	}
	return cleaned
}

// FormatMoney returns formatted Kenyan Shillings e.g. "KES 15,000"
func FormatMoney(amount int64) string {
	in := fmt.Sprintf("%d", amount)
	n := len(in)
	if n <= 3 {
		return "KES " + in
	}
	var out []byte
	rem := n % 3
	if rem > 0 {
		out = append(out, in[:rem]...)
		if n > rem {
			out = append(out, ',')
		}
	}
	for i := rem; i < n; i += 3 {
		out = append(out, in[i:i+3]...)
		if i+3 < n {
			out = append(out, ',')
		}
	}
	return "KES " + string(out)
}
