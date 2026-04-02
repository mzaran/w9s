package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// NewLogo creates a simple "w9s.sh" text brand in theme colors.
func NewLogo(theme *Theme) *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgColor)
	tv.SetBorder(false)
	hex := ColorToHex(theme.LogoColor)
	fmt.Fprintf(tv, "[#%06x::b] w9s.sh[-:-:-]", hex)
	return tv
}
