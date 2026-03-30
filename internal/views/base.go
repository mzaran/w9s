// Package views provides the view layer for the w9s TUI application.
package views

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// View is the interface that all w9s views must implement.
type View interface {
	Name() string
	Title() string
	Hints() []string
	Init(ctx context.Context) error
	Render() tview.Primitive
	Refresh() error
	OnKey(event *tcell.EventKey) *tcell.EventKey
	OnFocus() error
	OnLoseFocus() error
	Stop() error
	SetSwitchViewFn(func(string))
	SwitchToView(string)
}

// BaseView provides default implementations for the View interface.
// Embed this in concrete views to avoid boilerplate.
type BaseView struct {
	ctx         context.Context
	name        string
	title       string
	app         *tview.Application
	pages       *tview.Pages
	viewMgr     *ViewManager
	switchViewFn func(string)

	refreshing  atomic.Bool
	initialized atomic.Bool
	focused     atomic.Bool
	lastErrMu   sync.Mutex
	lastErr      error
}

// NewBaseView creates a new BaseView with the given name and title.
func NewBaseView(name, title string) BaseView {
	return BaseView{
		name:  name,
		title: title,
	}
}

func (b *BaseView) Name() string    { return b.name }
func (b *BaseView) Title() string   { return b.title }
func (b *BaseView) Hints() []string { return nil }

func (b *BaseView) Init(ctx context.Context) error {
	b.ctx = ctx
	b.initialized.Store(true)
	return nil
}

func (b *BaseView) Render() tview.Primitive { return nil }
func (b *BaseView) Refresh() error          { return nil }

func (b *BaseView) OnKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (b *BaseView) OnFocus() error {
	b.focused.Store(true)
	return nil
}

func (b *BaseView) OnLoseFocus() error {
	b.focused.Store(false)
	return nil
}

func (b *BaseView) Stop() error { return nil }

func (b *BaseView) SetSwitchViewFn(fn func(string)) {
	b.switchViewFn = fn
}

func (b *BaseView) SwitchToView(name string) {
	if b.switchViewFn != nil {
		b.switchViewFn(name)
	}
}

// SetApp sets the tview application reference.
func (b *BaseView) SetApp(app *tview.Application) { b.app = app }

// SetPages sets the tview pages reference.
func (b *BaseView) SetPages(pages *tview.Pages) { b.pages = pages }

// SetViewManager sets the view manager reference.
func (b *BaseView) SetViewManager(mgr *ViewManager) { b.viewMgr = mgr }

// App returns the tview application.
func (b *BaseView) App() *tview.Application { return b.app }

// Pages returns the tview pages.
func (b *BaseView) Pages() *tview.Pages { return b.pages }

// Ctx returns the view context.
func (b *BaseView) Ctx() context.Context { return b.ctx }

// IsRefreshing reports whether the view is currently refreshing.
func (b *BaseView) IsRefreshing() bool { return b.refreshing.Load() }

// SetRefreshing sets the refreshing state.
func (b *BaseView) SetRefreshing(v bool) { b.refreshing.Store(v) }

// IsFocused reports whether the view currently has focus.
func (b *BaseView) IsFocused() bool { return b.focused.Load() }

// LastError returns the last error encountered, or nil.
func (b *BaseView) LastError() error {
	b.lastErrMu.Lock()
	defer b.lastErrMu.Unlock()
	return b.lastErr
}

// SetLastError stores the last error. Pass nil to clear.
func (b *BaseView) SetLastError(err error) {
	b.lastErrMu.Lock()
	defer b.lastErrMu.Unlock()
	b.lastErr = err
}

// ViewManager manages the set of registered views and handles view switching.
type ViewManager struct {
	views       map[string]View
	viewOrder   []string
	currentView string
	mu          sync.RWMutex
}

// NewViewManager creates a new ViewManager.
func NewViewManager() *ViewManager {
	return &ViewManager{
		views: make(map[string]View),
	}
}

// Register adds a view to the manager. Views are displayed in registration order.
func (m *ViewManager) Register(view View) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := view.Name()
	m.views[name] = view
	m.viewOrder = append(m.viewOrder, name)
	if m.currentView == "" {
		m.currentView = name
	}
}

// SetCurrentView switches to the named view, calling OnLoseFocus on the
// current view and OnFocus on the new view.
func (m *ViewManager) SetCurrentView(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.views[name]; !ok {
		return nil
	}

	if m.currentView != "" {
		if cur, ok := m.views[m.currentView]; ok {
			if err := cur.OnLoseFocus(); err != nil {
				return err
			}
		}
	}

	m.currentView = name

	if next, ok := m.views[name]; ok {
		if err := next.OnFocus(); err != nil {
			return err
		}
	}
	return nil
}

// NextView cycles to the next view (Tab behavior).
func (m *ViewManager) NextView() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.viewOrder) == 0 {
		return ""
	}
	idx := m.indexOf(m.currentView)
	next := (idx + 1) % len(m.viewOrder)
	return m.viewOrder[next]
}

// PreviousView cycles to the previous view (Shift+Tab behavior).
func (m *ViewManager) PreviousView() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.viewOrder) == 0 {
		return ""
	}
	idx := m.indexOf(m.currentView)
	prev := (idx - 1 + len(m.viewOrder)) % len(m.viewOrder)
	return m.viewOrder[prev]
}

// CurrentView returns the currently active view, or nil.
func (m *ViewManager) CurrentView() View {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.views[m.currentView]
}

// CurrentViewName returns the name of the currently active view.
func (m *ViewManager) CurrentViewName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentView
}

// GetView returns the named view, or nil.
func (m *ViewManager) GetView(name string) View {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.views[name]
}

// Views returns all registered views in registration order.
func (m *ViewManager) Views() []View {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]View, 0, len(m.viewOrder))
	for _, name := range m.viewOrder {
		result = append(result, m.views[name])
	}
	return result
}

// ViewNames returns the ordered view names.
func (m *ViewManager) ViewNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.viewOrder))
	copy(out, m.viewOrder)
	return out
}

func (m *ViewManager) indexOf(name string) int {
	for i, n := range m.viewOrder {
		if n == name {
			return i
		}
	}
	return 0
}
