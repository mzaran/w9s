// Package ui provides reusable TUI components for the w9s application.
package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// Header is a single-row top bar: logo + cluster connection info.
type Header struct {
	*tview.Flex
	theme    *Theme
	logo     *tview.TextView
	info     *tview.TextView
	endpoint string
	version  string
	readOnly bool
}

// NewHeader creates a new single-row header with logo and info panel.
func NewHeader(theme *Theme) *Header {
	h := &Header{
		Flex:  tview.NewFlex().SetDirection(tview.FlexColumn),
		theme: theme,
	}

	h.logo = NewLogo(theme)
	h.info = tview.NewTextView().SetDynamicColors(true)
	h.info.SetBackgroundColor(theme.BgColor)
	h.info.SetBorder(false)

	h.Flex.AddItem(h.logo, 10, 0, false).
		AddItem(h.info, 0, 1, false)
	h.SetBackgroundColor(theme.BgColor)
	h.SetBorder(false)

	return h
}

// SetClusterInfo updates the endpoint and version display.
func (h *Header) SetClusterInfo(endpoint, version string) {
	h.endpoint = endpoint
	h.version = version
	h.renderInfo()
}

func (h *Header) renderInfo() {
	h.info.Clear()
	conn := ColorToHex(h.theme.ConnectedFg)
	dim := ColorToHex(h.theme.DimFg)

	fmt.Fprintf(h.info, "[#%06x]● Connected[-]  [white]%s[-]", conn, h.endpoint)
	if h.version != "" {
		fmt.Fprintf(h.info, "  [#%06x]%s[-]", dim, h.version)
	}
	if h.readOnly {
		fmt.Fprintf(h.info, "  [red::b][READ-ONLY][-:-:-]")
	}
}

// SetReadOnly enables or disables the read-only indicator.
func (h *Header) SetReadOnly(ro bool) {
	h.readOnly = ro
	h.renderInfo()
}

// SetClusterName is a compatibility method -- calls SetClusterInfo.
func (h *Header) SetClusterName(name string) {
	h.SetClusterInfo(name, h.version)
}
