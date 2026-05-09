# Development

This guide covers how to set up your development environment, run tests, and understand the project structure.

## Prerequisites

- Go 1.21 or later
- Git
- Tailscale CLI (for manual testing)

## Getting Started

Clone the repository and build:

```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/tailscale-portal.git
cd tailscale-portal
go build -o tailscale-portal
```

Run the application directly without building:

```bash
go run .
```

## Project Structure

```
tailscale-portal/
├── .github/
│   └── workflows/
│       └── release.yml       # GitHub Actions CI/CD pipeline
├── docs/
│   ├── development.md        # This file
│   ├── architecture.md       # Data model and parsing details
│   └── releasing.md          # Release process documentation
├── .goreleaser.yaml          # GoReleaser configuration
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── main.go                   # Application entry point
├── model.go                  # Bubble Tea model and data structures
├── update.go                 # Message handling and state updates
├── view.go                   # UI rendering functions
├── parser.go                 # JSON parsing logic
├── commands.go               # Tailscale CLI command wrappers
├── styles.go                 # Lipgloss styling definitions
├── LICENSE                   # MIT License
└── README.md                 # User-facing documentation
```

## File Overview

| File | Purpose |
|------|---------|
| `main.go` | Entry point. Initializes the Bubble Tea program and handles `--version` flag |
| `model.go` | Defines the `Model` struct (application state), `Service` struct, and `TailscaleStatus` types. Implements `list.Item` interface for the bubbles list component |
| `update.go` | The Elm architecture update function. Handles all keyboard input, window resize events, and async command completion messages |
| `view.go` | Renders the entire UI: header, table, list items, footer, help dialog, and confirmation dialog. Uses responsive column widths |
| `parser.go` | Parses `tailscale serve status --json` output into Go structs. Flattens the nested Web/TCP structure into a linear list of `Service` objects |
| `commands.go` | Wraps external CLI calls (`tailscale serve status`, `pbcopy`, `xclip`, `wl-copy`). Includes a mockable `commandRunner` interface for testing |
| `styles.go` | Centralized Lipgloss style definitions. Color palette and all visual styles are defined here |

## Running Tests

Run the full test suite:

```bash
go test -v ./...
```

Run with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

The project follows standard Go conventions:

- `gofmt` for formatting
- `go vet` for static analysis
- Descriptive variable names
- Comments on all exported types and functions

Run before committing:

```bash
go fmt ./...
go vet ./...
go test ./...
```

## Dependencies

Key dependencies managed via Go modules:

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/bubbles` | Reusable Bubble Tea components (list) |
| `github.com/charmbracelet/lipgloss` | Declarative styling |

View all dependencies:

```bash
go mod graph
```

## Adding a Feature

Typical flow for adding new functionality:

1. **Data layer** (`model.go`): Add fields to `Model` or `Service` structs if needed
2. **Commands** (`commands.go`): Add async command wrappers (e.g., new tailscale CLI operations)
3. **Update** (`update.go`): Handle new messages and keyboard shortcuts
4. **View** (`view.go`): Render new UI elements
5. **Styles** (`styles.go`): Add new styles if needed
6. **Tests**: Add unit tests for parser logic and command behavior

## Debugging

Since Bubble Tea uses the alternate screen buffer, standard `fmt.Println` debugging won't work. Use a log file instead:

```go
// In your command or update function
f, _ := os.OpenFile("/tmp/tailscale-portal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
f.WriteString(fmt.Sprintf("debug: %v\n", someValue))
f.Close()
```

Then tail the log in another terminal:

```bash
tail -f /tmp/tailscale-portal.log
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes with tests
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## License

MIT
