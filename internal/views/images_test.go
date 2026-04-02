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

func TestImagesViewNameAndTitle(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewImagesView(app, client)

	assert.Equal(t, "images", v.Name())
	assert.Equal(t, "Images", v.Title())
}

func TestImagesViewInitAndRefresh(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewImagesView(app, client)

	err := v.Init(context.Background())
	require.NoError(t, err)

	err = v.Refresh()
	require.NoError(t, err)
}

func TestImagesViewHints(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewImagesView(app, client)

	hints := v.Hints()
	assert.Contains(t, hints, "i Import")
	assert.Contains(t, hints, "b Build")
	assert.Contains(t, hints, "d Delete")
	assert.Contains(t, hints, "/ Filter")
	assert.Contains(t, hints, "r Refresh")
}

func TestImagesViewRenderAfterInit(t *testing.T) {
	app := tview.NewApplication()
	client := mock.NewFastMockClient()
	v := views.NewImagesView(app, client)

	require.NoError(t, v.Init(context.Background()))
	assert.NotNil(t, v.Render())
}
