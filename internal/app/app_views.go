package app

import "github.com/mzaran/w9s/internal/views"

// registerViews creates and registers all TUI views with the view manager.
// Views are registered in display order, which maps to number key shortcuts 1-7.
func (a *App) registerViews() {
	// Dashboard: cluster overview.
	a.initView(views.NewDashboardView(a.tviewApp, a.client))

	// Nodes: node management table.
	a.initView(views.NewNodesView(a.tviewApp, a.client))

	// Profiles: profile management with inheritance tree.
	a.initView(views.NewProfilesView(a.tviewApp, a.client))

	// Images: container image management.
	a.initView(views.NewImagesView(a.tviewApp, a.client))

	// Overlays: overlay management with file browsing.
	a.initView(views.NewOverlaysView(a.tviewApp, a.client))

	// Power: IPMI power management (only if client supports it).
	if a.client.HasPower() {
		a.initView(views.NewPowerView(a.tviewApp, a.client))
	}

	// Help: keyboard shortcuts and documentation.
	a.initView(views.NewHelpView(a.tviewApp))
}
