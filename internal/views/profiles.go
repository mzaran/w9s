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

// ProfilesView displays profiles with an inheritance tree panel.
type ProfilesView struct {
	BaseView

	client    dao.WarewulfClient
	profiles  map[string]*dao.WwProfile
	sorted    []string
	mu        sync.RWMutex
	table     *tview.Table
	tree      *tview.TreeView
	container *tview.Flex
	modalOpen bool
}

// NewProfilesView creates a new profiles view.
func NewProfilesView(app *tview.Application, client dao.WarewulfClient) *ProfilesView {
	pv := &ProfilesView{
		BaseView: NewBaseView("profiles", "Profiles"),
		client:   client,
		profiles: make(map[string]*dao.WwProfile),
	}
	pv.SetApp(app)
	return pv
}

func (pv *ProfilesView) Init(ctx context.Context) error {
	if err := pv.BaseView.Init(ctx); err != nil {
		return err
	}

	pv.table = tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSeparator(tview.Borders.Vertical)
	pv.table.SetBorder(false)
	pv.table.SetSelectionChangedFunc(func(row, _ int) {
		pv.onSelectionChanged(row)
	})

	pv.tree = tview.NewTreeView()
	pv.tree.SetBorder(true).SetTitle(" Inheritance ")

	tablePanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pv.table, 0, 1, true)

	pv.container = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tablePanel, 0, 3, true).
		AddItem(pv.tree, 0, 1, false)

	return nil
}

func (pv *ProfilesView) Render() tview.Primitive {
	return pv.container
}

func (pv *ProfilesView) Hints() []string {
	return []string{"r Refresh", "Enter Detail", "/ Filter"}
}

// IsFiltering reports whether a modal is open (implements Filterable).
func (pv *ProfilesView) IsFiltering() bool {
	return pv.modalOpen
}

func (pv *ProfilesView) OnKey(event *tcell.EventKey) *tcell.EventKey {
	if pv.modalOpen {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			pv.modalOpen = false
			pv.Pages().RemovePage("profile-detail")
			pv.Pages().SwitchToPage(pv.Name())
			pv.App().SetFocus(pv.table)
			return nil
		}
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		pv.showDetail()
		return nil
	case tcell.KeyRune:
		if event.Rune() == 'r' {
			_ = pv.Refresh()
			return nil
		}
	}
	return event
}

func (pv *ProfilesView) showDetail() {
	pv.mu.RLock()
	row, _ := pv.table.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(pv.sorted) {
		pv.mu.RUnlock()
		return
	}
	name := pv.sorted[idx]
	p := pv.profiles[name]
	pv.mu.RUnlock()

	var b strings.Builder
	fmt.Fprintf(&b, "Profile: %s\n", name)
	fmt.Fprintf(&b, "Comment: %s\n", p.Comment)
	fmt.Fprintf(&b, "Cluster: %s\n", p.ClusterName)
	fmt.Fprintf(&b, "Image: %s\n", p.ImageName)
	fmt.Fprintf(&b, "Parent Profiles: %s\n", strings.Join(p.Profiles, ", "))
	fmt.Fprintf(&b, "System Overlays: %s\n", strings.Join(p.SystemOverlay, ", "))
	fmt.Fprintf(&b, "Runtime Overlays: %s\n", strings.Join(p.RuntimeOverlay, ", "))
	if p.Kernel != nil {
		fmt.Fprintf(&b, "Kernel: %s\n", p.Kernel.Version)
		fmt.Fprintf(&b, "Kernel Args: %s\n", strings.Join(p.Kernel.Args, " "))
	}
	fmt.Fprintf(&b, "Init: %s\n", p.Init)
	fmt.Fprintf(&b, "Root: %s\n", p.Root)

	pv.modalOpen = true
	ui.ShowDetail(pv.Pages(), pv.App(), "profile-detail", name, b.String(), func() {
		pv.modalOpen = false
		pv.Pages().RemovePage("profile-detail")
		pv.Pages().SwitchToPage(pv.Name())
		pv.App().SetFocus(pv.table)
	})
}

func (pv *ProfilesView) OnFocus() error {
	if err := pv.BaseView.OnFocus(); err != nil {
		return err
	}
	if pv.table != nil {
		pv.App().SetFocus(pv.table)
	}
	return pv.Refresh()
}

func (pv *ProfilesView) Refresh() error {
	if pv.IsRefreshing() {
		return nil
	}
	pv.SetRefreshing(true)

	go func() {
		defer pv.SetRefreshing(false)

		profiles, err := pv.client.Profiles().List()
		if err != nil {
			pv.SetLastError(err)
			return
		}

		pv.mu.Lock()
		pv.profiles = profiles
		pv.sorted = make([]string, 0, len(profiles))
		for k := range profiles {
			pv.sorted = append(pv.sorted, k)
		}
		sort.Strings(pv.sorted)
		pv.mu.Unlock()

		pv.App().QueueUpdateDraw(func() {
			pv.renderTable()
		})
	}()

	return nil
}

func (pv *ProfilesView) renderTable() {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	pv.table.Clear()

	headers := []string{"NAME", "IMAGE", "SYSTEM OVERLAYS", "RUNTIME OVERLAYS", "PARENT PROFILES", "COMMENT"}
	for col, h := range headers {
		pv.table.SetCell(0, col, tview.NewTableCell(h).
			SetSelectable(false).
			SetExpansion(1).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold))
	}

	for row, name := range pv.sorted {
		p := pv.profiles[name]
		vals := []string{
			name,
			p.ImageName,
			strings.Join(p.SystemOverlay, ","),
			strings.Join(p.RuntimeOverlay, ","),
			strings.Join(p.Profiles, ","),
			p.Comment,
		}
		for col, v := range vals {
			pv.table.SetCell(row+1, col, tview.NewTableCell(v).SetExpansion(1))
		}
	}

	if len(pv.sorted) > 0 {
		pv.table.Select(1, 0)
	}
}

func (pv *ProfilesView) onSelectionChanged(row int) {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	idx := row - 1
	if idx < 0 || idx >= len(pv.sorted) {
		return
	}

	name := pv.sorted[idx]
	root := tview.NewTreeNode(name).SetColor(tcell.ColorGreen)
	pv.buildInheritanceTree(root, name, make(map[string]bool))
	pv.tree.SetRoot(root).SetCurrentNode(root)
}

func (pv *ProfilesView) buildInheritanceTree(node *tview.TreeNode, name string, visited map[string]bool) {
	if visited[name] {
		node.AddChild(tview.NewTreeNode("(circular)").SetColor(tcell.ColorRed))
		return
	}
	visited[name] = true

	p, ok := pv.profiles[name]
	if !ok {
		return
	}

	for _, parent := range p.Profiles {
		child := tview.NewTreeNode(parent).SetColor(tcell.ColorWhite)
		pv.buildInheritanceTree(child, parent, visited)
		node.AddChild(child)
	}

	// Show key fields inline.
	if p.ImageName != "" {
		node.AddChild(tview.NewTreeNode(fmt.Sprintf("image: %s", p.ImageName)).
			SetColor(tcell.ColorDarkCyan).SetSelectable(false))
	}
	if len(p.SystemOverlay) > 0 {
		node.AddChild(tview.NewTreeNode(fmt.Sprintf("sys: %s", strings.Join(p.SystemOverlay, ","))).
			SetColor(tcell.ColorDarkCyan).SetSelectable(false))
	}
}
