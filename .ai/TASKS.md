# Common Tasks

## How to Add a New Keyboard Action

**Example: Add `o` key to open a service URL in the browser**

### 1. Add the command (commands.go)

```go
func openURL(url string) tea.Cmd {
    return func() tea.Msg {
        // Try different open commands by OS
        if err := cmdRunner.Run("open", url); err == nil { // macOS
            return clearSuccessMsg{}
        }
        if err := cmdRunner.Run("xdg-open", url); err == nil { // Linux
            return clearSuccessMsg{}
        }
        if err := cmdRunner.Run("cmd", "/c", "start", url); err == nil { // Windows
            return clearSuccessMsg{}
        }
        return stopCompleteMsg{
            err: fmt.Errorf("failed to open URL: no browser launcher available"),
        }
    }
}
```

### 2. Handle the key (update.go)

In the `tea.KeyMsg` switch, after the `c` case:

```go
case "o":
    if item, ok := m.list.SelectedItem().(Service); ok {
        m.successMsg = "Opening URL in browser"
        return m, openURL(item.FullURL)
    }
    return m, nil
```

### 3. Show the key in the footer (view.go)

In `renderFooter()`, add to hints:

```go
hints = append(hints, keyStyle.Render("o")+" open")
```

### 4. Add to help dialog (view.go)

In `renderHelpView()`, add to helpItems:

```go
{"o", "Open selected service URL in browser"},
```

### 5. Test

```go
func TestOpenURL(t *testing.T) {
    mock := &mockCommandRunner{
        RunFunc: func(name string, arg ...string) ([]byte, error) {
            if name == "open" && arg[0] == "https://example.com" {
                return []byte{}, nil
            }
            return nil, fmt.Errorf("not found")
        },
    }
    SetCommandRunner(mock)
    defer resetCommandRunner()

    cmd := openURL("https://example.com")
    msg := cmd()
    if _, ok := msg.(clearSuccessMsg); !ok {
        t.Errorf("expected clearSuccessMsg, got %T", msg)
    }
}
```

## How to Add a New Field to the Service Display

**Example: Show the "last updated" time for each service**

### 1. Add field to Service struct (model.go)

```go
type Service struct {
    // ... existing fields ...
    LastUpdated time.Time
}
```

### 2. Populate it in parser (parser.go)

In `ServicesFromStatus()`, set the field:

```go
service := Service{
    // ... existing fields ...
    LastUpdated: time.Now(), // or parse from somewhere
}
```

### 3. Add a column in the table header (view.go)

In `renderTableHeader()`:

```go
// After URL column
b.WriteString(" ")
b.WriteString(tableHeaderStyle.Width(12).Render("UPDATED"))
```

### 4. Render it in the list delegate (view.go)

In `serviceDelegate.Render()`, after the URL:

```go
b.WriteString(" ")
updated := service.LastUpdated.Format("15:04")
updatedStyle := lipgloss.NewStyle().Width(12).Foreground(lipgloss.Color(colorMuted))
if isSelected {
    updatedStyle = updatedStyle.Bold(true)
}
b.WriteString(updatedStyle.Render(updated))
```

### 5. Adjust column widths

Recalculate the `serviceWidth`, `pathWidth`, and `urlWidth` formulas to account for the new column. Subtract the new column width from available space.

### 6. Update FilterValue (model.go)

If the new field should be searchable:

```go
func (s Service) FilterValue() string {
    return s.ServiceName + " " + s.Target + " " + s.Path + " " + s.LastUpdated.Format("15:04")
}
```

## How to Add a New Dialog

**Example: Add a "filter by protocol" dialog**

### 1. Add state to Model (model.go)

```go
type Model struct {
    // ... existing fields ...
    showFilter   bool
    filterProto  string // "", "HTTP", "HTTPS"
}
```

Initialize in `InitialModel()`:

```go
showFilter: false,
filterProto: "",
```

### 2. Add keyboard handler (update.go)

Add to global key handler:

