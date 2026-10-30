package tracker

import (
	"errors"
	"sync"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
	"persian_cal/internal/storage"
)

// WorktimeService manages timer operations and analytics.
type WorktimeService struct {
	mu          sync.RWMutex
	listenersMu sync.RWMutex
	store       *storage.Store
	listeners   []func(event string, timer models.ActiveTimer)
}

// NewWorktimeService creates a new worktime service instance.
func NewWorktimeService(store *storage.Store) *WorktimeService {
	return &WorktimeService{
		store:     store,
		listeners: make([]func(event string, timer models.ActiveTimer), 0),
	}
}

// OnChange registers a listener for timer state updates.
func (w *WorktimeService) OnChange(listener func(event string, timer models.ActiveTimer)) {
	w.listenersMu.Lock()
	defer w.listenersMu.Unlock()
	w.listeners = append(w.listeners, listener)
}

func (w *WorktimeService) notify(event string, timer models.ActiveTimer) {
	w.listenersMu.RLock()
	listeners := make([]func(event string, timer models.ActiveTimer), len(w.listeners))
	copy(listeners, w.listeners)
	w.listenersMu.RUnlock()

	for _, l := range listeners {
		go l(event, timer)
	}
}

// Start launches a new timer session.
func (w *WorktimeService) Start(project, description, color string, taskID *string) (*models.ActiveTimer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	current := w.store.GetActiveTimer()
	if current.IsRunning {
		return nil, errors.New("یک تایمر هم‌اکنون در حال اجراست")
	}

	if project == "" {
		project = "پروژه عمومی"
	}
	if color == "" {
		color = "#3b82f6"
	}

	now := time.Now()
	timer := models.ActiveTimer{
		IsRunning:          true,
		IsPaused:           false,
		Project:            project,
		TaskID:             taskID,
		Description:        description,
		Color:              color,
		StartedAt:          now,
		LastResumedAt:      now,
		AccumulatedSeconds: 0,
	}

	if err := w.store.SaveActiveTimer(timer); err != nil {
		return nil, err
	}

	w.notify("start", timer)
	return &timer, nil
}

// Pause pauses the running timer.
func (w *WorktimeService) Pause() (*models.ActiveTimer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	current := w.store.GetActiveTimer()
	if !current.IsRunning {
		return nil, errors.New("هیچ تایمر فعالی وجود ندارد")
	}
	if current.IsPaused {
		return &current, nil
	}

	elapsedMs := time.Since(current.LastResumedAt).Milliseconds()
	if current.AccumulatedMs == 0 && current.AccumulatedSeconds > 0 {
		current.AccumulatedMs = current.AccumulatedSeconds * 1000
	}
	current.AccumulatedMs += elapsedMs
	current.AccumulatedSeconds = current.AccumulatedMs / 1000
	current.IsPaused = true

	if err := w.store.SaveActiveTimer(current); err != nil {
		return nil, err
	}

	w.notify("pause", current)
	return &current, nil
}

// Resume unpauses a paused timer.
func (w *WorktimeService) Resume() (*models.ActiveTimer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	current := w.store.GetActiveTimer()
	if !current.IsRunning {
		return nil, errors.New("هیچ تایمر فعالی وجود ندارد")
	}
	if !current.IsPaused {
		return &current, nil
	}

	current.IsPaused = false
	current.LastResumedAt = time.Now()

	if err := w.store.SaveActiveTimer(current); err != nil {
		return nil, err
	}

	w.notify("resume", current)
	return &current, nil
}

