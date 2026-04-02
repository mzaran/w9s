package ui_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mzaran/w9s/internal/ui"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStyledTable(t *testing.T) {
	table := ui.NewStyledTable()
	require.NotNil(t, table)
}

func TestSetTableHeaders(t *testing.T) {
	table := tview.NewTable()
	headers := []string{"Name", "Status", "Image", "IP"}
	ui.SetTableHeaders(table, headers)

	for col, h := range headers {
		cell := table.GetCell(0, col)
		require.NotNil(t, cell, "header cell %d should exist", col)
		assert.Equal(t, h, cell.Text)
	}
}

func TestSetTableHeadersEmpty(t *testing.T) {
	table := tview.NewTable()
	assert.NotPanics(t, func() {
		ui.SetTableHeaders(table, nil)
	})
}

func TestAddTableRow(t *testing.T) {
	table := tview.NewTable()
	values := []string{"node01", "READY", "rocky9", "10.0.0.1"}
	ui.AddTableRow(table, 1, values)

	for col, v := range values {
		cell := table.GetCell(1, col)
		require.NotNil(t, cell)
		assert.Equal(t, v, cell.Text)
	}
}

func TestAddTableRowColored(t *testing.T) {
	table := tview.NewTable()
	values := []string{"node01", "DOWN", "rocky9"}
	colors := []tcell.Color{tcell.ColorWhite, tcell.ColorRed}

	ui.AddTableRowColored(table, 1, values, colors)

	// First two cells should have explicit colors.
	cell0 := table.GetCell(1, 0)
	require.NotNil(t, cell0)
	assert.Equal(t, "node01", cell0.Text)

	cell1 := table.GetCell(1, 1)
	require.NotNil(t, cell1)
	assert.Equal(t, "DOWN", cell1.Text)

	// Third cell has no color in the slice — should still exist.
	cell2 := table.GetCell(1, 2)
	require.NotNil(t, cell2)
	assert.Equal(t, "rocky9", cell2.Text)
}

func TestAddTableRowColoredEmpty(t *testing.T) {
	table := tview.NewTable()
	assert.NotPanics(t, func() {
		ui.AddTableRowColored(table, 1, nil, nil)
	})
}
