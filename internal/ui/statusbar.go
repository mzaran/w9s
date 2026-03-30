package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar is the bottom bar component showing keyboard hints and messages.
type StatusBar struct {
	*tview.TextView

	app   *tview.Application
	hints []string
}

// NewStatusBar creates a new StatusBar component.
func NewStatusBar(app *tview.Application) *StatusBar {
	sb := &StatusBar{
		TextView: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft),
		app:      app,
	}
	sb.SetBackgroundColor(tcell.ColorDarkBlue)
	return sb
}

// SetHints updates the keyboard hints displayed in the status bar.
func (sb *StatusBar) SetHints(hints []string) {
	sb.hints = hints
	sb.renderHints()
}

// SetMessage temporarily displays a message, reverting to hints after duration.
func (sb *StatusBar) SetMessage(msg string, duration time.Duration) {
	sb.SetText(" " + msg)
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			sb.app.QueueUpdateDraw(func() {
				sb.renderHints()
			})
		}()
	}
}

func (sb *StatusBar) renderHints() {
	var parts []string
	for _, h := range sb.hints {
		parts = append(parts, "[yellow]"+h+"[-]")
	}
	sb.SetText(" " + strings.Join(parts, "  |  "))
}
