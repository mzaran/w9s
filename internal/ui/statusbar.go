package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// spinnerChars are braille characters used for the animated spinner.
var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// StatusBar is the bottom bar showing keyboard hints, flash messages, and item count.
type StatusBar struct {
	*tview.TextView
	theme         *Theme
	app           *tview.Application
	lastHints     []string
	clearTimer    *time.Timer
	flashing      bool // true while a flash message is displayed
	spinnerMsg    string
	spinnerActive bool
	spinnerFrame  int
	spinnerStop   chan struct{}
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

// SetHints stores keyboard hints (no longer rendered — hints are in the header Menu).
func (s *StatusBar) SetHints(hints []string) {
	s.lastHints = hints
}

// ShowError displays a red error flash message that auto-clears after 4 seconds.
func (s *StatusBar) ShowError(msg string) {
	s.showFlash("ERR: "+msg, s.theme.StatusDown)
}

// ShowSuccess displays a green success flash message that auto-clears after 4 seconds.
func (s *StatusBar) ShowSuccess(msg string) {
	s.showFlash("OK: "+msg, s.theme.StatusReady)
}

func (s *StatusBar) showFlash(msg string, color tcell.Color) {
	// Auto-hide spinner when a result arrives.
	if s.spinnerActive {
		s.spinnerActive = false
		close(s.spinnerStop)
	}
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

// ShowSpinner starts an animated spinner with the given message.
// The spinner auto-cycles through braille characters at 80ms intervals.
// Call HideSpinner to stop, or it auto-stops when ShowSuccess/ShowError is called.
func (s *StatusBar) ShowSpinner(msg string) {
	s.spinnerMsg = msg
	s.spinnerActive = true
	s.spinnerFrame = 0
	s.spinnerStop = make(chan struct{})

	s.updateSpinnerDisplay()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.spinnerFrame = (s.spinnerFrame + 1) % len(spinnerChars)
				if s.app != nil {
					s.app.QueueUpdateDraw(func() {
						if s.spinnerActive {
							s.updateSpinnerDisplay()
						}
					})
				}
			case <-s.spinnerStop:
				return
			}
		}
	}()
}

// HideSpinner stops the spinner and restores the last hints.
func (s *StatusBar) HideSpinner() {
	if !s.spinnerActive {
		return
	}
	s.spinnerActive = false
	close(s.spinnerStop)
	s.SetHints(s.lastHints)
}

// IsSpinnerActive reports whether the spinner is currently running.
func (s *StatusBar) IsSpinnerActive() bool {
	return s.spinnerActive
}

func (s *StatusBar) updateSpinnerDisplay() {
	char := spinnerChars[s.spinnerFrame]
	s.Clear()
	hex := ColorToHex(s.theme.HintFg)
	fmt.Fprintf(s, " [#%06x::b]%c[-:-:-] [#%06x]%s[-]", hex, char, hex, s.spinnerMsg)
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

