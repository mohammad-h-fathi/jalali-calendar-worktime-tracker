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

func TestGetProjectsSummaryAndTimeline(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "data.json"))
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	svc := NewWorktimeService(store)

	base := time.Date(2026, 11, 10, 10, 0, 0, 0, time.Local)
	sessions := []models.WorkSession{
		{
			ID:              "s1",
			Project:         "توسعه نرم‌افزار",
			Description:     "API work",
			JalaliDate:      "1405-08-19",
			StartTime:       base,
			EndTime:         base.Add(30 * time.Minute),
			DurationSeconds: 1800,
			Color:           "#10b981",
		},
		{
			ID:              "s2",
			Project:         "توسعه نرم‌افزار",
			Description:     "UI polish",
			JalaliDate:      "1405-08-20",
			StartTime:       base.Add(24 * time.Hour),
			EndTime:         base.Add(25 * time.Hour),
			DurationSeconds: 3600,
			Color:           "#10b981",
		},
		{
			ID:              "s3",
			Project:         "توسعه نرم‌افزار",
			Description:     "Tests",
			JalaliDate:      "1405-08-20",
			StartTime:       base.Add(26 * time.Hour),
			EndTime:         base.Add(26*time.Hour + 15*time.Minute),
			DurationSeconds: 900,
			Color:           "#10b981",
		},
		{
			ID:              "s4",
			Project:         "پروژه جدید",
			Description:     "Kickoff",
			JalaliDate:      "1405-08-21",
			StartTime:       base.Add(48 * time.Hour),
			EndTime:         base.Add(48*time.Hour + 10*time.Minute),
			DurationSeconds: 600,
			Color:           "#ef4444",
		},
	}
	for _, s := range sessions {
		if err := store.SaveSession(s); err != nil {
			t.Fatalf("SaveSession failed: %v", err)
		}
	}

	summary := svc.GetProjectsSummary()
	if summary.TotalSeconds != 6900 || summary.SessionCount != 4 {
		t.Fatalf("Expected total 6900s and 4 sessions, got %d (%d sessions)", summary.TotalSeconds, summary.SessionCount)
	}
	if len(summary.Projects) == 0 {
		t.Fatalf("Expected projects in summary")
	}

	// Highest total project should be first: توسعه نرم‌افزار (6300s)
	top := summary.Projects[0]
	if top.Project != "توسعه نرم‌افزار" {
		t.Fatalf("Expected top project توسعه نرم‌افزار, got %s", top.Project)
	}
	if top.TotalSeconds != 6300 || top.SessionCount != 3 || top.ActiveDays != 2 {
		t.Fatalf("Unexpected top project stats: %+v", top)
	}
	if len(top.Timeline) != 2 {
		t.Fatalf("Expected 2 timeline days, got %d", len(top.Timeline))
	}
	// Newest day first: 1405-08-20 (4500s, 2 sessions newest first)
	if top.Timeline[0].JalaliDate != "1405-08-20" || top.Timeline[0].DurationSeconds != 4500 || top.Timeline[0].SessionCount != 2 {
		t.Fatalf("Unexpected first timeline day: %+v", top.Timeline[0])
	}
	if top.Timeline[0].Sessions[0].ID != "s3" || top.Timeline[0].Sessions[1].ID != "s2" {
		t.Fatalf("Expected sessions sorted newest-first (s3 then s2), got %s then %s", top.Timeline[0].Sessions[0].ID, top.Timeline[0].Sessions[1].ID)
	}
	if top.Timeline[1].JalaliDate != "1405-08-19" || top.Timeline[1].DurationSeconds != 1800 {
		t.Fatalf("Unexpected second timeline day: %+v", top.Timeline[1])
	}

	// Ad-hoc project not in definedProjects should also appear (second with 600s)
	second := summary.Projects[1]
	if second.Project != "پروژه جدید" || second.TotalSeconds != 600 || second.ActiveDays != 1 {
		t.Fatalf("Unexpected second project stats: %+v", second)
	}
}
