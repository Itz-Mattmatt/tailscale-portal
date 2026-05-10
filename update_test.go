package main

import (
	"errors"
	"fmt"
	"os"
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

func TestUpdateKeyHelpBlockedInForms(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		setupModel func(Model) Model
		checkField string
	}{
		{"h in new serve dialog", "h", func(m Model) Model { m.showNewServe = true; m.serveForm.focusIndex = 0; m.serveForm.inputs[0].Focus(); return m }, "showNewServe"},
		{"? in new serve dialog", "?", func(m Model) Model { m.showNewServe = true; m.serveForm.focusIndex = 0; m.serveForm.inputs[0].Focus(); return m }, "showNewServe"},
		{"h in edit fav dialog", "h", func(m Model) Model { m.showEditFav = true; m.editFavForm.focusIndex = 0; m.editFavForm.inputs[0].Focus(); return m }, "showEditFav"},
		{"? in edit fav dialog", "?", func(m Model) Model { m.showEditFav = true; m.editFavForm.focusIndex = 0; m.editFavForm.inputs[0].Focus(); return m }, "showEditFav"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createModelWithServices()
			m = tt.setupModel(m)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(msg)

			newM, ok := newModel.(Model)
			if !ok {
				t.Fatal("Expected Model type")
			}

			if newM.showHelp {
				t.Error("Expected showHelp to remain false when typing in form")
			}

			switch tt.checkField {
			case "showNewServe":
				if !newM.showNewServe {
					t.Error("Expected showNewServe to remain true")
				}
			case "showEditFav":
				if !newM.showEditFav {
					t.Error("Expected showEditFav to remain true")
				}
			}
		})
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

func TestUpdateKeyNewServe(t *testing.T) {
	m := createModelWithServices()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if !newM.showNewServe {
		t.Error("Expected showNewServe to be true")
	}
	if newM.serveForm.protocol != "https" {
		t.Error("Expected default protocol to be https")
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateKeyFavourites(t *testing.T) {
	m := createModelWithServices()
	m.favourites = []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if !newM.showFavourites {
		t.Error("Expected showFavourites to be true")
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateNewServeCancel(t *testing.T) {
	m := createModelWithServices()
	m.showNewServe = true
	m.serveForm.inputs[fieldPort].SetValue("9999")
	
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showNewServe {
		t.Error("Expected showNewServe to be false")
	}
	if newM.serveForm.inputs[fieldPort].Value() != "" {
		t.Error("Expected form to be reset")
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateNewServeSubmitValid(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)
	SetCommandRunner(mock)
	defer resetCommandRunner()

	m := createModelWithServices()
	m.showNewServe = true
	m.serveForm.inputs[fieldPort].SetValue("443")
	m.serveForm.inputs[fieldTarget].SetValue("http://localhost:3000")
	m.serveForm.inputs[fieldPath].SetValue("/")
	m.serveForm.protocol = "https"
	m.serveForm.mode = "serve"
	m.serveForm.saveToFav = false
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showNewServe {
		t.Error("Expected showNewServe to be false after submit")
	}
	if !newM.loading {
		t.Error("Expected loading to be true")
	}
	if cmd == nil {
		t.Error("Expected startService command")
	}
}

func TestUpdateNewServeSubmitInvalidPort(t *testing.T) {
	m := createModelWithServices()
	m.showNewServe = true
	m.serveForm.inputs[fieldPort].SetValue("abc")
	m.serveForm.inputs[fieldTarget].SetValue("http://localhost:3000")
	m.serveForm.inputs[fieldPath].SetValue("/")
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if !newM.showNewServe {
		t.Error("Expected showNewServe to still be true (validation failed)")
	}
	if newM.formError == "" {
		t.Error("Expected form error for invalid port")
	}
	if cmd != nil {
		t.Error("Expected nil command when validation fails")
	}
}

func TestUpdateNewServeSubmitInvalidTarget(t *testing.T) {
	m := createModelWithServices()
	m.showNewServe = true
	m.serveForm.inputs[fieldPort].SetValue("443")
	m.serveForm.inputs[fieldTarget].SetValue("not-a-url")
	m.serveForm.inputs[fieldPath].SetValue("/")
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.formError == "" {
		t.Error("Expected form error for invalid target")
	}
	if cmd != nil {
		t.Error("Expected nil command when validation fails")
	}
}

func TestUpdateAddToFavouritesFromInfo(t *testing.T) {
	m := createModelWithServices()
	m.showInfo = true
	m.infoIndex = 0
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showInfo {
		t.Error("Expected showInfo to be false")
	}
	if !newM.showNewServe {
		t.Error("Expected showNewServe to be true")
	}
	if newM.serveForm.inputs[fieldPort].Value() != "443" {
		t.Errorf("Expected port to be pre-populated with 443, got %s", newM.serveForm.inputs[fieldPort].Value())
	}
	if newM.serveForm.saveToFav != true {
		t.Error("Expected saveToFav to be true")
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateStartCompleteMsgSuccess(t *testing.T) {
	m := createModelWithServices()
	m.loading = true
	m.err = errors.New("previous error")
	
	msg := startCompleteMsg{}
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
	if newM.successMsg != "Service started successfully" {
		t.Errorf("Expected success message, got %s", newM.successMsg)
	}
	if cmd == nil {
		t.Error("Expected fetchStatus command")
	}
}

func TestUpdateStartCompleteMsgError(t *testing.T) {
	m := createModelWithServices()
	m.loading = true
	m.successMsg = "previous success"
	
	testErr := errors.New("start failed")
	msg := startCompleteMsg{err: testErr}
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

func TestUpdateFavouritesLoadedMsg(t *testing.T) {
	m := createModelWithServices()
	
	favourites := []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	msg := favouritesLoadedMsg{favourites: favourites}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if len(newM.favourites) != 1 {
		t.Errorf("Expected 1 favourite, got %d", len(newM.favourites))
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateFavouriteSavedMsg(t *testing.T) {
	m := createModelWithServices()
	
	msg := favouriteSavedMsg{}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.successMsg != "Favourite saved" {
		t.Errorf("Expected success message 'Favourite saved', got %s", newM.successMsg)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateFavouritesViewStart(t *testing.T) {
	mock := mockCommandRunnerForTests(
		func(name string, arg ...string) ([]byte, error) {
			if name == "tailscale" {
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("unexpected command")
		},
		nil,
	)
	SetCommandRunner(mock)
	defer resetCommandRunner()

	m := createModelWithServices()
	m.showFavourites = true
	m.favourites = []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	m.syncFavListItems()
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showFavourites {
		t.Error("Expected showFavourites to be false after starting")
	}
	if !newM.loading {
		t.Error("Expected loading to be true")
	}
	if cmd == nil {
		t.Error("Expected startService command")
	}
}

func TestUpdateFavouritesViewDeleteConfirm(t *testing.T) {
	m := createModelWithServices()
	m.showFavourites = true
	m.favourites = []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	m.syncFavListItems()
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if !newM.showFavConfirm {
		t.Error("Expected showFavConfirm to be true")
	}
	if newM.favConfirmIndex != 0 {
		t.Errorf("Expected favConfirmIndex to be 0, got %d", newM.favConfirmIndex)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateFavouritesViewDeleteYes(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}()

	m := createModelWithServices()
	m.showFavConfirm = true
	m.favConfirmIndex = 0
	m.favourites = []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	m.syncFavListItems()
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showFavConfirm {
		t.Error("Expected showFavConfirm to be false")
	}
	if len(newM.favourites) != 0 {
		t.Errorf("Expected 0 favourites after delete, got %d", len(newM.favourites))
	}
	if cmd == nil {
		t.Error("Expected saveFavouritesCmd command")
	}
}

func TestUpdateFavouritesViewEsc(t *testing.T) {
	m := createModelWithServices()
	m.showFavourites = true
	
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	
	newM, ok := newModel.(Model)
	if !ok {
		t.Fatal("Expected Model type")
	}
	if newM.showFavourites {
		t.Error("Expected showFavourites to be false")
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestUpdateValidateServeFormEmptyPort(t *testing.T) {
	form := serveForm{}
	result := validateServeForm(form)
	if result != "Port is required" {
		t.Errorf("Expected 'Port is required', got %s", result)
	}
}

func TestUpdateValidateServeFormInvalidPort(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("abc")
	result := validateServeForm(form)
	if result != "Port must be a number between 1 and 65535" {
		t.Errorf("Expected port error, got %s", result)
	}
}

func TestUpdateValidateServeFormInvalidTarget(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("443")
	form.inputs[fieldTarget].SetValue("not-a-url")
	result := validateServeForm(form)
	if result != "Target must be a URL (http://...) or host:port" {
		t.Errorf("Expected target error, got %s", result)
	}
}

func TestUpdateValidateServeFormMissingName(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("443")
	form.inputs[fieldTarget].SetValue("http://localhost:3000")
	form.inputs[fieldPath].SetValue("/")
	form.saveToFav = true
	form.inputs[fieldName].SetValue("")
	result := validateServeForm(form)
	if result != "Name is required when saving to favourites" {
		t.Errorf("Expected name error, got %s", result)
	}
}

func TestUpdateValidateServeFormValid(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("443")
	form.inputs[fieldTarget].SetValue("http://localhost:3000")
	form.inputs[fieldPath].SetValue("/")
	form.saveToFav = false
	result := validateServeForm(form)
	if result != "" {
		t.Errorf("Expected no error, got %s", result)
	}
}

func TestUpdateValidateServeFormValidHostPort(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("443")
	form.inputs[fieldTarget].SetValue("localhost:3000")
	form.inputs[fieldPath].SetValue("/")
	form.saveToFav = false
	result := validateServeForm(form)
	if result != "" {
		t.Errorf("Expected no error for host:port target, got %s", result)
	}
}

func TestUpdateResetServeForm(t *testing.T) {
	form := serveForm{}
	form.inputs[fieldPort].SetValue("443")
	form.inputs[fieldTarget].SetValue("http://localhost:3000")
	form.protocol = "http"
	form.mode = "funnel"
	form.saveToFav = true
	form.focusIndex = 3
	
	resetServeForm(&form)
	
	if form.inputs[fieldPort].Value() != "" {
		t.Error("Expected port to be reset")
	}
	if form.protocol != "https" {
		t.Error("Expected protocol to be reset to https")
	}
	if form.mode != "serve" {
		t.Error("Expected mode to be reset to serve")
	}
	if form.saveToFav != false {
		t.Error("Expected saveToFav to be reset")
	}
	if form.focusIndex != 0 {
		t.Error("Expected focusIndex to be reset")
	}
}
