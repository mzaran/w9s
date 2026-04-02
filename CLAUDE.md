# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

w9s (w9s.sh) — Terminal UI for Warewulf HPC cluster management. Apache 2.0. Module: `github.com/mzaran/w9s`.

## Build Commands

```bash
make build              # Build → build/w9s (static, CGO_ENABLED=0)
make test               # Run all tests
make lint               # golangci-lint
make fmt                # gofumpt
make ci                 # lint + test + build
make dev                # Run with race detector + --debug
make mock               # W9S_ENABLE_MOCK=1 ./w9s --mock
```

Run single test: `go test -v -run TestName ./internal/dao/...`

## Architecture

Layered design following k9s/s9s patterns:

```
cmd/w9s/main.go → cli.Execute()
  → internal/cli/root.go        Cobra CLI, config load, mock/real client
    → internal/app/app.go       tview lifecycle: initUI → registerViews → Run → Stop
      → internal/views/         View interface + ViewManager + ResourceView[T] generic
      → internal/ui/            Header, StatusBar, Table, confirm dialogs
      → internal/dao/           WarewulfClient interface + HTTP client + cache
      → internal/config/        Viper YAML + env overrides + auto-detect
pkg/mock/                       Full mock client with sample HPC cluster data
```

### Key Pattern: ResourceView[T] Generic

The DRY core. Each table-based view is ~50 lines of configuration instead of ~1500:

```go
// Concrete view = config, not code
func NewNodesView(app *tview.Application, client dao.WarewulfClient) View {
    return NewResourceView[*dao.WwNode]("nodes", "Nodes", app, ResourceViewConfig[*dao.WwNode]{
        Fetch:   func() (map[string]*dao.WwNode, error) { return client.Nodes().List() },
        Columns: []Column[*dao.WwNode]{ /* column definitions */ },
        Actions: []Action[*dao.WwNode]{ /* keyboard actions */ },
    })
}
```

Custom views (Dashboard, Profiles, Overlays) implement View directly via BaseView when they need non-table rendering.

### DAO Layer

`WarewulfClient` interface with sub-managers mapping 1:1 to warewulf REST API:
- NodeManager → `/api/nodes/`
- ProfileManager → `/api/profiles/`
- ImageManager → `/api/images/`
- OverlayManager → `/api/overlays/`
- PowerManager → `wwctl power` shell-out (conditionally available)

HTTP client uses functional options: `NewHTTPClient(WithEndpoint(...), WithBasicAuth(...), WithInsecure(true))`

### Warewulf API Quirks

- JSON field names contain spaces: `"cluster name"`, `"image name"`, `"network devices"`, `"runtime overlay"`, `"system overlay"`, `"asset key"`
- `WWBool` type serializes as string: `"true"`, `"yes"`, `"1"`, `"UNDEF"`, `""` — not JSON boolean
- API disabled by default (`config.api.enabled: false` in warewulf.conf)
- IP allowlist defaults to localhost only — remote clients need subnet added
- No pagination — all resources returned in one response

### Config

Location: `~/.config/w9s/config.yaml` (XDG) or `~/.w9s/config.yaml`
Env overrides: `W9S_ENDPOINT`, `W9S_USERNAME`, `W9S_PASSWORD`
Auto-detects `/etc/warewulf/warewulf.conf` and `wwctl` on PATH.
Mock mode: `W9S_ENABLE_MOCK=1 ./w9s --mock`
