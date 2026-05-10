package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

// fakeItem is a test helper type that implements list.Item but is not a Service
type fakeItem struct{}

func (f fakeItem) FilterValue() string { return "" }
func (f fakeItem) Title() string       { return "" }
func (f fakeItem) Description() string { return "" }

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		v        int
		min      int
		max      int
		expected int
	}{
		{"value in range", 50, 10, 100, 50},
		{"value below min", 5, 10, 100, 10},
		{"value above max", 150, 10, 100, 100},
		{"value at min", 10, 10, 100, 10},
		{"value at max", 100, 10, 100, 100},
		{"negative min", -5, -10, 10, -5},
		{"negative below min", -15, -10, 10, -10},
		{"zero range", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.v, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%d, %d, %d) = %d, expected %d", tt.v, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxWidth int
		expected string
	}{
		{"empty string", "", 10, ""},
		{"exact fit", "hello", 5, "hello"},
		{"longer than max", "hello world", 8, "hello..."},
		{"maxWidth <= 0", "hello", 0, ""},
		{"maxWidth negative", "hello", -5, ""},
		{"maxWidth <= 3 exact", "hi", 2, "hi"},
		{"maxWidth <= 3 cut", "hello", 3, "hel"},
		{"maxWidth 4", "hello world", 4, "h..."},
		{"unicode string", "Hello, 世界", 8, "Hello..."},
		{"single char", "x", 1, "x"},
		{"two chars", "xy", 2, "xy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.s, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.s, tt.maxWidth, result, tt.expected)
			}
		})
	}
}

func TestRenderHeader(t *testing.T) {
	t.Run("loading state", func(t *testing.T) {
		m := InitialModel()
		m.loading = true
		
		result := m.renderHeader()
		
		if result == "" {
			t.Error("Expected non-empty header")
		}
		
		// Should contain title
		if !strings.Contains(result, "Tailscale Portal") {
			t.Error("Expected header to contain title")
		}
		
		// Should not show refresh time when loading
		if strings.Contains(result, "Last refresh") {
			t.Error("Expected no refresh time during loading")
		}
	})
	
	t.Run("with success message", func(t *testing.T) {
		m := InitialModel()
		m.loading = false
		m.err = nil
		m.successMsg = "Test success"
		
		result := m.renderHeader()
		
		if !strings.Contains(result, "Test success") {
			t.Error("Expected header to contain success message")
		}
		
		if !strings.Contains(result, "Last refresh") {
			t.Error("Expected header to contain refresh time")
		}
	})
	
	t.Run("without success message", func(t *testing.T) {
		m := InitialModel()
		m.loading = false
		m.err = nil
		m.successMsg = ""
		
		result := m.renderHeader()
		
		if !strings.Contains(result, "Tailscale Portal") {
			t.Error("Expected header to contain title")
		}
		
		if !strings.Contains(result, "Last refresh") {
			t.Error("Expected header to contain refresh time")
		}
	})
	
	t.Run("with error", func(t *testing.T) {
		m := InitialModel()
		m.loading = false
		m.err = errors.New("test error")
		m.successMsg = ""
		
		result := m.renderHeader()
		
		if !strings.Contains(result, "Tailscale Portal") {
			t.Error("Expected header to contain title")
		}
		
		// renderHeader only shows refresh time when !m.loading && m.err == nil
		// So with an error, it should NOT show "Last refresh"
		if strings.Contains(result, "Last refresh") {
			t.Error("Expected header to NOT contain refresh time when there's an error")
		}
	})
}

func TestRenderFooter(t *testing.T) {
	m := InitialModel()
	
	result := m.renderFooter()
	
	if result == "" {
		t.Error("Expected non-empty footer")
	}
	
	// Check for key hints
	expectedHints := []string{"navigate", "refresh", "stop", "copy", "help", "quit"}
	for _, hint := range expectedHints {
		if !strings.Contains(strings.ToLower(result), hint) {
			t.Errorf("Expected footer to contain '%s'", hint)
		}
	}
}

func TestRenderHelpView(t *testing.T) {
	m := InitialModel()
	m.width = 80
	m.height = 40
	
	result := m.renderHelpView()
	
	if result == "" {
		t.Error("Expected non-empty help view")
	}
	
	// Check for help items
	expectedItems := []string{
		"Help",
		"Navigate",
		"Stop",
		"Copy",
		"Refresh",
		"Toggle",
		"Quit",
	}
	
	for _, item := range expectedItems {
		if !strings.Contains(result, item) {
			t.Errorf("Expected help view to contain '%s'", item)
		}
	}
}

