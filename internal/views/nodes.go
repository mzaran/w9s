package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/ui"
)

// NewNodesView creates a ResourceView for Warewulf nodes.
func NewNodesView(app *tview.Application, client dao.WarewulfClient) View {
	return NewResourceView[*dao.WwNode]("nodes", "Nodes", app, ResourceViewConfig[*dao.WwNode]{
		Fetch: func() (map[string]*dao.WwNode, error) {
			return client.Nodes().List()
		},
		Columns: []Column[*dao.WwNode]{
			{Name: "Name", Width: 20, Extract: func(name string, _ *dao.WwNode) string { return name }},
			{Name: "Cluster", Width: 15, Extract: func(_ string, n *dao.WwNode) string { return n.ClusterName }},
			{Name: "Image", Width: 20, Extract: func(_ string, n *dao.WwNode) string { return n.ImageName }},
			{Name: "Profiles", Width: 20, Extract: func(_ string, n *dao.WwNode) string { return strings.Join(n.Profiles, ",") }},
			{Name: "Primary IP", Width: 16, Extract: func(_ string, n *dao.WwNode) string { return primaryIP(n) }},
			{Name: "MAC", Width: 18, Extract: func(_ string, n *dao.WwNode) string { return primaryMAC(n) }},
			{Name: "Overlays", Width: 20, Extract: func(_ string, n *dao.WwNode) string { return strings.Join(n.SystemOverlay, ",") }},
			{Name: "Status", Width: 12, Extract: func(_ string, n *dao.WwNode) string {
				return nodeStatus(n)
			}},
		},
		FetchRaw: func(name string) (any, error) {
			return client.Nodes().Get(name)
		},
		ExtraHints: []string{"a Add", "e Edit"},
		OnKeyExtra: func(rv *ResourceView[*dao.WwNode], event *tcell.EventKey) *tcell.EventKey {
			if event.Key() != tcell.KeyRune {
				return event
			}
			switch event.Rune() {
			case 'a':
				showNodeAddForm(rv, client)
				return nil
			case 'e':
				showNodeEditForm(rv, client)
				return nil
			}
			return event
		},
		Actions: []Action[*dao.WwNode]{
			{Key: 'd', Label: "Delete", Destructive: true, Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Nodes().Delete(name)
			}},
			{Key: 'b', Label: "Build Overlays", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Nodes().BuildOverlays(name)
			}},
		},
		Detail: func(name string, n *dao.WwNode) string {
			var b strings.Builder
			fmt.Fprintf(&b, "Node: %s\n", name)
			fmt.Fprintf(&b, "Cluster: %s\n", n.ClusterName)
			fmt.Fprintf(&b, "Image: %s\n", n.ImageName)
			fmt.Fprintf(&b, "Profiles: %s\n", strings.Join(n.Profiles, ", "))
			fmt.Fprintf(&b, "System Overlays: %s\n", strings.Join(n.SystemOverlay, ", "))
			fmt.Fprintf(&b, "Runtime Overlays: %s\n", strings.Join(n.RuntimeOverlay, ", "))
			fmt.Fprintf(&b, "Primary IP: %s\n", primaryIP(n))
			fmt.Fprintf(&b, "Primary MAC: %s\n", primaryMAC(n))
			fmt.Fprintf(&b, "Comment: %s\n", n.Comment)
			fmt.Fprintf(&b, "Discoverable: %s\n", n.Discoverable)
			if n.Ipmi != nil {
				fmt.Fprintf(&b, "IPMI Address: %s\n", n.Ipmi.Ipaddr)
				fmt.Fprintf(&b, "IPMI User: %s\n", n.Ipmi.UserName)
			}
			return b.String()
		},
	})
}

func primaryIP(n *dao.WwNode) string {
	if n.PrimaryNetDev != "" {
		if nd, ok := n.NetDevs[n.PrimaryNetDev]; ok {
			return nd.Ipaddr
		}
	}
	for _, nd := range n.NetDevs {
		return nd.Ipaddr
	}
	return ""
}

