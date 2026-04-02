package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeader(t *testing.T) {
	theme := &ui.DefaultTheme
	h := ui.NewHeader(theme)
	require.NotNil(t, h)
	// Header embeds *tview.Flex, verify it's a valid primitive.
	assert.NotNil(t, h.Flex, "underlying Flex should be set")
}

func TestHeaderSetClusterInfo(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetClusterInfo("https://ww.example.com:9873", "v4.6.0")
	})
}

func TestHeaderSetMetrics(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetMetrics(10, 8, 2, 3, 4)
	})
}

func TestHeaderSetViews(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetViews([]string{"nodes", "images", "profiles", "overlays"})
	})
}

func TestHeaderSetCurrentView(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	h.SetViews([]string{"nodes", "images", "profiles"})
	assert.NotPanics(t, func() {
		h.SetCurrentView("images")
	})
}

func TestHeaderSetClusterName(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetClusterName("mycluster")
	})
}

func TestHeaderSetViewsEmpty(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetViews(nil)
	})
}

func TestHeaderSetCurrentViewNotInList(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	h.SetViews([]string{"nodes"})
	assert.NotPanics(t, func() {
		h.SetCurrentView("nonexistent")
	})
}
