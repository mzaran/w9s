// Package ui provides reusable TUI components for the w9s application.
package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Header is the top bar component showing cluster name, view tabs, and branding.
type Header struct {
	*tview.Flex

	clusterText *tview.TextView
	tabsText    *tview.TextView
	brandText   *tview.TextView
	views       []string
	currentView string
}

// NewHeader creates a new Header component.
func NewHeader() *Header {
	h := &Header{
		Flex:        tview.NewFlex().SetDirection(tview.FlexColumn),
		clusterText: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft),
		tabsText:    tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter),
		brandText:   tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
	}

	h.brandText.SetText("[blue::b]w9s.sh[-::-]")

	h.AddItem(h.clusterText, 20, 0, false).
		AddItem(h.tabsText, 0, 1, false).
		AddItem(h.brandText, 10, 0, false)

	h.SetBackgroundColor(tcell.ColorDarkBlue)
	h.clusterText.SetBackgroundColor(tcell.ColorDarkBlue)
	h.tabsText.SetBackgroundColor(tcell.ColorDarkBlue)
	h.brandText.SetBackgroundColor(tcell.ColorDarkBlue)

	return h
}

// SetClusterName updates the cluster name display.
func (h *Header) SetClusterName(name string) {
	h.clusterText.SetText(fmt.Sprintf("[white::b] %s[-::-]", name))
}

// SetCurrentView highlights the current view tab.
func (h *Header) SetCurrentView(name string) {
	h.currentView = name
	h.renderTabs()
}

// SetViews sets the list of view names for the tab bar.
func (h *Header) SetViews(views []string) {
	h.views = views
	h.renderTabs()
}

func (h *Header) renderTabs() {
	var parts []string
	for i, v := range h.views {
		label := fmt.Sprintf(" %d:%s ", i+1, v)
		if v == h.currentView {
			parts = append(parts, fmt.Sprintf("[black:white:b]%s[-:-:-]", label))
		} else {
			parts = append(parts, fmt.Sprintf("[white:-:-]%s[-:-:-]", label))
		}
	}
	h.tabsText.SetText(strings.Join(parts, " "))
}
