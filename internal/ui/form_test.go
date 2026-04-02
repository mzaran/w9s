package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestShowForm(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	fields := []ui.FormField{
		{Key: "name", Label: "Node Name", Default: "node01", Width: 30},
		{Key: "image", Label: "Image", Default: "rocky9"},
	}

	var savedValues map[string]string
	assert.NotPanics(t, func() {
		ui.ShowForm(pages, app, "edit", "main", "Edit Node", fields, func(values map[string]string) {
			savedValues = values
		}, nil)
	})

	// The form page should now be present.
	page := pages.HasPage("edit")
	assert.True(t, page, "form page should be added")
	_ = savedValues
}

func TestShowFormEmptyFields(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	assert.NotPanics(t, func() {
		ui.ShowForm(pages, app, "empty", "main", "Empty Form", nil, nil, nil)
	})
	assert.True(t, pages.HasPage("empty"))
}

func TestShowFormAutoWidth(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	// Width=0 should default to 40 internally.
	fields := []ui.FormField{
		{Key: "test", Label: "Test", Default: "", Width: 0},
	}

	assert.NotPanics(t, func() {
		ui.ShowForm(pages, app, "auto", "main", "Auto Width", fields, nil, nil)
	})
}
