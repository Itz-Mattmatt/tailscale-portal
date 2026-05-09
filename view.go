package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelpView()
	}

	if m.showConfirm {
		return m.renderConfirmDialog()
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Main content area
	if m.err != nil {
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'r' to retry or 'q' to quit"))
	} else if m.loading {
		b.WriteString(loadingStyle.Render("Loading Tailscale services..."))
		b.WriteString("\n")
	} else if len(m.services) == 0 {
		b.WriteString(emptyStyle.Render("No active Tailscale services"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'r' to refresh or 'q' to quit"))
	} else {
		// Table header
		b.WriteString(m.renderTableHeader())
		b.WriteString("\n")
		// Render the list
		b.WriteString(m.list.View())
	}

	// Footer with status and help hints
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	var b strings.Builder

	title := titleStyle.Render(" Tailscale Portal ")
	b.WriteString(title)
	b.WriteString("\n")

	if !m.loading && m.err == nil {
		status := statusStyle.Render(fmt.Sprintf(" Last refresh: %s ", m.lastRefresh.Format("15:04:05")))
		b.WriteString(status)
	}

	if m.successMsg != "" {
		b.WriteString(" ")
		b.WriteString(successStyle.Render(" " + m.successMsg + " "))
	}

	return b.String()
}

// renderTableHeader renders the column headers for the table view
func (m Model) renderTableHeader() string {
	var b strings.Builder

	// Indicator column
	b.WriteString(tableHeaderStyle.Width(2).Render(""))
	b.WriteString(" ")

	// Protocol column
	b.WriteString(tableHeaderStyle.Width(7).Align(lipgloss.Center).Render("PROTO"))
	b.WriteString(" ")

	// Service name column
	serviceWidth := clamp(m.width-60, 20, 35)
	b.WriteString(tableHeaderStyle.Width(serviceWidth).Render("SERVICE"))
	b.WriteString(" ")

	// Path → Target column
	pathWidth := clamp(m.width-serviceWidth-45, 20, 40)
	b.WriteString(tableHeaderStyle.Width(pathWidth).Render("PATH → TARGET"))
	b.WriteString(" ")

	// URL column (remaining space)
	urlWidth := m.width - serviceWidth - pathWidth - 18
	if urlWidth < 10 {
		urlWidth = 10
	}
	b.WriteString(tableHeaderStyle.Width(urlWidth).Render("URL"))

	return b.String()
}

// clamp returns v clamped between min and max
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// renderFooter renders the footer with help hints
func (m Model) renderFooter() string {
	var hints []string

	hints = append(hints, keyStyle.Render("↑/↓")+" navigate")
	hints = append(hints, keyStyle.Render("r")+" refresh")
	hints = append(hints, keyStyle.Render("s/enter")+" stop")
	hints = append(hints, keyStyle.Render("c")+" copy URL")
	hints = append(hints, keyStyle.Render("?/h")+" help")
	hints = append(hints, keyStyle.Render("q")+" quit")

	footer := strings.Join(hints, " • ")
	return footerStyle.Render(footer)
}

// renderHelpView renders the help screen
func (m Model) renderHelpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" Help "))
	b.WriteString("\n\n")

	helpItems := []struct {
		key  string
		desc string
	}{
		{"↑/↓ or j/k", "Navigate up/down through services"},
		{"Enter or s", "Stop/clear the selected service"},
		{"c", "Copy selected service URL to clipboard"},
		{"r or R", "Refresh the service list"},
		{"? or h", "Toggle this help dialog"},
		{"q or Ctrl+C", "Quit the application"},
	}

	for _, item := range helpItems {
		b.WriteString(keyStyle.Render(fmt.Sprintf(" %-12s ", item.key)))
		b.WriteString(" ")
		b.WriteString(item.desc)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press ? or h to close help"))

	// Center the help dialog
	helpContent := b.String()
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Render(helpContent))
}

// renderConfirmDialog renders the confirmation dialog
func (m Model) renderConfirmDialog() string {
	var b strings.Builder

	b.WriteString(warningStyle.Render(" ⚠ Stop Service "))
	b.WriteString("\n\n")
	b.WriteString(m.confirmMsg)
	b.WriteString("\n\n")
	b.WriteString(keyStyle.Render(" y ") + " Yes  ")
	b.WriteString(keyStyle.Render(" n ") + " No")

	dialog := b.String()

	// Center the dialog
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFA500")).
			Padding(1, 2).
			Render(dialog))
}

// serviceDelegate is a custom list item delegate for rendering services
type serviceDelegate struct{}

func (d serviceDelegate) Height() int                               { return 1 }
func (d serviceDelegate) Spacing() int                              { return 0 }
func (d serviceDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d serviceDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	service, ok := item.(Service)
	if !ok {
		return
	}

	var b strings.Builder
	width := m.Width()
	isSelected := index == m.Index()

	// Selection indicator
	if isSelected {
		b.WriteString(selectedIndicatorStyle.Render("▸"))
	} else {
		b.WriteString(" ")
	}
	b.WriteString(" ")

	// Protocol badge
	proto := service.Protocol
	if len(proto) < 5 {
		proto = " " + proto
	}
	protocolBadge := protocolHTTPStyle.Render(proto)
	if service.Protocol == "HTTPS" {
		protocolBadge = protocolHTTPSStyle.Render(proto)
	}
	b.WriteString(protocolBadge)
	b.WriteString(" ")

	// Calculate column widths based on available width
	serviceWidth := clamp(width-60, 20, 35)
	pathWidth := clamp(width-serviceWidth-45, 20, 40)
	urlWidth := width - serviceWidth - pathWidth - 18
	if urlWidth < 10 {
		urlWidth = 10
	}

	// Service name
	nameStyle := serviceNameStyle
	if isSelected {
		nameStyle = selectedServiceNameStyle
	}
	serviceName := truncate(service.ServiceName, serviceWidth)
	b.WriteString(nameStyle.Width(serviceWidth).Render(serviceName))
	b.WriteString(" ")

	// Path → Target
	pathTarget := truncate(service.Path+" → "+service.Target, pathWidth)
	ptStyle := pathTargetStyle
	if isSelected {
		ptStyle = selectedPathTargetStyle
	}
	b.WriteString(ptStyle.Width(pathWidth).Render(pathTarget))
	b.WriteString(" ")

	// URL
	urlStyle := urlStyle
	if isSelected {
		urlStyle = selectedUrlStyle
	}
	url := truncate(service.FullURL, urlWidth)
	b.WriteString(urlStyle.Width(urlWidth).Render(url))

	fmt.Fprint(w, b.String())
}

// truncate truncates a string to fit within maxWidth runes
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}
