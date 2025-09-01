package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

type tableModel struct {
	pm               *ProcessManager
	table            table.Model
	searchInput      *searchInputModel
	filterPicker     *filterPickerModel
	width            int
	height           int
	showSearch       bool
	showFilterPicker bool
	searchQuery      string
	activeFilter     string
	allProcesses     []Process
}

func newTableModel(pm *ProcessManager) *tableModel {
	columns := []table.Column{
		{Title: "PID", Width: 5},
		{Title: "Protocol", Width: 8},
		{Title: "Port", Width: 8},
		{Title: "Status", Width: 15},
		{Title: "Local Address", Width: 15},
		{Title: "Remote Address", Width: 15},
		{Title: "Process", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()

	// s.Header = s.Header.
	// 	BorderStyle(lipgloss.NormalBorder()).
	// 	BorderForeground(lipgloss.Color("240")).
	// 	BorderBottom(true).
	// 	Bold(false)

	// s.Selected = s.Selected.
	// 	Foreground(lipgloss.Color("229")).
	// 	Background(lipgloss.Color("57")).
	// 	Bold(false)

	t.SetStyles(s)

	// Initialize search input
	searchInput := newSearchInputModel()

	m := tableModel{
		pm:               pm,
		table:            t,
		searchInput:      searchInput,
		filterPicker:     newFilterPickerModel(),
		showSearch:       false,
		showFilterPicker: false,
		activeFilter:     "",
	}

	return &m
}

func (m *tableModel) updateTableSize() {
	// if m.width == 0 || m.height == 0 {
	// 	return
	// }

	// // Calculate available height (subtract padding and borders)
	// availableHeight := m.height - 4 // Account for padding, borders, and status line

	// // Account for search input if visible
	// if m.showSearch {
	// 	availableHeight -= 2 // Additional space for search input and results count
	// }

	// if availableHeight < 5 {
	// 	availableHeight = 5 // Minimum height
	// }

	// // Calculate column widths based on terminal width
	// availableWidth := m.width - 8 // Account for borders and padding

	// // Distribute width proportionally
	// pidWidth := 5
	// portWidth := 8
	// protocolWidth := 8
	// statusWidth := 15

	// // Calculate remaining width for process name and addresses
	// remainingWidth := availableWidth - pidWidth - portWidth - protocolWidth - statusWidth - 15 // spaces between columns

	// processWidth := remainingWidth / 3
	// localAddrWidth := remainingWidth / 3
	// remoteAddrWidth := remainingWidth / 3

	// // Ensure minimum widths
	// if processWidth < 15 {
	// 	processWidth = 15
	// }
	// if localAddrWidth < 15 {
	// 	localAddrWidth = 15
	// }
	// if remoteAddrWidth < 15 {
	// 	remoteAddrWidth = 15
	// }

	// columns := []table.Column{
	// 	{Title: "PID", Width: pidWidth},
	// 	{Title: "Protocol", Width: protocolWidth},
	// 	{Title: "Port", Width: portWidth},
	// 	{Title: "Status", Width: statusWidth},
	// 	{Title: "Local Address", Width: localAddrWidth},
	// 	{Title: "Remote Address", Width: remoteAddrWidth},
	// 	{Title: "Process", Width: processWidth},
	// }

	// m.table.SetColumns(columns)
	// m.table.SetHeight(availableHeight)
}

func (m *tableModel) Init() tea.Cmd {
	// Start with a tick to refresh the UI periodically.
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// tickMsg is a message that triggers a UI refresh.
type tickMsg time.Time

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Return a command to tick again.
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
		m.filterPicker.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if m.showSearch {
			switch msg.String() {
			case "esc":
				m.showSearch = false
				m.searchInput.SetValue("")
				m.searchQuery = ""
				m.table.Focus()
				return m, nil

			case "enter":
				m.showSearch = false
				m.searchQuery = m.searchInput.Value()
				m.table.Focus()
				return m, nil
			}
			// Update search input and apply search in real-time
			var updatedModel tea.Model
			updatedModel, cmd = m.searchInput.Update(msg)
			m.searchInput = updatedModel.(*searchInputModel)
			m.searchQuery = m.searchInput.Value()
			return m, cmd
		}

		if m.showFilterPicker {
			var updatedModel tea.Model
			updatedModel, cmd = m.filterPicker.Update(msg)
			m.filterPicker = updatedModel.(*filterPickerModel)

			switch msg.String() {
			case "enter":
				selectedItem := m.filterPicker.list.SelectedItem().(item)
				m.activeFilter = selectedItem.FilterValue()
				if m.activeFilter == "None" {
					m.activeFilter = ""
				}
				m.showFilterPicker = false
				return m, nil
			case "esc":
				m.showFilterPicker = false
				return m, nil
			}
			return m, cmd
		}

		switch msg.String() {
		case ":":
			m.showFilterPicker = true
			m.filterPicker.SetSize(m.width, m.height)
			return m, nil
		case "/":
			m.showSearch = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}

		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m *tableModel) View() string {
	if m.showFilterPicker {
		return m.filterPicker.View()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	processes, err := m.pm.Processes(ctx)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	// Store all processes for filtering
	m.allProcesses = processes

	// Filter processes based on search query
	filteredProcesses := m.filterProcesses(processes)

	rows := make([]table.Row, 0, len(filteredProcesses))

	for _, process := range filteredProcesses {
		rows = append(rows, table.Row{
			strconv.Itoa(process.PID),
			process.Protocol,
			strconv.Itoa(process.Port),
			process.Status,
			process.LocalAddr,
			process.RemoteAddr,
			process.Name,
		})
	}

	m.table.SetRows(rows)

	// Build the main view
	var mainView string
	if m.showSearch {
		tableWidth := 0
		for _, c := range m.table.Columns() {
			tableWidth += c.Width
		}
		tableWidth += baseStyle.GetHorizontalFrameSize()
		m.searchInput.SetWidth(tableWidth)

		mainView = lipgloss.JoinVertical(lipgloss.Left,
			m.searchInput.View(),
			baseStyle.Render(m.table.View()),
		)
	} else {
		mainView = baseStyle.Render(m.table.View())
	}

	// Add status bar
	statusBar := m.renderStatusBar()
	mainView = mainView + "\n" + statusBar

	return mainView
}

func (m *tableModel) filterProcesses(processes []Process) []Process {
	var filtered []Process

	if m.searchQuery == "" {
		filtered = processes
	} else {
		query := strings.ToLower(m.searchQuery)
		for _, process := range processes {
			// Search in process name, port, local address, remote address, and protocol.
			if strings.Contains(strings.ToLower(process.Name), query) ||
				strings.Contains(strings.ToLower(strconv.Itoa(process.Port)), query) ||
				strings.Contains(strings.ToLower(process.LocalAddr), query) ||
				strings.Contains(strings.ToLower(process.RemoteAddr), query) ||
				strings.Contains(strings.ToLower(process.Protocol), query) ||
				strings.Contains(strings.ToLower(process.Status), query) {
				filtered = append(filtered, process)
			}
		}
	}

	if m.activeFilter != "" {
		switch m.activeFilter {
		case "Only listening ports":
			var listeningOnly []Process
			for _, p := range filtered {
				if p.Status == "LISTEN" {
					listeningOnly = append(listeningOnly, p)
				}
			}
			filtered = listeningOnly
		case "Only established connections":
			var establishedOnly []Process
			for _, p := range filtered {
				if p.Status == "ESTABLISHED" {
					establishedOnly = append(establishedOnly, p)
				}
			}
			filtered = establishedOnly
		case "Show only TCP connections":
			var tcpOnly []Process
			for _, p := range filtered {
				if strings.HasPrefix(p.Protocol, "TCP") {
					tcpOnly = append(tcpOnly, p)
				}
			}
			filtered = tcpOnly
		case "Show only UDP connections":
			var udpOnly []Process
			for _, p := range filtered {
				if strings.HasPrefix(p.Protocol, "UDP") {
					udpOnly = append(udpOnly, p)
				}
			}
			filtered = udpOnly
		}
	}

	return filtered
}

func (m *tableModel) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	shortcuts := "[/] Search  [q] Quit"
	if m.activeFilter != "" {
		shortcuts += "  |  Filter: " + m.activeFilter
	}

	return statusStyle.Render(shortcuts)
}
