package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := InitialModel()

	// Test loading state
	if !m.loading {
		t.Error("Expected loading to be true")
	}

	// Test services slice is empty
	if m.services == nil {
		t.Error("Expected services to be initialized (not nil)")
	}
	if len(m.services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(m.services))
	}

	// Test showHelp is false
	if m.showHelp {
		t.Error("Expected showHelp to be false")
	}

	// Test showConfirm is false
	if m.showConfirm {
		t.Error("Expected showConfirm to be false")
	}

	// Test confirmIndex is -1
	if m.confirmIndex != -1 {
		t.Errorf("Expected confirmIndex to be -1, got %d", m.confirmIndex)
	}

	// Test list is initialized
	if m.list.Width() != 80 {
		t.Errorf("Expected list width to be 80, got %d", m.list.Width())
	}

	// Test lastRefresh is set to a recent time
	if time.Since(m.lastRefresh) > time.Second {
		t.Error("Expected lastRefresh to be set to current time")
	}

	// Test successMsg is empty
	if m.successMsg != "" {
		t.Errorf("Expected successMsg to be empty, got %s", m.successMsg)
	}

	// Test err is nil
	if m.err != nil {
		t.Errorf("Expected err to be nil, got %v", m.err)
	}
}

func TestInitReturnsCommand(t *testing.T) {
	m := InitialModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Expected Init() to return a non-nil command")
	}
}

func TestServiceFilterValue(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected string
	}{
		{
			name: "basic service",
			service: Service{
				ServiceName: "test.ts.net:443",
				Target:      "http://localhost:8080",
				Path:        "/",
			},
			expected: "test.ts.net:443 http://localhost:8080 /",
		},
		{
			name: "service with path",
			service: Service{
				ServiceName: "test.ts.net:3000",
				Target:      "http://localhost:5173",
				Path:        "/api",
			},
			expected: "test.ts.net:3000 http://localhost:5173 /api",
		},
		{
			name: "empty service",
			service: Service{
				ServiceName: "",
				Target:      "",
				Path:        "",
			},
			expected: "  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.FilterValue()
			if result != tt.expected {
				t.Errorf("FilterValue() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestServiceTitle(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected string
	}{
		{
			name: "full service name",
			service: Service{
				ServiceName: "test.ts.net:443",
			},
			expected: "test.ts.net:443",
		},
		{
			name: "localhost service",
			service: Service{
				ServiceName: "localhost:8080",
			},
			expected: "localhost:8080",
		},
		{
			name: "empty service name",
			service: Service{
				ServiceName: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.Title()
			if result != tt.expected {
				t.Errorf("Title() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestServiceDescription(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected string
	}{
		{
			name: "root path",
			service: Service{
				Path:   "/",
				Target: "http://localhost:8080",
			},
			expected: "/ → http://localhost:8080",
		},
		{
			name: "api path",
			service: Service{
				Path:   "/api/v1",
				Target: "http://localhost:3000",
			},
			expected: "/api/v1 → http://localhost:3000",
		},
		{
			name: "empty values",
			service: Service{
				Path:   "",
				Target: "",
			},
			expected: " → ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.Description()
			if result != tt.expected {
				t.Errorf("Description() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestServiceImplementsListItem(t *testing.T) {
	// Verify Service implements list.Item interface
	var _ list.Item = Service{}
}

func TestModelImplementsTeaModel(t *testing.T) {
	// Verify Model implements tea.Model interface
	var _ tea.Model = Model{}
}

func TestMessageTypes(t *testing.T) {
	// Test that message types can be instantiated
	_ = statusMsg{
		status: TailscaleStatus{},
		err:    nil,
	}

	_ = refreshCompleteMsg{
		err: nil,
	}

	_ = stopCompleteMsg{
		err: nil,
	}

	_ = clearSuccessMsg{}

	_ = tickMsg{
		Time: time.Now(),
	}
}
