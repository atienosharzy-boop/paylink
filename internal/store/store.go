package store

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"paylink/internal/models"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu        sync.RWMutex
	filePath  string
	School    models.School             `json:"school"`
	Students  map[string]models.Student `json:"students"` // Key: ID
	Payments  []models.Payment          `json:"payments"`
	Reminders []models.Reminder         `json:"reminders"`
}

// NewStore initializes the store from disk or creates default seed data
func NewStore(dbPath string, defaultSchool models.School) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	s := &Store{
		filePath:  dbPath,
		School:    defaultSchool,
		Students:  make(map[string]models.Student),
		Payments:  make([]models.Payment, 0),
		Reminders: make([]models.Reminder, 0),
	}

	// Try reading existing database file
	if data, err := os.ReadFile(dbPath); err == nil && len(data) > 0 {
		var saved struct {
			School    models.School             `json:"school"`
			Students  map[string]models.Student `json:"students"`
			Payments  []models.Payment          `json:"payments"`
			Reminders []models.Reminder         `json:"reminders"`
		}
		if err := json.Unmarshal(data, &saved); err == nil && len(saved.Students) > 0 {
			s.School = saved.School
			s.Students = saved.Students
			s.Payments = saved.Payments
			s.Reminders = saved.Reminders
			return s, nil
		}
	}

	// Seed realistic demo data if file does not exist or was empty
	s.seedDefaultData()
	if err := s.saveLocked(); err != nil {
		return nil, fmt.Errorf("failed to save initial seed database: %w", err)
	}

	return s, nil
}

func (s *Store) saveLocked() error {
	payload := struct {
		School    models.School             `json:"school"`
		Students  map[string]models.Student `json:"students"`
		Payments  []models.Payment          `json:"payments"`
		Reminders []models.Reminder         `json:"reminders"`
	}{
		School:    s.School,
		Students:  s.Students,
		Payments:  s.Payments,
		Reminders: s.Reminders,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, s.filePath)
}

func slugify(name, adm string) string {
	raw := strings.ToLower(name + "-" + adm)
	reg := regexp.MustCompile("[^a-z0-9]+")
	cleaned := reg.ReplaceAllString(raw, "-")
	return strings.Trim(cleaned, "-")
}

