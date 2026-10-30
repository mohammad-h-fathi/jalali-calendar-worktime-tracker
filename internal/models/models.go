package models

import "time"

// TaskPriority represents priority levels.
type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
	PriorityUrgent TaskPriority = "urgent"
)

// Task represents a to-do item tied to a Persian date or general backlog.
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	JalaliDate  string       `json:"jalali_date"` // YYYY-MM-DD in Solar Hijri
	Priority    TaskPriority `json:"priority"`
	Category    string       `json:"category"`
	Color       string       `json:"color"`
	Completed   bool         `json:"completed"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// WorkSession represents a completed or logged worktime record.
type WorkSession struct {
	ID              string    `json:"id"`
	Project         string    `json:"project"`
	TaskID          *string   `json:"task_id,omitempty"`
	Description     string    `json:"description"`
	JalaliDate      string    `json:"jalali_date"` // YYYY-MM-DD
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds int64     `json:"duration_seconds"`
	IsManual        bool      `json:"is_manual"`
	Color           string    `json:"color"`
	CreatedAt       time.Time `json:"created_at"`
}

// Project represents a workspace or category for worktime tracking.
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// ActiveTimer represents the current live punch clock state.
type ActiveTimer struct {
	IsRunning          bool      `json:"is_running"`
	IsPaused           bool      `json:"is_paused"`
	Project            string    `json:"project"`
	TaskID             *string   `json:"task_id,omitempty"`
	Description        string    `json:"description"`
	Color              string    `json:"color"`
	StartedAt          time.Time `json:"started_at"`
	LastResumedAt      time.Time `json:"last_resumed_at"`
	AccumulatedSeconds int64     `json:"accumulated_seconds"`
	AccumulatedMs      int64     `json:"accumulated_ms,omitempty"`
}

// ElapsedSeconds returns total elapsed time including current running interval.
func (t *ActiveTimer) ElapsedSeconds() int64 {
	if !t.IsRunning {
		return 0
	}
	accMs := t.AccumulatedMs
	if accMs == 0 && t.AccumulatedSeconds > 0 {
		accMs = t.AccumulatedSeconds * 1000
	}
	if t.IsPaused {
		return accMs / 1000
	}
	currentIntervalMs := time.Since(t.LastResumedAt).Milliseconds()
	if currentIntervalMs < 0 {
		currentIntervalMs = 0
	}
	return (accMs + currentIntervalMs) / 1000
}
