package ui

import (
	"github.com/rivo/tview"
)

// ShowConfirm displays a modal confirmation dialog with Yes/No buttons.
// When the user selects "Yes", onConfirm is called. The dialog is removed
// from pages in either case.
func ShowConfirm(app *tview.Application, pages *tview.Pages, title, message string, onConfirm func()) {
	const pageName = "confirm"

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage(pageName)
			if buttonLabel == "Yes" && onConfirm != nil {
				onConfirm()
			}
		})
	modal.SetTitle(" " + title + " ").SetBorder(true)

	pages.AddAndSwitchToPage(pageName, modal, true)
}
