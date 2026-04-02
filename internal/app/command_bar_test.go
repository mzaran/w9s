package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommandSwitchesToNodes(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")
	require.Equal(t, "dashboard", a.viewMgr.CurrentViewName())

	a.executeCommand("nodes")
	assert.Equal(t, "nodes", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToProfiles(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("profiles")
	assert.Equal(t, "profiles", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToImages(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("images")
	assert.Equal(t, "images", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToOverlays(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("overlays")
	assert.Equal(t, "overlays", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToPower(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("power")
	assert.Equal(t, "power", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToHelp(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("help")
	assert.Equal(t, "help", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandSwitchesToDashboard(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("nodes")

	a.executeCommand("dashboard")
	assert.Equal(t, "dashboard", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandByNumber1(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("nodes")

	a.executeCommand("1")
	assert.Equal(t, "dashboard", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandByNumber2(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("2")
	assert.Equal(t, "nodes", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandByNumber3(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("3")
	assert.Equal(t, "profiles", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandQuit(t *testing.T) {
	a := newTestApp(t)

	a.executeCommand("q")

	select {
	case <-a.ctx.Done():
		// OK - q called Stop which cancels context
	default:
		t.Fatal("context should be cancelled after :q")
	}
}

func TestExecuteCommandQuitFull(t *testing.T) {
	a := newTestApp(t)

	a.executeCommand("quit")

	select {
	case <-a.ctx.Done():
		// OK
	default:
		t.Fatal("context should be cancelled after :quit")
	}
}

func TestExecuteCommandUnknown(t *testing.T) {
	a := newTestApp(t)
	current := a.viewMgr.CurrentViewName()

	// Should not crash and view should not change.
	a.executeCommand("nonexistent")
	assert.Equal(t, current, a.viewMgr.CurrentViewName())
}

func TestExecuteCommandCaseInsensitive(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("NODES")
	assert.Equal(t, "nodes", a.viewMgr.CurrentViewName())
}

func TestExecuteCommandWithWhitespace(t *testing.T) {
	a := newTestApp(t)
	a.switchToView("dashboard")

	a.executeCommand("  nodes  ")
	assert.Equal(t, "nodes", a.viewMgr.CurrentViewName())
}

func TestShowCommandBarSetsActiveState(t *testing.T) {
	a := newTestApp(t)

	assert.False(t, a.commandActive)
	a.showCommandBar()
	assert.True(t, a.commandActive)
}

func TestHideCommandBarClearsActiveState(t *testing.T) {
	a := newTestApp(t)

	a.showCommandBar()
	assert.True(t, a.commandActive)

	a.hideCommandBar()
	assert.False(t, a.commandActive)
}

func TestCommandBarFieldsInitialized(t *testing.T) {
	a := newTestApp(t)

	assert.NotNil(t, a.commandInput, "commandInput should be initialized")
	assert.NotNil(t, a.mainLayout, "mainLayout should be initialized")
}

func TestCommandBarBypassesGlobalShortcuts(t *testing.T) {
	a := newTestApp(t)
	a.showCommandBar()

	// With command bar active, pressing 'q' should NOT stop the app.
	// Instead it should pass through to the input field.
	event := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	result := a.inputCapture(event)
	assert.NotNil(t, result, "'q' should pass through when command bar is active")

	// Context should still be alive.
	select {
	case <-a.ctx.Done():
		t.Fatal("context should NOT be cancelled when command bar is active")
	default:
		// OK
	}
}
