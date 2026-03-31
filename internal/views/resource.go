package views

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/ui"
)

// Column defines a single column in a resource table.
type Column[T any] struct {
	Name    string
	Width   int
	Extract func(name string, item T) string
	Color   func(name string, item T) tcell.Color // nil = default
}

// Action defines a keyboard action available on a resource view.
type Action[T any] struct {
	Key         rune
	Label       string
	Destructive bool
	Execute     func(ctx context.Context, name string, item T) error
}

// ResourceViewConfig configures a generic ResourceView.
type ResourceViewConfig[T any] struct {
	Fetch      func() (map[string]T, error)
	Columns    []Column[T]
	OnKeyExtra func(rv *ResourceView[T], event *tcell.EventKey) *tcell.EventKey // optional extra key handling
	ExtraHints []string                                                          // additional hints for OnKeyExtra actions
	Actions    []Action[T]
	Detail  func(name string, item T) string
	Filter  func(name string, item T, query string) bool
}

// ResourceView is a generic, DRY view for any Warewulf resource type.
type ResourceView[T any] struct {
	BaseView

	config      ResourceViewConfig[T]
	items       map[string]T
	sortedKeys  []string
	mu          sync.RWMutex
	table       *tview.Table
	container   *tview.Flex
	filterInput *tview.InputField
	filterQuery string
	filtering   bool
	modalOpen   bool
	sortCol     int  // -1 = default (by key name), 0..N = column index
	sortAsc     bool // true = ascending, false = descending
}

// IsFiltering reports whether the filter input or a modal is currently active.
func (rv *ResourceView[T]) IsFiltering() bool {
	return rv.filtering || rv.modalOpen
}

// NewResourceView creates a new ResourceView with the given configuration.
func NewResourceView[T any](name, title string, app *tview.Application, cfg ResourceViewConfig[T]) *ResourceView[T] {
	rv := &ResourceView[T]{
		BaseView: NewBaseView(name, title),
		config:   cfg,
		sortCol:  -1,
		sortAsc:  true,
		// items left nil — first Refresh() will do a synchronous load
	}
	rv.SetApp(app)
	return rv
}

// Init initializes the resource view, creating the table and layout.
func (rv *ResourceView[T]) Init(ctx context.Context) error {
	if err := rv.BaseView.Init(ctx); err != nil {
		return err
	}

	rv.table = tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSeparator(tview.Borders.Vertical)
	rv.table.SetBorder(false)

	// Note: Enter is handled in OnKey, not SetSelectedFunc,
	// to avoid re-entrancy when focus returns to the table after a modal closes.

	rv.filterInput = tview.NewInputField().
		SetLabel(" Filter: ").
		SetFieldWidth(0)
	rv.filterInput.SetChangedFunc(func(text string) {
		rv.filterQuery = text
		rv.renderTable()
	})

	rv.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rv.table, 0, 1, true).
		AddItem(rv.filterInput, 0, 0, false) // hidden by default; shown when / pressed

	return nil
}

// Render returns the view's root primitive.
func (rv *ResourceView[T]) Render() tview.Primitive {
	return rv.container
}

// Hints returns the keyboard hints for the status bar.
func (rv *ResourceView[T]) Hints() []string {
	hints := []string{
		"/ Filter",
		"s Sort",
		"S Reverse",
		"x Export",
		"Enter Detail",
		"r Refresh",
	}
	hints = append(hints, rv.config.ExtraHints...)
	for _, a := range rv.config.Actions {
		hints = append(hints, fmt.Sprintf("%c %s", a.Key, a.Label))
	}
	return hints
}

// Refresh fetches data and updates the table.
// The first call (when items is nil) is synchronous so the initial frame has data.
// Subsequent calls use a background goroutine for non-blocking refreshes.
func (rv *ResourceView[T]) Refresh() error {
	if rv.IsRefreshing() {
		return nil
	}
	rv.SetRefreshing(true)

	rv.mu.RLock()
	firstLoad := rv.items == nil
	rv.mu.RUnlock()

	if firstLoad {
		defer rv.SetRefreshing(false)
		items, err := rv.config.Fetch()
		if err != nil {
			rv.SetLastError(err)
			return nil
		}
		rv.mu.Lock()
		rv.items = items
		rv.sortedKeys = sortKeys(items)
		rv.mu.Unlock()
		rv.SetLastError(nil)
		rv.renderTable()
		return nil
	}

	go func() {
		defer rv.SetRefreshing(false)
		items, err := rv.config.Fetch()
		if err != nil {
			rv.SetLastError(err)
			return
		}
		rv.mu.Lock()
		rv.items = items
		rv.sortedKeys = sortKeys(items)
		rv.mu.Unlock()
		rv.SetLastError(nil)
		rv.renderTable()
	}()

	return nil
}

