//go:build windows

package tray

type windowsTray struct {
	title      string
	onActivate func()
}

// NewTray initializes a Windows tray controller.
func NewTray(initialTitle string) (TrayController, error) {
	return &windowsTray{
		title: initialTitle,
	}, nil
}

func (w *windowsTray) SetTitle(title string) {}

func (w *windowsTray) SetToolTip(title, description string) {}

func (w *windowsTray) SetOnActivate(fn func()) {
	w.onActivate = fn
}

func (w *windowsTray) SetOnToggleTimer(fn func()) {}

func (w *windowsTray) SetOnShowFloating(fn func()) {}

func (w *windowsTray) Close() {}
