//go:build darwin

package tray

import (
	"fmt"
	"log"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

// Selectors used by the macOS Darwin status bar implementation.
var (
	selAlloc                      = objc.RegisterName("alloc")
	selInit                       = objc.RegisterName("init")
	selSharedApplication          = objc.RegisterName("sharedApplication")
	selSetActivationPolicy        = objc.RegisterName("setActivationPolicy:")
	selFinishLaunching            = objc.RegisterName("finishLaunching")
	selRun                        = objc.RegisterName("run")
	selStop                       = objc.RegisterName("stop:")
	selSystemStatusBar            = objc.RegisterName("systemStatusBar")
	selStatusItemWithLen          = objc.RegisterName("statusItemWithLength:")
	selRemoveStatusItem           = objc.RegisterName("removeStatusItem:")
	selButton                     = objc.RegisterName("button")
	selSetTitle                   = objc.RegisterName("setTitle:")
	selSetToolTip                 = objc.RegisterName("setToolTip:")
	selSetMenu                    = objc.RegisterName("setMenu:")
	selInitWithTitle              = objc.RegisterName("initWithTitle:")
	selAddItem                    = objc.RegisterName("addItem:")
	selSeparatorItem              = objc.RegisterName("separatorItem")
	selInitWithTitleActionKeyEquiv = objc.RegisterName("initWithTitle:action:keyEquivalent:")
	selSetTarget                  = objc.RegisterName("setTarget:")
	selSetAction                  = objc.RegisterName("setAction:")
	selSetEnabled                 = objc.RegisterName("setEnabled:")
	selPerformSelectorOnMainThread = objc.RegisterName("performSelectorOnMainThread:withObject:waitUntilDone:")
	selDrainUpdates               = objc.RegisterName("drainUpdates:")
	selStopApp                    = objc.RegisterName("stopApp:")
	selOpenDashboard              = objc.RegisterName("onOpenDashboard:")
	selToggleTimer                = objc.RegisterName("onToggleTimer:")
	selShowFloating               = objc.RegisterName("onShowFloating:")
	selQuit                       = objc.RegisterName("onQuit:")
)

// Global reference for ObjC target callback routing.
var (
	activeDarwinTray   *darwinTray
	activeDarwinTrayMu sync.RWMutex
)

type darwinTray struct {
	mu             sync.RWMutex
	app            objc.ID
	statusBar      objc.ID
	statusItem     objc.ID
	btn            objc.ID
	menu           objc.ID
	target         objc.ID
	title          string
	tooltip        string
	titleChanged   bool
	tooltipChanged bool
	closed         bool
	done           chan struct{}

	onActivate     func()
	onToggleTimer  func()
	onShowFloating func()
	onExit         func()
}

// getClassID returns the ID for a named Objective-C class.
func getClassID(name string) objc.ID {
	return objc.ID(objc.GetClass(name))
}

// toNSString converts a Go string to a newly allocated NSString.
func toNSString(s string) objc.ID {
	cls := getClassID("NSString")
	if cls == 0 {
		return 0
	}
	b := append([]byte(s), 0)
	return cls.Send(selAlloc).Send(
		objc.RegisterName("initWithUTF8String:"),
		unsafe.Pointer(&b[0]),
	)
}

// registerDarwinTargetClass registers the custom Objective-C class for action callbacks.
func registerDarwinTargetClass() (objc.Class, error) {
	clsName := "PersianCalendarTrayTarget"
	existing := objc.GetClass(clsName)
	if existing != 0 {
		return existing, nil
	}

	superCls := objc.GetClass("NSObject")

	return objc.RegisterClass(
		clsName,
		superCls,
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: selDrainUpdates,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil {
						t.applyUpdatesOnMainThread()
					}
				},
			},
			{
				Cmd: selStopApp,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil && t.app != 0 {
						t.app.Send(selStop, objc.ID(0))
					}
				},
			},
			{
				Cmd: selOpenDashboard,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil {
						t.mu.RLock()
						fn := t.onActivate
						t.mu.RUnlock()
						if fn != nil {
							go fn()
						}
					}
				},
			},
			{
				Cmd: selToggleTimer,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil {
						t.mu.RLock()
						fn := t.onToggleTimer
						t.mu.RUnlock()
						if fn != nil {
							go fn()
						}
					}
				},
			},
			{
				Cmd: selShowFloating,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil {
						t.mu.RLock()
						fn := t.onShowFloating
						t.mu.RUnlock()
						if fn != nil {
							go fn()
						}
					}
				},
			},
			{
				Cmd: selQuit,
				Fn: func(self objc.ID, _cmd objc.SEL, sender objc.ID) {
					activeDarwinTrayMu.RLock()
					t := activeDarwinTray
					activeDarwinTrayMu.RUnlock()
					if t != nil {
						t.mu.RLock()
						fn := t.onExit
						t.mu.RUnlock()
						if fn != nil {
							go fn()
						}
					}
				},
			},
		},
	)
}