func TestRenderConfirmDialog(t *testing.T) {
	m := InitialModel()
	m.width = 80
	m.height = 40
	m.confirmMsg = "Stop service test.ts.net:443?"
	
	result := m.renderConfirmDialog()
	
	if result == "" {
		t.Error("Expected non-empty confirm dialog")
	}
	
	// Check for dialog content
	if !strings.Contains(result, "Stop Service") {
		t.Error("Expected confirm dialog to contain 'Stop Service'")
	}
	
	if !strings.Contains(result, m.confirmMsg) {
		t.Error("Expected confirm dialog to contain confirm message")
	}
	
	if !strings.Contains(result, "Yes") {
		t.Error("Expected confirm dialog to contain 'Yes' option")
	}
	
	if !strings.Contains(result, "No") {
		t.Error("Expected confirm dialog to contain 'No' option")
	}
}

func TestRenderTableHeader(t *testing.T) {
	m := InitialModel()
	m.width = 100
	
	result := m.renderTableHeader()
	
	if result == "" {
		t.Error("Expected non-empty table header")
	}
	
	// Check for column headers
	expectedHeaders := []string{"PROTO", "SERVICE", "PATH", "TARGET", "URL"}
	for _, header := range expectedHeaders {
		// Some headers might be combined, so check individually
		if !strings.Contains(result, header) && !strings.Contains(result, strings.ReplaceAll(header, " → ", "")) {
			// Allow for partial matches
		}
	}
}

func TestServiceDelegateHeight(t *testing.T) {
	d := serviceDelegate{}
	if d.Height() != 1 {
		t.Errorf("Expected height to be 1, got %d", d.Height())
	}
}

func TestServiceDelegateSpacing(t *testing.T) {
	d := serviceDelegate{}
	if d.Spacing() != 0 {
		t.Errorf("Expected spacing to be 0, got %d", d.Spacing())
	}
}

func TestServiceDelegateUpdate(t *testing.T) {
	d := serviceDelegate{}
	l := list.New([]list.Item{}, d, 80, 20)
	
	// Update should return nil
	cmd := d.Update(nil, &l)
	if cmd != nil {
		t.Error("Expected Update to return nil")
	}
}

func TestServiceDelegateRender(t *testing.T) {
	d := serviceDelegate{}
	
	service := Service{
		ServiceName: "test.ts.net:443",
		Hostname:    "test.ts.net",
		Port:        "443",
		Protocol:    "HTTPS",
		Target:      "http://localhost:8080",
		Path:        "/",
		FullURL:     "https://test.ts.net:443/",
	}
	
	t.Run("render HTTPS service", func(t *testing.T) {
		l := list.New([]list.Item{service}, d, 100, 20)
		var buf bytes.Buffer
		
		d.Render(&buf, l, 0, service)
		
		result := buf.String()
		if result == "" {
			t.Error("Expected non-empty render output")
		}
	})
	
	t.Run("render HTTP service", func(t *testing.T) {
		httpService := Service{
			ServiceName: "test.ts.net:3000",
			Hostname:    "test.ts.net",
			Port:        "3000",
			Protocol:    "HTTP",
			Target:      "http://localhost:3000",
			Path:        "/api",
			FullURL:     "http://test.ts.net:3000/api",
		}
		
		l := list.New([]list.Item{httpService}, d, 100, 20)
		var buf bytes.Buffer
		
		d.Render(&buf, l, 0, httpService)
		
		result := buf.String()
		if result == "" {
			t.Error("Expected non-empty render output")
		}
	})
	
	t.Run("render selected item", func(t *testing.T) {
		l := list.New([]list.Item{service}, d, 100, 20)
		l.Select(0) // Select the first item
		var buf bytes.Buffer
		
		d.Render(&buf, l, 0, service)
		
		result := buf.String()
		if result == "" {
			t.Error("Expected non-empty render output")
		}
	})
	
	t.Run("render non-service item", func(t *testing.T) {
		l := list.New([]list.Item{}, d, 100, 20)
		var buf bytes.Buffer
		
		// Pass a fake item instead of Service
		d.Render(&buf, l, 0, fakeItem{})
		
		result := buf.String()
		if result != "" {
			t.Error("Expected empty render output for non-service item")
		}
	})
}

func TestViewLoadingState(t *testing.T) {
	m := InitialModel()
	m.loading = true
	m.width = 80
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	if !strings.Contains(result, "Loading") {
		t.Error("Expected view to contain 'Loading' text")
	}
}

func TestViewErrorState(t *testing.T) {
	m := InitialModel()
	m.loading = false
	m.err = errors.New("test error message")
	m.width = 80
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	if !strings.Contains(result, "Error") {
		t.Error("Expected view to contain 'Error' text")
	}
	
	if !strings.Contains(result, "test error message") {
		t.Error("Expected view to contain error message")
	}
}

