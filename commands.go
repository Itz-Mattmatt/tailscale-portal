package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// commandRunner defines the interface for executing external commands
type commandRunner interface {
	Run(name string, arg ...string) ([]byte, error)
	RunWithStdin(name string, stdin string, arg ...string) error
}

// realCommandRunner is the production implementation
type realCommandRunner struct{}

func (r *realCommandRunner) Run(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.CombinedOutput()
}

func (r *realCommandRunner) RunWithStdin(name string, stdin string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.Run()
}

// Global runner that can be swapped for testing
var cmdRunner commandRunner = &realCommandRunner{}

// SetCommandRunner sets the command runner (used for testing)
func SetCommandRunner(r commandRunner) {
	cmdRunner = r
}

// fetchStatus runs the tailscale serve status command and returns the parsed status
func fetchStatus() tea.Cmd {
	return func() tea.Msg {
		output, err := cmdRunner.Run("tailscale", "serve", "status", "--json")

		if err != nil {
			// Check if tailscale is not found
			if strings.Contains(string(output), "not found") ||
				strings.Contains(err.Error(), "executable file not found") {
				return statusMsg{
					err: fmt.Errorf("tailscale CLI not found. Please install Tailscale: https://tailscale.com/download"),
				}
			}
			return statusMsg{
				err: fmt.Errorf("failed to get tailscale status: %s", string(output)),
			}
		}

		status, err := ParseStatus(output)
		if err != nil {
			return statusMsg{
				err: fmt.Errorf("failed to parse tailscale status: %w", err),
			}
		}

		return statusMsg{
			status: status,
		}
	}
}

// stopService stops a Tailscale service by running the appropriate off command
// For HTTPS: tailscale serve --https=<port> off
// For HTTP: tailscale serve --http=<port> off
func stopService(port string, protocol string) tea.Cmd {
	return func() tea.Msg {
		var args []string
		if protocol == "HTTPS" {
			args = []string{"serve", "--https=" + port, "off"}
		} else {
			args = []string{"serve", "--http=" + port, "off"}
		}

		output, err := cmdRunner.Run("tailscale", args...)

		if err != nil {
			return stopCompleteMsg{
				err: fmt.Errorf("failed to stop service on port %s: %s", port, string(output)),
			}
		}

		return stopCompleteMsg{}
	}
}

// refreshList refreshes the service list
func refreshList() tea.Cmd {
	return func() tea.Msg {
		// Small delay to show loading state
		time.Sleep(100 * time.Millisecond)
		return refreshCompleteMsg{}
	}
}

// copyToClipboard copies the given text to the system clipboard
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		// Try using pbcopy on macOS
		if err := cmdRunner.RunWithStdin("pbcopy", text); err == nil {
			return clearSuccessMsg{}
		}

		// Try using xclip on Linux
		if err := cmdRunner.RunWithStdin("xclip", text, "-selection", "clipboard"); err == nil {
			return clearSuccessMsg{}
		}

		// Try using wl-copy on Wayland
		if err := cmdRunner.RunWithStdin("wl-copy", text); err == nil {
			return clearSuccessMsg{}
		}

		// If all fail, return error
		return stopCompleteMsg{
			err: fmt.Errorf("failed to copy to clipboard: no clipboard tool available (tried pbcopy, xclip, wl-copy)"),
		}
	}
}

// mockCommandRunner is a mock implementation for testing
type mockCommandRunner struct {
	RunFunc           func(name string, arg ...string) ([]byte, error)
	RunWithStdinFunc  func(name string, stdin string, arg ...string) error
	runCalls          []runCall
	runWithStdinCalls []runWithStdinCall
}

type runCall struct {
	name string
	args []string
}

type runWithStdinCall struct {
	name  string
	stdin string
	args  []string
}

func (m *mockCommandRunner) Run(name string, arg ...string) ([]byte, error) {
	m.runCalls = append(m.runCalls, runCall{name: name, args: arg})
	if m.RunFunc != nil {
		return m.RunFunc(name, arg...)
	}
	return nil, fmt.Errorf("mock Run not implemented")
}

func (m *mockCommandRunner) RunWithStdin(name string, stdin string, arg ...string) error {
	m.runWithStdinCalls = append(m.runWithStdinCalls, runWithStdinCall{name: name, stdin: stdin, args: arg})
	if m.RunWithStdinFunc != nil {
		return m.RunWithStdinFunc(name, stdin, arg...)
	}
	return fmt.Errorf("mock RunWithStdin not implemented")
}

func (m *mockCommandRunner) getRunCalls() []runCall {
	return m.runCalls
}

func (m *mockCommandRunner) getRunWithStdinCalls() []runWithStdinCall {
	return m.runWithStdinCalls
}

// Helper function to reset the global runner to the real implementation
func resetCommandRunner() {
	cmdRunner = &realCommandRunner{}
}

// For async testing with context
type ctxKey string

const testRunnerKey ctxKey = "testRunner"

func withTestRunner(ctx context.Context, runner commandRunner) context.Context {
	return context.WithValue(ctx, testRunnerKey, runner)
}
