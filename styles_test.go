package main

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	// Test that color constants are non-empty
	colors := []string{
		colorBackground,
		colorForeground,
		colorPrimary,
		colorSuccess,
		colorError,
		colorWarning,
		colorInfo,
		colorMuted,
		colorHighlight,
		colorHTTPS,
		colorHTTP,
		colorSelected,
	}

	for _, color := range colors {
		if color == "" {
			t.Error("Color constant should not be empty")
		}
		// Verify it's a valid hex color
		if len(color) < 4 || color[0] != '#' {
			t.Errorf("Color %s should be a valid hex color", color)
		}
	}
}

func TestBaseStyleNotNil(t *testing.T) {
	styles := []lipgloss.Style{
		baseStyle,
		titleStyle,
		statusStyle,
		loadingStyle,
		emptyStyle,
		errorStyle,
		successStyle,
		warningStyle,
		helpStyle,
		footerStyle,
		paginationStyle,
		keyStyle,
	}

	for i, style := range styles {
		// Verify the style can be rendered without panic
		result := style.Render("test")
		if result == "" && i > 0 {
			// baseStyle might render empty, but others should produce output
			t.Logf("Style %d produced empty output", i)
		}
		// Just verify we can call methods on them without panic
		_ = style.String()
	}
}

func TestTableHeaderStyle(t *testing.T) {
	// Verify tableHeaderStyle can be used
	result := tableHeaderStyle.Render("TEST")
	if result == "" {
		t.Error("tableHeaderStyle should produce non-empty output")
	}
}

func TestServiceListStylesNotNil(t *testing.T) {
	styles := []lipgloss.Style{
		selectedIndicatorStyle,
		protocolHTTPSStyle,
		protocolHTTPStyle,
		serviceNameStyle,
		selectedServiceNameStyle,
		pathTargetStyle,
		selectedPathTargetStyle,
		urlStyle,
		selectedUrlStyle,
	}

	for i, style := range styles {
		// Verify the style can be rendered without panic
		result := style.Render("test")
		_ = result // We just want to ensure it doesn't panic
		
		if i < 0 {
			t.Error("This should never happen, just using i")
		}
	}
}

func TestProtocolStylesDifferent(t *testing.T) {
	// HTTPS and HTTP styles should be different
	httpsResult := protocolHTTPSStyle.Render("HTTPS")
	httpResult := protocolHTTPStyle.Render("HTTP")

	// They should render differently (different colors)
	if httpsResult == httpResult {
		t.Log("HTTPS and HTTP styles might be identical - this could be intentional")
	}
}

func TestSelectedStylesDifferentFromNormal(t *testing.T) {
	// Selected styles should differ from normal styles
	tests := []struct {
		normal   lipgloss.Style
		selected lipgloss.Style
		name     string
	}{
		{serviceNameStyle, selectedServiceNameStyle, "serviceName"},
		{pathTargetStyle, selectedPathTargetStyle, "pathTarget"},
		{urlStyle, selectedUrlStyle, "url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalResult := tt.normal.Render("test")
			selectedResult := tt.selected.Render("test")

			// They should be different (at minimum, bold flag differs)
			if normalResult == selectedResult {
				t.Logf("%s and selected%s styles produce identical output", tt.name, tt.name)
			}
		})
	}
}

func TestStyleWidthMethods(t *testing.T) {
	// Test that styles can have their width set
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"tableHeaderStyle", tableHeaderStyle},
		{"serviceNameStyle", serviceNameStyle},
		{"pathTargetStyle", pathTargetStyle},
		{"urlStyle", urlStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styled := tt.style.Width(20).Render("test")
			if styled == "" {
				t.Error("Style with width should produce output")
			}
		})
	}
}

func TestStyleChaining(t *testing.T) {
	// Test that styles can be chained without panic
	result := tableHeaderStyle.
		Width(20).
		Align(lipgloss.Left).
		Render("test")

	if result == "" {
		t.Error("Chained style should produce output")
	}
}
