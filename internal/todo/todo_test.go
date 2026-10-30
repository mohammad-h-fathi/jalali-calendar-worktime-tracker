package todo

import (
	"path/filepath"
	"testing"

	"persian_cal/internal/models"
	"persian_cal/internal/storage"
)

func TestCreateTaskDateValidation(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	svc := NewTaskService(store)

	// Valid task with ISO date
	task1, err := svc.CreateTask(models.Task{
		Title:      "Valid Task",
		JalaliDate: "1403-08-04",
	})
	if err != nil {
		t.Fatalf("Expected valid task creation, got error: %v", err)
	}
	if task1.JalaliDate != "1403-08-04" {
		t.Errorf("Expected JalaliDate 1403-08-04, got %s", task1.JalaliDate)
	}

	// Valid task with slash date (should be normalized)
	task2, err := svc.CreateTask(models.Task{
		Title:      "Slash Task",
		JalaliDate: "1403/8/4",
	})
	if err != nil {
		t.Fatalf("Expected normalized slash task creation, got error: %v", err)
	}
	if task2.JalaliDate != "1403-08-04" {
		t.Errorf("Expected normalized JalaliDate 1403-08-04, got %s", task2.JalaliDate)
	}

	// Invalid date
	_, err = svc.CreateTask(models.Task{
		Title:      "Invalid Date Task",
		JalaliDate: "1404-12-30", // 1404 is common year, Esfand has 29 days
	})
	if err == nil {
		t.Fatalf("Expected error for invalid date 1404-12-30, got nil")
	}
}
