package tray

// TrayController defines the cross-platform system tray / top-bar indicator interface.
type TrayController interface {
	// SetTitle updates the dynamic text shown in the top bar / tray.
	SetTitle(title string)
	// SetToolTip updates the tooltip text on hover.
	SetToolTip(title, description string)
	// SetOnActivate registers the callback triggered when the user clicks the indicator.
	SetOnActivate(fn func())
	// SetOnToggleTimer registers the callback triggered when toggling the timer from the menu.
	SetOnToggleTimer(fn func())
	// SetOnShowFloating registers the callback triggered to open the floating pill.
	SetOnShowFloating(fn func())
	// SetOnExit registers the callback triggered when the user clicks Exit in the menu.
	SetOnExit(fn func())
	// Run runs the platform event loop if required, or blocks until the tray is closed.
	Run()
	// Close unregisters the indicator.
	Close()
}
