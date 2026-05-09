package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchStatus runs the tailscale serve status command and returns the parsed status
func fetchStatus() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tailscale", "serve", "status", "--json")
		output, err := cmd.CombinedOutput()
		
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

		cmd := exec.Command("tailscale", args...)
		output, err := cmd.CombinedOutput()

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
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return clearSuccessMsg{}
		}

		// Try using xclip on Linux
		cmd = exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return clearSuccessMsg{}
		}

		// Try using wl-copy on Wayland
		cmd = exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return clearSuccessMsg{}
		}

		// If all fail, return error
		return stopCompleteMsg{
			err: fmt.Errorf("failed to copy to clipboard: no clipboard tool available (tried pbcopy, xclip, wl-copy)"),
		}
	}
}
