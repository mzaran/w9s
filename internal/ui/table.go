package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NewStyledTable creates a tview.Table with standard w9s styling:
// selectable rows, fixed header, and vertical separators.
func NewStyledTable() *tview.Table {
	t := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSeparator(tview.Borders.Vertical)
	t.SetBorder(false)
	return t
}

// SetTableHeaders populates the first row of a table with styled header cells.
func SetTableHeaders(table *tview.Table, headers []string) {
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell(h).
			SetSelectable(false).
			SetExpansion(1).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold))
	}
}

// AddTableRow adds a row of string values to the table at the given row index.
func AddTableRow(table *tview.Table, row int, values []string) {
	for col, v := range values {
		table.SetCell(row, col, tview.NewTableCell(v).SetExpansion(1))
	}
}

// AddTableRowColored adds a row with per-cell colors. If colors is shorter
// than values, remaining cells use the default color.
func AddTableRowColored(table *tview.Table, row int, values []string, colors []tcell.Color) {
	for col, v := range values {
		cell := tview.NewTableCell(v).SetExpansion(1)
		if col < len(colors) {
			cell.SetTextColor(colors[col])
		}
		table.SetCell(row, col, cell)
	}
}
