package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"persian_cal/internal/storage"
)

// WindowLauncher manages the desktop app window.
type WindowLauncher struct {
	mu      sync.Mutex
	url     string
	cmd     *exec.Cmd
	running bool
}

// NewLauncher creates a new WindowLauncher for the given local server URL.
func NewLauncher(url string) *WindowLauncher {
	return &WindowLauncher{url: url}
}

// Open launches or focuses the application window in App Mode.
func (l *WindowLauncher) Open() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If already running, on Linux we can bring it to foreground or relaunch
	if l.running && l.cmd != nil && l.cmd.Process != nil {
		// Process is still running
		return nil
	}

	profileDir := filepath.Join(storage.GetDataDir(), "web-profile")
	_ = os.MkdirAll(profileDir, 0755)

	candidates := getBrowserCandidates()
	var browserPath string
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			browserPath = path
			break
		}
	}

	var cmd *exec.Cmd
	if browserPath != "" {
		args := []string{
			fmt.Sprintf("--app=%s", l.url),
			fmt.Sprintf("--user-data-dir=%s", profileDir),
			"--window-size=460,720",
			"--class=persian-calendar",
			"--no-first-run",
			"--no-default-browser-check",
		}
		cmd = exec.Command(browserPath, args...)
	} else {
		// Fallback to system default opener
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "start", l.url)
		} else {
			cmd = exec.Command("xdg-open", l.url)
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	l.cmd = cmd
	l.running = true

	go func() {
		_ = cmd.Wait()
		l.mu.Lock()
		l.running = false
		l.mu.Unlock()
	}()

	return nil
}

func getBrowserCandidates() []string {
	if runtime.GOOS == "windows" {
		return []string{"msedge.exe", "chrome.exe", "brave.exe"}
	}
	return []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium-browser",
		"chromium",
		"brave-browser",
		"microsoft-edge-stable",
	}
}

// FloatingPillLauncher manages the native borderless floating timer pill.
type FloatingPillLauncher struct {
	mu      sync.Mutex
	url     string
	cmd     *exec.Cmd
	running bool
}

// NewFloatingLauncher creates a new FloatingPillLauncher for the given server URL.
func NewFloatingLauncher(url string) *FloatingPillLauncher {
	return &FloatingPillLauncher{url: url}
}

// Open launches the floating pill window if not already active.
func (f *FloatingPillLauncher) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running && f.cmd != nil && f.cmd.Process != nil {
		// Already running
		return nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		scriptPath := findWindowsFloatingScript()
		if scriptPath == "" {
			return fmt.Errorf("windows floating timer script not found")
		}
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath, f.url)
	} else {
		scriptPath := findLinuxFloatingScript()
		if scriptPath == "" {
			return fmt.Errorf("floating timer pill script not found")
		}
		cmd = exec.Command(scriptPath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PERSIAN_CAL_URL=%s", f.url))
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	f.cmd = cmd
	f.running = true

	go func() {
		_ = cmd.Wait()
		f.mu.Lock()
		f.running = false
		f.mu.Unlock()
	}()

	return nil
}

// Close terminates the floating pill window if running.
func (f *FloatingPillLauncher) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running && f.cmd != nil && f.cmd.Process != nil {
		_ = f.cmd.Process.Kill()
		f.running = false
	}
	return nil
}

// IsRunning returns whether the floating pill process is currently active.
func (f *FloatingPillLauncher) IsRunning() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.running
}

func findLinuxFloatingScript() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local/bin/persian-timer-pill"),
		"scripts/floating_timer.py",
		"../scripts/floating_timer.py",
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates,
			filepath.Join(execDir, "scripts/floating_timer.py"),
			filepath.Join(execDir, "../scripts/floating_timer.py"),
			filepath.Join(execDir, "persian-timer-pill"),
		)
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}

	if p, err := exec.LookPath("persian-timer-pill"); err == nil {
		return p
	}
	return ""
}

func findWindowsFloatingScript() string {
	candidates := []string{
		"scripts\\floating_timer.ps1",
		"..\\scripts\\floating_timer.ps1",
	}
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates,
			filepath.Join(execDir, "scripts\\floating_timer.ps1"),
			filepath.Join(execDir, "..\\scripts\\floating_timer.ps1"),
		)
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	return "scripts\\floating_timer.ps1"
}

