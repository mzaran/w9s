package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
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
				if n.ImageName != "" && len(n.Profiles) > 0 {
					return "● Ready"
				}
				return "● Pending"
			}},
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