func (s *Store) seedDefaultData() {
	now := time.Now()
	students := []models.Student{
		{
			ID:          "std-01",
			AdmNo:       "ADM-2024-012",
			FullName:    "Amani Kimani",
			ClassGrade:  "Grade 4 East",
			ParentName:  "Grace Kimani",
			ParentPhone: "254712345678",
			TermFee:     24500,
			PaidAmount:  24500,
			CreatedAt:   now.AddDate(0, -1, -10),
		},
		{
			ID:          "std-02",
			AdmNo:       "ADM-2024-045",
			FullName:    "Brian Ochieng",
			ClassGrade:  "Grade 4 East",
			ParentName:  "David Ochieng",
			ParentPhone: "254722890123",
			TermFee:     24500,
			PaidAmount:  14000,
			CreatedAt:   now.AddDate(0, -1, -9),
		},
		{
			ID:          "std-03",
			AdmNo:       "ADM-2024-088",
			FullName:    "Faith Wanjiku",
			ClassGrade:  "Grade 3 West",
			ParentName:  "Peter Mwangi",
			ParentPhone: "254733456789",
			TermFee:     22000,
			PaidAmount:  0,
			CreatedAt:   now.AddDate(0, -1, -8),
		},
		{
			ID:          "std-04",
			AdmNo:       "ADM-2024-102",
			FullName:    "Zawadi Mutua",
			ClassGrade:  "Grade 3 West",
			ParentName:  "Mercy Mutua",
			ParentPhone: "254701234567",
			TermFee:     22000,
			PaidAmount:  11000,
			CreatedAt:   now.AddDate(0, -1, -7),
		},
		{
			ID:          "std-05",
			AdmNo:       "ADM-2024-115",
			FullName:    "Kevin Kiprop",
			ClassGrade:  "Grade 2 North",
			ParentName:  "Joshua Kiprop",
			ParentPhone: "254718901234",
			TermFee:     19500,
			PaidAmount:  19500,
			CreatedAt:   now.AddDate(0, -1, -6),
		},
		{
			ID:          "std-06",
			AdmNo:       "ADM-2024-140",
			FullName:    "Halima Hassan",
			ClassGrade:  "Grade 2 North",
			ParentName:  "Amina Hassan",
			ParentPhone: "254726543210",
			TermFee:     19500,
			PaidAmount:  0,
			CreatedAt:   now.AddDate(0, -1, -5),
		},
		{
			ID:          "std-07",
			AdmNo:       "ADM-2024-162",
			FullName:    "Emmanuel Baraka",
			ClassGrade:  "Grade 1 Central",
			ParentName:  "Samuel Baraka",
			ParentPhone: "254790112233",
			TermFee:     18000,
			PaidAmount:  10000,
			CreatedAt:   now.AddDate(0, -1, -4),
		},
		{
			ID:          "std-08",
			AdmNo:       "ADM-2024-190",
			FullName:    "Precious Moraa",
			ClassGrade:  "Grade 1 Central",
			ParentName:  "Beatrice Moraa",
			ParentPhone: "254745678901",
			TermFee:     18000,
			PaidAmount:  0,
			CreatedAt:   now.AddDate(0, -1, -3),
		},
	}

	for _, std := range students {
		std.Slug = slugify(std.FullName, std.AdmNo)
		std.Recalculate()
		s.Students[std.ID] = std
	}

	// Seed some initial payments
	s.Payments = []models.Payment{
		{
			ID:           "pay-001",
			StudentID:    "std-01",
			StudentName:  "Amani Kimani",
			AdmNo:        "ADM-2024-012",
			ClassGrade:   "Grade 4 East",
			Amount:       24500,
			MpesaReceipt: "SKL4X9Q2TB",
			PhoneNumber:  "254712345678",
			Status:       "completed",
			PaidAt:       now.AddDate(0, 0, -4),
			BalanceAfter: 0,
			Notes:        "Term 1 2026 Full Clearance",
		},
		{
			ID:           "pay-002",
			StudentID:    "std-02",
			StudentName:  "Brian Ochieng",
			AdmNo:        "ADM-2024-045",
			ClassGrade:   "Grade 4 East",
			Amount:       14000,
			MpesaReceipt: "RKT892M3LA",
			PhoneNumber:  "254722890123",
			Status:       "completed",
			PaidAt:       now.AddDate(0, 0, -2),
			BalanceAfter: 10500,
			Notes:        "First installment",
		},
		{
			ID:           "pay-003",
			StudentID:    "std-04",
			StudentName:  "Zawadi Mutua",
			AdmNo:        "ADM-2024-102",
			ClassGrade:   "Grade 3 West",
			Amount:       11000,
			MpesaReceipt: "QJD114P0ZX",
			PhoneNumber:  "254701234567",
			Status:       "completed",
			PaidAt:       now.AddDate(0, 0, -1),
			BalanceAfter: 11000,
			Notes:        "Half term fee",
		},
		{
			ID:           "pay-004",
			StudentID:    "std-05",
			StudentName:  "Kevin Kiprop",
			AdmNo:        "ADM-2024-115",
			ClassGrade:   "Grade 2 North",
			Amount:       19500,
			MpesaReceipt: "TXN552B7KL",
			PhoneNumber:  "254718901234",
			Status:       "completed",
			PaidAt:       now.Add(-6 * time.Hour),
			BalanceAfter: 0,
			Notes:        "Full payment via link",
		},
	}
}

// GetStats calculates summary stats
func (s *Store) GetStats() models.DashboardStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var stats models.DashboardStats
	stats.TotalStudents = len(s.Students)

	for _, std := range s.Students {
		stats.TotalExpectedFees += std.TermFee
		stats.TotalCollected += std.PaidAmount
		stats.TotalOutstanding += std.Balance

		switch std.Status {
		case "cleared":
			stats.ClearedCount++
		case "partial":
			stats.PartialCount++
		case "owing":
			stats.OwingCount++
		}
	}

	if stats.TotalStudents > 0 {
		stats.ClearancePercentage = (float64(stats.ClearedCount) / float64(stats.TotalStudents)) * 100
	}

	// Copy payments sorted newest first
	n := len(s.Payments)
	recent := make([]models.Payment, n)
	copy(recent, s.Payments)
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].PaidAt.After(recent[j].PaidAt)
	})

	if len(recent) > 10 {
		recent = recent[:10]
	}
	stats.RecentPayments = recent

	return stats
}

