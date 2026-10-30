# Jalali Calendar & Worktime Tracker

<div align="center">

[🇬🇧 English](README.md) • [🇮🇷 فارسی](README.fa.md)

[![Release](https://img.shields.io/github/v/release/mohammad-h-fathi/jalali-calendar-worktime-tracker?style=flat-square&color=3b82f6)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases)
[![Build & Release](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/actions/workflows/release.yml/badge.svg)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-blueviolet?style=flat-square)](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

</div>

A native, lightweight, and modern desktop application for **Linux (Ubuntu 24.04 / 26.04 GNOME Wayland & X11)**, **Windows 10 / 11**, and **macOS**, written in **Go (Golang)**.

It displays a live Solar Hijri (Jalali) date and stopwatch indicator in the **Ubuntu GNOME top-bar**, features a sleek **borderless, always-on-top floating action timer pill**, provides daily **task management (To-Do)** tied to Persian calendar dates, and tracks **work hours per project** — rendered with the **Lahzeh (لحظه)** Persian typography.

---

## ✨ Key Features

### 1. GNOME Top-Bar Live Indicator (StatusNotifierItem & DBusMenu)
* **Always-visible date:** Shows today's Jalali date directly in the GNOME status panel (e.g. `۸ آبان ۱۴۰۵`).
* **Live Dynamic Stopwatch:** When a timer is running, elapsed time ticks live right in the top bar:  
  `۸ آبان | ⏱ ۰۰:۲۴:۱۰`
* **Right-Click Menu:** Quick access to open the dashboard, toggle/pause the timer, show the floating pill, or exit.

### 2. Borderless Always-on-Top Floating Action Timer Pill
* **Truly Borderless & Frameless:** No window titlebars, no minimize/maximize/close buttons. Styled as a compact rounded glassmorphic capsule (~270x44px).
* **Always-on-Top:** Floats above all open windows (browsers, IDEs, terminal, full-screen editors).
* **Draggable:** Easily dragged anywhere across monitors with mouse drag.
* **Auto-Removed on Stop:** Pressing **"Stop ⏹"** saves the session and **instantly removes and closes the floating window**.
* **Zero External Dependencies:** Built with native OS runtimes (GTK 3 on Linux, built-in PowerShell/WPF on Windows). Requires no Python installation on Windows!

### 3. Smart Jalali Calendar & Official Events
* Astronomical Solar Hijri leap year algorithm for 100% mathematical accuracy.
* **Live National Events API:** Automatically fetches official holidays and cultural occasions from `iran.ir/Calendar/GetEvents` with in-memory caching and offline fallback.
* Displays daily task badges and total work duration per day on the monthly grid.

### 4. Worktime Punch Clock & Analytics
* Categorize work sessions by project with custom color indicators.
* Live stopwatch with Pause / Resume and focus descriptions.
* Manual logging for past work sessions.
* Daily work summary with percentage breakdown per project.

### 5. Daily Task Management (To-Do List)
* Tasks tied directly to Persian calendar dates.
* Quick toggle and deletion with instant synchronization.
* Badge counter in the navigation bar for pending tasks.

---

## 📥 Downloads & Installation

### 🐧 Ubuntu & Debian (Debian Package `.deb`)

#### Recommended: 1-Click Debian Package
Download the latest `.deb` package from the [Releases](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases) page:
```bash
sudo apt install ./persian-calendar_*.deb
```
> This automatically installs the binary to `/usr/bin/`, sets up the desktop launcher in the GNOME App Grid, installs the app icon, and registers the Lahzeh fonts system-wide.

#### Option 2: Script-Based User Installation
```bash
git clone https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker.git
cd jalali-calendar-worktime-tracker
./scripts/install-desktop.sh
```

---

### 🪟 Windows (Windows 10 / 11)

1. Download `persian-calendar_*_windows_amd64.zip` from [Releases](https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker/releases).
2. Extract the ZIP archive.
3. Right-click **`install-windows.bat`** and click **Run**.
4. Shortcuts will be created on your Desktop and Start Menu, and autostart in the background on login will be enabled.

---

### 🍎 macOS

Download `persian-calendar_*_darwin_arm64.tar.gz` (for Apple Silicon M1–M4) or `amd64` (for Intel):
```bash
tar -xzf persian-calendar_*_darwin_*.tar.gz
./persian-calendar
```

---

## 🛠 Building from Source

Prerequisites: [Go 1.22+](https://go.dev/)

```bash
git clone https://github.com/mohammad-h-fathi/jalali-calendar-worktime-tracker.git
cd jalali-calendar-worktime-tracker

# Build Linux binary:
go build -o bin/persian-calendar ./cmd/persian-calendar

# Cross-compile for Windows (CGO-free):
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/persian-calendar.exe ./cmd/persian-calendar

# Cross-compile for macOS (Apple Silicon):
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/persian-calendar-darwin-arm64 ./cmd/persian-calendar
```

---

## 🚀 CLI Usage & Flags

```bash
# Launch full application (opens dashboard window):
persian-calendar

# Launch in background daemon mode (top-bar indicator only, without opening window):
persian-calendar --no-window

# Specify custom port (default: 41890):
persian-calendar --port=45000
```

---

## 📡 REST API & Realtime SSE Endpoints

The embedded server exposes a fast local HTTP & SSE API:

| Endpoint | Method | Description |
| :--- | :---: | :--- |
| `/api/today` | `GET` | Current Jalali date info, events, active timer state, and task counts |
| `/api/calendar/month?year=1405&month=8` | `GET` | Full monthly days grid with events and work seconds per day |
| `/api/tasks` | `GET / POST` | List tasks or create a new task (with date validation) |
| `/api/tasks/{id}/toggle` | `POST` | Toggle task completion status |
| `/api/tasks/{id}` | `DELETE` | Delete a task |
| `/api/timer` | `GET` | Active stopwatch status and elapsed seconds |
| `/api/timer/start` | `POST` | Start timer for a project |
| `/api/timer/pause` | `POST` | Pause running timer |
| `/api/timer/resume` | `POST` | Resume paused timer |
| `/api/timer/stop` | `POST` | Stop timer and save work session (with midnight crossing split) |
| `/api/timer/floating/open` | `POST` | Launch borderless floating pill widget |
| `/api/timer/floating/close` | `POST` | Close borderless floating pill widget |
| `/api/worktime/daily` | `GET` | Daily worktime summary and project breakdown |
| `/api/worktime/manual` | `POST` | Log a manual work session with date & time range |
| `/api/worktime/sessions/{id}` | `DELETE` | Delete a work session |
| `/api/window/open` | `POST` | Focus or open the dashboard window |
| `/api/events/stream` | `GET` | Real-time Server-Sent Events (SSE) broadcast channel |

---

## 📂 Modular Architecture

```text
jalali-calendar-worktime-tracker/
├── .github/workflows/
│   └── release.yml          # GitHub Actions multi-platform automated release matrix
├── cmd/
│   └── persian-calendar/
│       └── main.go          # Main entrypoint and lifecycle coordinator
├── internal/
│   ├── calendar/            # Astronomical Jalali engine, leap years, and iran.ir API
│   ├── models/              # Data models (Task, WorkSession, Project, ActiveTimer)
│   ├── server/              # HTTP server, REST handlers, and realtime SSE hub
│   ├── storage/             # Thread-safe atomic JSON persistence
│   ├── todo/                # Task management domain service
│   ├── tracker/             # Worktime stopwatch engine and daily analytics
│   ├── tray/                # D-Bus StatusNotifierItem (Linux) and Tray (Windows)
│   └── ui/                  # Native dashboard launcher and floating pill controller
├── scripts/
│   ├── floating_timer.py    # Borderless GTK 3 floating pill widget for Linux
│   ├── floating_timer.ps1   # Borderless WPF floating pill widget for Windows
│   ├── install-desktop.sh   # Linux user desktop installer
│   ├── install-windows.bat  # Windows 1-click installer
│   └── package-deb.sh       # Debian .deb package generator
├── fonts/                   # Full 9-weight Lahzeh (لحظه) Persian TTF font family
└── web/                     # Embedded static assets (HTML5, Vanilla CSS, JS)
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE). Free for personal and commercial use.