// OnKey handles key events for the resource view.
func (rv *ResourceView[T]) OnKey(event *tcell.EventKey) *tcell.EventKey {
	// When a modal is open, let tview handle it. Escape closes the modal.
	if rv.modalOpen {
		if event.Key() == tcell.KeyEscape {
			rv.closeModal()
			return nil
		}
		return event
	}

	// When filtering, only handle Escape and Enter here.
	// All other keys pass through to the focused filterInput.
	if rv.filtering {
		switch event.Key() {
		case tcell.KeyEscape:
			rv.filterQuery = ""
			rv.filterInput.SetText("")
			rv.hideFilter()
			return nil
		case tcell.KeyEnter:
			rv.filterQuery = rv.filterInput.GetText()
			rv.hideFilter()
			return nil
		}
		return event // let tview route to filterInput
	}

	// Normal (non-filter) mode.
	switch event.Key() {
	case tcell.KeyEnter:
		row, _ := rv.table.GetSelection()
		rv.showDetail(row)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			rv.showFilter()
			return nil
		case 'r':
			_ = rv.Refresh()
			return nil
		case 's':
			rv.nextSortColumn()
			return nil
		case 'S':
			rv.toggleSortDirection()
			return nil
		case 'x':
			rv.exportCSV()
			return nil
		default:
			if rv.config.OnKeyExtra != nil {
				if result := rv.config.OnKeyExtra(rv, event); result == nil {
					return nil
				}
			}
			return rv.handleAction(event)
		}
	case tcell.KeyEscape:
		if rv.sortCol >= 0 {
			rv.sortCol = -1
			rv.sortAsc = true
			rv.renderTable()
			return nil
		}
		if rv.filterQuery != "" {
			rv.filterQuery = ""
			rv.filterInput.SetText("")
			rv.renderTable()
			return nil
		}
	}
	return event
}

// OnFocus refreshes the view when it receives focus.
func (rv *ResourceView[T]) OnFocus() error {
	if err := rv.BaseView.OnFocus(); err != nil {
		return err
	}
	if rv.table != nil {
		rv.App().SetFocus(rv.table)
	}
	return rv.Refresh()
}

func (rv *ResourceView[T]) showFilter() {
	rv.filtering = true
	rv.container.ResizeItem(rv.filterInput, 1, 0)
	rv.App().SetFocus(rv.filterInput)
}

func (rv *ResourceView[T]) hideFilter() {
	rv.filtering = false
	rv.container.ResizeItem(rv.filterInput, 0, 0)
	rv.renderTable()
	rv.App().SetFocus(rv.table)
}

func (rv *ResourceView[T]) handleAction(event *tcell.EventKey) *tcell.EventKey {
	ch := event.Rune()
	for _, action := range rv.config.Actions {
		if action.Key != ch {
			continue
		}
		name, item, ok := rv.selectedItem()
		if !ok {
			return nil
		}
		if action.Destructive {
			act := action // capture
			rv.modalOpen = true
			ui.ShowConfirm(rv.App(), rv.Pages(), act.Label,
				fmt.Sprintf("Are you sure you want to %s '%s'?", strings.ToLower(act.Label), name),
				func() {
					go func() {
						if err := act.Execute(rv.Ctx(), name, item); err != nil {
							rv.SetLastError(err)
						}
						_ = rv.Refresh()
					}()
				},
				func() { rv.modalOpen = false }, // onDone: always clear modal state
			)
		} else {
			go func() {
				if err := action.Execute(rv.Ctx(), name, item); err != nil {
					rv.SetLastError(err)
				}
				_ = rv.Refresh()
			}()
		}
		return nil
	}
	return event
}

func (rv *ResourceView[T]) selectedItem() (string, T, bool) {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	filtered := rv.filteredKeys()
	row, _ := rv.table.GetSelection()
	idx := row - 1 // account for header row
	if idx < 0 || idx >= len(filtered) {
		var zero T
		return "", zero, false
	}
	key := filtered[idx]
	item, ok := rv.items[key]
	return key, item, ok
}

