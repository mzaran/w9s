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
