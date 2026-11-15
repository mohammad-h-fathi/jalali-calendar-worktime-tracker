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

func TestWorktimeSelectAndSwitchProject(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	svc := NewWorktimeService(store)

	// 1. Select a project while idle and verify Start("") uses it
	selected, err := svc.SelectProject("توسعه نرم‌افزار", "")
	if err != nil {
		t.Fatalf("SelectProject failed: %v", err)
	}
	if selected.Project != "توسعه نرم‌افزار" || selected.Color != "#10b981" {
		t.Fatalf("Expected selected project توسعه نرم‌افزار (#10b981), got %q (%q)", selected.Project, selected.Color)
	}

	started, err := svc.Start("", "", "", nil)
	if err != nil {
		t.Fatalf("Start with empty project failed: %v", err)
	}
	if !started.IsRunning || started.Project != "توسعه نرم‌افزار" {
		t.Fatalf("Expected running timer for توسعه نرم‌افزار, got running=%v project=%q", started.IsRunning, started.Project)
	}

	// Simulate 10 seconds elapsed on the first timer so switching saves a session
	running := store.GetActiveTimer()
	running.StartedAt = time.Now().Add(-10 * time.Second)
	running.LastResumedAt = running.StartedAt
	if err := store.SaveActiveTimer(running); err != nil {
		t.Fatalf("SaveActiveTimer failed: %v", err)
	}

	// 2. Calling Start with another project while running stops/saves first and starts second
	switched, err := svc.Start("Second Project", "Working on second", "#8b5cf6", nil)
	if err != nil {
		t.Fatalf("Switching project via Start failed: %v", err)
	}
	if !switched.IsRunning || switched.Project != "Second Project" || switched.Color != "#8b5cf6" {
		t.Fatalf("Expected running timer for Second Project, got running=%v project=%q color=%q", switched.IsRunning, switched.Project, switched.Color)
	}

	sessions := store.GetSessions("")
	if len(sessions) != 1 || sessions[0].Project != "توسعه نرم‌افزار" {
		t.Fatalf("Expected 1 recorded session for توسعه نرم‌افزار, got %+v", sessions)
	}

	// 3. Stopping preserves the last active project on the idle ActiveTimer
	if _, err := svc.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	idle, _ := svc.GetStatus()
	if idle.IsRunning {
		t.Errorf("Expected timer to be idle after Stop")
	}
	if idle.Project != "Second Project" || idle.Color != "#8b5cf6" {
		t.Errorf("Expected idle timer to preserve Second Project (#8b5cf6), got %q (%q)", idle.Project, idle.Color)
	}
}
