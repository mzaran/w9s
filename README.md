# W9s -- The Warewulf Dashboard For Your Terminal

W9s provides a terminal UI to manage your [Warewulf](https://github.com/warewulf/warewulf) HPC clusters --
a stylish companion to `wwctl`. It connects to the Warewulf v4 REST API (v4.4+, recommended v4.6.1+) and gives you a real-time dashboard for cluster provisioning and management. Inspired by [k9s](https://k9scli.io/) and [s9s](https://github.com/jontk/s9s).

<p align="center">
  <img src="docs/screenshots/w9s-demo.gif" alt="w9s demo" width="800">
</p>

---

## Installation

### Go Install

```bash
go install github.com/mzaran/w9s/cmd/w9s@latest
```

### Binary Download

```bash
curl -LO https://github.com/mzaran/w9s/releases/latest/download/w9s_Linux_x86_64.tar.gz
tar -xzf w9s_Linux_x86_64.tar.gz
sudo mv w9s /usr/local/bin/
```

### RPM / DEB

```bash
sudo rpm -i w9s_*.rpm       # Fedora, RHEL, Rocky
sudo dpkg -i w9s_*.deb      # Debian, Ubuntu
```

### Build from Source

```bash
git clone https://github.com/mzaran/w9s.git
cd w9s
make build
./build/w9s --version
```

---

## Getting Started

### 1. Configure

Create `~/.config/w9s/config.yaml`:

```yaml
refreshRate: "5s"
defaultCluster: "production"
clusters:
  - name: production
    cluster:
      endpoint: "http://warewulf-server:9873"
      username: "admin"
      password: "changeme"
      timeout: "10s"
  - name: staging
    cluster:
      endpoint: "http://ww-staging:9873"
      username: "admin"
      password: "changeme"
ui:
  enableMouse: true
  skin: "dracula"
```

### 2. Run

```bash
w9s                                    # connect to configured cluster
W9S_ENABLE_MOCK=1 w9s --mock          # mock mode (no server needed)
w9s version                            # show version
```

### 3. Environment Variables

| Variable | Description |
|----------|-------------|
| `W9S_ENDPOINT` | Warewulf API endpoint (overrides config) |
| `W9S_USERNAME` | API username |
| `W9S_PASSWORD` | API password |
| `W9S_ENABLE_MOCK` | Set to `1` to allow `--mock` flag |

---

## Features

**Full CRUD** -- Add, edit, and delete nodes and profiles directly from the TUI. Form-based dialogs guide you through field entry with validation, so you never have to remember YAML key names or shell out to `wwctl`.

**Image Management** -- Import OCI container images, trigger builds, and delete images you no longer need. The image view shows sizes, build status, and timestamps at a glance.

**Overlay Browser** -- Drill into overlay file trees, view file contents, and render `.ww` Go templates for specific nodes. See exactly what a node will receive at boot time without running `wwctl overlay show` in a loop.

**Column Sorting and Live Filter** -- Sort any table by any column with `s`/`S`. Press `/` for fuzzy search across all visible columns -- results update as you type.

**CSV Export** -- Press `x` to export the current table view to a CSV file. Useful for audits, inventory reports, or feeding data into other tools.

**Multi-Cluster** -- Configure multiple Warewulf clusters and switch between them with `Shift+C`. Each cluster maintains its own endpoint, credentials, and connection state.

**Keyboard-First** -- Every action is reachable from the keyboard. Works over SSH, in tmux, on headless servers. No browser, no port forwarding required.

**Read-Only Mode** -- Launch with `--readonly` to disable all write operations. Safe for monitoring dashboards and shared terminal sessions.

**Single Static Binary** -- Built with `CGO_ENABLED=0` and zero runtime dependencies. Drop it on any Linux box and run.

---

## Views

| # | View | Description |
|---|------|-------------|
| 1 | **Dashboard** | Cluster overview -- node counts, image summary, profile stats |
| 2 | **Nodes** | Node table with CRUD, sort, filter, detail, overlay build |
| 3 | **Profiles** | Profile list with inheritance tree, add/delete |
| 4 | **Images** | Image list with import, build, delete |
| 5 | **Overlays** | Overlay browser -- drill into files, render `.ww` templates |
| 6 | **Power** | IPMI power control (requires `wwctl` on PATH) |
| 7 | **Help** | Keyboard shortcuts reference |

---

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `1`-`7` | Jump to view |
| `Tab` / `Shift+Tab` | Next / previous view |
| `C` (Shift+C) | Switch cluster |
| `?` | Help |
| `q` | Quit |

### Table Views (Nodes, Images, Power)

| Key | Action |
|-----|--------|
| `/` | Open filter (Escape to close) |
| `s` | Sort by next column |
| `S` | Reverse sort |
| `Escape` | Reset sort / clear filter |
| `Enter` | View detail (scrollable pane) |
| `x` | Export table to CSV |
| `r` | Refresh data |

### Nodes

| Key | Action |
|-----|--------|
| `a` | Add node (form) |
| `e` | Edit node (form) |
| `d` | Delete node (confirmation) |
| `b` | Build overlays |

### Profiles

| Key | Action |
|-----|--------|
| `a` | Add profile (form) |
| `d` | Delete profile (confirmation) |
| `Enter` | View profile detail |

### Images

| Key | Action |
|-----|--------|
| `i` | Import from OCI registry |
| `b` | Build image |
| `d` | Delete image (confirmation) |

### Overlays

| Key | Action |
|-----|--------|
| `Enter` | Drill into files / view file content |
| `t` | Render `.ww` template for a node |
| `Escape` | Back to overlay list |

---

## Command Bar

Press `:` to open the command bar. Type a view name or number to switch views:

| Command | Action |
|---------|--------|
| `dashboard` / `1` | Switch to Dashboard |
| `nodes` / `2` | Switch to Nodes |
| `profiles` / `3` | Switch to Profiles |
| `images` / `4` | Switch to Images |
| `overlays` / `5` | Switch to Overlays |
| `power` / `6` | Switch to Power |
| `help` / `7` | Switch to Help |
| `q` / `quit` | Quit |

---

## Themes

w9s ships with 5 built-in skins: `default`, `dracula`, `gruvbox`, `nord`, `solarized`.

Set a skin in your config:

```yaml
ui:
  skin: "dracula"
```

### Custom Skins

Create `~/.config/w9s/skins/mytheme.yaml` with `name`, `header`, `table`, `status`, `border`, and `accent` color keys, then set `ui.skin: "mytheme"` in your config.

---

## CLI Reference

### `w9s`

Launch the TUI connected to your configured cluster.

| Flag | Description |
|------|-------------|
| `--config` | Path to config file (default: `~/.config/w9s/config.yaml`) |
| `--readonly` | Disable all write operations (add, edit, delete, build, import) |
| `--debug` | Enable debug logging |
| `--mock` | Use mock data (requires `W9S_ENABLE_MOCK=1`) |

### `w9s version`

Display version with ASCII logo.

```
w9s version        # full output with logo
w9s version -s     # short, just the version string
```

### `w9s info`

Display version, config file paths, active cluster, and skin.

---

## Configuration

Config file location: `~/.config/w9s/config.yaml` (XDG) or `~/.w9s/config.yaml`. Environment variables (`W9S_ENDPOINT`, `W9S_USERNAME`, `W9S_PASSWORD`) override config file values. See [Getting Started](#getting-started) for a full config example.

---

## Warewulf Server Setup

w9s connects to the [Warewulf v4 REST API](https://github.com/warewulf/warewulf) (v4.4+, recommended v4.6.1+). Enable it on your server:

```yaml
# /etc/warewulf/warewulf.conf
api:
  enabled: true
  tls: false          # set to true on v4.7+ for TLS
  allowed subnets:
    - "192.168.1.0/24"
    - "127.0.0.0/8"
    - "::1/128"
```

```bash
# Create API credentials
htpasswd -Bc /etc/warewulf/auth.conf admin

# Restart
sudo systemctl restart warewulfd

# Verify
curl -u admin:password http://localhost:9873/api/nodes/
```

> **TLS (Warewulf v4.7+):** Set `tls: true` in warewulf.conf and configure certificates. w9s will connect over HTTPS automatically. Use `--insecure` to skip certificate verification for self-signed certs.

---

## Troubleshooting

### Debug Logging

```bash
w9s --debug     # verbose logging to stderr
```

### Connection Issues

- Verify the API is enabled: `grep -A2 'api:' /etc/warewulf/warewulf.conf`
- Check credentials: `curl -u admin:pass http://server:9873/api/nodes/`
- Ensure your subnet is allowed in `warewulf.conf` under `api.allowed subnets`
- For TLS issues, try `--insecure` to skip certificate verification

### Mock Mode

Test w9s without a live server:

```bash
W9S_ENABLE_MOCK=1 w9s --mock
```

---

## Development

```bash
make build          # Build binary
make test           # Run all tests
make lint           # golangci-lint
make ci             # lint + test + build
make mock           # Run in mock mode
make dev            # Run with race detector

# Smoke tests (requires tmux)
./test/smoke/verify-views.sh
```

---

## Architecture

```
cmd/w9s/main.go -> cli.Execute()
  -> internal/cli/         Cobra CLI, config loading, mock/real client
  -> internal/app/         tview lifecycle, keyboard handling, cluster switcher
  -> internal/views/       View interface, ResourceView[T] generic, 7 views
  -> internal/ui/          Theme, Header, StatusBar, Table, Detail, Form, Confirm
  -> internal/dao/         WarewulfClient interface, HTTP client, cache, power
  -> internal/config/      Viper YAML, env overrides, auto-detect
  -> pkg/mock/             Full mock client with sample HPC cluster data
```

Key pattern: `ResourceView[T]` -- each table-based view is ~50 lines of column/action configuration instead of ~1500 lines of repeated boilerplate. Custom views (Dashboard, Profiles, Overlays) implement the `View` interface directly.

---

## License

Apache 2.0 -- see [LICENSE](LICENSE) for details.

---

[w9s.sh](https://w9s.sh)