// NewTray initializes the native macOS status bar indicator.
func NewTray(initialTitle string) (TrayController, error) {
	appCls := getClassID("NSApplication")
	if appCls == 0 {
		return nil, fmt.Errorf("macOS AppKit NSApplication class not found")
	}

	nsApp := appCls.Send(selSharedApplication)
	// NSApplicationActivationPolicyAccessory = 1 (does not appear in Dock)
	nsApp.Send(selSetActivationPolicy, 1)

	statusBarCls := getClassID("NSStatusBar")
	if statusBarCls == 0 {
		return nil, fmt.Errorf("macOS AppKit NSStatusBar class not found")
	}

	statusBar := statusBarCls.Send(selSystemStatusBar)

	// NSVariableStatusItemLength = -1.0
	statusItem := statusBar.Send(selStatusItemWithLen, float64(-1.0))
	if statusItem == 0 {
		return nil, fmt.Errorf("failed to allocate NSStatusItem on systemStatusBar")
	}

	btn := statusItem.Send(selButton)

	targetCls, err := registerDarwinTargetClass()
	if err != nil {
		log.Printf("Warning: failed to register Darwin target class: %v", err)
	}

	var targetInstance objc.ID
	if targetCls != 0 {
		targetInstance = objc.ID(targetCls).Send(selAlloc).Send(selInit)
	}

	t := &darwinTray{
		app:          nsApp,
		statusBar:    statusBar,
		statusItem:   statusItem,
		btn:          btn,
		target:       targetInstance,
		title:        initialTitle,
		titleChanged: true,
		done:         make(chan struct{}),
	}

	activeDarwinTrayMu.Lock()
	activeDarwinTray = t
	activeDarwinTrayMu.Unlock()

	// Initial title
	if initialTitle != "" && btn != 0 {
		nsTitle := toNSString(initialTitle)
		if nsTitle != 0 {
			btn.Send(selSetTitle, nsTitle)
		}
	}

	// Build context menu
	t.initMenu()

	return t, nil
}

func (t *darwinTray) initMenu() {
	if t.statusItem == 0 {
		return
	}

	menuCls := getClassID("NSMenu")
	itemCls := getClassID("NSMenuItem")
	if menuCls == 0 || itemCls == 0 {
		return
	}

	menu := menuCls.Send(selAlloc).Send(selInitWithTitle, toNSString("PersianCalendarMenu"))
	menu.Send(objc.RegisterName("setAutoenablesItems:"), false)

	// 1. Header (Application Title, disabled)
	hdr := itemCls.Send(selAlloc).Send(
		selInitWithTitleActionKeyEquiv,
		toNSString("گاه‌شمار و وظایف فارسی"),
		objc.SEL(0),
		toNSString(""),
	)
	hdr.Send(selSetEnabled, false)
	menu.Send(selAddItem, hdr)

	// Separator
	menu.Send(selAddItem, itemCls.Send(selSeparatorItem))

	// 2. Open Dashboard & Tasks
	itemDashboard := itemCls.Send(selAlloc).Send(
		selInitWithTitleActionKeyEquiv,
		toNSString("باز کردن گاه‌شمار و وظایف"),
		selOpenDashboard,
		toNSString(""),
	)
	itemDashboard.Send(selSetTarget, t.target)
	menu.Send(selAddItem, itemDashboard)

	// 3. Toggle Timer
	itemToggle := itemCls.Send(selAlloc).Send(
		selInitWithTitleActionKeyEquiv,
		toNSString("شروع / توقف زمان‌سنج"),
		selToggleTimer,
		toNSString(""),
	)
	itemToggle.Send(selSetTarget, t.target)
	menu.Send(selAddItem, itemToggle)

	// 4. Show Floating Pill
	itemPill := itemCls.Send(selAlloc).Send(
		selInitWithTitleActionKeyEquiv,
		toNSString("نمایش تایمر شناور (Pill)"),
		selShowFloating,
		toNSString(""),
	)
	itemPill.Send(selSetTarget, t.target)
	menu.Send(selAddItem, itemPill)

	// Separator
	menu.Send(selAddItem, itemCls.Send(selSeparatorItem))

	// 5. Quit
	itemQuit := itemCls.Send(selAlloc).Send(
		selInitWithTitleActionKeyEquiv,
		toNSString("خروج از برنامه"),
		selQuit,
		toNSString("q"),
	)
	itemQuit.Send(selSetTarget, t.target)
	menu.Send(selAddItem, itemQuit)

	t.statusItem.Send(selSetMenu, menu)
	t.menu = menu
}

