# Tailscale Portal

A TUI dashboard for managing Tailscale serve ports. View active services, copy URLs, and stop services with an intuitive terminal interface.

## Features

- **Service Overview**: View all active Tailscale serve services in a clean, scrollable list
- **Service Details**: See port, protocol (HTTP/HTTPS), target, path, and full URL for each service
- **Quick Actions**:
  - Copy service URLs to clipboard
  - Stop/clear services with confirmation
  - Refresh to get latest status
- **Keyboard Navigation**: Vim-style (j/k) and arrow key support
- **Dark Mode**: Beautiful dark theme with color-coded protocols

## Installation

### Prerequisites

- Go 1.21 or later
- Tailscale installed and configured
- `tailscale` CLI in your PATH

### Build from Source

```bash
git clone <repository>
cd tailscale-portal
go build -o tailscale-portal
```

Or run directly:

```bash
go run .
```

## Usage

Simply run the binary:

```bash
./tailscale-portal
```

### Key Bindings

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate up/down |
| `Enter` or `s` | Stop the selected service |
| `c` | Copy service URL to clipboard |
| `r` or `R` | Refresh service list |
| `?` or `h` | Toggle help dialog |
| `q` or `Ctrl+C` | Quit |

### Stopping a Service

1. Navigate to the service you want to stop
2. Press `Enter` or `s`
3. Confirm by pressing `y`
4. The list will automatically refresh

## Data Model

The app parses output from `tailscale serve status --json`:

```json
{
  "TCP": {
    "3000": { "HTTPS": true },
    "443": { "HTTPS": true }
  },
  "Web": {
    "hostname.ts.net:3000": {
      "Handlers": {
        "/": { "Proxy": "http://localhost:5173" }
      }
    }
  }
}
```

## Project Structure

```
tailscale-portal/
├── go.mod          # Go module definition
├── main.go         # Application entry point
├── model.go        # Bubble Tea model and data structures
├── update.go       # Message handling and state updates
├── view.go         # UI rendering functions
├── parser.go       # JSON parsing logic
├── commands.go     # Tailscale CLI command wrappers
├── styles.go       # Lipgloss styling definitions
└── README.md       # This file
```

## Troubleshooting

### "tailscale CLI not found"

Make sure Tailscale is installed and the `tailscale` command is in your PATH.

Install Tailscale:
- macOS: `brew install tailscale`
- Linux: https://tailscale.com/download/linux
- Windows: https://tailscale.com/download/windows

### Clipboard not working

The app tries multiple clipboard tools:
- macOS: `pbcopy`
- Linux (X11): `xclip`
- Linux (Wayland): `wl-copy`

Install the appropriate tool for your system.

### No services shown

If no services appear, ensure you have active Tailscale serve configurations:

```bash
tailscale serve status
```

To add a service:

```bash
tailscale serve --https=443 --set-path=/ http://localhost:3000
```

## License

MIT
