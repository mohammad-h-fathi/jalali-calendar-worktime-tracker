package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	log.Println("Starting Persian Calendar & Worktime Tracker...")

	// 1. Initialize persistent storage
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// 2. Initialize domain services
	workTracker := tracker.NewWorktimeService(store)
	taskTodo := todo.NewTaskService(store)

	// 3. Initialize Top-Bar Tray Indicator
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

	// 4. Initialize Desktop Window Launcher & Floating Pill Launcher
	appURL := fmt.Sprintf("http://127.0.0.1:%d", *port)
	launcher := ui.NewLauncher(appURL)
	floatingLauncher := ui.NewFloatingLauncher(appURL)

	// Automatically open borderless floating pill when timer starts, and close when stops
	workTracker.OnChange(func(event string, timer models.ActiveTimer) {
		if event == "start" {
			_ = floatingLauncher.Open()
		} else if event == "stop" {
			_ = floatingLauncher.Close()
		}
	})

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
			_ = floatingLauncher.Open()
		})
	}

	// 5. Initialize Embedded Server
	srv := server.NewServer(*port, store, workTracker, taskTodo, trayCtrl, floatingLauncher, web.FS)

	// 6. Launch window if requested
	if !*noWindow {
		go func() {
			// Small delay for server to bind
			time.Sleep(200 * time.Millisecond)
			if err := launcher.Open(); err != nil {
				log.Printf("Could not auto-open browser window: %v", err)
			}
		}()
	}

	// 7. Handle OS Interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		_ = srv.Stop()
		if trayCtrl != nil {
			trayCtrl.Close()
		}
		os.Exit(0)
	}()

	log.Printf("Server listening on %s", appURL)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
