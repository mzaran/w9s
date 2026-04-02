package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// LogoPanel displays ASCII art branding in the header.
type LogoPanel struct {
	*tview.TextView
	theme *Theme
}

// NewLogoPanel creates a logo panel with the given ASCII lines and brand text below.
func NewLogoPanel(theme *Theme, lines []string, brand string) *LogoPanel {
	lp := &LogoPanel{
		TextView: tview.NewTextView(),
		theme:    theme,
	}
	lp.SetDynamicColors(true)
	lp.SetBackgroundColor(theme.BgColor)
	lp.SetTextAlign(tview.AlignLeft)

	hex := ColorToHex(theme.ConnectedFg)
	brandHex := ColorToHex(theme.BrandColor)

	// Find widest line for centering the brand text.
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	for i, line := range lines {
		if i > 0 {
			fmt.Fprint(lp, "\n")
		}
		fmt.Fprintf(lp, "[#%06x]%s[-]", hex, line)
	}
	if brand != "" {
		// Center brand under the art.
		pad := (maxWidth - len(brand)) / 2
		if pad < 0 {
			pad = 0
		}
		spaces := ""
		for i := 0; i < pad; i++ {
			spaces += " "
		}
		fmt.Fprintf(lp, "\n[#%06x::b]%s%s[-:-:-]", brandHex, spaces, brand)
	}

	return lp
}
