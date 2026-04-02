package ui

import (
	"github.com/rivo/tview"
)

// ShowConfirm displays a modal confirmation dialog with Yes/No buttons.
// When the user selects "Yes", onConfirm is called. The dialog is removed
// from pages in either case.
// ShowConfirm displays a modal confirmation dialog with Yes/No buttons.
// onConfirm is called on "Yes". onDone is always called when the dialog closes (optional).
func ShowConfirm(app *tview.Application, pages *tview.Pages, title, message string, onConfirm func(), onDone ...func()) {
	const pageName = "confirm"

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage(pageName)
			if buttonLabel == "Yes" && onConfirm != nil {
				onConfirm()
			}
			for _, fn := range onDone {
				fn()
			}
		})
	modal.SetTitle(" " + title + " ").SetBorder(true)

	pages.AddAndSwitchToPage(pageName, modal, true)
}
