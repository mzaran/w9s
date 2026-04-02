package views_test

import (
	"context"
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mzaran/w9s/internal/views"
	"github.com/mzaran/w9s/pkg/mock"
)

func TestOverlaysViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewOverlaysView(app, client)

	assert.Equal(t, "overlays", v.Name())
	assert.Equal(t, "Overlays", v.Title())
}

func TestOverlaysViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewOverlaysView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestOverlaysViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewOverlaysView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}

func TestOverlaysViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewOverlaysView(app, client)

	hints := v.Hints()
	// Default mode is overlay list mode.
	assert.Contains(t, hints, "Enter Browse Files")
	assert.Contains(t, hints, "r Refresh")
}