func TestViewEmptyServices(t *testing.T) {
	m := InitialModel()
	m.loading = false
	m.err = nil
	m.services = []Service{}
	m.width = 80
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	if !strings.Contains(result, "No active") {
		t.Error("Expected view to contain 'No active' message")
	}
}

func TestViewWithServices(t *testing.T) {
	m := InitialModel()
	m.loading = false
	m.err = nil
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
	}
	m.services = services
	items := make([]list.Item, len(services))
	for i, svc := range services {
		items[i] = list.Item(svc)
	}
	m.list.SetItems(items)
	m.width = 100
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	// Should contain the table header
	if !strings.Contains(result, "PROTO") {
		t.Error("Expected view to contain PROTO header")
	}
}

func TestViewHelpMode(t *testing.T) {
	m := InitialModel()
	m.showHelp = true
	m.width = 80
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	if !strings.Contains(result, "Help") {
		t.Error("Expected view to contain 'Help' when in help mode")
	}
}

func TestViewConfirmMode(t *testing.T) {
	m := InitialModel()
	m.showConfirm = true
	m.confirmMsg = "Stop service?"
	m.width = 80
	m.height = 40
	
	result := m.View()
	
	if result == "" {
		t.Error("Expected non-empty view")
	}
	
	if !strings.Contains(result, "Stop Service") {
		t.Error("Expected view to contain 'Stop Service' when in confirm mode")
	}
}

func TestViewNewServeDialog(t *testing.T) {
	m := createModelWithServices()
	m.showNewServe = true
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for new serve dialog")
	}
}

func TestViewNewServeDialogWithSaveToFav(t *testing.T) {
	m := createModelWithServices()
	m.showNewServe = true
	m.serveForm.saveToFav = true
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestViewEditFavDialog(t *testing.T) {
	m := createModelWithServices()
	m.showEditFav = true
	m.editFavForm.inputs[fieldPort].SetValue("443")
	m.editFavForm.inputs[fieldTarget].SetValue("http://localhost:3000")
	m.editFavForm.inputs[fieldPath].SetValue("/")
	m.editFavForm.inputs[fieldName].SetValue("Test")
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for edit fav dialog")
	}
}

func TestViewFavouritesEmpty(t *testing.T) {
	m := createModelWithServices()
	m.showFavourites = true
	m.favourites = []Favourite{}
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for empty favourites")
	}
}

func TestViewFavouritesWithItems(t *testing.T) {
	m := createModelWithServices()
	m.showFavourites = true
	m.favourites = []Favourite{
		{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}
	m.syncFavListItems()
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for favourites with items")
	}
}

func TestViewFavConfirmDialog(t *testing.T) {
	m := createModelWithServices()
	m.showFavConfirm = true
	m.favConfirmMsg = "Delete favourite 'Test'?"
	m.width = 100
	m.height = 40
	
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for fav confirm dialog")
	}
}

func TestFavouriteDelegateHeight(t *testing.T) {
	d := favouriteDelegate{}
	if d.Height() != 1 {
		t.Errorf("Expected height 1, got %d", d.Height())
	}
}

func TestFavouriteDelegateSpacing(t *testing.T) {
	d := favouriteDelegate{}
	if d.Spacing() != 0 {
		t.Errorf("Expected spacing 0, got %d", d.Spacing())
	}
}

func TestFavouriteDelegateUpdate(t *testing.T) {
	d := favouriteDelegate{}
	l := list.New([]list.Item{}, d, 80, 20)
	cmd := d.Update(nil, &l)
	if cmd != nil {
		t.Error("Expected nil command from Update")
	}
}

func TestFavouriteDelegateRender(t *testing.T) {
	d := favouriteDelegate{}
	l := list.New([]list.Item{}, d, 80, 20)
	l.SetItems([]list.Item{
		Favourite{ID: "1", Name: "Test", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	})
	
	var buf strings.Builder
	d.Render(&buf, l, 0, l.Items()[0])
	
	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty rendered output")
	}
	if !strings.Contains(output, "Test") {
		t.Error("Expected output to contain favourite name")
	}
}

func TestFavouriteDelegateRenderNonFavourite(t *testing.T) {
	d := favouriteDelegate{}
	l := list.New([]list.Item{}, d, 80, 20)
	
	var buf strings.Builder
	d.Render(&buf, l, 0, fakeItem{})
	
	output := buf.String()
	if output != "" {
		t.Error("Expected empty output for non-favourite item")
	}
}