func primaryMAC(n *dao.WwNode) string {
	if n.PrimaryNetDev != "" {
		if nd, ok := n.NetDevs[n.PrimaryNetDev]; ok {
			return nd.Hwaddr
		}
	}
	for _, nd := range n.NetDevs {
		return nd.Hwaddr
	}
	return ""
}

func nodeStatus(n *dao.WwNode) string {
	if n.ImageName == "" {
		return "● No Image"
	}
	if len(n.SystemOverlay) > 0 || len(n.RuntimeOverlay) > 0 {
		return "● Built"
	}
	return "● Ready"
}

func showNodeAddForm(rv *ResourceView[*dao.WwNode], client dao.WarewulfClient) {
	rv.modalOpen = true
	fields := []ui.FormField{
		{Key: "name", Label: "Node Name", Width: 30},
		{Key: "profile", Label: "Profile", Default: "default", Width: 30},
		{Key: "ip", Label: "IP Address", Width: 20},
		{Key: "mac", Label: "MAC Address", Width: 20},
	}
	ui.ShowForm(rv.Pages(), rv.App(), "node-add", rv.Name(), "Add Node", fields,
		func(values map[string]string) {
			name := values["name"]
			if name == "" {
				return
			}
			node := &dao.WwNode{}
			node.Profiles = []string{values["profile"]}
			if values["ip"] != "" || values["mac"] != "" {
				node.NetDevs = map[string]*dao.NetDev{
					"default": {
						Ipaddr: values["ip"],
						Hwaddr: values["mac"],
					},
				}
				node.PrimaryNetDev = "default"
			}
			go func() {
				if err := client.Nodes().Add(name, node); err != nil {
					rv.ShowStatusError("Add node failed: " + err.Error())
				} else {
					rv.ShowStatusMessage("Node '" + name + "' added")
				}
				_ = rv.Refresh()
			}()
		},
		func() {
			rv.modalOpen = false
			rv.App().SetFocus(rv.table)
		},
	)
}

func showNodeEditForm(rv *ResourceView[*dao.WwNode], client dao.WarewulfClient) {
	name, node, ok := rv.selectedItem()
	if !ok {
		return
	}
	rv.modalOpen = true
	fields := []ui.FormField{
		{Key: "profile", Label: "Profile", Default: strings.Join(node.Profiles, ","), Width: 30},
		{Key: "cluster", Label: "Cluster", Default: node.ClusterName, Width: 30},
		{Key: "image", Label: "Image", Default: node.ImageName, Width: 30},
		{Key: "ip", Label: "IP Address", Default: primaryIP(node), Width: 20},
		{Key: "mac", Label: "MAC Address", Default: primaryMAC(node), Width: 20},
		{Key: "comment", Label: "Comment", Default: node.Comment, Width: 40},
	}
	ui.ShowForm(rv.Pages(), rv.App(), "node-edit", rv.Name(), fmt.Sprintf("Edit Node: %s", name), fields,
		func(values map[string]string) {
			updated := &dao.WwNode{}
			if values["profile"] != "" {
				updated.Profiles = strings.Split(values["profile"], ",")
			}
			updated.ClusterName = values["cluster"]
			updated.ImageName = values["image"]
			updated.Comment = values["comment"]
			if values["ip"] != "" || values["mac"] != "" {
				updated.NetDevs = map[string]*dao.NetDev{
					"default": {
						Ipaddr: values["ip"],
						Hwaddr: values["mac"],
					},
				}
				updated.PrimaryNetDev = "default"
			}
			go func() {
				if err := client.Nodes().Update(name, updated); err != nil {
					rv.ShowStatusError("Edit failed: " + err.Error())
				} else {
					rv.ShowStatusMessage("Node '" + name + "' updated")
				}
				_ = rv.Refresh()
			}()
		},
		func() {
			rv.modalOpen = false
			rv.App().SetFocus(rv.table)
		},
	)
}
