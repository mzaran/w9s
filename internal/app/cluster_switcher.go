package app

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
)

const clusterSwitcherPage = "cluster-switcher"

// showClusterSwitcher displays a list of configured clusters for switching.
func (a *App) showClusterSwitcher() {
	if len(a.config.Clusters) < 2 {
		a.statusBar.ShowError("Only one cluster configured")
		return
	}

	list := tview.NewList()
	list.SetBorder(true).SetTitle(" Switch Cluster ")
	list.SetBackgroundColor(tcell.ColorDefault)

	for i, c := range a.config.Clusters {
		label := c.Name
		secondary := c.Cluster.Endpoint
		if a.config.ActiveCluster != nil && a.config.ActiveCluster.Endpoint == c.Cluster.Endpoint {
			label += " (active)"
		}
		idx := i
		list.AddItem(label, secondary, rune('1'+i), func() {
			a.pages.RemovePage(clusterSwitcherPage)
			a.pages.SwitchToPage(a.viewMgr.CurrentViewName())

			if err := a.swapCluster(idx); err != nil {
				a.statusBar.ShowError("Switch failed: " + err.Error())
			} else {
				a.statusBar.ShowSuccess("Switched to " + a.config.Clusters[idx].Name)
			}
		})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.pages.RemovePage(clusterSwitcherPage)
			a.pages.SwitchToPage(a.viewMgr.CurrentViewName())
			return nil
		}
		return event
	})

	a.pages.AddAndSwitchToPage(clusterSwitcherPage, list, true)
	a.tviewApp.SetFocus(list)
}

// swapCluster connects to a different configured cluster.
func (a *App) swapCluster(idx int) error {
	if idx < 0 || idx >= len(a.config.Clusters) {
		return fmt.Errorf("invalid cluster index: %d", idx)
	}
	target := a.config.Clusters[idx]

	// For mock mode, just update config and refresh.
	if target.Cluster.Endpoint == "" || len(target.Cluster.Endpoint) > 4 && target.Cluster.Endpoint[:7] == "mock://" {
		a.config.ActiveCluster = &a.config.Clusters[idx].Cluster
		a.header.SetClusterInfo(target.Cluster.Endpoint, "")
		if v := a.viewMgr.CurrentView(); v != nil {
			_ = v.Refresh()
		}
		return nil
	}

	// Create new client, verify connectivity before closing old.
	timeout := 30 * time.Second
	if target.Cluster.Timeout != "" {
		if d, err := time.ParseDuration(target.Cluster.Timeout); err == nil {
			timeout = d
		}
	}
	newClient, err := dao.NewClientFromConfig(
		target.Cluster.Endpoint, target.Cluster.Username,
		target.Cluster.Password, target.Cluster.Insecure, timeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Verify reachable.
	if _, err := newClient.Nodes().List(); err != nil {
		newClient.Close()
		return fmt.Errorf("cluster unreachable: %w", err)
	}

	// Swap.
	a.client.Close()
	a.client = newClient
	a.config.ActiveCluster = &a.config.Clusters[idx].Cluster
	a.header.SetClusterInfo(target.Cluster.Endpoint, "")

	// Refresh current view.
	if v := a.viewMgr.CurrentView(); v != nil {
		_ = v.Refresh()
	}

	return nil
}
