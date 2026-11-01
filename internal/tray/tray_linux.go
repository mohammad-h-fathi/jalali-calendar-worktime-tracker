//go:build linux

package tray

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

type linuxTray struct {
	mu             sync.RWMutex
	conn           *dbus.Conn
	propsSpec      *prop.Properties
	menuPropsSpec  *prop.Properties
	onActivate     func()
	onToggleTimer  func()
	onShowFloating func()
	onExit         func()
	title          string
	tooltip        string
	path           dbus.ObjectPath
	menuPath       dbus.ObjectPath
	busName        string
	closed         bool
	done           chan struct{}
}

// ToolTipStruct represents the D-Bus structure for org.kde.StatusNotifierItem.ToolTip.
type ToolTipStruct struct {
	IconName    string
	IconPixmap  []struct{ Width, Height int32; Data []byte }
	Title       string
	Description string
}

// sniObject implements the D-Bus methods for org.kde.StatusNotifierItem.
type sniObject struct {
	tray *linuxTray
}

func (s *sniObject) Activate(x, y int32) *dbus.Error {
	s.tray.mu.RLock()
	cb := s.tray.onActivate
	s.tray.mu.RUnlock()

	if cb != nil {
		go cb()
	}
	return nil
}

func (s *sniObject) ContextMenu(x, y int32) *dbus.Error {
	return s.Activate(x, y)
}

func (s *sniObject) SecondaryActivate(x, y int32) *dbus.Error {
	return s.Activate(x, y)
}

func (s *sniObject) Scroll(delta int32, orientation string) *dbus.Error {
	return nil
}

// DBusMenu Types
type dbusMenuLayout struct {
	ID         int32
	Properties map[string]dbus.Variant
	Children   []dbus.Variant
}

type dbusMenuItemProps struct {
	ID         int32
	Properties map[string]dbus.Variant
}

type menuObject struct {
	tray *linuxTray
}

