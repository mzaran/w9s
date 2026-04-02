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

func TestHeaderSetClusterName(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetClusterName("mycluster")
	})
}

func TestHeaderSetReadOnly(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	assert.NotPanics(t, func() {
		h.SetReadOnly(true)
	})
}