// SetTitle updates the status bar text on macOS.
func (t *darwinTray) SetTitle(title string) {
	t.mu.Lock()
	t.title = title
	t.titleChanged = true
	t.mu.Unlock()

	t.requestMainThreadDrain()
}

// SetToolTip updates the tooltip text shown when hovering over the menu bar item.
func (t *darwinTray) SetToolTip(title, description string) {
	t.mu.Lock()
	if description != "" {
		t.tooltip = fmt.Sprintf("%s\n%s", title, description)
	} else {
		t.tooltip = title
	}
	t.tooltipChanged = true
	t.mu.Unlock()

	t.requestMainThreadDrain()
}

func (t *darwinTray) requestMainThreadDrain() {
	t.mu.RLock()
	target := t.target
	t.mu.RUnlock()

	if target == 0 {
		return
	}

	// Safely dispatch to main thread
	target.Send(selPerformSelectorOnMainThread, selDrainUpdates, objc.ID(0), false)
}

func (t *darwinTray) applyUpdatesOnMainThread() {
	t.mu.Lock()
	title := t.title
	titleChanged := t.titleChanged
	t.titleChanged = false

	tooltip := t.tooltip
	tooltipChanged := t.tooltipChanged
	t.tooltipChanged = false
	btn := t.btn
	t.mu.Unlock()

	if btn == 0 {
		return
	}

	if titleChanged && title != "" {
		nsTitle := toNSString(title)
		if nsTitle != 0 {
			btn.Send(selSetTitle, nsTitle)
		}
	}

	if tooltipChanged && tooltip != "" {
		nsTip := toNSString(tooltip)
		if nsTip != 0 {
			btn.Send(selSetToolTip, nsTip)
		}
	}
}

func (t *darwinTray) SetOnActivate(fn func()) {
	t.mu.Lock()
	t.onActivate = fn
	t.mu.Unlock()
}

func (t *darwinTray) SetOnToggleTimer(fn func()) {
	t.mu.Lock()
	t.onToggleTimer = fn
	t.mu.Unlock()
}

func (t *darwinTray) SetOnShowFloating(fn func()) {
	t.mu.Lock()
	t.onShowFloating = fn
	t.mu.Unlock()
}

func (t *darwinTray) SetOnExit(fn func()) {
	t.mu.Lock()
	t.onExit = fn
	t.mu.Unlock()
}

// Run executes the Cocoa AppKit event loop on the main thread.
func (t *darwinTray) Run() {
	if t.app != 0 {
		t.app.Send(selFinishLaunching)
		t.app.Send(selRun)
	} else {
		<-t.done
	}
}

// Close removes the status item and stops the Cocoa event loop.
func (t *darwinTray) Close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	statusBar := t.statusBar
	statusItem := t.statusItem
	target := t.target
	t.mu.Unlock()

	close(t.done)

	if statusBar != 0 && statusItem != 0 {
		statusBar.Send(selRemoveStatusItem, statusItem)
	}

	if target != 0 {
		target.Send(selPerformSelectorOnMainThread, selStopApp, objc.ID(0), false)
	}
}
