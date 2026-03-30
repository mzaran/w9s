# w9s — Terminal UI for Warewulf

<p align="center">
  <strong>w9s</strong> — a terminal interface for managing <a href="https://warewulf.org">Warewulf</a> HPC clusters
</p>

<p align="center">
  <a href="https://github.com/mzaran/w9s/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/mzaran/w9s/ci.yml?branch=main&style=flat-square&label=CI" alt="CI"></a>
  <a href="https://github.com/mzaran/w9s/releases/latest"><img src="https://img.shields.io/github/v/release/mzaran/w9s?style=flat-square" alt="Release"></a>
  <img src="https://img.shields.io/badge/go-1.24-007d9c?style=flat-square&logo=go&logoColor=white" alt="Go 1.24">
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square" alt="License"></a>
</p>

w9s provides a real-time terminal UI for Warewulf cluster provisioning and management. Monitor nodes, manage profiles, container images, and overlays — all from your terminal. Inspired by [k9s](https://k9scli.io/) and [s9s](https://github.com/jontk/s9s).

<p align="center">
  <img src="docs/screenshots/w9s-demo.gif" alt="w9s demo" width="800">
</p>

## Installation

### Go Install

```bash
go install github.com/mzaran/w9s/cmd/w9s@latest
```

### Binary Download

Download the latest release from the [releases page](https://github.com/mzaran/w9s/releases).

```bash
curl -LO https://github.com/mzaran/w9s/releases/download/v0.1.0/w9s_0.1.0_Linux_x86_64.tar.gz
tar -xzf w9s_0.1.0_Linux_x86_64.tar.gz
sudo mv w9s /usr/local/bin/
```

### RPM / DEB

```bash
# RPM (Fedora, RHEL, Rocky, Alma)
sudo rpm -i w9s_0.1.0_linux_amd64.rpm

# DEB (Debian, Ubuntu)
sudo dpkg -i w9s_0.1.0_linux_amd64.deb
```

### Docker

```bash
docker run --rm -it ghcr.io/mzaran/w9s:latest
```

### Build from Source

```bash
git clone https://github.com/mzaran/w9s.git
cd w9s
make build
sudo mv build/w9s /usr/local/bin/
```

## Quick Start

```bash
# Connect to a Warewulf server
w9s --server https://warewulf.example.com:9873

# Use mock mode for testing without a Warewulf server
W9S_ENABLE_MOCK=1 w9s --mock

# Show version
w9s version
```

## Configuration

w9s reads configuration from `~/.config/w9s/config.yaml`, `./config.yaml`, or environment variables.

```yaml
server:
  url: https://warewulf.example.com:9873
  token: "your-api-token"
  tls_skip_verify: false

ui:
  refresh_interval: 5s
  theme: default
```

Environment variables use the `W9S_` prefix:

| Variable | Description |
|---|---|
| `W9S_SERVER_URL` | Warewulf API endpoint |
| `W9S_SERVER_TOKEN` | Authentication token |
| `W9S_ENABLE_MOCK` | Set to `1` for mock mode |
| `W9S_LOG_LEVEL` | Log level (debug, info, warn, error) |

See [config.example.yaml](config.example.yaml) for a full reference.

## Views

| View | Description |
|---|---|
| **Dashboard** | Cluster overview with node status summary |
| **Nodes** | Node list with provisioning status, hardware info, and actions |
| **Profiles** | Warewulf node profiles and assignments |
| **Images** | Container images used for node provisioning |
| **Overlays** | System and runtime overlay management |
| **Power** | IPMI power control (on, off, cycle, status) |
| **Help** | Keyboard shortcuts and command reference |

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `1`-`7` | Switch views |
| `/` | Filter / search |
| `Enter` | View details |
| `d` | Delete selected |
| `e` | Edit selected |
| `r` | Refresh |
| `?` | Help |
| `q` | Quit |

## Warewulf Server Requirements

w9s communicates with Warewulf via its REST API. Ensure your Warewulf server is configured with:

- **API enabled**: `warewulfd` running with API access
- **TLS** (recommended): configure certificates for secure communication
- **Authentication**: API token or user credentials

Refer to the [Warewulf documentation](https://warewulf.org/docs/) for server setup.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Run tests (`make ci`)
4. Submit a pull request

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

---

[w9s.sh](https://w9s.sh)
