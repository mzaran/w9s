package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCrumbs(t *testing.T) {
	theme := &ui.DefaultTheme
	c := ui.NewCrumbs(theme)
	require.NotNil(t, c)
	assert.NotNil(t, c.TextView, "underlying TextView should be set")
}

func TestCrumbsSetCurrentView(t *testing.T) {
	c := ui.NewCrumbs(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		c.SetCurrentView("nodes")
	})
	text := c.GetText(false)
	assert.Contains(t, text, "nodes")
}

func TestCrumbsSetCurrentViewEmpty(t *testing.T) {
	c := ui.NewCrumbs(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		c.SetCurrentView("")
	})
}
