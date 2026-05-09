# Architecture

## Elm Architecture Pattern

This app follows Bubble Tea's Elm architecture:

```
Model (state) ──► View (render) ──► User Input ──► Update (new state + commands)
   ▲                                                              │
   └──────────────── Commands complete ───────────────────────────┘
```

## Data Flow

```
User presses 'r'
    │
    ▼
update.go: tea.KeyMsg "r"
    │
    ▼
update.go: fetchStatus()
    │
    ▼
commands.go: cmdRunner.Run("tailscale", "serve", "status", "--json")
    │
    ▼
parser.go: ParseStatus(output) → TailscaleStatus
    │
    ▼
parser.go: ServicesFromStatus(status) → []Service
    │
    ▼
update.go: statusMsg → update Model.services + Model.list.SetItems()
    │
    ▼
view.go: View() reads Model.services + Model.list → terminal output
```

## Module Dependencies

```
main.go
    ├──► model.go
    ├──► update.go ──► commands.go
    │       │              │
    │       │              ├──► parser.go
    │       │              │       │
    │       │              │       └──► model.go (TailscaleStatus, Service)
    │       │              │
    │       │              └──► model.go (Service)
    │       │
    │       └──► model.go (Model, Service)
    │
    ├──► view.go ──► styles.go
    │       │
    │       └──► model.go (Model, Service)
    │
    └──► styles.go
```

**No circular dependencies.** The dependency graph is a DAG.

## State Machine

```
[Loading]
    │ fetchStatus()
    ▼
[List] ◄────────────────────────┐
    │                           │
    │ r key                     │
    ▼                           │
[Loading]                       │
    │ fetchStatus()             │
    ▼                           │
[List] ─────────────────────────┘
    │
    │ Enter/s key
    ▼
[Confirm]
    │ y key
    ▼
[Loading] ──► stopService() ──► [List]
    │
    │ n/esc key
    ▼
[List]
    │
    │ ?/h key
    ▼
[Help] ──► ?/h key ──► [List]
```

## Key Design Decisions

### Why Flatten the JSON?

Tailscale's JSON is nested:
```
Web → serviceName → Handlers → path → Proxy
```

The Bubble Tea `list` component requires a flat `[]list.Item`. We flatten each handler into a standalone `Service` struct. This means one Tailscale Web entry with 3 handlers becomes 3 `Service` rows.

### Why commandRunner Interface?

All external commands (tailscale, clipboard) go through the `commandRunner` interface. In production, `realCommandRunner` uses `os/exec`. In tests, `mockCommandRunner` records calls and returns stubbed output. This allows 100% unit test coverage without requiring Tailscale to be installed.

### Why No Config File?

The app is intentionally stateless. It reads everything from `tailscale serve status --json` on startup and refresh. No persistent state means no config drift and simpler code.

### Responsive Column Math

The table has 4 columns: indicator, protocol badge, service name, path→target, URL. Column widths are calculated dynamically based on terminal width using `clamp(width - offset, min, max)`. This prevents columns from becoming unreadably narrow while maximizing space usage.

## Message Types and Handling

| Message | Sent By | Handled In | Action |
|---------|---------|-----------|--------|
| `tea.WindowSizeMsg` | Terminal | `update.go` | Resize list, recalculate columns |
| `tea.KeyMsg` | Keyboard | `update.go` | Route to dialog, list, or global handler |
| `statusMsg` | `fetchStatus()` | `update.go` | Parse → flatten → update list items |
| `stopCompleteMsg` | `stopService()` | `update.go` | Show success/error, trigger refresh |
| `clearSuccessMsg` | `copyToClipboard()` | `update.go` | No-op (success shown immediately) |
| `refreshCompleteMsg` | `refreshList()` | `update.go` | No-op (loading state already set) |

## Dialog Priority

When rendering, dialog visibility is checked in this order:
1. Help dialog (`showHelp`) — fullscreen overlay
2. Confirm dialog (`showConfirm`) — centered modal
3. Main view — table or error/empty state

When handling keyboard input, dialog state gates list navigation:
- If `showConfirm` is true: only y/n/esc are processed
- If `showHelp` is true: only ?/h/q are processed
- Otherwise: list navigation + action keys work
