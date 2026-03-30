// Package ui provides reusable TUI components for the w9s application.
package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Header is the top section: logo + cluster info + numbered tabs.
type Header struct {
	*tview.Flex
	theme    *Theme
	logo     *tview.TextView
	info     *tview.TextView
	tabs     *tview.TextView
	views    []string
	current  string
	endpoint string
	version  string
}

// NewHeader creates a new Design C header with logo, info panel, and tabs.
func NewHeader(theme *Theme) *Header {
	h := &Header{
		Flex:  tview.NewFlex().SetDirection(tview.FlexRow),
		theme: theme,
	}

	h.logo = NewLogo(theme)
	h.info = tview.NewTextView().SetDynamicColors(true)
	h.info.SetBackgroundColor(theme.BgColor)
	h.info.SetBorder(false)

	h.tabs = tview.NewTextView().SetDynamicColors(true)
	h.tabs.SetBackgroundColor(theme.BgColor)
	h.tabs.SetBorder(false)

	topRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(h.logo, 14, 0, false).
		AddItem(h.info, 0, 1, false)
	topRow.SetBackgroundColor(theme.BgColor)
	topRow.SetBorder(false)

	h.Flex.AddItem(topRow, 3, 0, false).
		AddItem(h.tabs, 1, 0, false)
	h.SetBackgroundColor(theme.BgColor)
	h.SetBorder(false)

	return h
}

// SetClusterInfo updates the endpoint and version display.
func (h *Header) SetClusterInfo(endpoint, version string) {
	h.endpoint = endpoint
	h.version = version
	h.renderInfo(0, 0, 0, 0, 0)
}

// SetMetrics updates the header with cluster metrics.
func (h *Header) SetMetrics(nodes, up, images, profiles, overlays int) {
	h.renderInfo(nodes, up, images, profiles, overlays)
}

func (h *Header) renderInfo(nodes, up, images, profiles, overlays int) {
	h.info.Clear()
	conn := ColorToHex(h.theme.ConnectedFg)
	dim := ColorToHex(h.theme.DimFg)
	val := ColorToHex(h.theme.MetricValue)
	lbl := ColorToHex(h.theme.MetricLabel)

	fmt.Fprintf(h.info, "[white::b]Warewulf Cluster Manager[-:-:-]")
	fmt.Fprintf(h.info, "       [#%06x]● Connected[-]  [white]%s[-]      [#%06x]%s[-]\n",
		conn, h.endpoint, dim, h.version)
	fmt.Fprintf(h.info, "[#%06x]w9s.sh[-]", dim)
	if nodes > 0 || images > 0 || profiles > 0 {
		fmt.Fprintf(h.info, "                          [#%06x]Nodes:[-] [#%06x]%d[-][#%06x]/%d[-]",
			lbl, val, up, lbl, nodes)
		fmt.Fprintf(h.info, "  [#%06x]Images:[-] [#%06x]%d[-]", lbl, val, images)
		fmt.Fprintf(h.info, "  [#%06x]Profiles:[-] [#%06x]%d[-]", lbl, val, profiles)
	}
}

// SetViews sets the list of view names for the tab bar.
func (h *Header) SetViews(views []string) {
	h.views = views
	h.renderTabs()
}

// SetCurrentView highlights the active tab.
func (h *Header) SetCurrentView(name string) {
	h.current = name
	h.renderTabs()
}

// SetClusterName is a compatibility method — calls SetClusterInfo.
func (h *Header) SetClusterName(name string) {
	h.SetClusterInfo(name, h.version)
}

func (h *Header) renderTabs() {
	h.tabs.Clear()
	activeBg := ColorToHex(h.theme.TabActiveBg)
	num := ColorToHex(h.theme.TabNumber)
	name := ColorToHex(h.theme.TabName)

	fmt.Fprint(h.tabs, " ")
	for i, v := range h.views {
		n := i + 1
		label := capitalize(v)
		if v == h.current {
			fmt.Fprintf(h.tabs, "[white:#%06x::b] %d %s [-:-:-] ", activeBg, n, label)
		} else {
			fmt.Fprintf(h.tabs, "[#%06x]%d[-] [#%06x]%s[-]  ", num, n, name, label)
		}
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
