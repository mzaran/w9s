package app

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCommandBar displays the ":" command input at the bottom of the screen,
// temporarily hiding the status bar.
func (a *App) showCommandBar() {
	a.commandActive = true
	a.commandInput.SetText("")

	// Swap visibility: hide status bar, show command input.
	a.mainLayout.ResizeItem(a.statusBar, 0, 0)
	a.mainLayout.ResizeItem(a.commandInput, 1, 0)
	a.tviewApp.SetFocus(a.commandInput)
}

// hideCommandBar hides the command input and restores the status bar.
func (a *App) hideCommandBar() {
	a.commandActive = false
	a.mainLayout.ResizeItem(a.commandInput, 0, 0)
	a.mainLayout.ResizeItem(a.statusBar, 1, 0)
}

// initCommandBar creates the command input field (hidden initially).
func (a *App) initCommandBar() {
	a.commandInput = tview.NewInputField()
	a.commandInput.SetLabel(":")
	a.commandInput.SetFieldWidth(0)
	a.commandInput.SetFieldBackgroundColor(tcell.ColorDefault)
	a.commandInput.SetLabelColor(tcell.ColorWhite)

	a.commandInput.SetDoneFunc(func(key tcell.Key) {
		cmd := strings.TrimSpace(a.commandInput.GetText())
		a.hideCommandBar()

		if key == tcell.KeyEscape || cmd == "" {
			// Cancelled: refocus the current view.
			a.refocusCurrentView()
			return
		}

		a.executeCommand(cmd)
	})
}

// executeCommand dispatches a command entered in the command bar.
func (a *App) executeCommand(cmd string) {
	cmd = strings.ToLower(strings.TrimSpace(cmd))

	viewCommands := map[string]string{
		"dashboard": "dashboard", "1": "dashboard",
		"nodes": "nodes", "2": "nodes",
		"profiles": "profiles", "3": "profiles",
		"images": "images", "4": "images",
		"overlays": "overlays", "5": "overlays",
		"power": "power", "6": "power",
		"help": "help", "7": "help",
	}

	switch cmd {
	case "q", "quit":
		a.Stop()
		return
	}

	if viewName, ok := viewCommands[cmd]; ok {
		a.switchToView(viewName)
		return
	}

	a.statusBar.ShowError("Unknown command: " + cmd)
}

// refocusCurrentView sets focus back to the current view's primitive.
func (a *App) refocusCurrentView() {
	if v := a.viewMgr.CurrentView(); v != nil {
		if prim := v.Render(); prim != nil {
			a.tviewApp.SetFocus(prim)
		}
	}
}
