package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/ui"
)

// NewImagesView creates a ResourceView for Warewulf container images.
func NewImagesView(app *tview.Application, client dao.WarewulfClient) View {
	return NewResourceView[*dao.WwImage]("images", "Images", app, ResourceViewConfig[*dao.WwImage]{
		Fetch: func() (map[string]*dao.WwImage, error) {
			return client.Images().List()
		},
		Columns: []Column[*dao.WwImage]{
			{Name: "Name", Width: 30, Extract: func(name string, _ *dao.WwImage) string { return name }},
			{Name: "Size", Width: 12, Extract: func(_ string, img *dao.WwImage) string { return humanSize(img.Size) }},
			{Name: "Kernels", Width: 10, Extract: func(_ string, img *dao.WwImage) string { return fmt.Sprintf("%d", len(img.Kernels)) }},
			{Name: "Build Time", Width: 22, Extract: func(_ string, img *dao.WwImage) string {
				if img.BuildTime == 0 {
					return "never"
				}
				return time.Unix(img.BuildTime, 0).Format("2006-01-02 15:04:05")
			}},
			{Name: "Writable", Width: 10, Extract: func(_ string, img *dao.WwImage) string {
				if img.Writable {
					return "yes"
				}
				return "no"
			}},
		},
		OnKeyExtra: func(rv *ResourceView[*dao.WwImage], event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyRune && event.Rune() == 'i' {
				showImageImportForm(rv, client)
				return nil
			}
			return event
		},
		Actions: []Action[*dao.WwImage]{
			{Key: 'b', Label: "Build", Execute: func(ctx context.Context, name string, _ *dao.WwImage) error {
				return client.Images().Build(name)
			}},
			{Key: 'd', Label: "Delete", Destructive: true, Execute: func(ctx context.Context, name string, _ *dao.WwImage) error {
				return client.Images().Delete(name)
			}},
		},
		Detail: func(name string, img *dao.WwImage) string {
			var b strings.Builder
			fmt.Fprintf(&b, "Image: %s\n", name)
			fmt.Fprintf(&b, "Size: %s\n", humanSize(img.Size))
			fmt.Fprintf(&b, "Writable: %v\n", img.Writable)
			bt := "never"
			if img.BuildTime > 0 {
				bt = time.Unix(img.BuildTime, 0).Format(time.RFC3339)
			}
			fmt.Fprintf(&b, "Build Time: %s\n", bt)
			fmt.Fprintf(&b, "Kernels: %s\n", strings.Join(img.Kernels, ", "))
			return b.String()
		},
	})
}

func showImageImportForm(rv *ResourceView[*dao.WwImage], client dao.WarewulfClient) {
	rv.modalOpen = true
	fields := []ui.FormField{
		{Key: "name", Label: "Image Name", Width: 30},
		{Key: "source", Label: "OCI Source", Default: "docker://", Width: 50},
	}
	ui.ShowForm(rv.Pages(), rv.App(), "image-import", rv.Name(), "Import Image", fields,
		func(values map[string]string) {
			name := values["name"]
			source := values["source"]
			if name == "" || source == "" {
				return
			}
			go func() {
				if err := client.Images().Import(name, source); err != nil {
					rv.SetLastError(err)
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

func humanSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
