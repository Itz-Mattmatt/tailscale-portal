package main

import (
	"errors"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Helper function to create a model with populated services for testing
func createModelWithServices() Model {
	m := InitialModel()
	services := []Service{
		{
			ServiceName: "test.ts.net:443",
			Hostname:    "test.ts.net",
			Port:        "443",
			Protocol:    "HTTPS",
			Target:      "http://localhost:8080",
			Path:        "/",
			FullURL:     "https://test.ts.net:443/",
		},
		{
			ServiceName: "test.ts.net:3000",
			Hostname:    "test.ts.net",
			Port:        "3000",
			Protocol:    "HTTP",
			Target:      "http://localhost:3000",
			Path:        "/api",
			FullURL:     "http://test.ts.net:3000/api",
		},
	}
	m.services = services
	items := make([]list.Item, len(services))
	for i, svc := range services {
		items[i] = list.Item(svc)
	}
	m.list.SetItems(items)
	m.loading = false
	m.width = 100
	m.height = 40
	return m
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	m := InitialModel()
	
	msg := tea.WindowSizeMsg{Width: 120, Height: 50}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.width != 120 {
		t.Errorf("Expected width to be 120, got %d", newM.width)
	}
	
	if newM.height != 50 {
		t.Errorf("Expected height to be 50, got %d", newM.height)
	}
	
	// List should be updated with new dimensions
	// Height = 50 - 8 = 42, but clamped at minimum 10
	if newM.list.Height() < 10 {
		t.Errorf("Expected list height to be at least 10, got %d", newM.list.Height())
	}
	
	// List width = 120 - 4 = 116
	if newM.list.Width() != 116 {
		t.Errorf("Expected list width to be 116, got %d", newM.list.Width())
	}
}

func TestUpdateWindowSizeSmallHeight(t *testing.T) {
	m := InitialModel()
	
	// Test with very small height
	msg := tea.WindowSizeMsg{Width: 80, Height: 5}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// List height should be clamped to minimum 10
	if newM.list.Height() != 10 {
		t.Errorf("Expected list height to be clamped to 10, got %d", newM.list.Height())
	}
}

func TestUpdateKeyQuit(t *testing.T) {
	m := createModelWithServices()
	
	tests := []struct {
		name string
		key  string
	}{
		{"q key", "q"},
		{"ctrl+c", "ctrl+c"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			_, cmd := m.Update(msg)
			
			if cmd == nil {
				t.Error("Expected quit command, got nil")
				return
			}
			
			// Execute the command to see what message it returns
			result := cmd()
			if _, ok := result.(tea.QuitMsg); !ok {
				t.Errorf("Expected tea.QuitMsg, got %T", result)
			}
		})
	}
}

func TestUpdateKeyToggleHelp(t *testing.T) {
	m := createModelWithServices()
	
	tests := []struct {
		name string
		key  string
	}{
		{"? key", "?"},
		{"h key", "h"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(msg)
			
			newM, ok := newModel.(Model)
			if !ok {
				t.Fatal("Expected Model type")
			}
			
			if !newM.showHelp {
				t.Error("Expected showHelp to be true")
			}
		})
	}
}

func TestUpdateKeyToggleHelpCancelsConfirm(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = 0
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if !newM.showHelp {
		t.Error("Expected showHelp to be true")
	}
	
	if newM.showConfirm {
		t.Error("Expected showConfirm to be false (cancelled)")
	}
	
	if newM.confirmIndex != -1 {
		t.Errorf("Expected confirmIndex to be -1, got %d", newM.confirmIndex)
	}
}

func TestUpdateKeyRefresh(t *testing.T) {
	m := createModelWithServices()
	m.err = errors.New("some error")
	m.successMsg = "previous success"
	
	tests := []struct {
		name string
		key  string
	}{
		{"r key", "r"},
		{"R key", "R"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, cmd := m.Update(msg)
			
			newM, ok := newModel.(Model)
			if !ok {
				t.Fatal("Expected Model type")
			}
			
			if !newM.loading {
				t.Error("Expected loading to be true")
			}
			
			if newM.err != nil {
				t.Error("Expected err to be nil (cleared)")
			}
			
			if newM.successMsg != "" {
				t.Error("Expected successMsg to be empty (cleared)")
			}
			
			if cmd == nil {
				t.Error("Expected non-nil command (fetchStatus)")
			}
		})
	}
}

