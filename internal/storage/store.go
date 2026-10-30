package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
)

// DataPayload represents the serializable root data structure.
type DataPayload struct {
	Tasks       []models.Task        `json:"tasks"`
	Sessions    []models.WorkSession `json:"sessions"`
	Projects    []models.Project     `json:"projects"`
	ActiveTimer models.ActiveTimer   `json:"active_timer"`
}

// Store coordinates persistent access to application data.
type Store struct {
	mu       sync.RWMutex
	filePath string
	data     DataPayload
}

// GetDataDir resolves the cross-platform user data directory.
func GetDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("HOME")
		if configDir == "" {
			configDir = "."
		}
	}
	dir := filepath.Join(configDir, "persian-calendar")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// NewStore initializes or loads the storage system.
func NewStore(customPath ...string) (*Store, error) {
	var path string
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		path = filepath.Join(GetDataDir(), "data.json")
	}

	s := &Store{
		filePath: path,
		data: DataPayload{
			Tasks:    make([]models.Task, 0),
			Sessions: make([]models.WorkSession, 0),
			Projects: []models.Project{
				{ID: "proj-1", Name: "پروژه عمومی", Color: "#3b82f6", CreatedAt: time.Now()},
				{ID: "proj-2", Name: "توسعه نرم‌افزار", Color: "#10b981", CreatedAt: time.Now()},
				{ID: "proj-3", Name: "جلسات و هماهنگی", Color: "#f59e0b", CreatedAt: time.Now()},
				{ID: "proj-4", Name: "مطالعه و پژوهش", Color: "#8b5cf6", CreatedAt: time.Now()},
			},
		},
	}

	if err := s.load(); err != nil {
		// If file doesn't exist, save defaults
		if os.IsNotExist(err) {
			_ = s.save()
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var payload DataPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		log.Printf("Warning: Corrupt data file %s: %v", s.filePath, err)
		// Check for .bak file
		bakFile := s.filePath + ".bak"
		if bakBytes, bakErr := os.ReadFile(bakFile); bakErr == nil {
			var bakPayload DataPayload
			if bakUnmarshalErr := json.Unmarshal(bakBytes, &bakPayload); bakUnmarshalErr == nil {
				log.Printf("Successfully restored store data from backup %s", bakFile)
				s.data = bakPayload
				corruptFile := fmt.Sprintf("%s.corrupt-%d", s.filePath, time.Now().Unix())
				_ = os.Rename(s.filePath, corruptFile)
				return nil
			}
		}

		// Quarantine corrupt file
		corruptFile := fmt.Sprintf("%s.corrupt-%d", s.filePath, time.Now().Unix())
		_ = os.Rename(s.filePath, corruptFile)
		log.Printf("Moved corrupt data file to %s, starting with defaults", corruptFile)

		s.data = DataPayload{
			Projects: []models.Project{
				{ID: "proj-1", Name: "پروژه عمومی", Color: "#3b82f6", CreatedAt: time.Now()},
				{ID: "proj-2", Name: "توسعه نرم‌افزار", Color: "#10b981", CreatedAt: time.Now()},
				{ID: "proj-3", Name: "جلسات و هماهنگی", Color: "#f59e0b", CreatedAt: time.Now()},
				{ID: "proj-4", Name: "مطالعه و پژوهش", Color: "#8b5cf6", CreatedAt: time.Now()},
			},
			Tasks:    make([]models.Task, 0),
			Sessions: make([]models.WorkSession, 0),
		}
		_ = s.saveLocked()
		return nil
	}

	s.data = payload
	if s.data.Projects == nil {
		s.data.Projects = make([]models.Project, 0)
	}
	if s.data.Tasks == nil {
		s.data.Tasks = make([]models.Task, 0)
	}
	if s.data.Sessions == nil {
		s.data.Sessions = make([]models.WorkSession, 0)
	}

	return nil
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Maintain a backup of the current valid file before overwriting
	bakFile := s.filePath + ".bak"
	if _, err := os.Stat(s.filePath); err == nil {
		_ = os.Rename(s.filePath, bakFile)
	}

	return os.Rename(tmpFile, s.filePath)
}

func (s *Store) save() error {
	return s.saveLocked()
}

// --- Task Operations ---

