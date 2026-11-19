package store_test

import (
	"os"
	"path/filepath"
	"paylink/internal/models"
	"paylink/internal/store"
	"testing"
)

func TestStore_PaymentAndBalanceRecalculation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_paylink.json")

	school := models.School{
		Name:    "Test Academy",
		Paybill: "123456",
	}

	st, err := store.NewStore(dbPath, school)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Add test student
	newStd := models.Student{
		FullName:    "Jane Doe",
		AdmNo:       "ADM-TEST-001",
		ClassGrade:  "Grade 1",
		ParentPhone: "0711223344",
		TermFee:     20000,
	}

	created, err := st.AddStudent(newStd)
	if err != nil {
		t.Fatalf("failed to add student: %v", err)
	}

	if created.Balance != 20000 {
		t.Fatalf("expected balance 20000, got %d", created.Balance)
	}
	if created.Status != "owing" {
		t.Fatalf("expected status owing, got %s", created.Status)
	}

	// Make partial payment of 8,000
	pay1, stdAfter1, err := st.RecordPayment(created.ID, "0711223344", "Installment 1", 8000, "REC-001")
	if err != nil {
		t.Fatalf("failed to record payment: %v", err)
	}

	if pay1.Amount != 8000 {
		t.Errorf("expected payment amount 8000, got %d", pay1.Amount)
	}
	if stdAfter1.Balance != 12000 {
		t.Errorf("expected balance 12000, got %d", stdAfter1.Balance)
	}
	if stdAfter1.Status != "partial" {
		t.Errorf("expected status partial, got %s", stdAfter1.Status)
	}

	// Overpayment should fail
	_, _, err = st.RecordPayment(created.ID, "0711223344", "Overpay", 15000, "REC-FAIL")
	if err == nil {
		t.Fatalf("expected error on overpayment, got nil")
	}

	// Clear remaining balance
	_, stdAfter2, err := st.RecordPayment(created.ID, "0711223344", "Final payment", 12000, "REC-002")
	if err != nil {
		t.Fatalf("failed to clear balance: %v", err)
	}

	if stdAfter2.Balance != 0 {
		t.Errorf("expected balance 0, got %d", stdAfter2.Balance)
	}
	if stdAfter2.Status != "cleared" {
		t.Errorf("expected status cleared, got %s", stdAfter2.Status)
	}

	// Verify persistence: load new store from same file
	stReloaded, err := store.NewStore(dbPath, school)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}

	reloadedStd, err := stReloaded.GetStudentByID(created.ID)
	if err != nil {
		t.Fatalf("failed to get reloaded student: %v", err)
	}

	if reloadedStd.Balance != 0 || reloadedStd.Status != "cleared" {
		t.Errorf("persistence mismatch: balance=%d, status=%s", reloadedStd.Balance, reloadedStd.Status)
	}
}

func TestStore_ExportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_csv.json")

	st, err := store.NewStore(dbPath, models.School{Name: "CSV Academy"})
	if err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	csvData, err := st.ExportCSV()
	if err != nil {
		t.Fatalf("failed to export CSV: %v", err)
	}

	if len(csvData) == 0 {
		t.Errorf("exported CSV is empty")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
