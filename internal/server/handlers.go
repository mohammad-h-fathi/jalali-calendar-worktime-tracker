package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
	"persian_cal/internal/storage"
	"persian_cal/internal/todo"
	"persian_cal/internal/tracker"
	"persian_cal/internal/ui"
)

type APIHandler struct {
	store            *storage.Store
	tracker          *tracker.WorktimeService
	todo             *todo.TaskService
	sseHub           *SSEHub
	floatingLauncher *ui.FloatingPillLauncher
}

func NewAPIHandler(store *storage.Store, tracker *tracker.WorktimeService, todo *todo.TaskService, sseHub *SSEHub, floatingLauncher *ui.FloatingPillLauncher) *APIHandler {
	return &APIHandler{
		store:            store,
		tracker:          tracker,
		todo:             todo,
		sseHub:           sseHub,
		floatingLauncher: floatingLauncher,
	}
}

func (h *APIHandler) handleToday(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	jDate := calendar.FromTime(now)
	isHoliday, holidayTitle := calendar.IsHoliday(jDate)
	events := calendar.GetEventsForDate(jDate.Year, jDate.Month, jDate.Day)

	timer, elapsed := h.tracker.GetStatus()
	dailySummary := h.tracker.GetDailySummary(jDate.FormatISO())
	incomplete := false
	todayTasks := h.todo.GetTasks(jDate.FormatISO(), &incomplete)

	resp := map[string]interface{}{
		"year":            jDate.Year,
		"month":           jDate.Month,
		"day":             jDate.Day,
		"weekday":         jDate.Weekday,
		"weekday_name":    jDate.WeekdayName(),
		"month_name":      jDate.MonthName(),
		"formatted":       jDate.Format(),
		"iso":             jDate.FormatISO(),
		"is_holiday":      isHoliday,
		"holiday_title":   holidayTitle,
		"events":          events,
		"active_timer":    timer,
		"elapsed_seconds": elapsed,
		"today_work_seconds": dailySummary.TotalSeconds,
		"today_tasks_count":  len(todayTasks),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type CalendarDayResponse struct {
	Year        int              `json:"year"`
	Month       int              `json:"month"`
	Day         int              `json:"day"`
	Weekday     int              `json:"weekday"`
	WeekdayName string           `json:"weekday_name"`
	IsToday     bool             `json:"is_today"`
	IsHoliday   bool             `json:"is_holiday"`
	Events      []calendar.Event `json:"events"`
	TaskCount   int              `json:"task_count"`
	WorkSeconds int64            `json:"work_seconds"`
}

func (h *APIHandler) handleMonthView(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	today := calendar.FromTime(time.Now())

	year := today.Year
	month := today.Month

	if y, err := strconv.Atoi(q.Get("year")); err == nil && y > 1300 && y < 1500 {
		year = y
	}
	if m, err := strconv.Atoi(q.Get("month")); err == nil && m >= 1 && m <= 12 {
		month = m
	}

	daysCount := calendar.DaysInMonth(year, month)
	taskCounts := h.todo.MonthTaskCounts(year, month)

	// Pre-fetch sessions for this month to compute work seconds per day
	prefix := fmt.Sprintf("%04d-%02d-", year, month)
	allSessions := h.store.GetSessions("")
	monthWorkMap := make(map[int]int64)
	for _, s := range allSessions {
		if strings.HasPrefix(s.JalaliDate, prefix) {
			dayStr := strings.TrimPrefix(s.JalaliDate, prefix)
			if d, err := strconv.Atoi(dayStr); err == nil {
				monthWorkMap[d] += s.DurationSeconds
			}
		}
	}

	// Calculate first day's weekday
	firstDayTime := (calendar.JalaliDate{Year: year, Month: month, Day: 1}).ToTime(time.Local)
	firstDayJalali := calendar.FromTime(firstDayTime)
	firstWeekday := firstDayJalali.Weekday

	days := make([]CalendarDayResponse, 0, daysCount)
	for d := 1; d <= daysCount; d++ {
		jDate := calendar.JalaliDate{
			Year:    year,
			Month:   month,
			Day:     d,
			Weekday: (firstWeekday + d - 1) % 7,
		}
		isHoliday, _ := calendar.IsHoliday(jDate)
		events := calendar.GetEventsForDate(year, month, d)
		isToday := (year == today.Year && month == today.Month && d == today.Day)

		days = append(days, CalendarDayResponse{
			Year:        year,
			Month:       month,
			Day:         d,
			Weekday:     jDate.Weekday,
			WeekdayName: jDate.WeekdayName(),
			IsToday:     isToday,
			IsHoliday:   isHoliday,
			Events:      events,
			TaskCount:   taskCounts[d],
			WorkSeconds: monthWorkMap[d],
		})
	}

	resp := map[string]interface{}{
		"year":          year,
		"month":         month,
		"month_name":    calendar.MonthNames[month-1],
		"days_count":    daysCount,
		"first_weekday": firstWeekday,
		"days":          days,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		date := r.URL.Query().Get("date")
		var comp *bool
		if c := r.URL.Query().Get("completed"); c != "" {
			val := (c == "true" || c == "1")
			comp = &val
		}
		tasks := h.todo.GetTasks(date, comp)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tasks)

	case http.MethodPost:
		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := h.todo.CreateTask(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.sseHub.Broadcast("task_update", created)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(created)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handleTaskItem(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/tasks/{id}/toggle or DELETE /api/tasks/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if len(parts) >= 2 && parts[1] == "toggle" && r.Method == http.MethodPost {
		updated, err := h.todo.ToggleTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.sseHub.Broadcast("task_update", updated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
		return
	}

	if r.Method == http.MethodDelete {
		if err := h.todo.DeleteTask(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.sseHub.Broadcast("task_deleted", id)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *APIHandler) handleTimer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		timer, elapsed := h.tracker.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"timer":           timer,
			"elapsed_seconds": elapsed,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handleTimerStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Project     string  `json:"project"`
		Description string  `json:"description"`
		Color       string  `json:"color"`
		TaskID      *string `json:"task_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	timer, err := h.tracker.Start(req.Project, req.Description, req.Color, req.TaskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(timer)
}

func (h *APIHandler) handleTimerPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	timer, err := h.tracker.Pause()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(timer)
}

func (h *APIHandler) handleTimerResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	timer, err := h.tracker.Resume()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(timer)
}

func (h *APIHandler) handleTimerStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, err := h.tracker.Stop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(session)
}

func (h *APIHandler) handleManualWorktime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var sess models.WorkSession
	if err := json.NewDecoder(r.Body).Decode(&sess); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess.IsManual = true
	if sess.DurationSeconds <= 0 && !sess.StartTime.IsZero() && !sess.EndTime.IsZero() {
		sess.DurationSeconds = int64(sess.EndTime.Sub(sess.StartTime).Seconds())
	}
	if err := h.store.SaveSession(sess); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.sseHub.Broadcast("session_added", sess)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sess)
}

func (h *APIHandler) handleDailyWorktime(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = calendar.FromTime(time.Now()).FormatISO()
	}
	summary := h.tracker.GetDailySummary(date)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *APIHandler) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projects := h.store.GetProjects()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(projects)
	case http.MethodPost:
		var p models.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if p.Name == "" {
			http.Error(w, "نام پروژه نمی‌تواند خالی باشد", http.StatusBadRequest)
			return
		}
		if p.Color == "" {
			p.Color = "#3b82f6"
		}
		if err := h.store.SaveProject(p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handleFloatingOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.floatingLauncher != nil {
		if err := h.floatingLauncher.Open(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *APIHandler) handleFloatingClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.floatingLauncher != nil {
		_ = h.floatingLauncher.Close()
	}
	w.WriteHeader(http.StatusOK)
}