// ListStudents returns filtered students
func (s *Store) ListStudents(search, classFilter, statusFilter string) []models.Student {
	s.mu.RLock()
	defer s.mu.RUnlock()

	search = strings.ToLower(strings.TrimSpace(search))
	classFilter = strings.TrimSpace(classFilter)
	statusFilter = strings.ToLower(strings.TrimSpace(statusFilter))

	list := make([]models.Student, 0, len(s.Students))
	for _, std := range s.Students {
		if search != "" {
			nameMatch := strings.Contains(strings.ToLower(std.FullName), search)
			admMatch := strings.Contains(strings.ToLower(std.AdmNo), search)
			parentMatch := strings.Contains(strings.ToLower(std.ParentName), search)
			phoneMatch := strings.Contains(std.ParentPhone, search)
			if !nameMatch && !admMatch && !parentMatch && !phoneMatch {
				continue
			}
		}

		if classFilter != "" && !strings.EqualFold(std.ClassGrade, classFilter) {
			continue
		}

		if statusFilter != "" && !strings.EqualFold(std.Status, statusFilter) {
			continue
		}

		list = append(list, std)
	}

	// Sort alphabetically by class and name
	sort.Slice(list, func(i, j int) bool {
		if list[i].ClassGrade == list[j].ClassGrade {
			return list[i].FullName < list[j].FullName
		}
		return list[i].ClassGrade < list[j].ClassGrade
	})

	return list
}

// GetDistinctClasses returns all classes
func (s *Store) GetDistinctClasses() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := make(map[string]struct{})
	for _, std := range s.Students {
		if std.ClassGrade != "" {
			set[std.ClassGrade] = struct{}{}
		}
	}

	classes := make([]string, 0, len(set))
	for c := range set {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	return classes
}

// GetStudentByID retrieves student by ID
func (s *Store) GetStudentByID(id string) (*models.Student, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	std, ok := s.Students[id]
	if !ok {
		return nil, fmt.Errorf("student not found")
	}
	return &std, nil
}

// GetStudentBySlug retrieves student by their unique payment link slug
func (s *Store) GetStudentBySlug(slug string) (*models.Student, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, std := range s.Students {
		if std.Slug == slug {
			return &std, nil
		}
	}
	return nil, fmt.Errorf("payment link not found")
}

// AddStudent registers a new student
func (s *Store) AddStudent(sIn models.Student) (*models.Student, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(sIn.FullName) == "" {
		return nil, fmt.Errorf("student name cannot be blank")
	}
	if strings.TrimSpace(sIn.AdmNo) == "" {
		return nil, fmt.Errorf("admission number cannot be blank")
	}

	// Check unique admission number
	for _, existing := range s.Students {
		if strings.EqualFold(existing.AdmNo, sIn.AdmNo) {
			return nil, fmt.Errorf("admission number %s already exists", sIn.AdmNo)
		}
	}

	id := fmt.Sprintf("std-%d", time.Now().UnixNano()%1000000)
	sIn.ID = id
	sIn.ParentPhone = models.NormalizeKenyanPhone(sIn.ParentPhone)
	sIn.Slug = slugify(sIn.FullName, sIn.AdmNo)
	sIn.CreatedAt = time.Now()
	sIn.Recalculate()

	s.Students[id] = sIn
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return &sIn, nil
}

// UpdateStudent edits an existing student
func (s *Store) UpdateStudent(sIn models.Student) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.Students[sIn.ID]
	if !ok {
		return fmt.Errorf("student not found")
	}

	existing.FullName = sIn.FullName
	existing.ClassGrade = sIn.ClassGrade
	existing.ParentName = sIn.ParentName
	existing.ParentPhone = models.NormalizeKenyanPhone(sIn.ParentPhone)
	existing.TermFee = sIn.TermFee
	existing.Recalculate()

	s.Students[existing.ID] = existing
	return s.saveLocked()
}

