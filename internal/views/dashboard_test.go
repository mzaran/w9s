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

func TestDashboardViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewDashboardView(app, client)

	assert.Equal(t, "dashboard", v.Name())
	assert.Equal(t, "Dashboard", v.Title())
}

func TestDashboardViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewDashboardView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestDashboardViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewDashboardView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}

func TestDashboardViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewDashboardView(app, client)

	hints := v.Hints()
	assert.Contains(t, hints, "r Refresh")
	assert.Contains(t, hints, "Tab Next View")
}
