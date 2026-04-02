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

func TestProfilesViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewProfilesView(app, client)

	assert.Equal(t, "profiles", v.Name())
	assert.Equal(t, "Profiles", v.Title())
}

func TestProfilesViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewProfilesView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestProfilesViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewProfilesView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}

func TestProfilesViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewProfilesView(app, client)

	hints := v.Hints()
	assert.Contains(t, hints, "a Add")
	assert.Contains(t, hints, "d Delete")
	assert.Contains(t, hints, "Enter Detail")
	assert.Contains(t, hints, "r Refresh")
}
