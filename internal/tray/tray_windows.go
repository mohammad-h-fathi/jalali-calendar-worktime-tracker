//go:build windows

package tray

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pRegisterClassExW    = user32.NewProc("RegisterClassExW")
	pCreateWindowExW     = user32.NewProc("CreateWindowExW")
	pDefWindowProcW      = user32.NewProc("DefWindowProcW")
	pDestroyWindow       = user32.NewProc("DestroyWindow")
	pPostQuitMessage     = user32.NewProc("PostQuitMessage")
	pGetMessageW         = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessageW    = user32.NewProc("DispatchMessageW")
	pPostMessageW        = user32.NewProc("PostMessageW")
	pCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	pAppendMenuW         = user32.NewProc("AppendMenuW")
	pTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	pGetCursorPos        = user32.NewProc("GetCursorPos")
	pSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	pLoadIconW              = user32.NewProc("LoadIconW")
	pRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")

	pShell_NotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandleW  = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmUser          = 0x0400
	wmTrayCallback  = wmUser + 1
	wmCommand       = 0x0111
	wmLButtonUp     = 0x0202
	wmLButtonDblClk = 0x0203
	wmRButtonUp     = 0x0205
	wmDestroy       = 0x0002

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	idiApplication = 32512
	mfString       = 0x00000000
	mfSeparator    = 0x00000800
	tpmRightButton = 0x0002
	tpmBottomAlign = 0x0020

	cmdOpen     = 1001
	cmdToggle   = 1002
	cmdFloating = 1003
	cmdExit     = 1004
)

type point struct {
	X int32
	Y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uTimeoutOrVersion uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type windowsTray struct {
	mu             sync.RWMutex
	hwnd           uintptr
	nid            notifyIconData
	title          string
	onActivate     func()
	onToggleTimer  func()
	onShowFloating func()
	onExit         func()
	closed         bool
}

var (
	globalTray       *windowsTray
	globalTrayMu     sync.Mutex
	wmTaskbarCreated uint32
)

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	globalTrayMu.Lock()
	t := globalTray
	globalTrayMu.Unlock()

	if wmTaskbarCreated != 0 && msg == wmTaskbarCreated {
		if t != nil {
			t.mu.RLock()
			pShell_NotifyIconW.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&t.nid)))
			t.mu.RUnlock()
		}
		return 0
	}

	switch msg {
	case wmTrayCallback:
		switch lParam {
		case wmLButtonUp, wmLButtonDblClk:
			if t != nil {
				t.mu.RLock()
				cb := t.onActivate
				t.mu.RUnlock()
				if cb != nil {
					go cb()
				}
			}
		case wmRButtonUp:
			if t != nil {
				t.showContextMenu()
			}
		}
		return 0

	case wmCommand:
		cmdId := int(wParam & 0xFFFF)
		if t != nil {
			t.mu.RLock()
			var cb func()
			switch cmdId {
			case cmdOpen:
				cb = t.onActivate
			case cmdToggle:
				cb = t.onToggleTimer
			case cmdFloating:
				cb = t.onShowFloating
			case cmdExit:
				cb = t.onExit
			}
			t.mu.RUnlock()

			if cb != nil {
				go cb()
			}
		}
		return 0

	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := pDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// NewTray initializes a Windows notification area tray icon.
func NewTray(initialTitle string) (TrayController, error) {
	t := &windowsTray{
		title: initialTitle,
	}

	ready := make(chan error, 1)

	go func() {
		runtime.LockOSThread()

		className, _ := syscall.UTF16PtrFromString("PersianCalendarTrayWndClass")
		hInstance, _, _ := pGetModuleHandleW.Call(0)

		wc := wndClassEx{
			cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
			lpfnWndProc:   syscall.NewCallback(wndProc),
			hInstance:     hInstance,
			lpszClassName: className,
		}
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		hwnd, _, _ := pCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(className)),
			0,
			0, 0, 0, 0,
			0, 0, hInstance, 0,
		)

		if hwnd == 0 {
			ready <- fmt.Errorf("failed to create tray window")
			return
		}

		t.hwnd = hwnd

		// Default application icon
		hIcon, _, _ := pLoadIconW.Call(0, uintptr(idiApplication))

		t.nid = notifyIconData{
			cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
			hWnd:             hwnd,
			uID:              1,
			uFlags:           nifMessage | nifIcon | nifTip,
			uCallbackMessage: wmTrayCallback,
			hIcon:            hIcon,
		}

		tipChars, _ := syscall.UTF16FromString(initialTitle)
		copy(t.nid.szTip[:], tipChars)

		pShell_NotifyIconW.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&t.nid)))

		globalTrayMu.Lock()
		globalTray = t
		taskbarMsgName, _ := syscall.UTF16PtrFromString("TaskbarCreated")
		retMsg, _, _ := pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(taskbarMsgName)))
		wmTaskbarCreated = uint32(retMsg)
		globalTrayMu.Unlock()

		ready <- nil

		// Windows message loop
		var m msg
		for {
			ret, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if int32(ret) <= 0 {
				break
			}
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
	}()

	if err := <-ready; err != nil {
		return nil, err
	}

	return t, nil
}

