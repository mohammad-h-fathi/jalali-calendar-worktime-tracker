//go:build !linux && !windows && !darwin

package tray

type dummyTray struct {
	title      string
	onActivate func()
}

// NewTray initializes a fallback dummy tray controller for other unsupported platforms.
func NewTray(initialTitle string) (TrayController, error) {
	return &dummyTray{
		title: initialTitle,
	}, nil
}

func (d *dummyTray) SetTitle(title string) {}

func (d *dummyTray) SetToolTip(title, description string) {}

func (d *dummyTray) SetOnActivate(fn func()) {
	d.onActivate = fn
}

func (d *dummyTray) SetOnToggleTimer(fn func()) {}

func (d *dummyTray) SetOnShowFloating(fn func()) {}

func (d *dummyTray) SetOnExit(fn func()) {}

func (d *dummyTray) Run() {
	select {}
}

func (d *dummyTray) Close() {}
