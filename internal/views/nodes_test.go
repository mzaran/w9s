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

func TestNodesViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewNodesView(app, client)

	assert.Equal(t, "nodes", v.Name())
	assert.Equal(t, "Nodes", v.Title())
}

func TestNodesViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewNodesView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestNodesViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewNodesView(app, client)

	hints := v.Hints()
	assert.Contains(t, hints, "a Add")
	assert.Contains(t, hints, "e Edit")
	assert.Contains(t, hints, "d Delete")
	assert.Contains(t, hints, "b Build Overlays")
	assert.Contains(t, hints, "/ Filter")
	assert.Contains(t, hints, "r Refresh")
}

func TestNodesViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewNodesView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}
