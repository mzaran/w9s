package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowDetail displays content in a full-screen, scrollable, left-aligned pane.
// Press Escape or Enter to close. Supports j/k and arrow key scrolling.
func ShowDetail(pages *tview.Pages, app *tview.Application, pageName, title, content string, onClose func()) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true)
	tv.SetBorder(true).SetTitle(" " + title + " ")
	tv.SetText(content)
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			onClose()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				onClose()
				return nil
			}
		}
		return event // arrows, j/k, pgup/pgdn handled by TextView
	})
	pages.AddAndSwitchToPage(pageName, tv, true)
	app.SetFocus(tv)
}
