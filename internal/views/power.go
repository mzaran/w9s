package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
)

// NewPowerView creates a ResourceView for IPMI power management.
// Should only be registered when client.HasPower() is true.
func NewPowerView(app *tview.Application, client dao.WarewulfClient) View {
	return NewResourceView[*dao.WwNode]("power", "Power", app, ResourceViewConfig[*dao.WwNode]{
		Fetch: func() (map[string]*dao.WwNode, error) {
			return client.Nodes().List()
		},
		Columns: []Column[*dao.WwNode]{
			{Name: "Name", Width: 20, Extract: func(name string, _ *dao.WwNode) string { return name }},
			{Name: "IPMI Address", Width: 16, Extract: func(_ string, n *dao.WwNode) string {
				if n.Ipmi != nil {
					return n.Ipmi.Ipaddr
				}
				return ""
			}},
			{Name: "IPMI User", Width: 15, Extract: func(_ string, n *dao.WwNode) string {
				if n.Ipmi != nil {
					return n.Ipmi.UserName
				}
				return ""
			}},
			{Name: "IPMI Interface", Width: 15, Extract: func(_ string, n *dao.WwNode) string {
				if n.Ipmi != nil {
					return n.Ipmi.Interface
				}
				return ""
			}},
		},
		Actions: []Action[*dao.WwNode]{
			{Key: 'o', Label: "Power On", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Power().On(name)
			}},
			{Key: 'f', Label: "Power Off", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Power().Off(name)
			}},
			{Key: 'c', Label: "Cycle", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Power().Cycle(name)
			}},
			{Key: 'r', Label: "Reset", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				return client.Power().Reset(name)
			}},
			{Key: 's', Label: "Status", Execute: func(ctx context.Context, name string, _ *dao.WwNode) error {
				_, err := client.Power().Status(name)
				return err
			}},
		},
		Detail: func(name string, n *dao.WwNode) string {
			var b strings.Builder
			fmt.Fprintf(&b, "Node: %s\n", name)
			if n.Ipmi != nil {
				fmt.Fprintf(&b, "IPMI Address: %s\n", n.Ipmi.Ipaddr)
				fmt.Fprintf(&b, "IPMI User: %s\n", n.Ipmi.UserName)
				fmt.Fprintf(&b, "IPMI Interface: %s\n", n.Ipmi.Interface)
				fmt.Fprintf(&b, "IPMI Gateway: %s\n", n.Ipmi.Gateway)
				fmt.Fprintf(&b, "IPMI Netmask: %s\n", n.Ipmi.Netmask)
			}
			return b.String()
		},
	})
}
