package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// ClusterInfo is the left header panel showing cluster context and metrics.
type ClusterInfo struct {
	*tview.TextView
	theme       *Theme
	clusterName string
	endpoint    string
	username    string
	w9sVersion  string
	wwVersion   string
	readOnly    bool
	nodes       int
	images      int
	overlays    int
	hasMetrics  bool
}

// NewClusterInfo creates a new ClusterInfo panel.
func NewClusterInfo(theme *Theme) *ClusterInfo {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgColor)
	tv.SetBorder(false)

	ci := &ClusterInfo{
		TextView: tv,
		theme:    theme,
	}
	ci.render()
	return ci
}

// SetClusterName sets the cluster display name.
func (ci *ClusterInfo) SetClusterName(name string) {
	ci.clusterName = name
	ci.render()
}

// SetEndpoint sets the cluster endpoint URL.
func (ci *ClusterInfo) SetEndpoint(ep string) {
	ci.endpoint = ep
	ci.render()
}

// SetReadOnly sets the read-only indicator.
func (ci *ClusterInfo) SetReadOnly(ro bool) {
	ci.readOnly = ro
	ci.render()
}

// SetUsername sets the API username display.
func (ci *ClusterInfo) SetUsername(u string) {
	ci.username = u
	ci.render()
}

// SetW9sVersion sets the w9s version display.
func (ci *ClusterInfo) SetW9sVersion(v string) {
	ci.w9sVersion = v
	ci.render()
}

// SetWWVersion sets the Warewulf server version display.
func (ci *ClusterInfo) SetWWVersion(v string) {
	ci.wwVersion = v
	ci.render()
}

// SetMetrics updates the cluster metric counters.
func (ci *ClusterInfo) SetMetrics(nodes, images, overlays int) {
	ci.nodes = nodes
	ci.images = images
	ci.overlays = overlays
	ci.hasMetrics = true
	ci.render()
}

func (ci *ClusterInfo) render() {
	ci.Clear()
	lbl := ColorToHex(ci.theme.MetricLabel)
	val := ColorToHex(ci.theme.MetricValue)
	conn := ColorToHex(ci.theme.ConnectedFg)

	or := func(s, fallback string) string {
		if s == "" {
			return fallback
		}
		return s
	}

	// k9s-style aligned key:value table.
	const fmat = " [#%06x]%-12s[-] [#%06x]%s[-]\n"

	fmt.Fprintf(ci, " [#%06x]%-12s[-] [#%06x]●[-] [white]%s[-]\n", lbl, "Cluster:", conn, or(ci.clusterName, "-"))
	fmt.Fprintf(ci, fmat, lbl, "Endpoint:", val, or(ci.endpoint, "-"))
	fmt.Fprintf(ci, fmat, lbl, "User:", val, or(ci.username, "-"))
	fmt.Fprintf(ci, fmat, lbl, "W9s Rev:", val, or(ci.w9sVersion, "-"))
	fmt.Fprintf(ci, fmat, lbl, "WW Rev:", val, or(ci.wwVersion, "-"))
	if ci.hasMetrics {
		fmt.Fprintf(ci, fmat, lbl, "Nodes:", val, fmt.Sprintf("%d", ci.nodes))
		fmt.Fprintf(ci, fmat, lbl, "Images:", val, fmt.Sprintf("%d", ci.images))
	} else {
		fmt.Fprintf(ci, fmat, lbl, "Nodes:", val, "-")
		fmt.Fprintf(ci, fmat, lbl, "Images:", val, "-")
	}
	if ci.readOnly {
		fmt.Fprintf(ci, " [red::b][READ-ONLY][-:-:-]")
	}
}
