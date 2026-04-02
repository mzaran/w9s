package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Menu is the header panel showing keyboard hint shortcuts.
type Menu struct {
	*tview.TextView
	theme *Theme
	hints []string
}

// NewMenu creates a new Menu panel.
func NewMenu(theme *Theme) *Menu {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgColor)
	tv.SetBorder(false)
	tv.SetBorderPadding(0, 0, 2, 0) // left padding for spacing from ClusterInfo

	return &Menu{
		TextView: tv,
		theme:    theme,
	}
}

// SetHints updates the keyboard hints in a column-major grid.
func (m *Menu) SetHints(hints []string) {
	m.hints = hints
	m.render()
}

func (m *Menu) render() {
	m.Clear()
	keyClr := ColorToHex(m.theme.HotkeyFg)
	hintClr := ColorToHex(m.theme.HintFg)

	const (
		maxRows = 7
		maxCols = 2
		keyW    = 4  // width for key portion (e.g. "Tab " or "s   ")
		descW   = 12 // width for description (e.g. "Next View   ")
	)

	total := len(m.hints)
	if total > maxRows*maxCols {
		total = maxRows * maxCols
	}

	cols := (total + maxRows - 1) / maxRows
	if cols > maxCols {
		cols = maxCols
	}

	for row := 0; row < maxRows; row++ {
		if row > 0 {
			fmt.Fprint(m, "\n")
		}
		for col := 0; col < cols; col++ {
			idx := col*maxRows + row
			if idx >= total {
				continue
			}
			if col > 0 {
				fmt.Fprint(m, "  ") // gap between columns
			}
			k, a := splitMenuHint(m.hints[idx])
			// Pad key and description separately for clean alignment.
			paddedKey := padRight(k, keyW)
			paddedDesc := padRight(a, descW)
			fmt.Fprintf(m, "[#%06x::b]%s[-:-:-][#%06x]%s[-]", keyClr, paddedKey, hintClr, paddedDesc)
		}
	}
}

// padRight pads s with spaces to the given width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// splitMenuHint splits "key action" on first space.
func splitMenuHint(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
