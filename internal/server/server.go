package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
	"persian_cal/internal/storage"
	"persian_cal/internal/todo"
	"persian_cal/internal/tracker"
	"persian_cal/internal/tray"
	"persian_cal/internal/ui"
)

// SSEHub handles server-sent events broadcasting to browser windows and PiP pills.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]bool
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]bool),
	}
}

func (h *SSEHub) Register(ch chan []byte) {
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()
}

func (h *SSEHub) Unregister(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *SSEHub) Broadcast(eventType string, data interface{}) {
	b, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		return
	}
	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(b)))

	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// Server holds the web server configuration and lifecycle.
type Server struct {
	httpServer *http.Server
	port       int
	store      *storage.Store
	tracker    *tracker.WorktimeService
	todo       *todo.TaskService
	tray       tray.TrayController
	sseHub     *SSEHub
	stopChan   chan struct{}
}

// FormatSeconds formats seconds into HH:MM:SS with Persian numerals.
func FormatSeconds(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	str := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	return calendar.ToPersianDigits(str)
}

// NewServer builds the HTTP & SSE server.
func NewServer(port int, store *storage.Store, workTracker *tracker.WorktimeService, taskTodo *todo.TaskService, trayCtrl tray.TrayController, floatingLauncher *ui.FloatingPillLauncher, windowLauncher *ui.WindowLauncher, embeddedFS fs.FS) *Server {
	hub := NewSSEHub()
	handler := NewAPIHandler(store, workTracker, taskTodo, hub, floatingLauncher, windowLauncher)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/today", handler.handleToday)
	mux.HandleFunc("/api/calendar/month", handler.handleMonthView)
	mux.HandleFunc("/api/tasks", handler.handleTasks)
	mux.HandleFunc("/api/tasks/", handler.handleTaskItem)
	mux.HandleFunc("/api/timer", handler.handleTimer)
	mux.HandleFunc("/api/timer/start", handler.handleTimerStart)
	mux.HandleFunc("/api/timer/pause", handler.handleTimerPause)
	mux.HandleFunc("/api/timer/resume", handler.handleTimerResume)
	mux.HandleFunc("/api/timer/stop", handler.handleTimerStop)
	mux.HandleFunc("/api/timer/floating/open", handler.handleFloatingOpen)
	mux.HandleFunc("/api/timer/floating/close", handler.handleFloatingClose)
	mux.HandleFunc("/api/window/open", handler.handleWindowOpen)
	mux.HandleFunc("/api/worktime/manual", handler.handleManualWorktime)
	mux.HandleFunc("/api/worktime/daily", handler.handleDailyWorktime)
	mux.HandleFunc("/api/worktime/sessions/", handler.handleWorkSessionItem)
	mux.HandleFunc("/api/projects", handler.handleProjects)

	// SSE event stream
	mux.HandleFunc("/api/events/stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientChan := make(chan []byte, 32)
		hub.Register(clientChan)
		defer hub.Unregister(clientChan)

		// Send initial hello
		_, _ = fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-clientChan:
				_, _ = w.Write(msg)
				flusher.Flush()
			}
		}
	})

	// Static files serving: Prefer embedded FS; only use local ./web when developer explicitly opts in
	var fileServer http.Handler
	if os.Getenv("PERSIAN_CAL_DEV") == "1" {
		if info, err := os.Stat("web"); err == nil && info.IsDir() {
			fileServer = http.FileServer(http.Dir("web"))
		}
	}
	if fileServer == nil && embeddedFS != nil {
		if _, err := fs.Stat(embeddedFS, "index.html"); err == nil {
			fileServer = http.FileServer(http.FS(embeddedFS))
		} else if sub, err := fs.Sub(embeddedFS, "web"); err == nil {
			fileServer = http.FileServer(http.FS(sub))
		} else {
			fileServer = http.FileServer(http.FS(embeddedFS))
		}
	}

	if fileServer != nil {
		mux.Handle("/", fileServer)
	}

	s := &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", port),
			Handler: mux,
		},
		port:     port,
		store:    store,
		tracker:  workTracker,
		todo:     taskTodo,
		tray:     trayCtrl,
		sseHub:   hub,
		stopChan: make(chan struct{}),
	}

	// Register tracker change callback for tray & SSE
	workTracker.OnChange(func(event string, t models.ActiveTimer) {
		s.updateTrayAndBroadcast()
	})

	return s
}

func (s *Server) updateTrayAndBroadcast() {
	timer, elapsed := s.tracker.GetStatus()
	now := time.Now()
	jDate := calendar.FromTime(now)
	baseDate := fmt.Sprintf("%s %s", calendar.ToPersianDigits(fmt.Sprintf("%d", jDate.Day)), jDate.MonthName())

	var trayTitle string
	var tooltip string

	if timer.IsRunning {
		elapsedStr := FormatSeconds(elapsed)
		if timer.IsPaused {
			trayTitle = fmt.Sprintf("%s | ⏸ %s", baseDate, elapsedStr)
			tooltip = fmt.Sprintf("تقویم: %s\nتایمر متوقف شده: %s (%s)", jDate.Format(), elapsedStr, timer.Project)
		} else {
			trayTitle = fmt.Sprintf("%s | ⏱ %s", baseDate, elapsedStr)
			tooltip = fmt.Sprintf("تقویم: %s\nدر حال کار: %s (%s)", jDate.Format(), elapsedStr, timer.Project)
		}
	} else {
		trayTitle = fmt.Sprintf("%s %s", baseDate, calendar.ToPersianDigits(fmt.Sprintf("%d", jDate.Year)))
		tooltip = fmt.Sprintf("تقویم فارسی: %s\nروز: %s", jDate.Format(), jDate.WeekdayName())
	}

	if s.tray != nil {
		s.tray.SetTitle(trayTitle)
		s.tray.SetToolTip(trayTitle, tooltip)
	}

	s.sseHub.Broadcast("timer_update", map[string]interface{}{
		"timer":             timer,
		"elapsed_seconds":   elapsed,
		"formatted_elapsed": FormatSeconds(elapsed),
		"tray_title":        trayTitle,
	})
}

// Start runs the HTTP server and background tick loop.
func (s *Server) Start() error {
	// Background ticker for live seconds, tray update, and midnight day transitions
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastDate := calendar.FromTime(time.Now()).FormatISO()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				currentDate := calendar.FromTime(time.Now()).FormatISO()
				if currentDate != lastDate {
					lastDate = currentDate
					s.updateTrayAndBroadcast()
					s.sseHub.Broadcast("day_changed", map[string]string{"date": currentDate})
				} else {
					timer, _ := s.tracker.GetStatus()
					if timer.IsRunning && !timer.IsPaused {
						s.updateTrayAndBroadcast()
					}
				}
			}
		}
	}()

	// Initial tray setup
	s.updateTrayAndBroadcast()

	return s.httpServer.ListenAndServe()
}

// Stop shuts down the server.
func (s *Server) Stop() error {
	close(s.stopChan)
	return s.httpServer.Close()
}
