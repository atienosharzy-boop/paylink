package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"paylink/internal/config"
	"paylink/internal/models"
	"paylink/internal/sms"
	"paylink/internal/store"
	"strconv"
	"strings"
)

type DashboardHandler struct {
	cfg   *config.Config
	store *store.Store
	sms   *sms.Service
	tmpl  *template.Template
}

func NewDashboardHandler(cfg *config.Config, s *store.Store, smsSvc *sms.Service, tmpl *template.Template) *DashboardHandler {
	return &DashboardHandler{
		cfg:   cfg,
		store: s,
		sms:   smsSvc,
		tmpl:  tmpl,
	}
}

type DashboardViewData struct {
	School     models.School
	Stats      models.DashboardStats
	Students   []models.Student
	Classes    []string
	Search     string
	ClassFilter string
	StatusFilter string
	BaseURL    string
	DemoMode   bool
}

// ServeDashboard renders the school fee management dashboard
func (h *DashboardHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	classFilter := r.URL.Query().Get("class")
	statusFilter := r.URL.Query().Get("status")

	data := DashboardViewData{
		School:       h.store.School,
		Stats:        h.store.GetStats(),
		Students:     h.store.ListStudents(search, classFilter, statusFilter),
		Classes:      h.store.GetDistinctClasses(),
		Search:       search,
		ClassFilter:  classFilter,
		StatusFilter: statusFilter,
		BaseURL:      h.cfg.BaseURL,
		DemoMode:     h.cfg.DemoMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// AddStudent handles creating a new student
func (h *DashboardHandler) AddStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fee, _ := strconv.ParseInt(r.FormValue("term_fee"), 10, 64)
	std := models.Student{
		FullName:    strings.TrimSpace(r.FormValue("full_name")),
		AdmNo:       strings.TrimSpace(r.FormValue("adm_no")),
		ClassGrade:  strings.TrimSpace(r.FormValue("class_grade")),
		ParentName:  strings.TrimSpace(r.FormValue("parent_name")),
		ParentPhone: strings.TrimSpace(r.FormValue("parent_phone")),
		TermFee:     fee,
	}

	created, err := h.store.AddStudent(std)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(created)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// BulkImport handles batch pasting of students: Name, AdmNo, ParentPhone, Class, Fee
func (h *DashboardHandler) BulkImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawText := r.FormValue("bulk_text")
	lines := strings.Split(rawText, "\n")
	var records []models.Student

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			adm := strings.TrimSpace(parts[1])
			phone := ""
			class := ""
			fee := int64(20000) // default

			if len(parts) >= 3 {
				phone = strings.TrimSpace(parts[2])
			}
			if len(parts) >= 4 {
				class = strings.TrimSpace(parts[3])
			}
			if len(parts) >= 5 {
				if parsed, err := strconv.ParseInt(strings.TrimSpace(parts[4]), 10, 64); err == nil && parsed > 0 {
					fee = parsed
				}
			}

			records = append(records, models.Student{
				FullName:    name,
				AdmNo:       adm,
				ParentPhone: phone,
				ClassGrade:  class,
				TermFee:     fee,
			})
		}
	}

	imported, err := h.store.BulkImport(records)
	if err != nil {
		http.Error(w, fmt.Sprintf("Import error: %v", err), http.StatusBadRequest)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"imported": imported,
		})
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// DeleteStudent removes a student
func (h *DashboardHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.FormValue("id")
	}

	if err := h.store.DeleteStudent(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// SendReminder triggers an SMS reminder to a single student's parent
func (h *DashboardHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.FormValue("id")
	}

	std, err := h.store.GetStudentByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if std.Balance <= 0 {
		http.Error(w, "Student has fully cleared fees, no reminder needed", http.StatusBadRequest)
		return
	}

	payLinkURL := fmt.Sprintf("%s/pay/%s", h.cfg.BaseURL, std.Slug)
	message := h.sms.BuildReminderMessage(h.store.School, *std, payLinkURL)

	if err := h.sms.SendSMS(std.ParentPhone, message); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send SMS: %v", err), http.StatusInternalServerError)
		return
	}

	h.store.LogReminder(models.Reminder{
		StudentID:   std.ID,
		StudentName: std.FullName,
		ParentPhone: std.ParentPhone,
		AmountOwed:  std.Balance,
		Message:     message,
		Status:      "sent",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Reminder sent successfully to " + std.ParentPhone,
		"whatsapp_url":  h.sms.BuildWhatsAppLink(std.ParentPhone, message),
	})
}

// RemindAllOwed sends reminders to every student with a remaining balance
func (h *DashboardHandler) RemindAllOwed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	students := h.store.ListStudents("", "", "owing")
	partial := h.store.ListStudents("", "", "partial")
	owedStudents := append(students, partial...)

	sentCount := 0
	for _, std := range owedStudents {
		if std.Balance > 0 && std.ParentPhone != "" {
			payLinkURL := fmt.Sprintf("%s/pay/%s", h.cfg.BaseURL, std.Slug)
			msg := h.sms.BuildReminderMessage(h.store.School, std, payLinkURL)
			_ = h.sms.SendSMS(std.ParentPhone, msg)
			_ = h.store.LogReminder(models.Reminder{
				StudentID:   std.ID,
				StudentName: std.FullName,
				ParentPhone: std.ParentPhone,
				AmountOwed:  std.Balance,
				Message:     msg,
				Status:      "sent",
			})
			sentCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"sent_count": sentCount,
		"message":    fmt.Sprintf("Sent reminders to %d parents with outstanding balances", sentCount),
	})
}

// ExportCSV downloads the student roster and fee ledger
func (h *DashboardHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.ExportCSV()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("PayLink_Fees_%s.csv", strings.ReplaceAll(h.store.School.CurrentTerm, " ", "_"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(data)
}
