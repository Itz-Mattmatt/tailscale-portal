package main

import (
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

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
			
		case "?", "h":
			if m.showConfirm {
				m.showConfirm = false
				m.confirmIndex = -1
			}
			if m.showInfo {
				m.showInfo = false
				m.infoIndex = -1
			}
			m.showHelp = !m.showHelp
			return m, nil
			
		case "r", "R":
			if m.showConfirm {
				m.showConfirm = false
				m.confirmIndex = -1
			}
			if m.showInfo {
				m.showInfo = false
				m.infoIndex = -1
			}
			m.loading = true
			m.err = nil
			m.successMsg = ""
			return m, fetchStatus()
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
					m.confirmIndex = m.infoIndex
					m.showConfirm = true
					m.confirmMsg = "Stop service " + service.ServiceName + "?"
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
						m.confirmIndex = i
						m.showConfirm = true
						m.confirmMsg = "Stop service " + item.ServiceName + "?"
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
	}

	// Pass other messages to the list
	if !m.showHelp && !m.showConfirm && !m.showInfo {
		newListModel, cmd := m.list.Update(msg)
		m.list = newListModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
