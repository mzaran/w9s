package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar is the bottom bar showing keyboard hints, flash messages, and item count.
type StatusBar struct {
	*tview.TextView
	theme      *Theme
	app        *tview.Application
	lastHints  []string
	clearTimer *time.Timer
	flashing   bool // true while a flash message is displayed
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

// SetApp sets the tview application reference for QueueUpdateDraw.
func (s *StatusBar) SetApp(app *tview.Application) {
	s.app = app
}

// SetHints displays keyboard hints. Each hint is "key action" format.
// If a flash message is active, hints are stored but not rendered until the flash clears.
func (s *StatusBar) SetHints(hints []string) {
	s.lastHints = hints
	if s.flashing {
		return // don't overwrite flash message
	}
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

// ShowError displays a red error flash message that auto-clears after 4 seconds.
func (s *StatusBar) ShowError(msg string) {
	s.showFlash("✗ "+msg, s.theme.StatusDown)
}

// ShowSuccess displays a green success flash message that auto-clears after 4 seconds.
func (s *StatusBar) ShowSuccess(msg string) {
	s.showFlash("✓ "+msg, s.theme.StatusReady)
}

func (s *StatusBar) showFlash(msg string, color tcell.Color) {
	if s.clearTimer != nil {
		s.clearTimer.Stop()
	}
	s.flashing = true
	s.Clear()
	hex := ColorToHex(color)
	fmt.Fprintf(s, " [#%06x::b]%s[-:-:-]", hex, msg)

	s.clearTimer = time.AfterFunc(4*time.Second, func() {
		if s.app != nil {
			s.app.QueueUpdateDraw(func() {
				s.flashing = false
				s.SetHints(s.lastHints)
			})
		}
	})
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
