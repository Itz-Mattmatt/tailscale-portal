# Project Context

## What This Is

Tailscale Portal is a terminal UI (TUI) application built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework. It provides a dashboard for viewing and managing Tailscale serve ports.

## Core Purpose

Parse `tailscale serve status --json` output and display it as an interactive, scrollable table in the terminal. Users can copy URLs to clipboard and stop services without memorizing CLI flags.

## Target Users

Developers and DevOps engineers who use Tailscale serve to expose local services.

## Key Constraints

1. **Single binary** — No configuration files, no database, no daemon. Everything is stateless.
2. **Depends on `tailscale` CLI** — The app shells out to `tailscale serve status --json` and `tailscale serve --https=<port> off`. It cannot function without the Tailscale CLI installed.
3. **Read-only by default** — The only mutating operation is "stop service", which requires explicit confirmation.
4. **Cross-platform** — Must work on macOS, Linux, and Windows (though clipboard support varies).
5. **No network calls** — All data comes from local CLI execution.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21+ |
| TUI Framework | Bubble Tea (Elm architecture) |
| Components | bubbles (list) |
| Styling | Lipgloss |
| Build | GoReleaser |
| CI/CD | GitHub Actions |

## File Layout

```
tailscale-portal/
├── main.go          # Entry point, --version flag
├── model.go         # Data structures: Model, Service, TailscaleStatus
├── update.go        # Message handling, keyboard input, state changes
├── view.go          # All rendering: table, dialogs, footer
├── parser.go        # JSON parsing from tailscale output
├── commands.go      # CLI wrappers: tailscale, clipboard tools
├── styles.go        # Lipgloss style definitions
├── *_test.go        # Unit tests for each module
├── .goreleaser.yaml # Release configuration
└── .github/workflows/release.yml  # CI/CD pipeline
```

## External Dependencies

- `tailscale` CLI binary (runtime)
- `pbcopy` | `xclip` | `wl-copy` (runtime, for clipboard)

## Invariants

- `Service` implements `list.Item` (required by `bubbles/list`)
- `Model.list` must always be kept in sync with `Model.services`
- `commandRunner` interface is the only way to execute external commands (enables mocking in tests)
- All styles are defined in `styles.go` (never inline Lipgloss in view.go)
- The app uses the alternate screen buffer (`tea.WithAltScreen()`)

## Testing Approach

- `mockCommandRunner` swaps the global `cmdRunner` to stub CLI calls
- Parser tests verify JSON edge cases (empty, malformed, nested handlers)
- Update tests verify keyboard handling and state transitions
- No integration tests against real Tailscale (unit tests only)
