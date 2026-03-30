package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// StatusBar is the bottom bar showing keyboard hints and item count.
type StatusBar struct {
	*tview.TextView
	theme *Theme
}

// NewStatusBar creates a new themed status bar.
func NewStatusBar(theme *Theme) *StatusBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgColor)
	tv.SetBorder(false)
	return &StatusBar{TextView: tv, theme: theme}
}

// SetHints displays keyboard hints. Each hint is "key action" format.
func (s *StatusBar) SetHints(hints []string) {
	s.Clear()
	key := ColorToHex(s.theme.HotkeyFg)
	hint := ColorToHex(s.theme.HintFg)

	fmt.Fprint(s, " ")
	for _, h := range hints {
		k, a := splitHint(h)
		if a != "" {
			fmt.Fprintf(s, "[#%06x::b]%s[-:-:-] [#%06x]%s[-]  ", key, k, hint, a)
		} else {
			fmt.Fprintf(s, "[#%06x]%s[-]  ", hint, h)
		}
	}
}

// SetCount appends a right-aligned count (e.g., "10 nodes").
func (s *StatusBar) SetCount(label string, count int) {
	cnt := ColorToHex(s.theme.CountFg)
	fmt.Fprintf(s, "[#%06x::d]%d %s[-:-:-]", cnt, count, label)
}

// SetMessage displays a temporary message (replaces hints).
func (s *StatusBar) SetMessage(msg string) {
	s.Clear()
	fmt.Fprintf(s, " [white]%s[-]", msg)
}

// splitHint splits "key action" on first space.
func splitHint(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
