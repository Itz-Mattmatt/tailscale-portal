package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// mockCommandRunnerForTests creates a mock with predefined responses
func mockCommandRunnerForTests(runFunc func(name string, arg ...string) ([]byte, error), 
	runWithStdinFunc func(name string, stdin string, arg ...string) error) *mockCommandRunner {
	return &mockCommandRunner{
		RunFunc:          runFunc,
		RunWithStdinFunc: runWithStdinFunc,
	}
}

func TestFetchStatusSuccess(t *testing.T) {
	jsonData := `{
		"TCP": {
			"443": { "HTTPS": true }
		},
		"Web": {
			"test.ts.net:443": {
				"Handlers": {
					"/": { "Proxy": "http://localhost:8080" }
				}
			}
		}
	}`

	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" && len(arg) == 3 && arg[0] == "serve" && arg[1] == "status" && arg[2] == "--json" {
				return []byte(jsonData), nil
			}
			return nil, fmt.Errorf("unexpected command: %s %v", name, arg)
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := fetchStatus()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	msg := cmd()
	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("Expected statusMsg, got %T", msg)
	}

	if statusMsg.err != nil {
		t.Errorf("Expected no error, got %v", statusMsg.err)
	}

	if len(statusMsg.status.TCP) != 1 {
		t.Errorf("Expected 1 TCP entry, got %d", len(statusMsg.status.TCP))
	}

	if len(statusMsg.status.Web) != 1 {
		t.Errorf("Expected 1 Web entry, got %d", len(statusMsg.status.Web))
	}
}

func TestFetchStatusTailscaleNotFound(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("tailscale: command not found"), errors.New("exit status 127")
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := fetchStatus()
	msg := cmd()

	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("Expected statusMsg, got %T", msg)
	}

	if statusMsg.err == nil {
		t.Error("Expected error for tailscale not found")
	}

	if statusMsg.err != nil && statusMsg.err.Error() != "tailscale CLI not found. Please install Tailscale: https://tailscale.com/download" {
		t.Errorf("Expected specific error message, got: %v", statusMsg.err)
	}
}

func TestFetchStatusExecutableNotFound(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return nil, errors.New("executable file not found in $PATH")
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := fetchStatus()
	msg := cmd()

	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("Expected statusMsg, got %T", msg)
	}

	if statusMsg.err == nil {
		t.Error("Expected error for executable not found")
	}

	if statusMsg.err != nil && statusMsg.err.Error() != "tailscale CLI not found. Please install Tailscale: https://tailscale.com/download" {
		t.Errorf("Expected specific error message, got: %v", statusMsg.err)
	}
}

func TestFetchStatusInvalidJSON(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("not valid json"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := fetchStatus()
	msg := cmd()

	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("Expected statusMsg, got %T", msg)
	}

	if statusMsg.err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if statusMsg.err != nil && !errors.Is(statusMsg.err, errors.New("failed to parse JSON")) {
		// Error message should contain "failed to parse"
		errStr := statusMsg.err.Error()
		if len(errStr) == 0 {
			t.Error("Expected non-empty error message")
		}
	}
}

func TestFetchStatusCommandFailure(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("some tailscale error"), errors.New("exit status 1")
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := fetchStatus()
	msg := cmd()

	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("Expected statusMsg, got %T", msg)
	}

	if statusMsg.err == nil {
		t.Error("Expected error for command failure")
	}
}

func TestStopServiceHTTPS(t *testing.T) {
	var capturedArgs []string
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				capturedArgs = arg
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := stopService("443", "HTTPS")
	msg := cmd()

	stopMsg, ok := msg.(stopCompleteMsg)
	if !ok {
		t.Fatalf("Expected stopCompleteMsg, got %T", msg)
	}

	if stopMsg.err != nil {
		t.Errorf("Expected no error, got %v", stopMsg.err)
	}

	// Verify correct arguments were used
	if len(capturedArgs) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(capturedArgs), capturedArgs)
	}

	if capturedArgs[0] != "serve" {
		t.Errorf("Expected first arg to be 'serve', got %s", capturedArgs[0])
	}

	if capturedArgs[1] != "--https=443" {
		t.Errorf("Expected second arg to be '--https=443', got %s", capturedArgs[1])
	}

	if capturedArgs[2] != "off" {
		t.Errorf("Expected third arg to be 'off', got %s", capturedArgs[2])
	}
}