func (rv *ResourceView[T]) renderTable() {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	// Save current selection before clearing.
	savedRow, _ := rv.table.GetSelection()

	rv.table.Clear()

	// Header row with sort indicator.
	for col, c := range rv.config.Columns {
		header := strings.ToUpper(c.Name)
		if col == rv.sortCol {
			if rv.sortAsc {
				header += " ▲"
			} else {
				header += " ▼"
			}
		}
		cell := tview.NewTableCell(header).
			SetSelectable(false).
			SetExpansion(1).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold)
		if c.Width > 0 {
			cell.SetMaxWidth(c.Width)
		}
		rv.table.SetCell(0, col, cell)
	}

	// Data rows.
	filtered := rv.filteredKeys()
	for rowIdx, key := range filtered {
		item := rv.items[key]
		for col, c := range rv.config.Columns {
			text := c.Extract(key, item)
			cell := tview.NewTableCell(text).SetExpansion(1)
			if c.Width > 0 {
				cell.SetMaxWidth(c.Width)
			}
			if c.Color != nil {
				cell.SetTextColor(c.Color(key, item))
			}
			rv.table.SetCell(rowIdx+1, col, cell)
		}
	}

	// Restore selection.
	if savedRow > 0 && savedRow <= len(filtered) {
		rv.table.Select(savedRow, 0)
	} else if len(filtered) > 0 {
		rv.table.Select(1, 0)
	}
}

func (rv *ResourceView[T]) filteredKeys() []string {
	var keys []string
	if rv.filterQuery == "" {
		keys = make([]string, len(rv.sortedKeys))
		copy(keys, rv.sortedKeys)
	} else {
		query := strings.ToLower(rv.filterQuery)
		for _, key := range rv.sortedKeys {
			item := rv.items[key]
			if rv.config.Filter != nil {
				if rv.config.Filter(key, item, query) {
					keys = append(keys, key)
				}
			} else {
				for _, c := range rv.config.Columns {
					if strings.Contains(strings.ToLower(c.Extract(key, item)), query) {
						keys = append(keys, key)
						break
					}
				}
			}
		}
	}

	// Apply column sort if active.
	if rv.sortCol >= 0 && rv.sortCol < len(rv.config.Columns) {
		col := rv.config.Columns[rv.sortCol]
		asc := rv.sortAsc
		sort.SliceStable(keys, func(i, j int) bool {
			vi := col.Extract(keys[i], rv.items[keys[i]])
			vj := col.Extract(keys[j], rv.items[keys[j]])
			if asc {
				return vi < vj
			}
			return vi > vj
		})
	}

	return keys
}

func (rv *ResourceView[T]) nextSortColumn() {
	rv.sortCol++
	if rv.sortCol >= len(rv.config.Columns) {
		rv.sortCol = 0
	}
	rv.sortAsc = true
	rv.renderTable()
}

func (rv *ResourceView[T]) toggleSortDirection() {
	if rv.sortCol < 0 {
		rv.sortCol = 0
	}
	rv.sortAsc = !rv.sortAsc
	rv.renderTable()
}

func (rv *ResourceView[T]) exportCSV() {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	filtered := rv.filteredKeys()
	if len(filtered) == 0 {
		return
	}

	// Build CSV content.
	var b strings.Builder

	// Header row.
	for i, c := range rv.config.Columns {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(c.Name)
	}
	b.WriteByte('\n')

	// Data rows.
	for _, key := range filtered {
		item := rv.items[key]
		for i, c := range rv.config.Columns {
			if i > 0 {
				b.WriteByte(',')
			}
			val := c.Extract(key, item)
			// Quote values containing commas or newlines.
			if strings.ContainsAny(val, ",\n\"") {
				val = "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
			}
			b.WriteString(val)
		}
		b.WriteByte('\n')
	}

	// Write to /tmp with timestamp.
	filename := fmt.Sprintf("/tmp/w9s-export-%s-%d.csv", rv.Name(), time.Now().Unix())
	if err := os.WriteFile(filename, []byte(b.String()), 0644); err != nil {
		rv.ShowStatusError("Export failed: " + err.Error())
		return
	}

	// Show flash and detail pane.
	rv.ShowStatusMessage(fmt.Sprintf("Exported %d rows to %s", len(filtered), filename))
	msg := fmt.Sprintf("Exported %d rows to:\n%s\n\n%s", len(filtered), filename, b.String())
	rv.modalOpen = true
	ui.ShowDetail(rv.Pages(), rv.App(), "export", "Export Complete", msg, func() {
		rv.closeModal()
	})
}

func (rv *ResourceView[T]) showDetail(row int) {
	if rv.config.Detail == nil {
		return
	}
	rv.mu.RLock()
	filtered := rv.filteredKeys()
	idx := row - 1
	if idx < 0 || idx >= len(filtered) {
		rv.mu.RUnlock()
		return
	}
	key := filtered[idx]
	item := rv.items[key]
	rv.mu.RUnlock()

	text := rv.config.Detail(key, item)

	rv.modalOpen = true
	ui.ShowDetail(rv.Pages(), rv.App(), "detail", key, text, func() {
		rv.closeModal()
	})
}

func (rv *ResourceView[T]) closeModal() {
	rv.modalOpen = false
	rv.Pages().RemovePage("detail")
	rv.Pages().SwitchToPage(rv.Name())
	rv.App().SetFocus(rv.table)
}

func sortKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
