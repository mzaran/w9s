package app

import "github.com/gdamore/tcell/v2"

// setupKeyboard installs the global input capture handler.
func (a *App) setupKeyboard() {
	a.tviewApp.SetInputCapture(a.inputCapture)
}

// Filterable is implemented by views that support filter mode.
type Filterable interface {
	IsFiltering() bool
}

// inputCapture handles global keyboard shortcuts and delegates
// unhandled keys to the current view.
func (a *App) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	// Let the current view handle the key first.
	if v := a.viewMgr.CurrentView(); v != nil {
		result := v.OnKey(event)
		if result == nil {
			return nil // consumed by view
		}
		// If the view is filtering, skip global shortcuts so keystrokes
		// reach the filter input (e.g., don't let 'q' quit the app).
		if f, ok := v.(Filterable); ok && f.IsFiltering() {
			return event
		}
	}

	switch event.Key() {
	case tcell.KeyCtrlC:
		a.Stop()
		return nil

	case tcell.KeyTab:
		name := a.viewMgr.NextView()
		if name != "" {
			a.switchToView(name)
		}
		return nil

	case tcell.KeyBacktab: // Shift+Tab
		name := a.viewMgr.PreviousView()
		if name != "" {
			a.switchToView(name)
		}
		return nil

	case tcell.KeyRune:
		return a.handleRuneKey(event)
	}

	return event
}

// handleRuneKey handles regular character key presses.
func (a *App) handleRuneKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		a.Stop()
		return nil

	case '?':
		a.switchToView("help")
		return nil

	case 'C':
		a.showClusterSwitcher()
		return nil

	case '/':
		// Filter is handled by individual views via OnKey.
		// Delegate back to the current view.
		return event

	case ':':
		// TODO: focus command input (future)
		return nil

	case '1', '2', '3', '4', '5', '6', '7':
		idx := int(event.Rune() - '1')
		names := a.viewMgr.ViewNames()
		if idx >= 0 && idx < len(names) {
			a.switchToView(names[idx])
		}
		return nil
	}

	return event
}
