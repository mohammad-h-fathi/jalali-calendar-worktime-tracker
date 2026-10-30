package todo

import (
	"fmt"
	"strconv"
	"strings"

	"persian_cal/internal/models"
	"persian_cal/internal/storage"
)

// TaskService coordinates to-do and backlog management.
type TaskService struct {
	store *storage.Store
}

// NewTaskService initializes a new task service.
func NewTaskService(store *storage.Store) *TaskService {
	return &TaskService{store: store}
}

// GetTasks returns filtered tasks.
func (s *TaskService) GetTasks(jalaliDate string, completed *bool) []models.Task {
	return s.store.GetTasks(jalaliDate, completed)
}

// CreateTask saves a new task.
func (s *TaskService) CreateTask(t models.Task) (models.Task, error) {
	if t.Title == "" {
		return t, fmt.Errorf("عنوان وظیفه نمی‌تواند خالی باشد")
	}
	if t.Priority == "" {
		t.Priority = models.PriorityMedium
	}
	if t.Color == "" {
		switch t.Priority {
		case models.PriorityUrgent:
			t.Color = "#ef4444"
		case models.PriorityHigh:
			t.Color = "#f97316"
		case models.PriorityMedium:
			t.Color = "#3b82f6"
		default:
			t.Color = "#10b981"
		}
	}

	err := s.store.SaveTask(t)
	return t, err
}

// ToggleTask toggles completion status.
func (s *TaskService) ToggleTask(id string) (*models.Task, error) {
	return s.store.ToggleTask(id)
}

// DeleteTask removes a task.
func (s *TaskService) DeleteTask(id string) error {
	return s.store.DeleteTask(id)
}

// MonthTaskCounts returns a map of day -> pending task count for a given Jalali year & month.
func (s *TaskService) MonthTaskCounts(jy, jm int) map[int]int {
	prefix := fmt.Sprintf("%04d-%02d-", jy, jm)
	incomplete := false
	tasks := s.store.GetTasks("", &incomplete)

	counts := make(map[int]int)
	for _, t := range tasks {
		if strings.HasPrefix(t.JalaliDate, prefix) {
			dayStr := strings.TrimPrefix(t.JalaliDate, prefix)
			if day, err := strconv.Atoi(dayStr); err == nil {
				counts[day]++
			}
		}
	}
	return counts
}
