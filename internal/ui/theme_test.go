package ui_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mzaran/w9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestDefaultThemeInitialized(t *testing.T) {
	theme := ui.DefaultTheme

	// Spot-check that key colors are set (not the zero value for tcell.Color).
	assert.NotEqual(t, tcell.Color(0), theme.LogoColor, "LogoColor should be set")
	assert.NotEqual(t, tcell.Color(0), theme.ConnectedFg, "ConnectedFg should be set")
	assert.NotEqual(t, tcell.Color(0), theme.StatusReady, "StatusReady should be set")
	assert.NotEqual(t, tcell.Color(0), theme.StatusDown, "StatusDown should be set")
	assert.NotEqual(t, tcell.Color(0), theme.HotkeyFg, "HotkeyFg should be set")
	assert.NotEqual(t, tcell.Color(0), theme.TabActiveBg, "TabActiveBg should be set")
	assert.NotEqual(t, tcell.Color(0), theme.SelectionBg, "SelectionBg should be set")
}

func TestColorToHex(t *testing.T) {
	tests := []struct {
		name  string
		color tcell.Color
		want  int32
	}{
		{"red", tcell.NewRGBColor(255, 0, 0), 0xFF0000},
		{"green", tcell.NewRGBColor(0, 255, 0), 0x00FF00},
		{"blue", tcell.NewRGBColor(0, 0, 255), 0x0000FF},
		{"white", tcell.NewRGBColor(255, 255, 255), 0xFFFFFF},
		{"black", tcell.NewRGBColor(0, 0, 0), 0x000000},
		{"cyan", tcell.NewRGBColor(0, 188, 212), 0x00BCD4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.ColorToHex(tt.color)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultThemeFieldsAreDistinct(t *testing.T) {
	theme := ui.DefaultTheme
	// StatusReady (green) and StatusDown (red) must be different.
	assert.NotEqual(t, theme.StatusReady, theme.StatusDown,
		"StatusReady and StatusDown should be different colors")
}