func priorityRank(p models.TaskPriority) int {
	switch p {
	case models.PriorityUrgent:
		return 4
	case models.PriorityHigh:
		return 3
	case models.PriorityMedium:
		return 2
	case models.PriorityLow:
		return 1
	default:
		return 0
	}
}

func (s *Store) GetTasks(filterDate string, completed *bool) []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Task
	for _, t := range s.data.Tasks {
		if filterDate != "" && t.JalaliDate != filterDate {
			continue
		}
		if completed != nil && t.Completed != *completed {
			continue
		}
		result = append(result, t)
	}

	// Sort: incomplete first, then by priority (urgent -> high -> medium -> low), then by creation
	sort.Slice(result, func(i, j int) bool {
		if result[i].Completed != result[j].Completed {
			return !result[i].Completed
		}
		pI := priorityRank(result[i].Priority)
		pJ := priorityRank(result[j].Priority)
		if pI != pJ {
			return pI > pJ
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func (s *Store) SaveTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", now.UnixNano())
		task.CreatedAt = now
	}
	task.UpdatedAt = now

	found := false
	for i, t := range s.data.Tasks {
		if t.ID == task.ID {
			s.data.Tasks[i] = task
			found = true
			break
		}
	}
	if !found {
		s.data.Tasks = append(s.data.Tasks, task)
	}

	return s.save()
}

func (s *Store) ToggleTask(id string) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.data.Tasks {
		if t.ID == id {
			s.data.Tasks[i].Completed = !s.data.Tasks[i].Completed
			now := time.Now()
			if s.data.Tasks[i].Completed {
				s.data.Tasks[i].CompletedAt = &now
			} else {
				s.data.Tasks[i].CompletedAt = nil
			}
			s.data.Tasks[i].UpdatedAt = now
			updated := s.data.Tasks[i]
			return &updated, s.save()
		}
	}
	return nil, fmt.Errorf("task not found: %s", id)
}

func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]models.Task, 0, len(s.data.Tasks))
	for _, t := range s.data.Tasks {
		if t.ID != id {
			filtered = append(filtered, t)
		}
	}
	s.data.Tasks = filtered
	return s.save()
}

// --- WorkSession Operations ---

func (s *Store) GetSessions(filterDate string) []models.WorkSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.WorkSession
	for _, sess := range s.data.Sessions {
		if filterDate != "" && sess.JalaliDate != filterDate {
			continue
		}
		result = append(result, sess)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.After(result[j].StartTime)
	})

	return result
}

func (s *Store) SaveSession(sess models.WorkSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if sess.ID == "" {
		sess.ID = fmt.Sprintf("sess-%d", now.UnixNano())
		sess.CreatedAt = now
	}
	if sess.JalaliDate == "" {
		sess.JalaliDate = calendar.FromTime(sess.StartTime).FormatISO()
	}

	found := false
	for i, item := range s.data.Sessions {
		if item.ID == sess.ID {
			s.data.Sessions[i] = sess
			found = true
			break
		}
	}
	if !found {
		s.data.Sessions = append(s.data.Sessions, sess)
	}

	return s.save()
}

func (s *Store) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]models.WorkSession, 0, len(s.data.Sessions))
	for _, item := range s.data.Sessions {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}
	s.data.Sessions = filtered
	return s.save()
}

// --- Project Operations ---

func (s *Store) GetProjects() []models.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copied := make([]models.Project, len(s.data.Projects))
	copy(copied, s.data.Projects)
	return copied
}

func (s *Store) SaveProject(proj models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if proj.ID == "" {
		proj.ID = fmt.Sprintf("proj-%d", time.Now().UnixNano())
		proj.CreatedAt = time.Now()
	}

	found := false
	for i, p := range s.data.Projects {
		if p.ID == proj.ID {
			s.data.Projects[i] = proj
			found = true
			break
		}
	}
	if !found {
		s.data.Projects = append(s.data.Projects, proj)
	}

	return s.save()
}

// --- ActiveTimer Operations ---

func (s *Store) GetActiveTimer() models.ActiveTimer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.ActiveTimer
}

func (s *Store) SaveActiveTimer(timer models.ActiveTimer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.ActiveTimer = timer
	return s.save()
}
