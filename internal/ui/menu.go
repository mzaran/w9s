package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// Menu is the right header panel showing keyboard hint shortcuts.
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

	return &Menu{
		TextView: tv,
		theme:    theme,
	}
}

// SetHints updates the keyboard hints displayed in a 2-column grid (max 4 rows).
func (m *Menu) SetHints(hints []string) {
	m.hints = hints
	m.render()
}

func (m *Menu) render() {
	m.Clear()
	keyClr := ColorToHex(m.theme.HotkeyFg)
	hintClr := ColorToHex(m.theme.HintFg)

	// Fixed-width cells: each hint occupies exactly cellWidth characters.
	// 3 columns x 7 rows = 21 slots (enough for ~16 hints).
	const (
		maxRows  = 7
		maxCols  = 3
		cellWidth = 16 // e.g. "s Sort          " or "Tab Next View   "
	)

	total := len(m.hints)
	if total > maxRows*maxCols {
		total = maxRows * maxCols
	}

	// Column-major layout: fill down, then across.
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
			k, a := splitMenuHint(m.hints[idx])
			fmt.Fprintf(m, "[#%06x::b]%s[-:-:-] [#%06x]%-*s[-]", keyClr, k, hintClr, cellWidth-len(k)-1, a)
		}
	}
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
