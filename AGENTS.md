# Agent Instructions

# CRITICAL RULES - MUST FOLLOW

## RESPONSES

- Keep responses concise and to the point - unless the user asks otherwise

## PLANNING MODE

- Always ask clarifying questions
- Never assume design, tech stack or features
- Use deep-dive sub-agents to assist with research
- Use deep-dive sub-agents to review the different aspects of your plan before presenting to the user

## CHANGE / EDIT MODE

- Never implement features yourself when possible - use sub-agents!
- Identify changes from the plan that can be implemented in parallel, and use sub-agents to implement the features efficiently
- When using sub-agents to implement features, act as a coordinator only
- Use the best model for the task - premium models for complex tasks (like coding) and mid-tier models for simpler tasks, like documentation
- After completing features (large or small), always run check commands to make sure project can build
- When following plans always keep plan files up to date with progress and status

## TESTING

- Use any testing tools, libraries available to the project for testing your changes
- Never assume your changes simply work, always test!
- If the project does not have any testing tools, scripts, MCP tools, skills, etc. available for testing, ask the user whether testing should be skipped.

## Quick Commands

```bash
# Build and run
go build -o tailscale-portal && ./tailscale-portal

# Run tests
go test -v ./...

# Run a single test
go test -run TestUpdateKeyQuit -v

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture That Isn't Obvious

**External commands are never called directly.** Everything goes through the `commandRunner` interface in `commands.go`. The global `cmdRunner` variable is swapped to `mockCommandRunner` in tests. **Always call `defer resetCommandRunner()` after `SetCommandRunner(mock)`** or subsequent tests will fail with cryptic errors.

**Nested JSON is flattened.** `tailscale serve status --json` returns a tree (`Web` → `Handlers` → paths). The `ServicesFromStatus()` function in `parser.go` unrolls this into a flat `[]Service` slice because the Bubble Tea `list` component requires linear `[]list.Item`. One Web entry with 3 handlers becomes 3 `Service` rows.

**Version is embedded at link time.** `main.go` declares `var version = "dev"`. GoReleaser injects the real version via `-X main.version={{.Version}}` ldflag. Don't hardcode version strings anywhere else.

## File Boundaries (Never Cross These)

| File | Contains | Never Contains |
|------|----------|----------------|
| `model.go` | Structs, types, `list.Item` interface impl | Logic, rendering, CLI calls |
| `parser.go` | JSON unmarshal + flattening | External calls, UI code |
| `commands.go` | All `os/exec`, all `tea.Cmd` factories, mock infrastructure | State mutations, rendering |
| `update.go` | All `tea.Msg` handling, all state transitions | `fmt.Print`, Lipgloss, `os/exec` |
| `view.go` | All `strings.Builder`, all `lipgloss.Render` | State mutation, CLI calls |
| `styles.go` | All `lipgloss.NewStyle()` | Any other code |
| `main.go` | Entry point, `--version` flag | Everything else |

## Testing Patterns

### Mocking external commands

```go
mock := &mockCommandRunner{
    RunFunc: func(name string, arg ...string) ([]byte, error) {
        return []byte(`{"TCP":{},"Web":{}}`), nil
    },
}
SetCommandRunner(mock)
defer resetCommandRunner() // MANDATORY
```

### Test structure

Tests are table-driven where possible. Tests use `t.Run("subtest name", func(t *testing.T) {...})` for grouping. No external dependencies required — all tests run offline via mocks.

### What to test

- **Parser**: Empty JSON, malformed service names, missing TCP section, multiple handlers per service
- **Commands**: Verify correct CLI arguments are passed to `tailscale serve`, verify fallback order for clipboard tools
- **Update**: Keyboard routing, dialog state machine, confirmation flow, error/success state transitions
- **View**: Render doesn't panic with nil/empty/loading/error states
- **Model**: Interface compliance (`Service` implements `list.Item`, `Model` implements `tea.Model`)

## Key Gotchas

1. **Dialog priority matters.** `View()` checks `showHelp` → `showConfirm` → main view in that order. `Update()` gates list navigation when dialogs are open. Adding a new dialog requires adding to both chains.

2. **Clipboard errors reuse `stopCompleteMsg`.** This is intentional — the UI handler treats it the same way (shows error in the status bar). Don't create a separate message type unless the UI needs different behavior.

3. **`Model.list` and `Model.services` must stay synced.** After parsing, update both: `m.services = ServicesFromStatus(...)` and `m.list.SetItems(items)`. The list is what renders; `services` is what the confirmation dialog indexes into.

4. **`buildFullURL` prepends `/` to paths.** A path of `"api"` becomes `"/api"`. A path of `"/api"` stays `"/api"`.

5. **The app uses alternate screen buffer.** `tea.WithAltScreen()` in `main.go` means `fmt.Println` debugging won't work. Use a log file if you need to debug.

## Style Rules

- All colors are hex constants in `styles.go` (never inline in `view.go`)
- Style variable names end in `Style`
- HTTPS badge = green (`#50C878`), HTTP badge = orange (`#FFA500`)
- Selected items use `colorHighlight` (`#FFE66D`)

## Release & CI

- Release triggers on `v*` tag push via `.github/workflows/release.yml`
- GoReleaser config is `.goreleaser.yaml` (version 2 syntax)
- Tests **must** pass before GoReleaser runs
- Update `YOUR_GITHUB_USERNAME` placeholders in `.goreleaser.yaml` and `README.md` before first release
- Build requires Go 1.25 (per `go.mod`), but CI uses `go-version: stable`

## Deeper Reference

The `.ai/` folder contains AI-optimized documentation:
- `.ai/PROJECT.md` — Constraints, invariants, testing approach
- `.ai/ARCHITECTURE.md` — Data flow, state machine, message types, module dependencies
- `.ai/CONVENTIONS.md` — Naming rules, file organization, code patterns
- `.ai/TASKS.md` — Step-by-step recipes for common modifications
