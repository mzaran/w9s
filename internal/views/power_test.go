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

func TestPowerViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewPowerView(app, client)

	assert.Equal(t, "power", v.Name())
	assert.Equal(t, "Power", v.Title())
}

func TestPowerViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewPowerView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestPowerViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewPowerView(app, client)

	hints := v.Hints()
	assert.Contains(t, hints, "o Power On")
	assert.Contains(t, hints, "f Power Off")
	assert.Contains(t, hints, "c Cycle")
	assert.Contains(t, hints, "r Reset")
	assert.Contains(t, hints, "s Status")
	assert.Contains(t, hints, "/ Filter")
}

func TestPowerViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewPowerView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}
