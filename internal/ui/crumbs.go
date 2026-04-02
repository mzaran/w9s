package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Crumbs shows the current view as a breadcrumb bar.
type Crumbs struct {
	*tview.TextView
	theme   *Theme
	current string
}

// NewCrumbs creates a new breadcrumb component.
func NewCrumbs(theme *Theme) *Crumbs {
	c := &Crumbs{
		TextView: tview.NewTextView(),
		theme:    theme,
	}
	c.SetDynamicColors(true)
	c.SetBackgroundColor(theme.BgColor)
	c.SetBorderPadding(0, 0, 1, 1)
	return c
}

// SetCurrentView updates the breadcrumb to show the active view name.
func (c *Crumbs) SetCurrentView(name string) {
	c.current = name
	c.Clear()
	activeBg := ColorToHex(c.theme.CrumbActiveBg)
	fg := ColorToHex(c.theme.CrumbFg)
	fmt.Fprintf(c, "[#%06x:#%06x:b] <%s> [-:-:-]", fg, activeBg, strings.ToLower(name))
}