func TestStopServiceHTTP(t *testing.T) {
	var capturedArgs []string
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				capturedArgs = arg
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := stopService("3000", "HTTP")
	msg := cmd()

	stopMsg, ok := msg.(stopCompleteMsg)
	if !ok {
		t.Fatalf("Expected stopCompleteMsg, got %T", msg)
	}

	if stopMsg.err != nil {
		t.Errorf("Expected no error, got %v", stopMsg.err)
	}

	// Verify correct arguments were used
	if len(capturedArgs) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(capturedArgs), capturedArgs)
	}

	if capturedArgs[1] != "--http=3000" {
		t.Errorf("Expected second arg to be '--http=3000', got %s", capturedArgs[1])
	}
}

func TestStopServiceFailure(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("permission denied"), errors.New("exit status 1")
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := stopService("443", "HTTPS")
	msg := cmd()

	stopMsg, ok := msg.(stopCompleteMsg)
	if !ok {
		t.Fatalf("Expected stopCompleteMsg, got %T", msg)
	}

	if stopMsg.err == nil {
		t.Error("Expected error for stop failure")
	}

	if stopMsg.err != nil && !contains(stopMsg.err.Error(), "failed to stop service") {
		t.Errorf("Expected error to contain 'failed to stop service', got: %v", stopMsg.err)
	}
}

func TestRefreshList(t *testing.T) {
	cmd := refreshList()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Measure time
	start := time.Now()
	msg := cmd()
	elapsed := time.Since(start)

	// Should take approximately 100ms
	if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected ~100ms delay, got %v", elapsed)
	}

	_, ok := msg.(refreshCompleteMsg)
	if !ok {
		t.Fatalf("Expected refreshCompleteMsg, got %T", msg)
	}
}