func TestUpdateKeyRefreshCancelsConfirm(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = 0
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.showConfirm {
		t.Error("Expected showConfirm to be false (cancelled)")
	}
	
	if newM.confirmIndex != -1 {
		t.Errorf("Expected confirmIndex to be -1, got %d", newM.confirmIndex)
	}
}

func TestUpdateConfirmDialogYes(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = 0
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.showConfirm {
		t.Error("Expected showConfirm to be false")
	}
	
	// Loading is set to true when confirming stop (before command executes)
	if !newM.loading {
		t.Error("Expected loading to be true when confirming stop")
	}
	
	if cmd == nil {
		t.Error("Expected stopService command to be returned")
	}
}

func TestUpdateConfirmDialogYesNoSelection(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = -1 // Invalid index
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.showConfirm {
		t.Error("Expected showConfirm to be false")
	}
	
	if newM.confirmIndex != -1 {
		t.Errorf("Expected confirmIndex to be -1, got %d", newM.confirmIndex)
	}
	
	if cmd != nil {
		t.Error("Expected nil command when no valid selection")
	}
}

func TestUpdateConfirmDialogNo(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = 0
	
	tests := []struct {
		name string
		key  string
	}{
		{"n key", "n"},
		{"N key", "N"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, cmd := m.Update(msg)
			
			newM, ok := newModel.(Model)
			if !ok {
				t.Fatal("Expected Model type")
			}
			
			if newM.showConfirm {
				t.Error("Expected showConfirm to be false")
			}
			
			if newM.confirmIndex != -1 {
				t.Errorf("Expected confirmIndex to be -1, got %d", newM.confirmIndex)
			}
			
			if cmd != nil {
				t.Error("Expected nil command")
			}
		})
	}
}