// DeleteStudent removes student
func (s *Store) DeleteStudent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Students[id]; !ok {
		return fmt.Errorf("student not found")
	}
	delete(s.Students, id)
	return s.saveLocked()
}

// BulkImport accepts lines of CSV or text (Format: Name, AdmNo, ParentPhone, Class, TermFee)
func (s *Store) BulkImport(records []models.Student) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	imported := 0
	for _, rec := range records {
		if strings.TrimSpace(rec.FullName) == "" || strings.TrimSpace(rec.AdmNo) == "" {
			continue
		}
		// check exists
		exists := false
		for _, e := range s.Students {
			if strings.EqualFold(e.AdmNo, rec.AdmNo) {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		id := fmt.Sprintf("std-%d", time.Now().UnixNano()%10000000)
		rec.ID = id
		rec.ParentPhone = models.NormalizeKenyanPhone(rec.ParentPhone)
		rec.Slug = slugify(rec.FullName, rec.AdmNo)
		rec.CreatedAt = time.Now()
		rec.Recalculate()

		s.Students[id] = rec
		imported++
	}

	if imported > 0 {
		if err := s.saveLocked(); err != nil {
			return imported, err
		}
	}
	return imported, nil
}

// RecordPayment records a partial or full payment, updates the student balance, and creates an audit record
func (s *Store) RecordPayment(studentID, phone, notes string, amount int64, receipt string) (*models.Payment, *models.Student, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	std, ok := s.Students[studentID]
	if !ok {
		return nil, nil, fmt.Errorf("student not found")
	}

	if amount <= 0 {
		return nil, nil, fmt.Errorf("payment amount must be greater than zero")
	}
	if amount > std.Balance {
		return nil, nil, fmt.Errorf("payment amount of %s exceeds remaining balance of %s",
			models.FormatMoney(amount), models.FormatMoney(std.Balance))
	}

	std.PaidAmount += amount
	std.Recalculate()
	s.Students[studentID] = std

	payment := models.Payment{
		ID:           fmt.Sprintf("pay-%d", time.Now().UnixNano()%1000000),
		StudentID:    std.ID,
		StudentName:  std.FullName,
		AdmNo:        std.AdmNo,
		ClassGrade:   std.ClassGrade,
		Amount:       amount,
		MpesaReceipt: receipt,
		PhoneNumber:  models.NormalizeKenyanPhone(phone),
		Status:       "completed",
		PaidAt:       time.Now(),
		BalanceAfter: std.Balance,
		Notes:        notes,
	}

	s.Payments = append(s.Payments, payment)
	if err := s.saveLocked(); err != nil {
		return nil, nil, err
	}

	return &payment, &std, nil
}

// ListPayments returns all payments
func (s *Store) ListPayments() []models.Payment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.Payment, len(s.Payments))
	copy(list, s.Payments)
	sort.Slice(list, func(i, j int) bool {
		return list[i].PaidAt.After(list[j].PaidAt)
	})
	return list
}

// GetPaymentByID finds a specific payment
func (s *Store) GetPaymentByID(id string) (*models.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.Payments {
		if p.ID == id || p.MpesaReceipt == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("payment receipt not found")
}

// LogReminder stores reminder audit
func (s *Store) LogReminder(r models.Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r.ID = fmt.Sprintf("rem-%d", time.Now().UnixNano()%1000000)
	r.SentAt = time.Now()
	s.Reminders = append(s.Reminders, r)
	return s.saveLocked()
}

// ExportCSV creates a downloadable spreadsheet of all students and fee statuses
func (s *Store) ExportCSV() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Write header
	header := []string{
		"Admission No", "Student Name", "Class", "Parent Name", "Parent Phone",
		"Term Fee (KES)", "Amount Paid (KES)", "Balance (KES)", "Status", "Payment Slug",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, std := range s.Students {
		row := []string{
			std.AdmNo,
			std.FullName,
			std.ClassGrade,
			std.ParentName,
			std.ParentPhone,
			fmt.Sprintf("%d", std.TermFee),
			fmt.Sprintf("%d", std.PaidAmount),
			fmt.Sprintf("%d", std.Balance),
			std.Status,
			std.Slug,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}
