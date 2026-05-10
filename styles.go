package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	colorBackground = "#1a1a1a"
	colorForeground = "#e0e0e0"
	colorPrimary    = "#7B68EE" // MediumSlateBlue
	colorSuccess    = "#50C878" // Emerald
	colorError      = "#FF6B6B" // Light Red
	colorWarning    = "#FFA500" // Orange
	colorInfo       = "#4ECDC4" // Teal
	colorMuted      = "#666666"
	colorHighlight  = "#FFE66D" // Yellow
	colorHTTPS      = "#50C878" // Green for HTTPS
	colorHTTP       = "#FFA500" // Orange for HTTP
	colorSelected   = "#2d2d2d"
)

// Base styles
var (
	baseStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBackground)).
		Foreground(lipgloss.Color(colorForeground))

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorPrimary)).
		Background(lipgloss.Color(colorBackground)).
		Padding(0, 1).
		MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Italic(true)

	loadingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorInfo)).
		Italic(true)

	emptyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Italic(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorError)).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorSuccess)).
		Background(lipgloss.Color(colorBackground)).
		Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorWarning)).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted))

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		MarginTop(1)

	paginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted))

	keyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHighlight)).
		Bold(true)
)

// Table header style
var tableHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colorMuted)).
	Bold(true)

// Service list styles
var (
	selectedIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorPrimary)).
		Bold(true)

	protocolHTTPSStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBackground)).
		Background(lipgloss.Color(colorHTTPS)).
		Bold(true).
		Padding(0, 1)

	protocolHTTPStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBackground)).
		Background(lipgloss.Color(colorHTTP)).
		Bold(true).
		Padding(0, 1)

	serviceNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorForeground)).
		Bold(true)

	selectedServiceNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHighlight)).
		Bold(true)

	pathTargetStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorInfo))

	selectedPathTargetStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorInfo)).
		Bold(true)

	urlStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Italic(true)

	selectedUrlStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorPrimary)).
		Italic(true)
)

// Dialog styles
var (
	dialogBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	dialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorPrimary))

	formLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorForeground)).
		Bold(true)

	formInputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorForeground))

	formInputFocusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHighlight))

	toggleActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBackground)).
		Background(lipgloss.Color(colorPrimary)).
		Bold(true).
		Padding(0, 1)

	toggleInactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Padding(0, 1)

	checkboxCheckedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorSuccess))

	checkboxUncheckedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted))

	favouriteStarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorWarning))

	runningBadgeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorSuccess))

	foregroundBadgeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBackground)).
		Background(lipgloss.Color(colorInfo)).
		Bold(true).
		Padding(0, 1)

	dialogFooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Italic(true)
)
