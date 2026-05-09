# Conventions

## Code Style

- Standard Go formatting (`gofmt`)
- No line length limits
- Exported types/functions have comments
- Unexported types do not need comments unless complex

## Naming

| Pattern | Example | Rule |
|---------|---------|------|
| Types | `TailscaleStatus`, `Service`, `Model` | PascalCase, descriptive |
| Interfaces | `commandRunner` | lowerCamelCase (unexported) |
| Functions | `fetchStatus()`, `buildFullURL()` | lowerCamelCase, verb-first |
| Messages | `statusMsg`, `stopCompleteMsg` | lowerCamelCase, noun+Msg suffix |
| Styles | `titleStyle`, `errorStyle` | descriptive+Style suffix |
| Booleans | `showHelp`, `loading`, `isSelected` | prefix with verb or "is" |

## File Organization Rules

1. **model.go** — All structs and types. No logic.
2. **parser.go** — JSON parsing only. No external calls.
3. **commands.go** — All `os/exec` calls. All `tea.Cmd` factories. Mock infrastructure.
4. **update.go** — All message handling. All state transitions. No rendering.
5. **view.go** — All `strings.Builder` and `lipgloss.Render` calls. No state mutation.
6. **styles.go** — All `lipgloss.NewStyle()` calls. No other code.
7. **main.go** — Entry point only. Minimal logic.

**Never mix concerns across these boundaries.**

## Bubble Tea Patterns

### Adding a New Command

If you need to run an external command asynchronously:

```go
// commands.go
func myNewCommand(args string) tea.Cmd {
    return func() tea.Msg {
        output, err := cmdRunner.Run("tailscale", "my-command", args)
        if err != nil {
            return myCompleteMsg{err: fmt.Errorf("...")}
        }
        return myCompleteMsg{data: output}
    }
}

// model.go
type myCompleteMsg struct {
    data []byte
    err  error
}

// update.go
case myCompleteMsg:
    m.loading = false
    if msg.err != nil {
        m.err = msg.err
    } else {
        // process msg.data
    }
    return m, nil
```

### Adding a New Keyboard Shortcut

```go
// update.go, inside tea.KeyMsg handler
switch msg.String() {
case "x":
    if m.showConfirm || m.showHelp {
        return m, nil
    }
    // your action
    return m, myCommand()
}
```

Always check `showConfirm` and `showHelp` first to avoid triggering actions inside dialogs.

### Adding a New Dialog

1. Add bool field to `Model` in `model.go`
2. Add keyboard handler in `update.go` (check bool, toggle)
3. Add render function in `view.go` (check bool at top of `View()`)
4. Add styles in `styles.go` if needed
5. Follow the pattern: dialog bools are checked in order of priority

## Error Handling

- External command errors: wrap with context, return as `statusMsg{err: ...}` or `stopCompleteMsg{err: ...}`
- JSON parse errors: return as `statusMsg{err: ...}`, displayed in UI
- Clipboard failures: return as `stopCompleteMsg{err: ...}` (reused message type for simplicity)
- Never `panic()` or `log.Fatal()`

## Testing Patterns

### Mocking a Command

```go
mock := &mockCommandRunner{
    RunFunc: func(name string, arg ...string) ([]byte, error) {
        if name == "tailscale" && arg[0] == "serve" {
            return []byte(`{"TCP":{},"Web":{}}`), nil
        }
        return nil, fmt.Errorf("unexpected command")
    },
}
SetCommandRunner(mock)
defer resetCommandRunner() // CRITICAL: reset after test
```

Always call `resetCommandRunner()` in `defer` to avoid polluting other tests.

### Testing Parser Edge Cases

Test these scenarios:
- Empty JSON `{}`
- Valid with multiple handlers
- Malformed service name (missing `:`)
- Missing TCP section (defaults to HTTP)
- Null fields

## Style Rules

- All colors are hex codes, defined as constants in `styles.go`
- Background color is always `#1a1a1a`
- Selected items use `colorHighlight` (`#FFE66D`)
- HTTPS = green (`#50C878`), HTTP = orange (`#FFA500`)
- Errors = red (`#FF6B6B`)
- Never create styles inline in `view.go`
- Style variable names end in `Style`

## Version Embedding

Version is embedded at build time via ldflags:
```
-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```

The `--version` flag is handled in `main.go` before starting the TUI.
