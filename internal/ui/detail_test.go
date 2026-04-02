package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestShowDetail(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	closeCalled := false
	assert.NotPanics(t, func() {
		ui.ShowDetail(pages, app, "detail", "Node Details", "Name: node01\nIP: 10.0.0.1", func() {
			closeCalled = true
		})
	})

	assert.True(t, pages.HasPage("detail"), "detail page should be added")
	_ = closeCalled
}

func TestShowDetailEmptyContent(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	assert.NotPanics(t, func() {
		ui.ShowDetail(pages, app, "empty", "Empty", "", func() {})
	})
	assert.True(t, pages.HasPage("empty"))
}

func TestShowDetailLongContent(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	pages.AddPage("main", tview.NewBox(), true, true)

	// Generate multi-line content to verify scrollable view handles it.
	content := ""
	for i := 0; i < 100; i++ {
		content += "line of detail content\n"
	}

	assert.NotPanics(t, func() {
		ui.ShowDetail(pages, app, "long", "Long Content", content, func() {})
	})
}
