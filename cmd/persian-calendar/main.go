package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"persian_cal/internal/calendar"
	"persian_cal/internal/models"
	"persian_cal/internal/server"
	"persian_cal/internal/storage"
	"persian_cal/internal/todo"
	"persian_cal/internal/tracker"
	"persian_cal/internal/tray"
	"persian_cal/internal/ui"
	"persian_cal/web"
)

func main() {
	port := flag.Int("port", 41890, "Port for the embedded web application")
	noWindow := flag.Bool("no-window", false, "Start in background top-bar daemon mode without opening window")
	flag.Parse()

	appURL := fmt.Sprintf("http://127.0.0.1:%d", *port)
	launcher := ui.NewLauncher(appURL)

	// 1. Single-instance check: if app is already running, open window for the running instance and exit
	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("%s/api/today", appURL))
	if err == nil && resp.StatusCode == http.StatusOK {
		_ = resp.Body.Close()
		log.Printf("An instance of Persian Calendar is already running on %s", appURL)
		if !*noWindow {
			postResp, postErr := client.Post(fmt.Sprintf("%s/api/window/open", appURL), "application/json", nil)
			if postErr == nil && postResp.StatusCode == http.StatusOK {
				_ = postResp.Body.Close()
			} else {
				if err := launcher.Open(); err != nil {
					log.Printf("Could not auto-open window: %v", err)
				}
			}
		}
		os.Exit(0)
	}

	log.Println("Starting Persian Calendar & Worktime Tracker...")

	// 2. Initialize persistent storage
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// 3. Initialize domain services
	workTracker := tracker.NewWorktimeService(store)
	taskTodo := todo.NewTaskService(store)

	// 4. Initialize Top-Bar Tray Indicator
	now := time.Now()
	jDate := calendar.FromTime(now)
	initialTitle := fmt.Sprintf("%s %s %s", calendar.ToPersianDigits(fmt.Sprintf("%d", jDate.Day)), jDate.MonthName(), calendar.ToPersianDigits(fmt.Sprintf("%d", jDate.Year)))

	trayCtrl, err := tray.NewTray(initialTitle)
	if err != nil {
		log.Printf("Warning: Could not register top-bar indicator: %v", err)
	} else {
		defer trayCtrl.Close()
		log.Printf("Top-bar indicator registered with title: %s", initialTitle)
	}

	// 5. Initialize Desktop Window Launcher & Floating Pill Launcher
	floatingLauncher := ui.NewFloatingLauncher(appURL)

	// Automatically open borderless floating pill when timer starts, and close when stops
	workTracker.OnChange(func(event string, timer models.ActiveTimer) {
		if event == "start" {
			_ = floatingLauncher.Open()
		} else if event == "stop" {
			_ = floatingLauncher.Close()
		}
	})

	var srv *server.Server

	var shutdownOnce sync.Once
	cleanup := func() {
		shutdownOnce.Do(func() {
			log.Println("Shutting down...")
			if floatingLauncher != nil {
				_ = floatingLauncher.Close()
			}
			if srv != nil {
				_ = srv.Stop()
			}
			if trayCtrl != nil {
				trayCtrl.Close()
			}
			os.Exit(0)
		})
	}

	if trayCtrl != nil {
		trayCtrl.SetOnActivate(func() {
			log.Println("Top-bar indicator clicked: focusing/opening dashboard window")
			if err := launcher.Open(); err != nil {
				log.Printf("Error opening window: %v", err)
			}
		})
		trayCtrl.SetOnToggleTimer(func() {
			timer, _ := workTracker.GetStatus()
			if timer.IsRunning {
				if timer.IsPaused {
					_, _ = workTracker.Resume()
				} else {
					_, _ = workTracker.Pause()
				}
			} else {
				_, _ = workTracker.Start("پروژه عمومی", "", "", nil)
			}
		})
		trayCtrl.SetOnShowFloating(func() {
			log.Println("Top-bar indicator menu: opening floating pill")
			if err := floatingLauncher.Open(); err != nil {
				log.Printf("Error opening floating pill: %v", err)
			}
		})
		trayCtrl.SetOnExit(cleanup)
	}

	// 6. Initialize Embedded Server
	srv = server.NewServer(*port, store, workTracker, taskTodo, trayCtrl, floatingLauncher, launcher, web.FS)

	// 7. Launch window if requested
	if !*noWindow {
		go func() {
			// Small delay for server to bind
			time.Sleep(200 * time.Millisecond)
			if err := launcher.Open(); err != nil {
				log.Printf("Could not auto-open browser window: %v", err)
			}
		}()
	}

	// 8. Handle OS Interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
	}()

	log.Printf("Server listening on %s", appURL)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
