package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/ui"
)

const (
	overlayModeList = iota
	overlayModeFiles
)

// OverlaysView provides a two-mode view: overlay list and file tree drill-down.
type OverlaysView struct {
	BaseView

	client          dao.WarewulfClient
	overlays        map[string]*dao.WwOverlay
	sortedOverlays  []string
	mu              sync.RWMutex
	mode            int
	selectedOverlay string
	table           *tview.Table
	container       *tview.Flex
	breadcrumb      *tview.TextView
	modalOpen       bool
}

// NewOverlaysView creates a new overlays view.
func NewOverlaysView(app *tview.Application, client dao.WarewulfClient) *OverlaysView {
	ov := &OverlaysView{
		BaseView: NewBaseView("overlays", "Overlays"),
		client:   client,
		overlays: make(map[string]*dao.WwOverlay),
	}
	ov.SetApp(app)
	return ov
}

func (ov *OverlaysView) Init(ctx context.Context) error {
	if err := ov.BaseView.Init(ctx); err != nil {
		return err
	}

	ov.table = tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSeparator(tview.Borders.Vertical)
	ov.table.SetBorder(false)
	// Enter is handled in OnKey to avoid re-entrancy on focus restore.

	ov.breadcrumb = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	ov.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ov.breadcrumb, 1, 0, false).
		AddItem(ov.table, 0, 1, true)

	return nil
}

func (ov *OverlaysView) Render() tview.Primitive {
	return ov.container
}

func (ov *OverlaysView) Hints() []string {
	if ov.mode == overlayModeFiles {
		return []string{"Esc Back", "Enter View File", "r Refresh"}
	}
	return []string{"Enter Browse Files", "r Refresh"}
}

// IsFiltering reports whether a modal is open (implements Filterable).
func (ov *OverlaysView) IsFiltering() bool {
	return ov.modalOpen
}

func (ov *OverlaysView) OnKey(event *tcell.EventKey) *tcell.EventKey {
	// When file content modal is open
	if ov.modalOpen {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			ov.modalOpen = false
			ov.Pages().RemovePage("filecontent")
			ov.Pages().SwitchToPage(ov.Name())
			ov.App().SetFocus(ov.table)
			return nil
		}
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		row, _ := ov.table.GetSelection()
		ov.onEnter(row)
		return nil
	case tcell.KeyEscape:
		if ov.mode == overlayModeFiles {
			ov.mode = overlayModeList
			ov.selectedOverlay = ""
			ov.renderOverlayList()
			return nil
		}
	case tcell.KeyRune:
		if event.Rune() == 'r' {
			_ = ov.Refresh()
			return nil
		}
	}
	return event
}

func (ov *OverlaysView) OnFocus() error {
	if err := ov.BaseView.OnFocus(); err != nil {
		return err
	}
	if ov.table != nil {
		ov.App().SetFocus(ov.table)
	}
	return ov.Refresh()
}

func (ov *OverlaysView) Refresh() error {
	if ov.IsRefreshing() {
		return nil
	}
	ov.SetRefreshing(true)

	go func() {
		defer ov.SetRefreshing(false)

		overlays, err := ov.client.Overlays().List()
		if err != nil {
			ov.SetLastError(err)
			return
		}

		ov.mu.Lock()
		ov.overlays = overlays
		ov.sortedOverlays = make([]string, 0, len(overlays))
		for k := range overlays {
			ov.sortedOverlays = append(ov.sortedOverlays, k)
		}
		sort.Strings(ov.sortedOverlays)
		ov.mu.Unlock()

		ov.App().QueueUpdateDraw(func() {
			if ov.mode == overlayModeList {
				ov.renderOverlayList()
			} else {
				ov.renderFileList()
			}
		})
	}()

	return nil
}

func (ov *OverlaysView) renderOverlayList() {
	ov.mu.RLock()
	defer ov.mu.RUnlock()

	ov.breadcrumb.SetText("[yellow]Overlays[-]")
	ov.table.Clear()

	headers := []string{"NAME", "FILES", "SITE"}
	for col, h := range headers {
		ov.table.SetCell(0, col, tview.NewTableCell(h).
			SetSelectable(false).
			SetExpansion(1).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold))
	}

	for row, name := range ov.sortedOverlays {
		ovl := ov.overlays[name]
		site := "no"
		if ovl.Site {
			site = "yes"
		}
		vals := []string{
			name,
			fmt.Sprintf("%d", len(ovl.Files)),
			site,
		}
		for col, v := range vals {
			ov.table.SetCell(row+1, col, tview.NewTableCell(v).SetExpansion(1))
		}
	}

	if len(ov.sortedOverlays) > 0 {
		ov.table.Select(1, 0)
	}
}

func (ov *OverlaysView) renderFileList() {
	ov.mu.RLock()
	defer ov.mu.RUnlock()

	ov.breadcrumb.SetText(fmt.Sprintf("[yellow]Overlays[-] > [green]%s[-]", ov.selectedOverlay))
	ov.table.Clear()

	headers := []string{"FILE"}
	for col, h := range headers {
		ov.table.SetCell(0, col, tview.NewTableCell(h).
			SetSelectable(false).
			SetExpansion(1).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold))
	}

	ovl, ok := ov.overlays[ov.selectedOverlay]
	if !ok {
		return
	}

	files := make([]string, len(ovl.Files))
	copy(files, ovl.Files)
	sort.Strings(files)

	for row, f := range files {
		color := tcell.ColorWhite
		if strings.HasSuffix(f, ".ww") {
			color = tcell.ColorAquaMarine
		}
		ov.table.SetCell(row+1, 0, tview.NewTableCell(f).
			SetExpansion(1).
			SetTextColor(color))
	}

	if len(files) > 0 {
		ov.table.Select(1, 0)
	}
}

func (ov *OverlaysView) onEnter(row int) {
	idx := row - 1
	if idx < 0 {
		return
	}

	ov.mu.RLock()
	mode := ov.mode

	switch mode {
	case overlayModeList:
		if idx >= len(ov.sortedOverlays) {
			ov.mu.RUnlock()
			return
		}
		name := ov.sortedOverlays[idx]
		ov.mu.RUnlock()

		ov.selectedOverlay = name
		ov.mode = overlayModeFiles
		ov.renderFileList()

	case overlayModeFiles:
		ovl, ok := ov.overlays[ov.selectedOverlay]
		if !ok {
			ov.mu.RUnlock()
			return
		}
		files := make([]string, len(ovl.Files))
		copy(files, ovl.Files)
		sort.Strings(files)
		overlay := ov.selectedOverlay
		ov.mu.RUnlock()

		if idx >= len(files) {
			return
		}
		ov.showFileContent(overlay, files[idx])

	default:
		ov.mu.RUnlock()
	}
}

func (ov *OverlaysView) showFileContent(overlay, path string) {
	go func() {
		file, err := ov.client.Overlays().GetFile(overlay, path, "")
		if err != nil {
			ov.SetLastError(err)
			return
		}

		ov.App().QueueUpdateDraw(func() {
			content := file.Contents
			title := fmt.Sprintf("%s: %s", overlay, path)

			ov.modalOpen = true
			ui.ShowDetail(ov.Pages(), ov.App(), "filecontent", title, content, func() {
				ov.modalOpen = false
				ov.Pages().RemovePage("filecontent")
				ov.Pages().SwitchToPage(ov.Name())
				ov.App().SetFocus(ov.table)
			})
		})
	}()
}