```go
case "f":
    if m.showConfirm || m.showHelp {
        return m, nil
    }
    m.showFilter = !m.showFilter
    return m, nil
```

### 3. Gate list navigation (update.go)

In the list navigation section, add:

```go
if m.showFilter {
    // Handle filter keys
    switch msg.String() {
    case "h":
        m.filterProto = "HTTP"
        m.showFilter = false
        return m, applyFilter(m.filterProto)
    case "s":
        m.filterProto = "HTTPS"
        m.showFilter = false
        return m, applyFilter(m.filterProto)
    case "a", "esc":
        m.filterProto = ""
        m.showFilter = false
        return m, applyFilter("")
    }
    return m, nil
}
```

### 4. Add applyFilter command (commands.go)

```go
func applyFilter(proto string) tea.Cmd {
    return func() tea.Msg {
        return refreshCompleteMsg{}
    }
}
```

### 5. Render the dialog (view.go)

At the top of `View()`:

```go
if m.showFilter {
    return m.renderFilterDialog()
}
```

Add render function:

```go
func (m Model) renderFilterDialog() string {
    var b strings.Builder
    b.WriteString(titleStyle.Render(" Filter by Protocol "))
    b.WriteString("\n\n")
    b.WriteString(keyStyle.Render(" h ") + " HTTP only\n")
    b.WriteString(keyStyle.Render(" s ") + " HTTPS only\n")
    b.WriteString(keyStyle.Render(" a ") + " All protocols\n")
    b.WriteString(keyStyle.Render(" esc ") + " Cancel")
    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Render(b.String()))
}
```

### 6. Update footer and help

Add "f filter" to footer hints and help dialog.

## How to Add a New Test

**Example: Test that empty Tailscale status shows empty state**

### 1. Write the test in the appropriate `*_test.go` file

```go
func TestEmptyStatus(t *testing.T) {
    status := TailscaleStatus{
        TCP: map[string]TCPConfig{},
        Web: map[string]WebConfig{},
    }
    services := ServicesFromStatus(status)
    if len(services) != 0 {
        t.Errorf("expected 0 services, got %d", len(services))
    }
}
```

### 2. Run the test

```bash
go test -run TestEmptyStatus -v
```

### 3. Test patterns to follow

- Name: `Test<FunctionName>_<Scenario>`
- Table-driven tests for multiple cases
- Always reset `cmdRunner` after mocking
- Assert both happy path and error path

## How to Modify the Release Process

### Add a new package format

Edit `.goreleaser.yaml`, find the `nfpms` section, add to `formats`:

```yaml
formats:
  - deb
  - rpm
  - apk
  - archlinux  # new
```

### Add a new build target

In `.goreleaser.yaml` `builds` section, add to `goos`/`goarch`:

```yaml
goos:
  - linux
  - darwin
  - windows
  - freebsd  # new
goarch:
  - amd64
  - arm64
  - "386"
  - arm      # new
```

### Change version numbering

The version comes from Git tags. To change how versions are parsed, modify the `snapshot.version_template` in `.goreleaser.yaml`.

## How to Fix a Bug

### Clipboard not working on a new platform

1. Find the `copyToClipboard()` function in `commands.go`
2. Add the new clipboard tool before the final error return:

```go
// Try using new-tool on NewOS
if err := cmdRunner.RunWithStdin("new-tool", text); err == nil {
    return clearSuccessMsg{}
}
```

3. Add a test case in `commands_test.go`

### UI rendering broken at small terminal sizes

1. Check `clamp()` calls in `view.go`
2. Ensure minimum widths are reasonable (at least 10-15 chars per column)
3. Check `listHeight` calculation in `update.go` — ensure it never goes below a minimum

### Parser fails on new Tailscale JSON format

1. Add the new field to `TailscaleStatus`, `WebConfig`, or `HandlerConfig` in `model.go`
2. Update `ServicesFromStatus()` in `parser.go` to read the new field
3. Add test JSON in `parser_test.go` with the new format
4. Ensure backward compatibility (old JSON without the field should still parse)
