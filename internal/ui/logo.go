package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// LogoLines contains the 3-line block-letter W9S logo.
var LogoLines = []string{
	" ╦ ╦╔═╗╔═╗",
	" ║║║╠═╣╚═╗",
	" ╚╩╝╩ ╩╚═╝",
}

// NewLogo creates a tview.TextView rendering the W9S logo in theme colors.
func NewLogo(theme *Theme) *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgColor)
	tv.SetBorder(false)
	hex := ColorToHex(theme.LogoColor)
	for _, line := range LogoLines {
		fmt.Fprintf(tv, "[#%06x::b]%s[-:-:-]\n", hex, line)
	}
	return tv
}