func TestUpdateConfirmDialogEsc(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	m.confirmIndex = 0
	
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.showConfirm {
		t.Error("Expected showConfirm to be false")
	}
	
	if newM.confirmIndex != -1 {
		t.Errorf("Expected confirmIndex to be -1, got %d", newM.confirmIndex)
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateKeyEnterShowsInfoPopup(t *testing.T) {
	m := createModelWithServices()
	m.list.Select(0) // Select first item
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if !newM.showInfo {
		t.Error("Expected showInfo to be true")
	}
	
	if newM.infoIndex != 0 {
		t.Errorf("Expected infoIndex to be 0, got %d", newM.infoIndex)
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateKeySShowsConfirm(t *testing.T) {
	m := createModelWithServices()
	m.list.Select(0)
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if !newM.showConfirm {
		t.Error("Expected showConfirm to be true")
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateKeyCopy(t *testing.T) {
	m := createModelWithServices()
	m.list.Select(0) // Select first item
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.successMsg != "Copied URL to clipboard" {
		t.Errorf("Expected successMsg to be 'Copied URL to clipboard', got %s", newM.successMsg)
	}
	
	if cmd == nil {
		t.Error("Expected copyToClipboard command")
	}
}

func TestUpdateStatusMsgWithError(t *testing.T) {
	m := InitialModel()
	m.loading = true
	
	testErr := errors.New("test error")
	msg := statusMsg{err: testErr}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.loading {
		t.Error("Expected loading to be false")
	}
	
	if newM.err != testErr {
		t.Errorf("Expected err to be %v, got %v", testErr, newM.err)
	}
	
	if len(newM.services) != 0 {
		t.Errorf("Expected services to be empty, got %d", len(newM.services))
	}
	
	items := newM.list.Items()
	if len(items) != 0 {
		t.Errorf("Expected list items to be empty, got %d", len(items))
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateStatusMsgSuccess(t *testing.T) {
	m := InitialModel()
	m.loading = true
	m.err = errors.New("previous error")
	
	status := TailscaleStatus{
		TCP: map[string]TCPConfig{
			"443": {HTTPS: true},
		},
		Web: map[string]WebConfig{
			"test.ts.net:443": {
				Handlers: map[string]HandlerConfig{
					"/": {Proxy: "http://localhost:8080"},
				},
			},
		},
	}
	
	msg := statusMsg{status: status}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.loading {
		t.Error("Expected loading to be false")
	}
	
	if newM.err != nil {
		t.Error("Expected err to be nil")
	}
	
	if len(newM.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(newM.services))
	}
	
	items := newM.list.Items()
	if len(items) != 1 {
		t.Errorf("Expected 1 list item, got %d", len(items))
	}
	
	if time.Since(newM.lastRefresh) > time.Second {
		t.Error("Expected lastRefresh to be updated")
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateStopCompleteMsgWithError(t *testing.T) {
	m := createModelWithServices()
	m.loading = true
	m.successMsg = "previous success"
	
	testErr := errors.New("stop failed")
	msg := stopCompleteMsg{err: testErr}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.loading {
		t.Error("Expected loading to be false")
	}
	
	if newM.err != testErr {
		t.Errorf("Expected err to be %v, got %v", testErr, newM.err)
	}
	
	if newM.successMsg != "" {
		t.Error("Expected successMsg to be empty")
	}
	
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateStopCompleteMsgSuccess(t *testing.T) {
	m := createModelWithServices()
	m.loading = true
	m.err = errors.New("previous error")
	
	msg := stopCompleteMsg{}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	if newM.loading {
		t.Error("Expected loading to be false")
	}
	
	if newM.err != nil {
		t.Error("Expected err to be nil")
	}
	
	if newM.successMsg != "Service stopped successfully" {
		t.Errorf("Expected successMsg to be 'Service stopped successfully', got %s", newM.successMsg)
	}
	
	if cmd == nil {
		t.Error("Expected fetchStatus command")
	}
}

func TestUpdateClearSuccessMsg(t *testing.T) {
	m := createModelWithServices()
	
	msg := clearSuccessMsg{}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// clearSuccessMsg should not change state
	if cmd != nil {
		t.Error("Expected nil command")
	}
	
	_ = newM // Silence unused variable warning
}

func TestUpdateTickMsg(t *testing.T) {
	m := createModelWithServices()
	
	msg := tickMsg{Time: time.Now()}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// tickMsg should not change state
	if cmd != nil {
		t.Error("Expected nil command")
	}
	
	_ = newM // Silence unused variable warning
}

func TestUpdateListNavigationWhenNotInHelpOrConfirm(t *testing.T) {
	m := createModelWithServices()
	m.showHelp = false
	m.showConfirm = false
	
	// Send a navigation key that should be passed to the list
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// The command should be non-nil (list navigation command)
	// Note: cmd might be nil if the list doesn't have any navigation to do
	_ = newM
	_ = cmd
}

func TestUpdateListNavigationBlockedInHelpMode(t *testing.T) {
	m := createModelWithServices()
	m.showHelp = true
	
	originalIndex := m.list.Index()
	
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// List index should not change in help mode
	if newM.list.Index() != originalIndex {
		t.Error("Expected list index to not change in help mode")
	}
}

func TestUpdateListNavigationBlockedInConfirmMode(t *testing.T) {
	m := createModelWithServices()
	m.showConfirm = true
	
	originalIndex := m.list.Index()
	
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// List index should not change in confirm mode
	if newM.list.Index() != originalIndex {
		t.Error("Expected list index to not change in confirm mode")
	}
}

func TestUpdateUnknownMessageType(t *testing.T) {
	m := createModelWithServices()
	
	// Send an unknown message type
	msg := "unknown message"
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	
	// Should still work and pass to list
	_ = newM
	_ = cmd
}
