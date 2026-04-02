package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
)

// HeaderMetricsFunc is a callback to update header metrics.
type HeaderMetricsFunc func(nodes, up, images, profiles, overlays int)

// DashboardView shows a cluster summary with aggregated statistics.
type DashboardView struct {
	BaseView

	client          dao.WarewulfClient
	container       *tview.Flex
	summary         *tview.TextView
	nodeStats       *tview.TextView
	imageList       *tview.TextView
	headerMetricsFn HeaderMetricsFunc
}

// SetHeaderMetricsFn sets the callback used to update header metrics after refresh.
func (dv *DashboardView) SetHeaderMetricsFn(fn HeaderMetricsFunc) {
	dv.headerMetricsFn = fn
}

// NewDashboardView creates a new dashboard view.
func NewDashboardView(app *tview.Application, client dao.WarewulfClient) *DashboardView {
	dv := &DashboardView{
		BaseView: NewBaseView("dashboard", "Dashboard"),
		client:   client,
	}
	dv.SetApp(app)
	return dv
}

func (dv *DashboardView) Init(ctx context.Context) error {
	if err := dv.BaseView.Init(ctx); err != nil {
		return err
	}

	dv.summary = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	dv.summary.SetBorder(true).SetTitle(" Cluster Summary ")

	dv.nodeStats = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	dv.nodeStats.SetBorder(true).SetTitle(" Nodes ")

	dv.imageList = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	dv.imageList.SetBorder(true).SetTitle(" Images ")

	top := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(dv.summary, 0, 1, false).
		AddItem(dv.nodeStats, 0, 1, false)

	dv.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(top, 0, 1, false).
		AddItem(dv.imageList, 0, 1, false)

	return nil
}

func (dv *DashboardView) Render() tview.Primitive {
	return dv.container
}

func (dv *DashboardView) Hints() []string {
	return []string{"r Refresh", "Tab Next View"}
}

func (dv *DashboardView) OnKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune && event.Rune() == 'r' {
		_ = dv.Refresh()
		return nil
	}
	return event
}

func (dv *DashboardView) OnFocus() error {
	if err := dv.BaseView.OnFocus(); err != nil {
		return err
	}
	return dv.Refresh()
}

func (dv *DashboardView) Refresh() error {
	if dv.IsRefreshing() {
		return nil
	}
	dv.SetRefreshing(true)

	go func() {
		defer dv.SetRefreshing(false)

		nodes, nodeErr := dv.client.Nodes().List()
		images, imgErr := dv.client.Images().List()
		overlays, ovlErr := dv.client.Overlays().List()

		dv.App().QueueUpdateDraw(func() {
			dv.renderSummary(nodes, nodeErr, images, imgErr, overlays, ovlErr)
			dv.renderNodeStats(nodes)
			dv.renderImageList(images)

			// Update header metrics.
			if dv.headerMetricsFn != nil {
				nodeCount, imgCount, ovlCount := 0, 0, 0
				if nodeErr == nil {
					nodeCount = len(nodes)
				}
				if imgErr == nil {
					imgCount = len(images)
				}
				if ovlErr == nil {
					ovlCount = len(overlays)
				}
				dv.headerMetricsFn(nodeCount, nodeCount, imgCount, 0, ovlCount)
			}
		})
	}()

	return nil
}

func (dv *DashboardView) renderSummary(
	nodes map[string]*dao.WwNode, nodeErr error,
	images map[string]*dao.WwImage, imgErr error,
	overlays map[string]*dao.WwOverlay, ovlErr error,
) {
	dv.summary.Clear()
	var b strings.Builder

	nodeCount := 0
	if nodeErr == nil {
		nodeCount = len(nodes)
	}
	imgCount := 0
	if imgErr == nil {
		imgCount = len(images)
	}
	ovlCount := 0
	if ovlErr == nil {
		ovlCount = len(overlays)
	}

	fmt.Fprintf(&b, "[yellow]Total Nodes:[-]    %d\n", nodeCount)
	fmt.Fprintf(&b, "[yellow]Total Images:[-]   %d\n", imgCount)
	fmt.Fprintf(&b, "[yellow]Total Overlays:[-] %d\n", ovlCount)

	if nodeErr != nil {
		fmt.Fprintf(&b, "\n[red]Node error: %v[-]\n", nodeErr)
	}
	if imgErr != nil {
		fmt.Fprintf(&b, "\n[red]Image error: %v[-]\n", imgErr)
	}

	fmt.Fprint(dv.summary, b.String())
}

func (dv *DashboardView) renderNodeStats(nodes map[string]*dao.WwNode) {
	dv.nodeStats.Clear()
	if nodes == nil {
		return
	}
	var b strings.Builder

	// Nodes by image.
	byImage := make(map[string]int)
	byProfileCount := make(map[int]int)
	for _, n := range nodes {
		img := n.ImageName
		if img == "" {
			img = "(none)"
		}
		byImage[img]++
		byProfileCount[len(n.Profiles)]++
	}

	fmt.Fprintf(&b, "[yellow]By Image:[-]\n")
	for _, img := range sortedMapKeys(byImage) {
		fmt.Fprintf(&b, "  %-25s %d\n", img, byImage[img])
	}

	fmt.Fprintf(&b, "\n[yellow]By Profile Count:[-]\n")
	counts := make([]int, 0, len(byProfileCount))
	for c := range byProfileCount {
		counts = append(counts, c)
	}
	sort.Ints(counts)
	for _, c := range counts {
		fmt.Fprintf(&b, "  %d profiles: %d nodes\n", c, byProfileCount[c])
	}

	fmt.Fprint(dv.nodeStats, b.String())
}

func (dv *DashboardView) renderImageList(images map[string]*dao.WwImage) {
	dv.imageList.Clear()
	if images == nil {
		return
	}
	var b strings.Builder

	fmt.Fprintf(&b, "[yellow]%-30s  %-12s  %-8s[-]\n", "NAME", "SIZE", "KERNELS")
	for _, name := range sortedMapKeys(images) {
		img := images[name]
		fmt.Fprintf(&b, "%-30s  %-12s  %d\n", name, humanSize(img.Size), len(img.Kernels))
	}

	fmt.Fprint(dv.imageList, b.String())
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
