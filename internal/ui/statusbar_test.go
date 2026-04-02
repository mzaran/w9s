package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusBar(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	require.NotNil(t, sb)
	assert.NotNil(t, sb.TextView, "underlying TextView should be set")
}

func TestStatusBarSetHints(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.SetHints([]string{"q Quit", "/ Filter", "? Help"})
	})
}

func TestStatusBarSetHintsEmpty(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.SetHints(nil)
	})
}

func TestStatusBarSetHintsSingleWord(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	// Hint without a space should be handled gracefully.
	assert.NotPanics(t, func() {
		sb.SetHints([]string{"nospace"})
	})
}

func TestStatusBarShowError(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.ShowError("something went wrong")
	})
}

func TestStatusBarShowSuccess(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.ShowSuccess("node updated")
	})
}

func TestStatusBarSetApp(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	app := tview.NewApplication()
	assert.NotPanics(t, func() {
		sb.SetApp(app)
	})
}

func TestStatusBarSetCount(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.SetCount("nodes", 42)
	})
}

func TestStatusBarSetMessage(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		sb.SetMessage("Loading...")
	})
}

func TestStatusBarFlashBlocksHints(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	sb.ShowError("error!")
	// While flashing, SetHints should not panic (hints are stored but not rendered).
	assert.NotPanics(t, func() {
		sb.SetHints([]string{"q Quit"})
	})
}

func TestStatusBarShowSpinner(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	sb.ShowSpinner("Building image...")
	assert.True(t, sb.IsSpinnerActive(), "spinner should be active after ShowSpinner")
	sb.HideSpinner()
	assert.False(t, sb.IsSpinnerActive(), "spinner should be inactive after HideSpinner")
}

func TestStatusBarHideSpinnerIdempotent(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	// HideSpinner when no spinner is active should not panic.
	assert.NotPanics(t, func() {
		sb.HideSpinner()
	})
}

func TestStatusBarShowErrorAutoHidesSpinner(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	sb.ShowSpinner("Building...")
	assert.True(t, sb.IsSpinnerActive())
	sb.ShowError("build failed")
	assert.False(t, sb.IsSpinnerActive(), "ShowError should auto-hide spinner")
}

func TestStatusBarShowSuccessAutoHidesSpinner(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	sb.ShowSpinner("Building...")
	assert.True(t, sb.IsSpinnerActive())
	sb.ShowSuccess("build completed")
	assert.False(t, sb.IsSpinnerActive(), "ShowSuccess should auto-hide spinner")
}

func TestStatusBarSpinnerRestoresHints(t *testing.T) {
	sb := ui.NewStatusBar(&ui.DefaultTheme)
	hints := []string{"q Quit", "/ Filter"}
	sb.SetHints(hints)
	sb.ShowSpinner("Loading...")
	sb.HideSpinner()
	// After HideSpinner, the bar should not panic and hints should be restored.
	// We can't easily check rendered text, but we verify no panic.
	assert.False(t, sb.IsSpinnerActive())
}
