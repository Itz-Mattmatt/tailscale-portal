package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// TailscaleStatus represents the JSON output from `tailscale serve status --json`
type TailscaleStatus struct {
	TCP map[string]TCPConfig `json:"TCP"`
	Web map[string]WebConfig `json:"Web"`
}

// TCPConfig represents a TCP port configuration
type TCPConfig struct {
	HTTPS bool `json:"HTTPS"`
}

// WebConfig represents a Web service configuration
type WebConfig struct {
	Handlers map[string]HandlerConfig `json:"Handlers"`
}

// HandlerConfig represents a single handler configuration
type HandlerConfig struct {
	Proxy string `json:"Proxy"`
}

// Service represents a flattened view of a Tailscale service handler
type Service struct {
	ServiceName string // hostname:port
	Hostname    string
	Port        string
	Protocol    string // HTTPS or HTTP
	Target      string // Proxy target (e.g., http://localhost:5173)
	Path        string // Handler path (e.g., / or /opencode)
	FullURL     string // Generated access URL
}

// FilterValue implements list.Item interface
func (s Service) FilterValue() string {
	return s.ServiceName + " " + s.Target + " " + s.Path
}

// Title returns the title for list items
func (s Service) Title() string {
	return s.ServiceName
}

// Description returns the description for list items
func (s Service) Description() string {
	return s.Path + " → " + s.Target
}

// Model represents the Bubble Tea application model
type Model struct {
	services     []Service
	list         list.Model
	lastRefresh  time.Time
	loading      bool
	err          error
	successMsg   string
	showHelp     bool
	showConfirm  bool // Show stop confirmation dialog
	confirmIndex int  // Index of service to stop
	confirmMsg   string
	showInfo     bool // Show info popup
	infoIndex    int  // Index of service being viewed
	width        int
	height       int
}

// Messages for Bubble Tea

type statusMsg struct {
	status TailscaleStatus
	err    error
}

type refreshCompleteMsg struct {
	err error
}

type stopCompleteMsg struct {
	err error
}

type clearSuccessMsg struct{}

type tickMsg struct {
	time.Time
}

// InitialModel creates the initial model state
func InitialModel() Model {
	// Create list with empty items
	const defaultWidth = 80
	const listHeight = 24

	l := list.New([]list.Item{}, serviceDelegate{}, defaultWidth, listHeight)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return Model{
		list:         l,
		loading:      true,
		lastRefresh:  time.Now(),
		services:     []Service{},
		showHelp:     false,
		showConfirm:  false,
		confirmIndex: -1,
		showInfo:     false,
		infoIndex:    -1,
	}
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatus(),
	)
}
