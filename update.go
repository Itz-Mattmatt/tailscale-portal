package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update list dimensions
		listHeight := msg.Height - 8 // Reserve space for header and footer
		if listHeight < 10 {
			listHeight = 10
		}
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(listHeight)
		m.favList.SetWidth(msg.Width - 8)
		m.favList.SetHeight(listHeight - 4)

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.String() {
		case "q", "ctrl+c":
			// Close dialogs before quitting
			if m.showNewServe {
				m.showNewServe = false
				resetServeForm(&m.serveForm)
				return m, nil
			}
			if m.showFavourites {
				m.showFavourites = false
				return m, nil
			}
			if m.showFavConfirm {
				m.showFavConfirm = false
				m.favConfirmIndex = -1
				return m, nil
			}
			if m.showEditFav {
				m.showEditFav = false
				resetServeForm(&m.editFavForm)
				return m, nil
			}
			return m, tea.Quit
			
		case "?", "h":
			// Don't toggle help when typing in form inputs
			if m.showNewServe || m.showEditFav {
				break
			}
			if m.showConfirm {
				m.showConfirm = false
				m.confirmIndex = -1
			}
			if m.showInfo {
				m.showInfo = false
				m.infoIndex = -1
			}
			if m.showNewServe {
				m.showNewServe = false
				resetServeForm(&m.serveForm)
			}
			if m.showFavourites {
				m.showFavourites = false
			}
			if m.showFavConfirm {
				m.showFavConfirm = false
				m.favConfirmIndex = -1
			}
			if m.showEditFav {
				m.showEditFav = false
				resetServeForm(&m.editFavForm)
			}
			m.showHelp = !m.showHelp
			return m, nil
			
		case "r", "R":
			if m.showNewServe || m.showEditFav {
				return m, nil
			}
			if m.showConfirm {
				m.showConfirm = false
				m.confirmIndex = -1
			}
			if m.showInfo {
				m.showInfo = false
				m.infoIndex = -1
			}
			if m.showNewServe {
				m.showNewServe = false
				resetServeForm(&m.serveForm)
			}
			if m.showFavourites {
				m.showFavourites = false
			}
			if m.showFavConfirm {
				m.showFavConfirm = false
				m.favConfirmIndex = -1
			}
			if m.showEditFav {
				m.showEditFav = false
				resetServeForm(&m.editFavForm)
			}
			m.loading = true
			m.err = nil
			m.successMsg = ""
			return m, fetchStatus()

		case "n":
			if !m.showConfirm && !m.showHelp && !m.showInfo && !m.showNewServe && !m.showFavourites && !m.showFavConfirm && !m.showEditFav {
				m.showNewServe = true
				m.formError = ""
				resetServeForm(&m.serveForm)
				focusServeFormField(&m.serveForm)
				return m, nil
			}

		case "f":
			if m.showHelp || m.showInfo || m.showNewServe || m.showEditFav {
				return m, nil
			}
			m.showFavourites = true
			m.syncFavListItems()
			return m, nil
		}

		// Handle confirmation dialog
		if m.showConfirm {
			switch msg.String() {
			case "y", "Y":
				// Confirm stop
				m.showConfirm = false
				if m.confirmIndex >= 0 && m.confirmIndex < len(m.services) {
					service := m.services[m.confirmIndex]
					m.loading = true
					return m, tea.Batch(
						stopService(service.Port, service.Protocol),
					)
				}
				m.confirmIndex = -1
				return m, nil
				
			case "n", "N", "esc":
				// Cancel
				m.showConfirm = false
				m.confirmIndex = -1
				return m, nil
			}
			return m, nil
		}

		// Handle info popup
		if m.showInfo {
			switch msg.String() {
		case "s":
			// Stop from info popup
			m.showInfo = false
			if m.infoIndex >= 0 && m.infoIndex < len(m.services) {
				service := m.services[m.infoIndex]
				if service.IsForeground {
					m.successMsg = "Foreground services must be stopped in their terminal (Ctrl+C)"
				} else {
					m.confirmIndex = m.infoIndex
					m.showConfirm = true
					m.confirmMsg = "Stop service " + service.ServiceName + "?"
				}
			}
			m.infoIndex = -1
			return m, nil
		case "a":
			m.showInfo = false
			if m.infoIndex >= 0 && m.infoIndex < len(m.services) {
				service := m.services[m.infoIndex]
				setServeFormFromService(&m.serveForm, service)
				m.showNewServe = true
				m.formError = ""
				focusServeFormField(&m.serveForm)
			}
			m.infoIndex = -1
			return m, nil
			case "enter", "esc", "q":
				// Close info popup
				m.showInfo = false
				m.infoIndex = -1
				return m, nil
			}
			return m, nil
		}

		// Handle new serve dialog
		if m.showNewServe {
			switch msg.String() {
			case "esc":
				m.showNewServe = false
				resetServeForm(&m.serveForm)
				m.formError = ""
				return m, nil
			case "tab", "down":
				m.serveForm.focusIndex++
				if m.serveForm.focusIndex > int(fieldCount)+2 { // fields + protocol + mode + checkbox
					m.serveForm.focusIndex = 0
				}
				// Skip name field if not saving
				if m.serveForm.focusIndex == int(fieldName) && !m.serveForm.saveToFav {
					m.serveForm.focusIndex++
				}
				focusServeFormField(&m.serveForm)
				return m, nil
			case "shift+tab", "up":
				m.serveForm.focusIndex--
				if m.serveForm.focusIndex < 0 {
					m.serveForm.focusIndex = int(fieldCount) + 2
				}
				// Skip name field if not saving
				if m.serveForm.focusIndex == int(fieldName) && !m.serveForm.saveToFav {
					m.serveForm.focusIndex--
				}
				focusServeFormField(&m.serveForm)
				return m, nil
			case "left":
				if m.serveForm.focusIndex == int(fieldCount) { // protocol
					if m.serveForm.protocol == "https" {
						m.serveForm.protocol = "http"
					} else {
						m.serveForm.protocol = "https"
					}
					return m, nil
				}
			case "right":
				if m.serveForm.focusIndex == int(fieldCount) { // protocol
					if m.serveForm.protocol == "https" {
						m.serveForm.protocol = "http"
					} else {
						m.serveForm.protocol = "https"
					}
					return m, nil
				}
			case " ":
				if m.serveForm.focusIndex == int(fieldCount)+1 { // mode
					if m.serveForm.mode == "serve" {
						m.serveForm.mode = "funnel"
					} else {
						m.serveForm.mode = "serve"
					}
					return m, nil
				}
				if m.serveForm.focusIndex == int(fieldCount)+2 { // checkbox
					m.serveForm.saveToFav = !m.serveForm.saveToFav
					return m, nil
				}
			case "enter":
				if m.serveForm.focusIndex == int(fieldCount) { // protocol toggle
					if m.serveForm.protocol == "https" {
						m.serveForm.protocol = "http"
					} else {
						m.serveForm.protocol = "https"
					}
					return m, nil
				}
				if m.serveForm.focusIndex == int(fieldCount)+1 { // mode toggle
					if m.serveForm.mode == "serve" {
						m.serveForm.mode = "funnel"
					} else {
						m.serveForm.mode = "serve"
					}
					return m, nil
				}
				cmd := m.submitServeForm(m.serveForm, false, -1)
				if cmd != nil {
					m.showNewServe = false
					resetServeForm(&m.serveForm)
				}
				return m, cmd
			}

			// Pass key messages to the focused text input
			if m.serveForm.focusIndex >= 0 && m.serveForm.focusIndex < int(fieldCount) {
				newInput, inputCmd := m.serveForm.inputs[m.serveForm.focusIndex].Update(msg)
				m.serveForm.inputs[m.serveForm.focusIndex] = newInput
				return m, inputCmd
			}

			return m, nil
		}

		// Handle edit favourite dialog
		if m.showEditFav {
			switch msg.String() {
			case "esc":
				m.showEditFav = false
				resetServeForm(&m.editFavForm)
				m.formError = ""
				return m, nil
			case "tab", "down":
				m.editFavForm.focusIndex++
				if m.editFavForm.focusIndex > int(fieldCount)+1 {
					m.editFavForm.focusIndex = 0
				}
				focusServeFormField(&m.editFavForm)
				return m, nil
			case "shift+tab", "up":
				m.editFavForm.focusIndex--
				if m.editFavForm.focusIndex < 0 {
					m.editFavForm.focusIndex = int(fieldCount) + 1
				}
				focusServeFormField(&m.editFavForm)
				return m, nil
			case "left":
				if m.editFavForm.focusIndex == int(fieldCount) {
					if m.editFavForm.protocol == "https" {
						m.editFavForm.protocol = "http"
					} else {
						m.editFavForm.protocol = "https"
					}
					return m, nil
				}
			case "right":
				if m.editFavForm.focusIndex == int(fieldCount) {
					if m.editFavForm.protocol == "https" {
						m.editFavForm.protocol = "http"
					} else {
						m.editFavForm.protocol = "https"
					}
					return m, nil
				}
			case " ":
				if m.editFavForm.focusIndex == int(fieldCount)+1 {
					if m.editFavForm.mode == "serve" {
						m.editFavForm.mode = "funnel"
					} else {
						m.editFavForm.mode = "serve"
					}
					return m, nil
				}
			case "enter":
				if m.editFavForm.focusIndex == int(fieldCount) {
					if m.editFavForm.protocol == "https" {
						m.editFavForm.protocol = "http"
					} else {
						m.editFavForm.protocol = "https"
					}
					return m, nil
				}
				if m.editFavForm.focusIndex == int(fieldCount)+1 {
					if m.editFavForm.mode == "serve" {
						m.editFavForm.mode = "funnel"
					} else {
						m.editFavForm.mode = "serve"
					}
					return m, nil
				}
				cmd := m.submitServeForm(m.editFavForm, true, m.editFavIndex)
				if cmd != nil {
					m.showEditFav = false
					resetServeForm(&m.editFavForm)
				}
				return m, cmd
			}

			// Pass key messages to the focused text input
			if m.editFavForm.focusIndex >= 0 && m.editFavForm.focusIndex < int(fieldCount) {
				newInput, inputCmd := m.editFavForm.inputs[m.editFavForm.focusIndex].Update(msg)
				m.editFavForm.inputs[m.editFavForm.focusIndex] = newInput
				return m, inputCmd
			}

			return m, nil
		}

		// Handle favourites confirm delete
		if m.showFavConfirm {
			switch msg.String() {
			case "y", "Y":
				m.showFavConfirm = false
				if m.favConfirmIndex >= 0 && m.favConfirmIndex < len(m.favourites) {
					m.favourites = append(m.favourites[:m.favConfirmIndex], m.favourites[m.favConfirmIndex+1:]...)
					m.syncFavListItems()
					return m, saveFavouritesCmd(m.favourites)
				}
				m.favConfirmIndex = -1
				return m, nil
			case "n", "N", "esc":
				m.showFavConfirm = false
				m.favConfirmIndex = -1
				return m, nil
			}
			return m, nil
		}

		// Handle favourites view
		if m.showFavourites {
			switch msg.String() {
			case "esc", "q":
				m.showFavourites = false
				return m, nil
			case "enter":
				if item, ok := m.favList.SelectedItem().(Favourite); ok {
					m.showFavourites = false
					m.loading = true
					return m, startService(item.Port, item.Target, item.Path, item.Protocol, item.Mode)
				}
				return m, nil
			case "e":
				if item, ok := m.favList.SelectedItem().(Favourite); ok {
					for i, fav := range m.favourites {
						if fav.ID == item.ID {
							m.editFavIndex = i
							m.editFavForm.inputs[fieldPort].SetValue(fmt.Sprintf("%d", fav.Port))
							m.editFavForm.inputs[fieldTarget].SetValue(fav.Target)
							m.editFavForm.inputs[fieldPath].SetValue(fav.Path)
							m.editFavForm.inputs[fieldName].SetValue(fav.Name)
							m.editFavForm.protocol = fav.Protocol
							m.editFavForm.mode = fav.Mode
							m.editFavForm.focusIndex = 0
							focusServeFormField(&m.editFavForm)
							m.showFavourites = false
							m.showEditFav = true
							m.formError = ""
							break
						}
					}
				}
				return m, nil
			case "d":
				if item, ok := m.favList.SelectedItem().(Favourite); ok {
					for i, fav := range m.favourites {
						if fav.ID == item.ID {
							m.favConfirmIndex = i
							m.showFavConfirm = true
							m.favConfirmMsg = fmt.Sprintf("Delete favourite '%s'?", fav.Name)
							break
						}
					}
				}
				return m, nil
			}
			// Pass navigation to the favourites list
			newFavList, cmd := m.favList.Update(msg)
			m.favList = newFavList
			return m, cmd
		}

		// Handle list navigation and actions
		switch msg.String() {
		case "enter":
			// Open info popup
			if item, ok := m.list.SelectedItem().(Service); ok {
				// Find the index of this service
				for i, svc := range m.services {
					if svc.ServiceName == item.ServiceName && svc.Path == item.Path {
						m.infoIndex = i
						m.showInfo = true
						break
					}
				}
			}
			return m, nil

		case "s":
			// Stop service - open confirm dialog
			if item, ok := m.list.SelectedItem().(Service); ok {
				// Find the index of this service
				for i, svc := range m.services {
					if svc.ServiceName == item.ServiceName && svc.Path == item.Path {
						if svc.IsForeground {
							m.successMsg = "Foreground services must be stopped in their terminal (Ctrl+C)"
						} else {
							m.confirmIndex = i
							m.showConfirm = true
							m.confirmMsg = "Stop service " + item.ServiceName + "?"
						}
						break
					}
				}
			}
			return m, nil
			
		case "c":
			if item, ok := m.list.SelectedItem().(Service); ok {
				m.successMsg = "Copied URL to clipboard"
				return m, copyToClipboard(item.FullURL)
			}
			return m, nil
		}

	case statusMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.services = []Service{}
			m.list.SetItems(make([]list.Item, 0))
		} else {
			m.err = nil
			m.services = ServicesFromStatus(msg.status)
			items := make([]list.Item, len(m.services))
			for i, svc := range m.services {
				items[i] = list.Item(svc)
			}
			m.list.SetItems(items)
			m.lastRefresh = time.Now()
		}
		return m, nil

	case stopCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.successMsg = ""
		} else {
			m.successMsg = "Service stopped successfully"
			m.err = nil
			// Refresh the list after stopping
			return m, tea.Batch(
				fetchStatus(),
			)
		}
		return m, nil

	case clearSuccessMsg:
		// Clipboard success
		return m, nil

	case tickMsg:
		// Update last refresh time display
		return m, nil

	case startCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.successMsg = ""
		} else {
			m.successMsg = "Service started successfully"
			m.err = nil
			return m, tea.Batch(
				fetchStatus(),
			)
		}
		return m, nil

	case favouritesLoadedMsg:
		if msg.err != nil {
			// Non-fatal: just don't load favourites
			m.err = msg.err
		} else {
			m.favourites = msg.favourites
			m.syncFavListItems()
		}
		return m, nil

	case favouriteSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.successMsg = ""
		} else {
			m.successMsg = "Favourite saved"
		}
		return m, nil
	}

	// Pass other messages to the list
	if !m.showHelp && !m.showConfirm && !m.showInfo && !m.showNewServe && !m.showFavourites && !m.showFavConfirm && !m.showEditFav {
		newListModel, cmd := m.list.Update(msg)
		m.list = newListModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// focusServeFormField blurs all inputs and focuses the one at form.focusIndex
func focusServeFormField(form *serveForm) {
	for i := range form.inputs {
		form.inputs[i].Blur()
	}
	if form.focusIndex >= 0 && form.focusIndex < int(fieldCount) {
		form.inputs[form.focusIndex].Focus()
	}
}

// resetServeForm resets the serve form to default values
func resetServeForm(form *serveForm) {
	for i := range form.inputs {
		form.inputs[i].SetValue("")
		form.inputs[i].Blur()
	}
	form.focusIndex = 0
	form.protocol = "https"
	form.mode = "serve"
	form.saveToFav = false
}

// setServeFormFromService pre-populates the form from a running service
func setServeFormFromService(form *serveForm, s Service) {
	form.inputs[fieldPort].SetValue(s.Port)
	form.inputs[fieldTarget].SetValue(s.Target)
	form.inputs[fieldPath].SetValue(s.Path)
	form.inputs[fieldName].SetValue("")
	form.protocol = strings.ToLower(s.Protocol)
	form.mode = "serve"
	form.saveToFav = true
	form.focusIndex = 0
}

// validateServeForm validates the form fields and returns an error message or empty string
func validateServeForm(form serveForm) string {
	portStr := form.inputs[fieldPort].Value()
	if portStr == "" {
		return "Port is required"
	}
	port := 0
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil || port < 1 || port > 65535 {
		return "Port must be a number between 1 and 65535"
	}

	target := form.inputs[fieldTarget].Value()
	if target == "" {
		return "Target is required"
	}
	// Accept http://, https://, or bare host:port
	isHTTP := strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")
	isHostPort := strings.Contains(target, ":") && !strings.Contains(target, "/")
	if !isHTTP && !isHostPort {
		return "Target must be a URL (http://...) or host:port"
	}

	path := form.inputs[fieldPath].Value()
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "Path must start with /"
	}

	if form.saveToFav {
		name := form.inputs[fieldName].Value()
		if name == "" {
			return "Name is required when saving to favourites"
		}
	}

	return ""
}

