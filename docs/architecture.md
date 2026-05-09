# Architecture

This document describes how Tailscale Portal works under the hood: the data model, parsing logic, and UI architecture.

## Overview

Tailscale Portal is a [Bubble Tea](https://github.com/charmbracelet/bubbletea) application that follows the Elm architecture pattern:

```
Model → View (renders UI)
  ↑         ↓
Update ← Msg (keyboard, async commands)
```

1. **Model** holds the application state
2. **View** renders the state to the terminal
3. **Update** handles messages (keyboard input, command completions) and returns a new model + commands
4. **Commands** run asynchronously and send messages back to Update

## Data Flow

```
tailscale serve status --json
        │
        ▼
   [commands.go]  ──►  [parser.go]
        │                    │
        │              Parse JSON into
        │              TailscaleStatus struct
        │                    │
        │                    ▼
        │         ServicesFromStatus()
        │         Flatten into []Service
        │                    │
        ▼                    ▼
   statusMsg ──────────►  [update.go]
                              │
                              ▼
                        Update model.list
                              │
                              ▼
                         [view.go]
                        Render UI
```

## Tailscale JSON Format

The `tailscale serve status --json` command returns a nested structure:

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

### TCP Section

Maps port numbers to protocol configuration:

| Field | Type | Meaning |
|-------|------|---------|
| `TCP` | `map[string]TCPConfig` | Port → config mapping |
| `TCPConfig.HTTPS` | `bool` | Whether this port uses HTTPS |

### Web Section

Maps service identifiers (hostname:port) to handler configurations:

| Field | Type | Meaning |
|-------|------|---------|
| `Web` | `map[string]WebConfig` | Service name → web config |
| `WebConfig.Handlers` | `map[string]HandlerConfig` | Path → handler mapping |
| `HandlerConfig.Proxy` | `string` | Backend URL (e.g. `http://localhost:3000`) |

## Internal Data Model

### Service Struct

The nested Tailscale JSON is flattened into a linear list of `Service` objects:

```go
type Service struct {
    ServiceName string // hostname:port (the Web map key)
    Hostname    string // Just the hostname part
    Port        string // Just the port part
    Protocol    string // "HTTPS" or "HTTP" (looked up from TCP section)
    Target      string // Proxy target (e.g. http://localhost:5173)
    Path        string // Handler path (e.g. "/" or "/api")
    FullURL     string // Generated access URL
}
```

**Why flatten?** The Bubble Tea `list` component expects a flat slice of items. The nested structure (Web → Handlers) is unrolled so each handler becomes its own row in the UI.

### Model Struct

```go
type Model struct {
    services     []Service       // Flattened service list
    list         list.Model      // Bubble Tea list component
    lastRefresh  time.Time       // Timestamp for status bar
    loading      bool            // Show loading spinner/message
    err          error           // Last error (shown in UI)
    successMsg   string          // Ephemeral success notification
    showHelp     bool            // Help dialog visible?
    showConfirm  bool            // Stop confirmation dialog visible?
    confirmIndex int             // Which service to stop
    confirmMsg   string          // Confirmation dialog text
    width        int             // Terminal width
    height       int             // Terminal height
}
```

## Parsing Logic

### ParseStatus

`ParseStatus(data []byte)` unmarshals the raw JSON into `TailscaleStatus`. Returns an error if the JSON is malformed.

### ServicesFromStatus

`ServicesFromStatus(status TailscaleStatus)` performs the flattening:

1. Iterate over `status.Web` (each entry is a service name → web config)
2. Parse the service name (`hostname:port`) into separate fields
3. Look up the port in `status.TCP` to determine HTTP vs HTTPS
4. Iterate over handlers within the web config
5. Build the full URL using `buildFullURL()`
6. Append a `Service` to the result slice

### URL Building

`buildFullURL()` constructs the access URL:

```
scheme://hostname:port/path
```

- Scheme is `https` for HTTPS services, `http` otherwise
- Path is prepended with `/` if it doesn't start with one

## Command Execution

### commandRunner Interface

To make the code testable, external CLI calls go through an interface:

```go
type commandRunner interface {
    Run(name string, arg ...string) ([]byte, error)
    RunWithStdin(name string, stdin string, arg ...string) error
}
```

**Production**: `realCommandRunner` uses `os/exec`
**Testing**: `mockCommandRunner` records calls and returns stubbed output

### Tailscale Commands

| Function | Command | Purpose |
|----------|---------|---------|
| `fetchStatus()` | `tailscale serve status --json` | Get current serve configuration |
| `stopService(port, protocol)` | `tailscale serve --https=<port> off` or `--http=<port> off` | Stop a service |

### Clipboard Commands

`copyToClipboard()` tries clipboard tools in order:

1. `pbcopy` (macOS)
2. `xclip -selection clipboard` (Linux X11)
3. `wl-copy` (Linux Wayland)

## UI Architecture

### Responsive Layout

The view adapts to terminal size using a clamp function:

```
total width = terminal width
service column   = clamp(width-60, 20, 35)
path column      = clamp(width-service-45, 20, 40)
url column       = width - service - path - 18 (minimum 10)
```

This ensures columns don't get too narrow while using available space.

### Rendering Pipeline

```
View()
├── renderHeader()        # Title + last refresh time + success message
├── renderTableHeader()   # PROTO | SERVICE | PATH → TARGET | URL
├── list.View()           # Bubble Tea list (delegated to serviceDelegate)
│   └── serviceDelegate.Render()
│       ├── Selection indicator (▸)
│       ├── Protocol badge (green=HTTPS, orange=HTTP)
│       ├── Service name (hostname:port)
│       ├── Path → Target
│       └── Full URL
└── renderFooter()        # Key hints
```

### Dialogs

Two modal dialogs overlay the main view:

- **Help dialog** (`showHelp`): Centered box with all key bindings
- **Stop confirmation** (`showConfirm`): Centered box asking "Stop service X?" with y/n options

Both use `lipgloss.Place()` for centering and have rounded borders.

### Styling

All visual styles are defined in `styles.go` using a centralized color palette:

```
Background:  #1a1a1a    (dark)
Foreground:  #e0e0e0    (light gray)
Primary:     #7B68EE    (purple - title, selection)
Success:     #50C878    (green - HTTPS, success messages)
Error:       #FF6B6B    (red - errors)
Warning:     #FFA500    (orange - HTTP, warnings)
Info:        #4ECDC4    (teal - paths)
Muted:       #666666    (gray - footer, URLs)
Highlight:   #FFE66D    (yellow - keys, selected items)
```

## Message Types

Bubble Tea messages used in the application:

| Message | Source | Purpose |
|---------|--------|---------|
| `tea.WindowSizeMsg` | Terminal | Window resized |
| `tea.KeyMsg` | Keyboard | Key pressed |
| `statusMsg` | `fetchStatus()` | Tailscale status loaded |
| `refreshCompleteMsg` | `refreshList()` | Refresh animation done |
| `stopCompleteMsg` | `stopService()` | Service stopped |
| `clearSuccessMsg` | `copyToClipboard()` | Clipboard operation done |
| `tickMsg` | Timer | Periodic UI updates |

## Testing Strategy

### Unit Tests

- **Parser tests** (`parser_test.go`): JSON parsing edge cases, malformed input, empty status
- **Command tests** (`commands_test.go`): Mock runner verifies correct CLI arguments are passed
- **Model tests** (`model_test.go`): Initial model state, message handling
- **View tests** (`view_test.go`): Rendering with various states (empty, error, loading)
- **Styles tests** (`styles_test.go`): Style definitions
- **Update tests** (`update_test.go`): Keyboard handling, state transitions

### Mocking

The `mockCommandRunner` records every call:

```go
mock := &mockCommandRunner{
    RunFunc: func(name string, arg ...string) ([]byte, error) {
        return []byte(`{"TCP":{},"Web":{}}`), nil
    },
}
SetCommandRunner(mock)
// ... run code ...
calls := mock.getRunCalls()
// Verify tailscale serve status --json was called
```

This allows testing without requiring Tailscale to be installed.
