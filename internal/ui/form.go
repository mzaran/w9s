package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FormField defines a field in a form.
type FormField struct {
	Key     string // programmatic key for reading the value
	Label   string // display label
	Default string // pre-filled value
	Width   int    // field width (0 = auto)
}

// ShowForm displays a full-screen form with Save and Cancel buttons.
// onSave receives the field values keyed by FormField.Key.
// onClose is called on both save and cancel to restore focus.
func ShowForm(pages *tview.Pages, app *tview.Application, pageName, returnPage, title string, fields []FormField, onSave func(values map[string]string), onClose func()) {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" " + title + " ")

	for _, f := range fields {
		w := f.Width
		if w == 0 {
			w = 40
		}
		form.AddInputField(f.Label, f.Default, w, nil, nil)
	}

	closeForm := func() {
		pages.RemovePage(pageName)
		pages.SwitchToPage(returnPage)
		if onClose != nil {
			onClose()
		}
	}

	form.AddButton("Save", func() {
		values := make(map[string]string)
		for _, f := range fields {
			item := form.GetFormItemByLabel(f.Label)
			if input, ok := item.(*tview.InputField); ok {
				values[f.Key] = input.GetText()
			}
		}
		if onSave != nil {
			onSave(values)
		}
		closeForm()
	})

	form.AddButton("Cancel", func() {
		closeForm()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeForm()
			return nil
		}
		return event
	})

	pages.AddAndSwitchToPage(pageName, form, true)
	app.SetFocus(form)
}
