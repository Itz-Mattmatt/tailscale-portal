package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TailscaleStatus represents the JSON output from `tailscale serve status --json`
type TailscaleStatus struct {
	TCP        map[string]TCPConfig       `json:"TCP"`
	Web        map[string]WebConfig       `json:"Web"`
	Foreground map[string]TailscaleStatus `json:"Foreground"`
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
	ServiceName  string // hostname:port
	Hostname     string
	Port         string
	Protocol     string // HTTPS or HTTP
	Target       string // Proxy target (e.g., http://localhost:5173)
	Path         string // Handler path (e.g., / or /opencode)
	FullURL      string // Generated access URL
	IsForeground bool   // True if this is a foreground (non-background) service
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

// Favourite represents a saved service configuration
type Favourite struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Target   string `json:"target"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	Mode     string `json:"mode"`
}

// FilterValue implements list.Item interface
func (f Favourite) FilterValue() string {
	return f.Name
}

// Title returns the title for list items
func (f Favourite) Title() string {
	return f.Name
}

// Description returns the description for list items
func (f Favourite) Description() string {
	return f.Target + f.Path
}

// FavouritesFile is the root JSON structure
type FavouritesFile struct {
	Version    string      `json:"version"`
	Favourites []Favourite `json:"favourites"`
}

// formField identifies which field is focused in the serve form
type formField int

const (
	fieldPort formField = iota
	fieldTarget
	fieldPath
	fieldName
	fieldCount
)

// serveForm holds the state of the new serve / edit favourite dialog
type serveForm struct {
	inputs     [fieldCount]textinput.Model
	focusIndex int
	protocol   string
	mode       string
	saveToFav  bool
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

	// New serve dialog
	showNewServe bool
	serveForm    serveForm

	// Favourites
	favourites     []Favourite
	showFavourites bool
	favList        list.Model

	// Favourites confirm delete
	showFavConfirm  bool
	favConfirmIndex int
	favConfirmMsg   string

	// Edit favourite
	showEditFav  bool
	editFavIndex int
	editFavForm  serveForm

	// Form error
	formError string
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

type startCompleteMsg struct {
	err error
}

type favouritesLoadedMsg struct {
	favourites []Favourite
	err        error
}

type favouriteSavedMsg struct {
	err error
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

	favList := list.New([]list.Item{}, favouriteDelegate{}, defaultWidth, listHeight)
	favList.Title = ""
	favList.SetShowTitle(false)
	favList.SetShowStatusBar(false)
	favList.SetFilteringEnabled(false)
	favList.SetShowPagination(false)
	favList.Styles.PaginationStyle = paginationStyle
	favList.Styles.HelpStyle = helpStyle

	var serveFormInputs [fieldCount]textinput.Model
	for i := range serveFormInputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colorHighlight))
		t.CharLimit = 156
		switch formField(i) {
		case fieldPort:
			t.Placeholder = "443"
			t.CharLimit = 5
		case fieldTarget:
			t.Placeholder = "localhost:3000"
		case fieldPath:
			t.Placeholder = "/"
		case fieldName:
			t.Placeholder = "My Dev Server"
		}
		serveFormInputs[i] = t
	}

	var editFavFormInputs [fieldCount]textinput.Model
	for i := range editFavFormInputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colorHighlight))
		t.CharLimit = 156
		switch formField(i) {
		case fieldPort:
			t.Placeholder = "443"
			t.CharLimit = 5
		case fieldTarget:
			t.Placeholder = "localhost:3000"
		case fieldPath:
			t.Placeholder = "/"
		case fieldName:
			t.Placeholder = "My Dev Server"
		}
		editFavFormInputs[i] = t
	}

	return Model{
		list:            l,
		favList:         favList,
		loading:         true,
		lastRefresh:     time.Now(),
		services:        []Service{},
		showHelp:        false,
		showConfirm:     false,
		confirmIndex:    -1,
		showInfo:        false,
		infoIndex:       -1,
		showNewServe:    false,
		serveForm: serveForm{
			inputs:     serveFormInputs,
			focusIndex: 0,
			protocol:   "https",
			mode:       "serve",
			saveToFav:  false,
		},
		favourites:      []Favourite{},
		showFavourites:  false,
		showFavConfirm:  false,
		favConfirmIndex: -1,
		showEditFav:     false,
		editFavIndex:    -1,
		editFavForm: serveForm{
			inputs:     editFavFormInputs,
			focusIndex: 0,
			protocol:   "https",
			mode:       "serve",
			saveToFav:  false,
		},
	}
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatus(),
		loadFavouritesCmd(),
	)
}
