package ui_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogo(t *testing.T) {
	logo := ui.NewLogo(&ui.DefaultTheme)
	require.NotNil(t, logo)
}

func TestLogoContainsW9s(t *testing.T) {
	logo := ui.NewLogo(&ui.DefaultTheme)
	text := logo.GetText(true) // stripTags=true
	assert.Contains(t, text, "w9s", "logo should contain 'w9s'")
}

func TestLogoNonEmpty(t *testing.T) {
	logo := ui.NewLogo(&ui.DefaultTheme)
	text := logo.GetText(false)
	assert.NotEmpty(t, text, "logo text should not be empty")
}
