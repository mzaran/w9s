package views

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpView displays keyboard shortcuts and help text.
type HelpView struct {
	BaseView

	textView  *tview.TextView
	container *tview.Flex
}

// NewHelpView creates a new help view.
func NewHelpView(app *tview.Application) *HelpView {
	hv := &HelpView{
		BaseView: NewBaseView("help", "Help"),
	}
	hv.SetApp(app)
	return hv
}

func (hv *HelpView) Init(ctx context.Context) error {
	if err := hv.BaseView.Init(ctx); err != nil {
		return err
	}

	hv.textView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetScrollable(true)
	hv.textView.SetBorder(true).SetTitle(" w9s Help ")

	hv.textView.SetText(helpText)

	hv.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(hv.textView, 0, 1, true)

	return nil
}

func (hv *HelpView) Render() tview.Primitive {
	return hv.container
}

func (hv *HelpView) Hints() []string {
	return []string{"Tab Next View", "q Quit"}
}

func (hv *HelpView) OnKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (hv *HelpView) OnFocus() error {
	if err := hv.BaseView.OnFocus(); err != nil {
		return err
	}
	if hv.textView != nil {
		hv.App().SetFocus(hv.textView)
	}
	return nil
}

const helpText = `[yellow]Global Keys[-]

  [green]Tab[-]          Next view
  [green]Shift+Tab[-]    Previous view
  [green]1-9[-]          Jump to view by number
  [green]q[-]            Quit
  [green]?[-]            Show this help
  [green]C[-]            Switch cluster (Shift+C)
  [green]:[-]            Command bar (type :nodes, :q, etc.)

[yellow]Command Bar Commands[-]

  [green]:dashboard[-]   or [green]:1[-]   Switch to Dashboard
  [green]:nodes[-]       or [green]:2[-]   Switch to Nodes
  [green]:profiles[-]    or [green]:3[-]   Switch to Profiles
  [green]:images[-]      or [green]:4[-]   Switch to Images
  [green]:overlays[-]    or [green]:5[-]   Switch to Overlays
  [green]:power[-]       or [green]:6[-]   Switch to Power
  [green]:help[-]        or [green]:7[-]   Switch to Help
  [green]:q[-]           or [green]:quit[-] Quit

[yellow]Resource Views (Nodes, Images, Power)[-]

  [green]/[-]            Open filter (fuzzy match)
  [green]Escape[-]       Clear filter / cancel
  [green]Enter[-]        Show detail for selected item
  [green]r[-]            Refresh data

[yellow]Nodes View[-]

  [green]d[-]            Delete node (with confirmation)
  [green]b[-]            Build overlays for node

[yellow]Images View[-]

  [green]b[-]            Build image
  [green]i[-]            Import image (TODO)
  [green]d[-]            Delete image (with confirmation)

[yellow]Power View[-]

  [green]o[-]            Power on
  [green]f[-]            Power off
  [green]c[-]            Power cycle
  [green]r[-]            Reset
  [green]s[-]            Check status

[yellow]Profiles View[-]

  [green]Enter[-]        Show inheritance tree
  [green]r[-]            Refresh

[yellow]Overlays View[-]

  [green]Enter[-]        Browse files in overlay / view file
  [green]Escape[-]       Back to overlay list
  [green]r[-]            Refresh
`
