# Tailscale Portal

A beautiful terminal dashboard for managing your Tailscale serve ports. No more squinting at `tailscale serve status` output — see all your services at a glance, copy URLs instantly, and stop services with a keystroke.

```
┌────────────────────────────────────────────── Tailscale Portal ──────────────────────────────────────────────┐
│ 5 services active                                                                                              │
│                                                                                                                │
│  > dashboard    :8443   HTTPS  -> http://localhost:3000          https://myhost.ts.net:8443                    │
│    api          :8080   HTTPS  -> http://localhost:8080          https://myhost.ts.net:8080                    │
│    docs         :3001   HTTP   -> http://localhost:3001          http://myhost.ts.net:3001                     │
│    blog         :443    HTTPS  -> http://localhost:4000          https://myhost.ts.net                         │
│    metrics      :9090   HTTPS  -> http://localhost:9090          https://myhost.ts.net:9090                    │
│                                                                                                                │
│  ↑/↓ navigate  •  Enter/s stop  •  c copy URL  •  r refresh  •  ? help  •  q quit                              │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## What is this?

If you use [Tailscale serve](https://tailscale.com/kb/1242/tailscale-serve) to expose local services over your tailnet, you've probably typed `tailscale serve status` a hundred times. **Tailscale Portal** gives you a live, interactive view of all your served ports in a clean terminal interface.

**Perfect for:**
- Quickly seeing which ports are being served
- Copying URLs to share with teammates
- Cleaning up stale services without memoring CLI flags

## Installation

### macOS / Linux (Homebrew)

```bash
brew tap YOUR_GITHUB_USERNAME/tap
brew install tailscale-portal
```

### Debian / Ubuntu

```bash
wget https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal/releases/latest/download/tailscale-portal_linux_amd64.deb
sudo dpkg -i tailscale-portal_linux_amd64.deb
```

### Fedora / RHEL / CentOS

```bash
wget https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal/releases/latest/download/tailscale-portal_linux_amd64.rpm
sudo rpm -i tailscale-portal_linux_amd64.rpm
```

### Alpine Linux

```bash
wget https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal/releases/latest/download/tailscale-portal_linux_amd64.apk
sudo apk add --allow-untrusted tailscale-portal_linux_amd64.apk
```

### Direct Download

Grab the binary for your platform from the [latest release](https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal/releases/latest) and move it to your `PATH`:

```bash
# macOS (Apple Silicon)
curl -L -o tailscale-portal.tar.gz \
  https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal/releases/latest/download/tailscale-portal_darwin_arm64.tar.gz
tar -xzf tailscale-portal.tar.gz
sudo mv tailscale-portal /usr/local/bin/
```

### Build from Source

Requires Go 1.21+:

```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal.git
cd tailscale-portal
go build -o tailscale-portal
```

## Quick Start

Just run it:

```bash
tailscale-portal
```

That's it. It opens a full-screen dashboard showing all your active Tailscale serve services.

## Controls

| Key | What it does |
|-----|-------------|
| `↑` `↓` or `j` `k` | Move up / down the list |
| `Enter` or `s` | Stop the selected service |
| `c` | Copy the service URL to your clipboard |
| `r` or `R` | Refresh the list |
| `?` or `h` | Show / hide help |
| `q` or `Ctrl+C` | Quit |

### Stopping a Service

1. Navigate to the service you want to stop
2. Press `Enter` or `s`
3. Confirm with `y`
4. The list refreshes automatically

### Copying a URL

1. Navigate to the service
2. Press `c`
3. The full URL (e.g. `https://yourhost.ts.net:8443`) is now in your clipboard

## Requirements

- [Tailscale](https://tailscale.com/download) installed and logged in
- The `tailscale` CLI in your `PATH`
- At least one active `tailscale serve` configuration (or the list will be empty — that's normal!)

## Troubleshooting

**"tailscale CLI not found"**

Tailscale needs to be installed and the `tailscale` command available in your shell. Install it from [tailscale.com/download](https://tailscale.com/download) if you haven't already.

**Clipboard doesn't work**

The app tries to use your system's clipboard tool:
- **macOS:** `pbcopy` (built-in)
- **Linux (X11):** `xclip` — install with `sudo apt install xclip`
- **Linux (Wayland):** `wl-copy` — install with `sudo apt install wl-clipboard`

**No services are showing**

If the list is empty, you likely don't have any `tailscale serve` configurations active. Check with:

```bash
tailscale serve status
```

To serve a local app:

```bash
tailscale serve --https=443 --set-path=/ http://localhost:3000
```

Then run `tailscale-portal` again and it will appear.

## Documentation

For developers and contributors:

- [`docs/development.md`](docs/development.md) — Development setup, project structure, running tests
- [`docs/architecture.md`](docs/architecture.md) — Data model, parsing logic, UI architecture
- [`docs/releasing.md`](docs/releasing.md) — CI/CD pipeline, GoReleaser config, creating releases

## License

MIT