func (w *windowsTray) showContextMenu() {
	hMenu, _, _ := pCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}

	txtOpen, _ := syscall.UTF16PtrFromString("باز کردن تقویم (Open)")
	txtToggle, _ := syscall.UTF16PtrFromString("شروع / توقف تایمر (Toggle Timer)")
	txtFloating, _ := syscall.UTF16PtrFromString("ویجت شناور (Floating Pill)")
	txtExit, _ := syscall.UTF16PtrFromString("خروج (Exit)")

	pAppendMenuW.Call(hMenu, uintptr(mfString), uintptr(cmdOpen), uintptr(unsafe.Pointer(txtOpen)))
	pAppendMenuW.Call(hMenu, uintptr(mfString), uintptr(cmdToggle), uintptr(unsafe.Pointer(txtToggle)))
	pAppendMenuW.Call(hMenu, uintptr(mfString), uintptr(cmdFloating), uintptr(unsafe.Pointer(txtFloating)))
	pAppendMenuW.Call(hMenu, uintptr(mfSeparator), 0, 0)
	pAppendMenuW.Call(hMenu, uintptr(mfString), uintptr(cmdExit), uintptr(unsafe.Pointer(txtExit)))

	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(w.hwnd)
	pTrackPopupMenu.Call(hMenu, uintptr(tpmRightButton|tpmBottomAlign), uintptr(pt.X), uintptr(pt.Y), 0, w.hwnd, 0)
	pPostMessageW.Call(w.hwnd, 0, 0, 0)
}

func (w *windowsTray) SetTitle(title string) {
	w.mu.Lock()
	w.title = title
	w.mu.Unlock()
	w.SetToolTip(title, "")
}

func (w *windowsTray) SetToolTip(title, description string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed || w.hwnd == 0 {
		return
	}

	tip := title
	if description != "" {
		if tip != "" {
			tip += " - " + description
		} else {
			tip = description
		}
	}
	if len(tip) > 120 {
		tip = tip[:120]
	}

	tipChars, _ := syscall.UTF16FromString(tip)
	var newTip [128]uint16
	copy(newTip[:], tipChars)
	w.nid.szTip = newTip

	pShell_NotifyIconW.Call(uintptr(nimModify), uintptr(unsafe.Pointer(&w.nid)))
}

func (w *windowsTray) SetOnActivate(fn func()) {
	w.mu.Lock()
	w.onActivate = fn
	w.mu.Unlock()
}

func (w *windowsTray) SetOnToggleTimer(fn func()) {
	w.mu.Lock()
	w.onToggleTimer = fn
	w.mu.Unlock()
}

func (w *windowsTray) SetOnShowFloating(fn func()) {
	w.mu.Lock()
	w.onShowFloating = fn
	w.mu.Unlock()
}

func (w *windowsTray) SetOnExit(fn func()) {
	w.mu.Lock()
	w.onExit = fn
	w.mu.Unlock()
}

func (w *windowsTray) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return
	}
	w.closed = true

	pShell_NotifyIconW.Call(uintptr(nimDelete), uintptr(unsafe.Pointer(&w.nid)))
	if w.hwnd != 0 {
		pDestroyWindow.Call(w.hwnd)
	}
}
