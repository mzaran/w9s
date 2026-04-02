package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestShowConfirm(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	confirmed := false
	assert.NotPanics(t, func() {
		ui.ShowConfirm(app, pages, "Delete Node", "Delete node01?", func() {
			confirmed = true
		})
	})

	assert.True(t, pages.HasPage("confirm"), "confirm page should be added")
	_ = confirmed
}

func TestShowConfirmWithOnDone(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	doneCalled := false
	assert.NotPanics(t, func() {
		ui.ShowConfirm(app, pages, "Reboot", "Reboot node01?", func() {}, func() {
			doneCalled = true
		})
	})

	assert.True(t, pages.HasPage("confirm"))
	_ = doneCalled
}

func TestShowConfirmNilCallback(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	assert.NotPanics(t, func() {
		ui.ShowConfirm(app, pages, "Test", "Test message", nil)
	})
}
