// Package ui provides reusable TUI components for the w9s application.
package ui

import "github.com/rivo/tview"

// WolfSmall is a compact wolf ASCII art for the header logo panel.
// WolfFace is the w9s mascot.
var WolfFace = []string{
	"(O.o)",
}

// Header is a 3-column top bar: ClusterInfo (left) + Menu (middle) + Logo (right).
type Header struct {
	*tview.Flex
	clusterInfo *ClusterInfo
	menu        *Menu
	logo        *LogoPanel
}

// NewHeader creates a new 3-column header.
func NewHeader(theme *Theme) *Header {
	h := &Header{
		Flex:        tview.NewFlex().SetDirection(tview.FlexColumn),
		clusterInfo: NewClusterInfo(theme),
		menu:        NewMenu(theme),
		logo:        NewLogoPanel(theme, WolfFace, "w9s.sh"),
	}

	h.Flex.AddItem(h.clusterInfo, 35, 0, false).
		AddItem(h.menu, 0, 1, false).
		AddItem(h.logo, 10, 0, false)
	h.SetBackgroundColor(theme.BgColor)
	h.SetBorder(false)

	return h
}

// ClusterInfo returns the left cluster info panel.
func (h *Header) ClusterInfo() *ClusterInfo {
	return h.clusterInfo
}

// Menu returns the middle menu/hints panel.
func (h *Header) Menu() *Menu {
	return h.menu
}

// Logo returns the right logo panel.
func (h *Header) Logo() *LogoPanel {
	return h.logo
}
