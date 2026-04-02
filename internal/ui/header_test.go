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
	assert.NotNil(t, h.Flex, "underlying Flex should be set")
}

func TestHeaderClusterInfo(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	require.NotNil(t, h.ClusterInfo())
	assert.NotPanics(t, func() {
		h.ClusterInfo().SetClusterName("mycluster")
		h.ClusterInfo().SetEndpoint("https://ww:9873")
		h.ClusterInfo().SetReadOnly(true)
		h.ClusterInfo().SetMetrics(10, 2, 4)
	})
}

func TestHeaderMenu(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	require.NotNil(t, h.Menu())
	assert.NotPanics(t, func() {
		h.Menu().SetHints([]string{"q Quit", "/ Filter", "s Sort"})
	})
}

func TestHeaderLogo(t *testing.T) {
	h := ui.NewHeader(&ui.DefaultTheme)
	require.NotNil(t, h.Logo())
}
