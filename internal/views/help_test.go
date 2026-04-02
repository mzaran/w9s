package views_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mzaran/w9s/internal/views"
)

func TestHelpViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	v := views.NewHelpView(app)

	assert.Equal(t, "help", v.Name())
	assert.Equal(t, "Help", v.Title())
}

func TestHelpViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	v := views.NewHelpView(app)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}

func TestHelpViewHints(t *testing.T) {
	app := tview.NewApplication()
	v := views.NewHelpView(app)

	hints := v.Hints()
	assert.Contains(t, hints, "Tab Next View")
	assert.Contains(t, hints, "q Quit")
}

func TestHelpViewHintsNoTODO(t *testing.T) {
	app := tview.NewApplication()
	v := views.NewHelpView(app)

	hints := v.Hints()
	for _, h := range hints {
		assert.False(t, strings.Contains(h, "TODO"),
			"hint should not contain TODO: %q", h)
	}
}