func TestCopyToClipboardPbcopySuccess(t *testing.T) {
	mock := mockCommandRunnerForTests(
		nil,
		func(name string, stdin string, arg ...string) error {
			if name == "pbcopy" {
				return nil // Success
			}
			return errors.New("not pbcopy")
		},
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := copyToClipboard("https://test.ts.net:443/")
	msg := cmd()

	_, ok := msg.(clearSuccessMsg)
	if !ok {
		t.Fatalf("Expected clearSuccessMsg, got %T", msg)
	}
}

func TestCopyToClipboardXclipSuccess(t *testing.T) {
	callCount := 0
	mock := mockCommandRunnerForTests(
		nil,
		func(name string, stdin string, arg ...string) error {
			callCount++
			if name == "pbcopy" {
				return errors.New("not available")
			}
			if name == "xclip" {
				// Verify correct args
				if len(arg) == 2 && arg[0] == "-selection" && arg[1] == "clipboard" {
					return nil // Success
				}
			}
			return errors.New("not xclip")
		},
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := copyToClipboard("https://test.ts.net:443/")
	msg := cmd()

	_, ok := msg.(clearSuccessMsg)
	if !ok {
		t.Fatalf("Expected clearSuccessMsg, got %T", msg)
	}

	if callCount < 2 {
		t.Errorf("Expected at least 2 calls (pbcopy then xclip), got %d", callCount)
	}
}

func TestCopyToClipboardWlCopySuccess(t *testing.T) {
	callCount := 0
	mock := mockCommandRunnerForTests(
		nil,
		func(name string, stdin string, arg ...string) error {
			callCount++
			if name == "pbcopy" || name == "xclip" {
				return errors.New("not available")
			}
			if name == "wl-copy" {
				return nil // Success
			}
			return errors.New("unknown tool")
		},
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := copyToClipboard("https://test.ts.net:443/")
	msg := cmd()

	_, ok := msg.(clearSuccessMsg)
	if !ok {
		t.Fatalf("Expected clearSuccessMsg, got %T", msg)
	}

	if callCount < 3 {
		t.Errorf("Expected at least 3 calls (pbcopy, xclip, wl-copy), got %d", callCount)
	}
}

func TestCopyToClipboardAllFail(t *testing.T) {
	mock := mockCommandRunnerForTests(
		nil,
		func(name string, stdin string, arg ...string) error {
			return errors.New("not available")
		},
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := copyToClipboard("https://test.ts.net:443/")
	msg := cmd()

	stopMsg, ok := msg.(stopCompleteMsg)
	if !ok {
		t.Fatalf("Expected stopCompleteMsg (error), got %T", msg)
	}

	if stopMsg.err == nil {
		t.Error("Expected error when all clipboard tools fail")
	}

	if stopMsg.err != nil && !contains(stopMsg.err.Error(), "no clipboard tool available") {
		t.Errorf("Expected error about no clipboard tool, got: %v", stopMsg.err)
	}
}

func TestCopyToClipboardPassesCorrectText(t *testing.T) {
	testURL := "https://example.com:443/test-path"
	var capturedStdin string

	mock := mockCommandRunnerForTests(
		nil,
		func(name string, stdin string, arg ...string) error {
			if name == "pbcopy" {
				capturedStdin = stdin
				return nil
			}
			return errors.New("not available")
		},
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := copyToClipboard(testURL)
	cmd()

	if capturedStdin != testURL {
		t.Errorf("Expected stdin to be %q, got %q", testURL, capturedStdin)
	}
}

func TestSetCommandRunner(t *testing.T) {
	// Save original runner
	originalRunner := cmdRunner

	// Create a mock
	mock := &mockCommandRunner{}

	// Set the mock
	SetCommandRunner(mock)

	// Verify it was set
	if cmdRunner != mock {
		t.Error("SetCommandRunner did not set the runner correctly")
	}

	// Restore original
	cmdRunner = originalRunner
}

func TestStartServiceHTTPS(t *testing.T) {
	var capturedArgs []string
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				capturedArgs = arg
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := startService(443, "http://localhost:3000", "/", "https", "serve")
	msg := cmd()

	startMsg, ok := msg.(startCompleteMsg)
	if !ok {
		t.Fatalf("Expected startCompleteMsg, got %T", msg)
	}
	if startMsg.err != nil {
		t.Errorf("Expected no error, got %v", startMsg.err)
	}

	if len(capturedArgs) != 4 {
		t.Fatalf("Expected 4 args, got %d: %v", len(capturedArgs), capturedArgs)
	}
	if capturedArgs[0] != "serve" {
		t.Errorf("Expected first arg to be 'serve', got %s", capturedArgs[0])
	}
	if capturedArgs[1] != "--bg" {
		t.Errorf("Expected second arg to be '--bg', got %s", capturedArgs[1])
	}
	if capturedArgs[2] != "--https=443" {
		t.Errorf("Expected third arg to be '--https=443', got %s", capturedArgs[2])
	}
	if capturedArgs[3] != "localhost:3000" {
		t.Errorf("Expected fourth arg to be 'localhost:3000', got %s", capturedArgs[3])
	}
}

func TestStartServiceHTTP(t *testing.T) {
	var capturedArgs []string
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				capturedArgs = arg
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := startService(8080, "http://localhost:8080", "/api", "http", "serve")
	cmd()

	if len(capturedArgs) != 5 {
		t.Fatalf("Expected 5 args, got %d: %v", len(capturedArgs), capturedArgs)
	}
	if capturedArgs[0] != "serve" {
		t.Errorf("Expected first arg to be 'serve', got %s", capturedArgs[0])
	}
	if capturedArgs[1] != "--bg" {
		t.Errorf("Expected second arg to be '--bg', got %s", capturedArgs[1])
	}
	if capturedArgs[2] != "--http=8080" {
		t.Errorf("Expected third arg to be '--http=8080', got %s", capturedArgs[2])
	}
	if capturedArgs[3] != "/api" {
		t.Errorf("Expected fourth arg to be '/api', got %s", capturedArgs[3])
	}
	if capturedArgs[4] != "localhost:8080" {
		t.Errorf("Expected fifth arg to be 'localhost:8080', got %s", capturedArgs[4])
	}
}

func TestStartServiceFunnel(t *testing.T) {
	var capturedArgs []string
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				capturedArgs = arg
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := startService(443, "http://localhost:4000", "/", "https", "funnel")
	cmd()

	if len(capturedArgs) != 4 {
		t.Fatalf("Expected 4 args, got %d: %v", len(capturedArgs), capturedArgs)
	}
	if capturedArgs[0] != "funnel" {
		t.Errorf("Expected first arg to be 'funnel', got %s", capturedArgs[0])
	}
	if capturedArgs[1] != "--bg" {
		t.Errorf("Expected second arg to be '--bg', got %s", capturedArgs[1])
	}
	if capturedArgs[2] != "--https=443" {
		t.Errorf("Expected third arg to be '--https=443', got %s", capturedArgs[2])
	}
	if capturedArgs[3] != "localhost:4000" {
		t.Errorf("Expected fourth arg to be 'localhost:4000', got %s", capturedArgs[3])
	}
}

func TestStartServiceFailure(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("permission denied"), errors.New("exit status 1")
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)

	SetCommandRunner(mock)
	defer resetCommandRunner()

	cmd := startService(443, "http://localhost:3000", "/", "https", "serve")
	msg := cmd()

	startMsg, ok := msg.(startCompleteMsg)
	if !ok {
		t.Fatalf("Expected startCompleteMsg, got %T", msg)
	}
	if startMsg.err == nil {
		t.Error("Expected error for start failure")
	}
	if startMsg.err != nil && !contains(startMsg.err.Error(), "failed to start service") {
		t.Errorf("Expected error to contain 'failed to start service', got: %v", startMsg.err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(s[:len(substr)] == substr) || 
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
