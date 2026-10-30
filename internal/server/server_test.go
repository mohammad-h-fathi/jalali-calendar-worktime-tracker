package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"persian_cal/internal/models"
	"persian_cal/internal/storage"
	"persian_cal/internal/todo"
	"persian_cal/internal/tracker"
)

func TestDeleteWorkSessionRoute(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	// Add a session
	now := time.Now()
	sess := models.WorkSession{
		ID:              "test-sess-1",
		Project:         "Test Project",
		JalaliDate:      "1403-08-04",
		StartTime:       now.Add(-1 * time.Hour),
		EndTime:         now,
		DurationSeconds: 3600,
	}
	if err := store.SaveSession(sess); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	workTracker := tracker.NewWorktimeService(store)
	taskTodo := todo.NewTaskService(store)
	hub := NewSSEHub()
	handler := NewAPIHandler(store, workTracker, taskTodo, hub, nil, nil)

	// Send DELETE request
	req := httptest.NewRequest(http.MethodDelete, "/api/worktime/sessions/test-sess-1", nil)
	w := httptest.NewRecorder()
	handler.handleWorkSessionItem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// Verify session is deleted
	sessions := store.GetSessions("")
	for _, s := range sessions {
		if s.ID == "test-sess-1" {
			t.Errorf("Session was not deleted from store")
		}
	}
}

func TestManualWorktimeValidationAndTimestamps(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	workTracker := tracker.NewWorktimeService(store)
	taskTodo := todo.NewTaskService(store)
	hub := NewSSEHub()
	handler := NewAPIHandler(store, workTracker, taskTodo, hub, nil, nil)

	// Test 1: Invalid time range (end <= start)
	badBody := map[string]interface{}{
		"jalali_date": "1403-08-04",
		"start_time":  "12:00",
		"end_time":    "10:00",
		"project":     "Test",
	}
	b, _ := json.Marshal(badBody)
	req := httptest.NewRequest(http.MethodPost, "/api/worktime/manual", bytes.NewReader(b))
	w := httptest.NewRecorder()
	handler.handleManualWorktime(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for end <= start, got %d", w.Code)
	}

	// Test 2: Valid manual session
	goodBody := map[string]interface{}{
		"jalali_date": "1403-08-04",
		"start_time":  "09:00",
		"end_time":    "11:30",
		"project":     "Coding",
		"description": "Fixing issues",
	}
	b2, _ := json.Marshal(goodBody)
	req2 := httptest.NewRequest(http.MethodPost, "/api/worktime/manual", bytes.NewReader(b2))
	w2 := httptest.NewRecorder()
	handler.handleManualWorktime(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid manual log, got %d: %s", w2.Code, w2.Body.String())
	}

	var created models.WorkSession
	if err := json.Unmarshal(w2.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if created.DurationSeconds != 2*3600+30*60 { // 2.5 hours = 9000s
		t.Errorf("Expected duration 9000s, got %d", created.DurationSeconds)
	}

	// Verify StartTime corresponds to the Jalali date (1403-08-04 = 2024-10-25)
	if created.StartTime.Hour() != 9 || created.StartTime.Minute() != 0 {
		t.Errorf("Expected start time 09:00, got %v", created.StartTime)
	}
	if created.EndTime.Hour() != 11 || created.EndTime.Minute() != 30 {
		t.Errorf("Expected end time 11:30, got %v", created.EndTime)
	}
}
