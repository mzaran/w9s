package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

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
	Fetch   func() (map[string]T, error)
	Columns []Column[T]
	Actions []Action[T]
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
	selectedRow int
}

// NewResourceView creates a new ResourceView with the given configuration.
func NewResourceView[T any](name, title string, app *tview.Application, cfg ResourceViewConfig[T]) *ResourceView[T] {
	rv := &ResourceView[T]{
		BaseView: NewBaseView(name, title),
		config:   cfg,
		items:    make(map[string]T),
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

	rv.table.SetSelectedFunc(func(row, _ int) {
		rv.showDetail(row)
	})

	rv.filterInput = tview.NewInputField().
		SetLabel(" Filter: ").
		SetFieldWidth(0).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				rv.filterQuery = ""
				rv.filterInput.SetText("")
			case tcell.KeyEnter:
				rv.filterQuery = rv.filterInput.GetText()
			}
			rv.renderTable()
			rv.App().SetFocus(rv.table)
		})
	rv.filterInput.SetChangedFunc(func(text string) {
		rv.filterQuery = text
		rv.renderTable()
	})

	rv.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rv.table, 0, 1, true).
		AddItem(rv.filterInput, 1, 0, false)

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
		"Enter Detail",
		"r Refresh",
	}
	for _, a := range rv.config.Actions {
		hints = append(hints, fmt.Sprintf("%c %s", a.Key, a.Label))
	}
	return hints
}

// Refresh fetches data in a background goroutine and updates the table.
func (rv *ResourceView[T]) Refresh() error {
	if rv.IsRefreshing() {
		return nil
	}
	rv.SetRefreshing(true)

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
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			rv.App().SetFocus(rv.filterInput)
			return nil
		case 'r':
			_ = rv.Refresh()
			return nil
		default:
			return rv.handleAction(event)
		}
	case tcell.KeyEscape:
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

	rv.table.Clear()

	// Header row.
	for col, c := range rv.config.Columns {
		cell := tview.NewTableCell(strings.ToUpper(c.Name)).
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
	if rv.selectedRow > 0 && rv.selectedRow <= len(filtered) {
		rv.table.Select(rv.selectedRow, 0)
	} else if len(filtered) > 0 {
		rv.table.Select(1, 0)
	}
}

func (rv *ResourceView[T]) filteredKeys() []string {
	if rv.filterQuery == "" {
		return rv.sortedKeys
	}
	var result []string
	query := strings.ToLower(rv.filterQuery)
	for _, key := range rv.sortedKeys {
		item := rv.items[key]
		if rv.config.Filter != nil {
			if rv.config.Filter(key, item, query) {
				result = append(result, key)
			}
		} else {
			// Default: substring match on any column value.
			for _, c := range rv.config.Columns {
				if strings.Contains(strings.ToLower(c.Extract(key, item)), query) {
					result = append(result, key)
					break
				}
			}
		}
	}
	return result
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

	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(_ int, _ string) {
			rv.Pages().RemovePage("detail")
			rv.App().SetFocus(rv.table)
		})
	rv.Pages().AddAndSwitchToPage("detail", modal, true)
}

func sortKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