// Stop finishes the current timer and saves it as a WorkSession.
// If the session crossed midnight, it splits it across calendar days.
func (w *WorktimeService) Stop() (*models.WorkSession, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	current := w.store.GetActiveTimer()
	if !current.IsRunning {
		return nil, errors.New("هیچ تایمر فعالی برای پایان دادن وجود ندارد")
	}

	totalSeconds := current.ElapsedSeconds()
	now := time.Now()

	var lastSession *models.WorkSession
	// If at least 5 seconds were tracked, save session(s)
	if totalSeconds >= 5 {
		startDay := calendar.FromTime(current.StartedAt).FormatISO()
		endDay := calendar.FromTime(now).FormatISO()

		if startDay == endDay {
			sess := models.WorkSession{
				Project:         current.Project,
				TaskID:          current.TaskID,
				Description:     current.Description,
				JalaliDate:      startDay,
				StartTime:       current.StartedAt,
				EndTime:         now,
				DurationSeconds: totalSeconds,
				IsManual:        false,
				Color:           current.Color,
			}
			if err := w.store.SaveSession(sess); err != nil {
				return nil, err
			}
			lastSession = &sess
		} else {
			// Crosses midnight: split session into daily intervals
			totalWallSec := int64(now.Sub(current.StartedAt).Seconds())
			if totalWallSec <= 0 {
				totalWallSec = totalSeconds
			}

			curStart := current.StartedAt
			remainingTracked := totalSeconds

			for curStart.Before(now) {
				nextMidnight := time.Date(curStart.Year(), curStart.Month(), curStart.Day(), 0, 0, 0, 0, curStart.Location()).AddDate(0, 0, 1)
				sliceEnd := now
				if nextMidnight.Before(now) {
					sliceEnd = nextMidnight
				}

				sliceWallSec := int64(sliceEnd.Sub(curStart).Seconds())
				sliceTrackedSec := int64(0)
				if nextMidnight.Before(now) {
					sliceTrackedSec = sliceWallSec * totalSeconds / totalWallSec
					if sliceTrackedSec > remainingTracked {
						sliceTrackedSec = remainingTracked
					}
					remainingTracked -= sliceTrackedSec
				} else {
					sliceTrackedSec = remainingTracked
				}

				if sliceTrackedSec > 0 {
					jDate := calendar.FromTime(curStart).FormatISO()
					sess := models.WorkSession{
						Project:         current.Project,
						TaskID:          current.TaskID,
						Description:     current.Description,
						JalaliDate:      jDate,
						StartTime:       curStart,
						EndTime:         sliceEnd,
						DurationSeconds: sliceTrackedSec,
						IsManual:        false,
						Color:           current.Color,
					}
					if err := w.store.SaveSession(sess); err != nil {
						return nil, err
					}
					lastSession = &sess
				}

				curStart = sliceEnd
			}
		}
	}

	cleared := models.ActiveTimer{IsRunning: false}
	_ = w.store.SaveActiveTimer(cleared)

	w.notify("stop", cleared)
	return lastSession, nil
}

// GetStatus returns the active timer and current elapsed seconds.
func (w *WorktimeService) GetStatus() (models.ActiveTimer, int64) {
	timer := w.store.GetActiveTimer()
	return timer, timer.ElapsedSeconds()
}

// ProjectStats holds aggregated work hours per project.
type ProjectStats struct {
	Project         string `json:"project"`
	Color           string `json:"color"`
	DurationSeconds int64  `json:"duration_seconds"`
}

// DailySummary holds total worktime and breakdown for a given date.
type DailySummary struct {
	JalaliDate   string               `json:"jalali_date"`
	TotalSeconds int64                `json:"total_seconds"`
	SessionCount int                  `json:"session_count"`
	Breakdown    []ProjectStats       `json:"breakdown"`
	Sessions     []models.WorkSession `json:"sessions"`
}

// GetDailySummary computes worktime statistics for a specific Persian date.
func (w *WorktimeService) GetDailySummary(jalaliDate string) DailySummary {
	sessions := w.store.GetSessions(jalaliDate)
	var total int64
	projMap := make(map[string]*ProjectStats)

	for _, s := range sessions {
		total += s.DurationSeconds
		if stat, ok := projMap[s.Project]; ok {
			stat.DurationSeconds += s.DurationSeconds
		} else {
			projMap[s.Project] = &ProjectStats{
				Project:         s.Project,
				Color:           s.Color,
				DurationSeconds: s.DurationSeconds,
			}
		}
	}

	breakdown := make([]ProjectStats, 0, len(projMap))
	for _, p := range projMap {
		breakdown = append(breakdown, *p)
	}

	return DailySummary{
		JalaliDate:   jalaliDate,
		TotalSeconds: total,
		SessionCount: len(sessions),
		Breakdown:    breakdown,
		Sessions:     sessions,
	}
}