func (m *menuObject) getItems() map[int32]map[string]dbus.Variant {
	return map[int32]map[string]dbus.Variant{
		0: {
			"children-display": dbus.MakeVariant("submenu"),
		},
		1: {
			"label":   dbus.MakeVariant("📅 باز کردن گاه‌شمار و وظایف"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
			"type":    dbus.MakeVariant("standard"),
		},
		2: {
			"label":   dbus.MakeVariant("⏱ شروع / توقف زمان‌سنج"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
			"type":    dbus.MakeVariant("standard"),
		},
		3: {
			"label":   dbus.MakeVariant("🔘 نمایش تایمر شناور (Pill)"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
			"type":    dbus.MakeVariant("standard"),
		},
		4: {
			"label":   dbus.MakeVariant("❌ خروج از برنامه"),
			"enabled": dbus.MakeVariant(true),
			"visible": dbus.MakeVariant(true),
			"type":    dbus.MakeVariant("standard"),
		},
	}
}

func (m *menuObject) GetLayout(parentId int32, recursionDepth int32, propertyNames []string) (uint32, dbusMenuLayout, *dbus.Error) {
	items := m.getItems()

	children := []dbus.Variant{
		dbus.MakeVariant(dbusMenuLayout{
			ID:         1,
			Properties: items[1],
			Children:   []dbus.Variant{},
		}),
		dbus.MakeVariant(dbusMenuLayout{
			ID:         2,
			Properties: items[2],
			Children:   []dbus.Variant{},
		}),
		dbus.MakeVariant(dbusMenuLayout{
			ID:         3,
			Properties: items[3],
			Children:   []dbus.Variant{},
		}),
		dbus.MakeVariant(dbusMenuLayout{
			ID:         4,
			Properties: items[4],
			Children:   []dbus.Variant{},
		}),
	}

	layout := dbusMenuLayout{
		ID:         0,
		Properties: items[0],
		Children:   children,
	}

	return 1, layout, nil
}

func (m *menuObject) GetGroupProperties(ids []int32, propertyNames []string) ([]dbusMenuItemProps, *dbus.Error) {
	items := m.getItems()
	var result []dbusMenuItemProps

	for _, id := range ids {
		if props, ok := items[id]; ok {
			result = append(result, dbusMenuItemProps{
				ID:         id,
				Properties: props,
			})
		}
	}
	return result, nil
}

func (m *menuObject) GetProperty(id int32, name string) (dbus.Variant, *dbus.Error) {
	items := m.getItems()
	if props, ok := items[id]; ok {
		if val, exists := props[name]; exists {
			return val, nil
		}
	}
	return dbus.MakeVariant(""), nil
}

func (m *menuObject) Event(id int32, eventId string, data dbus.Variant, timestamp uint32) *dbus.Error {
	if eventId == "clicked" {
		switch id {
		case 1:
			m.tray.mu.RLock()
			cb := m.tray.onActivate
			m.tray.mu.RUnlock()
			if cb != nil {
				go cb()
			}
		case 2:
			m.tray.mu.RLock()
			cb := m.tray.onToggleTimer
			m.tray.mu.RUnlock()
			if cb != nil {
				go cb()
			}
		case 3:
			m.tray.mu.RLock()
			cb := m.tray.onShowFloating
			m.tray.mu.RUnlock()
			if cb != nil {
				go cb()
			}
		case 4:
			m.tray.mu.RLock()
			cb := m.tray.onExit
			m.tray.mu.RUnlock()
			if cb != nil {
				go cb()
			} else {
				os.Exit(0)
			}
		}
	}
	return nil
}

func (m *menuObject) EventGroup(events []struct {
	Id        int32
	EventId   string
	Data      dbus.Variant
	Timestamp uint32
}) ([]int32, *dbus.Error) {
	for _, e := range events {
		_ = m.Event(e.Id, e.EventId, e.Data, e.Timestamp)
	}
	return []int32{}, nil
}

func (m *menuObject) AboutToShow(id int32) (bool, *dbus.Error) {
	return false, nil
}

func (m *menuObject) AboutToShowGroup(ids []int32) ([]int32, []int32, *dbus.Error) {
	return []int32{}, []int32{}, nil
}

// NewTray initializes the D-Bus StatusNotifierItem in Ubuntu GNOME top bar.
func NewTray(initialTitle string) (TrayController, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	pid := os.Getpid()
	busName := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", pid)
	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		busName = fmt.Sprintf("org.kde.StatusNotifierItem-%d-%d", pid, os.Getuid())
		_, _ = conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	}

	menuPath := dbus.ObjectPath("/MenuBar")

	t := &linuxTray{
		conn:     conn,
		path:     "/StatusNotifierItem",
		menuPath: menuPath,
		busName:  busName,
		title:    initialTitle,
		tooltip:  "تقویم فارسی و زمان کار",
		done:     make(chan struct{}),
	}

	// 1. Export StatusNotifierItem Methods
	sni := &sniObject{tray: t}
	if err := conn.Export(sni, t.path, "org.kde.StatusNotifierItem"); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to export StatusNotifierItem: %w", err)
	}

	// 2. Export DBusMenu Methods on /MenuBar
	menu := &menuObject{tray: t}
	if err := conn.Export(menu, t.menuPath, "com.canonical.dbusmenu"); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to export com.canonical.dbusmenu: %w", err)
	}

	// 3. Export DBusMenu Properties
	menuPropsMap := map[string]map[string]*prop.Prop{
		"com.canonical.dbusmenu": {
			"Version": {
				Value:    uint32(3),
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"TextDirection": {
				Value:    "rtl",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"Status": {
				Value:    "normal",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"IconThemePath": {
				Value:    []string{},
				Writable: false,
				Emit:     prop.EmitFalse,
			},
		},
	}
	menuProps, err := prop.Export(conn, t.menuPath, menuPropsMap)
	if err == nil {
		t.menuPropsSpec = menuProps
	}

	// 4. Export Introspection for GNOME Shell and GDBus
	sniIntrospect := &introspect.Node{
		Name: string(t.path),
		Interfaces: []introspect.Interface{
			prop.IntrospectData,
			{
				Name:    "org.kde.StatusNotifierItem",
				Methods: introspect.Methods(sni),
			},
		},
	}
	_ = conn.Export(introspect.NewIntrospectable(sniIntrospect), t.path, "org.freedesktop.DBus.Introspectable")

	menuIntrospect := &introspect.Node{
		Name: string(t.menuPath),
		Interfaces: []introspect.Interface{
			prop.IntrospectData,
			{
				Name:    "com.canonical.dbusmenu",
				Methods: introspect.Methods(menu),
			},
		},
	}
	_ = conn.Export(introspect.NewIntrospectable(menuIntrospect), t.menuPath, "org.freedesktop.DBus.Introspectable")

	rootIntrospect := &introspect.Node{
		Name: "/",
		Children: []introspect.Node{
			{Name: "StatusNotifierItem"},
			{Name: "MenuBar"},
		},
	}
	_ = conn.Export(introspect.NewIntrospectable(rootIntrospect), "/", "org.freedesktop.DBus.Introspectable")

	homeDir, _ := os.UserHomeDir()
	iconPath := filepath.Join(homeDir, ".local/share/icons/hicolor/scalable/apps")

	// 5. Export All StatusNotifierItem Properties required by GNOME ubuntu-appindicators
	propsMap := map[string]map[string]*prop.Prop{
		"org.kde.StatusNotifierItem": {
			"Category": {
				Value:    "ApplicationStatus",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Id": {
				Value:    "persian-calendar",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Title": {
				Value:    initialTitle,
				Writable: true,
				Emit:     prop.EmitTrue,
			},
			"Status": {
				Value:    "Active",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"WindowId": {
				Value:    int32(0),
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"IconThemePath": {
				Value:    iconPath,
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"Menu": {
				Value:    t.menuPath,
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"ItemIsMenu": {
				Value:    false,
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"IconName": {
				Value:    "persian-calendar",
				Writable: true,
				Emit:     prop.EmitTrue,
			},
			"IconPixmap": {
				Value:    []struct{ Width, Height int32; Data []byte }{},
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"OverlayIconName": {
				Value:    "",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"OverlayIconPixmap": {
				Value:    []struct{ Width, Height int32; Data []byte }{},
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"AttentionIconName": {
				Value:    "",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"AttentionIconPixmap": {
				Value:    []struct{ Width, Height int32; Data []byte }{},
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"AttentionMovieName": {
				Value:    "",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"ToolTip": {
				Value: ToolTipStruct{
					IconName:    "persian-calendar",
					Title:       initialTitle,
					Description: "تقویم خورشیدی و زمان‌سنج کارکرد",
				},
				Writable: true,
				Emit:     prop.EmitTrue,
			},
			// Ayatana / Ubuntu Indicators extensions for text display on GNOME panel
			"XAyatanaLabel": {
				Value:    initialTitle,
				Writable: true,
				Emit:     prop.EmitTrue,
			},
			"XAyatanaLabelGuide": {
				Value:    "۰۰ آبان ۰۰۰۰",
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			"XAyatanaOrderingIndex": {
				Value:    uint32(0),
				Writable: false,
				Emit:     prop.EmitFalse,
			},
		},
	}

	props, err := prop.Export(conn, t.path, propsMap)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to export properties: %w", err)
	}
	t.propsSpec = props

	// 6. Register with org.kde.StatusNotifierWatcher
	registerWithWatcher := func() {
		watcher := conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
		call := watcher.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, t.busName)
		if call.Err != nil {
			_ = watcher.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, string(t.path))
		}
	}

	registerWithWatcher()

	// Automatic re-registration if GNOME Shell restarts or starts after autostart
	go func() {
		for _, delay := range []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond, 3000 * time.Millisecond} {
			time.Sleep(delay)
			registerWithWatcher()
		}

		_ = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			"type='signal',interface='org.kde.StatusNotifierWatcher',member='StatusNotifierHostRegistered'")
		_ = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			"type='signal',interface='org.freedesktop.DBus',member='NameOwnerChanged',arg0='org.kde.StatusNotifierWatcher'")

		ch := make(chan *dbus.Signal, 10)
		conn.Signal(ch)
		for sig := range ch {
			if sig == nil {
				return
			}
			if sig.Name == "org.kde.StatusNotifierWatcher.StatusNotifierHostRegistered" {
				registerWithWatcher()
			} else if sig.Name == "org.freedesktop.DBus.NameOwnerChanged" && len(sig.Body) >= 3 {
				if newOwner, ok := sig.Body[2].(string); ok && newOwner != "" {
					registerWithWatcher()
				}
			}
		}
	}()

	return t, nil
}

func (t *linuxTray) SetTitle(title string) {
	t.mu.Lock()
	t.title = title
	t.mu.Unlock()

	if t.propsSpec != nil {
		_ = t.propsSpec.Set("org.kde.StatusNotifierItem", "Title", dbus.MakeVariant(title))
		_ = t.propsSpec.Set("org.kde.StatusNotifierItem", "XAyatanaLabel", dbus.MakeVariant(title))
	}
	_ = t.conn.Emit(t.path, "org.kde.StatusNotifierItem.NewTitle")
	_ = t.conn.Emit(t.path, "org.kde.StatusNotifierItem.XAyatanaNewTitle")
	_ = t.conn.Emit(t.path, "org.kde.StatusNotifierItem.XAyatanaNewLabel")
}

func (t *linuxTray) SetToolTip(title, description string) {
	t.mu.Lock()
	t.tooltip = description
	t.mu.Unlock()

	if t.propsSpec != nil {
		val := ToolTipStruct{
			IconName:    "persian-calendar",
			Title:       title,
			Description: description,
		}
		_ = t.propsSpec.Set("org.kde.StatusNotifierItem", "ToolTip", dbus.MakeVariant(val))
	}
	_ = t.conn.Emit(t.path, "org.kde.StatusNotifierItem.NewToolTip")
}

func (t *linuxTray) SetOnActivate(fn func()) {
	t.mu.Lock()
	t.onActivate = fn
	t.mu.Unlock()
}

func (t *linuxTray) SetOnToggleTimer(fn func()) {
	t.mu.Lock()
	t.onToggleTimer = fn
	t.mu.Unlock()
}

func (t *linuxTray) SetOnShowFloating(fn func()) {
	t.mu.Lock()
	t.onShowFloating = fn
	t.mu.Unlock()
}

func (t *linuxTray) SetOnExit(fn func()) {
	t.mu.Lock()
	t.onExit = fn
	t.mu.Unlock()
}

func (t *linuxTray) Run() {
	<-t.done
}

func (t *linuxTray) Close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	t.mu.Unlock()

	close(t.done)
	if t.conn != nil {
		_ = t.conn.Close()
	}
}

