package tracker

import (
	"path/filepath"
	"testing"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
	"persian_cal/internal/storage"
)

func TestWorktimePausePrecision(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	svc := NewWorktimeService(store)

	_, err = svc.Start("Test Project", "", "", nil)
	if err != nil {
		t.Fatalf("Failed to start timer: %v", err)
	}

	// Fast pauses
	for i := 0; i < 5; i++ {
		time.Sleep(120 * time.Millisecond)
		_, _ = svc.Pause()
		time.Sleep(50 * time.Millisecond)
		_, _ = svc.Resume()
	}

	status, elapsed := svc.GetStatus()
	if !status.IsRunning {
		t.Errorf("Expected timer to be running")
	}
	if elapsed < 0 {
		t.Errorf("Elapsed seconds cannot be negative")
	}
}

func TestWorktimeMidnightCrossing(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	svc := NewWorktimeService(store)

	// Simulate active timer that started yesterday before midnight
	now := time.Now()
	currentDayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := currentDayMidnight.Add(-2 * time.Hour)
	timer := models.ActiveTimer{
		IsRunning:          true,
		IsPaused:           false,
		Project:            "Night Owl Project",
		StartedAt:          yesterday,
		LastResumedAt:      yesterday,
		AccumulatedSeconds: 0,
	}
	_ = store.SaveActiveTimer(timer)

	// Stop timer
	sess, err := svc.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if sess == nil {
		t.Fatalf("Expected non-nil session")
	}

	// Verify sessions in store
	sessions := store.GetSessions("")
	if len(sessions) < 2 {
		t.Fatalf("Expected at least 2 sessions for midnight crossing, got %d", len(sessions))
	}

	startDay := calendar.FromTime(yesterday).FormatISO()
	endDay := calendar.FromTime(now).FormatISO()
	// GetSessions sorts newest first (StartTime.After)
	if sessions[0].JalaliDate != endDay {
		t.Errorf("Newest session date expected %s, got %s", endDay, sessions[0].JalaliDate)
	}
	if sessions[len(sessions)-1].JalaliDate != startDay {
		t.Errorf("Oldest session date expected %s, got %s", startDay, sessions[len(sessions)-1].JalaliDate)
	}
}
