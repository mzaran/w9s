// Package app implements the w9s TUI application lifecycle.
package app

import (
	"context"
	"time"

	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/ui"
	"github.com/mzaran/w9s/internal/views"
)

// App is the main w9s TUI application.
type App struct {
	ctx           context.Context
	cancel        context.CancelFunc
	client        dao.WarewulfClient
	tviewApp      *tview.Application
	viewMgr       *views.ViewManager
	pages         *tview.Pages
	header        *ui.Header
	statusBar     *ui.StatusBar
	mainLayout    *tview.Flex
	commandInput  *tview.InputField
	commandActive bool
	config        *config.Config
	refreshTicker *time.Ticker
}

// New creates a new App instance.
func New(client dao.WarewulfClient, cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	tviewApp := tview.NewApplication()
	if cfg.UI.EnableMouse {
		tviewApp.EnableMouse(true)
	}

	a := &App{
		ctx:      ctx,
		cancel:   cancel,
		client:   client,
		tviewApp: tviewApp,
		config:   cfg,
	}

	a.initUI()
	a.registerViews()
	a.setupKeyboard()

	return a, nil
}

// Client returns the Warewulf client.
func (a *App) Client() dao.WarewulfClient {
	return a.client
}

// Config returns the app configuration.
func (a *App) Config() *config.Config {
	return a.config
}

// TviewApp returns the underlying tview application.
func (a *App) TviewApp() *tview.Application {
	return a.tviewApp
}

// ViewMgr returns the view manager.
func (a *App) ViewMgr() *views.ViewManager {
	return a.viewMgr
}

// Context returns the app context.
func (a *App) Context() context.Context {
	return a.ctx
}

// initUI creates the main layout: header + pages + status bar.
func (a *App) initUI() {
	theme := &ui.DefaultTheme

	// Header: info (1 row) + tabs (1 row) = 2 rows.
	a.header = ui.NewHeader(theme)

	// Content pages.
	a.pages = tview.NewPages()

	// Status bar (1 row).
	a.statusBar = ui.NewStatusBar(theme)
	a.statusBar.SetApp(a.tviewApp)
	a.statusBar.SetHints([]string{"q Quit", "Tab Next", "? Help", ": Command", "1-7 Views"})

	// Set cluster info in header.
	if a.config.ActiveCluster != nil {
		a.header.SetClusterInfo(a.config.ActiveCluster.Endpoint, "")
	}

	// Command bar (hidden initially, 0 height).
	a.initCommandBar()

	// Main layout: vertical flex.
	a.mainLayout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 2, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false).
		AddItem(a.commandInput, 0, 0, false)

	a.viewMgr = views.NewViewManager()

	a.tviewApp.SetRoot(a.mainLayout, true)
}

// updateHeader refreshes the header bar to show the current view tab.
func (a *App) updateHeader() {
	if a.viewMgr == nil {
		return
	}
	a.header.SetCurrentView(a.viewMgr.CurrentViewName())
}

// updateStatusBar updates the status bar hints for the current view.
func (a *App) updateStatusBar() {
	if a.viewMgr == nil {
		return
	}
	current := a.viewMgr.CurrentView()
	if current != nil {
		hints := current.Hints()
		if hints == nil {
			hints = []string{"q Quit", "Tab Next View", "? Help"}
		}
		a.statusBar.SetHints(hints)
	}
}

// switchToView changes the active view by name.
func (a *App) switchToView(name string) {
	if err := a.viewMgr.SetCurrentView(name); err != nil {
		return
	}
	a.pages.SwitchToPage(name)
	a.updateHeader()
	a.updateStatusBar()
}

// Run starts the TUI application. Blocks until the app exits.
func (a *App) Run() error {
	// Parse refresh rate.
	refreshDuration := 5 * time.Second
	if a.config.RefreshRate != "" {
		if d, err := time.ParseDuration(a.config.RefreshRate); err == nil {
			refreshDuration = d
		}
	}

	// Start refresh ticker.
	a.refreshTicker = time.NewTicker(refreshDuration)
	go func() {
		for {
			select {
			case <-a.refreshTicker.C:
				if v := a.viewMgr.CurrentView(); v != nil {
					a.tviewApp.QueueUpdateDraw(func() {
						_ = v.Refresh()
					})
				}
			case <-a.ctx.Done():
				return
			}
		}
	}()

	// Set initial view to first registered view.
	viewNames := a.viewMgr.ViewNames()
	if len(viewNames) > 0 {
		a.switchToView(viewNames[0])
	}

	// Set the header view tabs.
	a.header.SetViews(viewNames)
	a.updateHeader()

	// Run the TUI (blocks).
	return a.tviewApp.Run()
}

// Stop gracefully shuts down the application.
func (a *App) Stop() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}

	// Stop all views.
	for _, v := range a.viewMgr.Views() {
		_ = v.Stop()
	}

	// Close the client.
	if a.client != nil {
		_ = a.client.Close()
	}

	a.cancel()
	a.tviewApp.Stop()
}

// initView initializes a view and registers it with the view manager and pages.
func (a *App) initView(v views.View) {
	if err := v.Init(a.ctx); err != nil {
		return
	}
	v.SetSwitchViewFn(func(name string) {
		a.tviewApp.QueueUpdateDraw(func() {
			a.switchToView(name)
		})
	})

	// If the view embeds BaseView, set its pages and viewMgr references
	// so views can show modals and navigate.
	type pagesAware interface {
		SetPages(*tview.Pages)
	}
	if pa, ok := v.(pagesAware); ok {
		pa.SetPages(a.pages)
	}
	type viewMgrAware interface {
		SetViewManager(*views.ViewManager)
	}
	if va, ok := v.(viewMgrAware); ok {
		va.SetViewManager(a.viewMgr)
	}

	// Wire status message callback for flash messages.
	type statusMessenger interface {
		SetStatusMessageFn(func(string, bool))
	}
	if sm, ok := v.(statusMessenger); ok {
		sm.SetStatusMessageFn(func(msg string, isError bool) {
			if isError {
				a.statusBar.ShowError(msg)
			} else {
				a.statusBar.ShowSuccess(msg)
			}
		})
	}

	a.viewMgr.Register(v)
	prim := v.Render()
	if prim != nil {
		a.pages.AddPage(v.Name(), prim, true, false)
	}
}