// submitServeForm validates and submits the form, returning a command or nil
func (m *Model) submitServeForm(form serveForm, isEdit bool, editIndex int) tea.Cmd {
	if errMsg := validateServeForm(form); errMsg != "" {
		m.formError = errMsg
		return nil
	}

	portStr := form.inputs[fieldPort].Value()
	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	target := form.inputs[fieldTarget].Value()
	path := form.inputs[fieldPath].Value()
	if path == "" {
		path = "/"
	}
	protocol := form.protocol
	mode := form.mode

	var cmds []tea.Cmd

	// Start the service
	cmds = append(cmds, startService(port, target, path, protocol, mode))

	// Save to favourites if checked
	if form.saveToFav && !isEdit {
		name := form.inputs[fieldName].Value()
		newFav := Favourite{
			ID:       generateID(),
			Name:     name,
			Port:     port,
			Target:   target,
			Path:     path,
			Protocol: protocol,
			Mode:     mode,
		}
		m.favourites = append(m.favourites, newFav)
		sortFavourites(m.favourites)
		cmds = append(cmds, saveFavouritesCmd(m.favourites))
	}

	if isEdit && editIndex >= 0 && editIndex < len(m.favourites) {
		name := form.inputs[fieldName].Value()
		m.favourites[editIndex].Name = name
		m.favourites[editIndex].Port = port
		m.favourites[editIndex].Target = target
		m.favourites[editIndex].Path = path
		m.favourites[editIndex].Protocol = protocol
		m.favourites[editIndex].Mode = mode
		sortFavourites(m.favourites)
		cmds = append(cmds, saveFavouritesCmd(m.favourites))
	}

	m.loading = true
	m.formError = ""
	return tea.Batch(cmds...)
}

// syncFavListItems updates the favList items from m.favourites
func (m *Model) syncFavListItems() {
	items := make([]list.Item, len(m.favourites))
	for i, fav := range m.favourites {
		items[i] = list.Item(fav)
	}
	m.favList.SetItems(items)
}
