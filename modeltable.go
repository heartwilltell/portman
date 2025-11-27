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
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

var headerLeftStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("205")).
	Padding(0, 2)

var headerRightStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240")).
	Padding(0, 2)

var confirmBoxStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("205")).
	Padding(1, 3)

type filterState struct {
	tcpOnly         bool
	udpOnly         bool
	listenOnly      bool
	establishedOnly bool
}

type statusKind int

const (
	statusKindInfo statusKind = iota
	statusKindError
)

func (f *filterState) toggleTCP() {
	f.tcpOnly = !f.tcpOnly
	if f.tcpOnly {
		f.udpOnly = false
	}
}

func (f *filterState) toggleUDP() {
	f.udpOnly = !f.udpOnly
	if f.udpOnly {
		f.tcpOnly = false
		f.listenOnly = false
	}
}

func (f *filterState) toggleListen() {
	f.listenOnly = !f.listenOnly
	if f.listenOnly {
		f.establishedOnly = false
	}
}

func (f *filterState) toggleEstablished() {
	f.establishedOnly = !f.establishedOnly
	if f.establishedOnly {
		f.listenOnly = false
	}
}

func (f *filterState) clear() {
	f.tcpOnly = false
	f.udpOnly = false
	f.listenOnly = false
	f.establishedOnly = false
}

func (f filterState) allows(p Process) bool {
	protocol := strings.ToUpper(p.Protocol)
	status := strings.ToUpper(p.Status)

	if f.tcpOnly && !strings.HasPrefix(protocol, "TCP") {
		return false
	}
	if f.udpOnly && !strings.HasPrefix(protocol, "UDP") {
		return false
	}
	if f.listenOnly && status != "LISTEN" {
		return false
	}
	if f.establishedOnly && status != "ESTABLISHED" {
		return false
	}

	return true
}

func (f filterState) activeLabels() []string {
	labels := make([]string, 0, 4)
	if f.tcpOnly {
		labels = append(labels, "TCP")
	}
	if f.udpOnly {
		labels = append(labels, "UDP")
	}
	if f.listenOnly {
		labels = append(labels, "LISTEN")
	}
	if f.establishedOnly {
		labels = append(labels, "ESTABLISHED")
	}
	return labels
}

type tableModel struct {
	pm               *ProcessManager
	table            table.Model
	searchInput      *searchInputModel
	width            int
	height           int
	tableViewWidth   int
	showSearch       bool
	searchQuery      string
	allProcesses     []Process
	filteredRowCount int
	filters          filterState
	statusMessage    string
	statusKind       statusKind
	statusExpires    time.Time
	confirmKill      bool
	confirmTarget    Process
	horizontalScroll int
}

func newTableModel(pm *ProcessManager) *tableModel {
	columns := []table.Column{
		{Title: "PID", Width: 5},
		{Title: "Protocol", Width: 8},
		{Title: "Port", Width: 8},
		{Title: "Status", Width: 15},
		{Title: "Local Address", Width: 15},
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
		pm:          pm,
		table:       t,
		searchInput: searchInput,
		showSearch:  false,
	}
	m.filters.tcpOnly = true
	m.filters.listenOnly = true

	return &m
}

func (m *tableModel) setStatusMessage(text string, kind statusKind) {
	m.statusMessage = text
	m.statusKind = kind
	m.statusExpires = time.Now().Add(4 * time.Second)
}

func (m *tableModel) updateTableSize() {
	if m.width <= 0 {
		return
	}

	type columnSpec struct {
		title  string
		min    int
		weight int
	}

	specs := []columnSpec{
		{title: "PID", min: 6, weight: 0},
		{title: "Protocol", min: 8, weight: 0},
		{title: "Port", min: 6, weight: 0},
		{title: "Status", min: 12, weight: 1},
		{title: "Local Address", min: 18, weight: 0},
		{title: "Process", min: 18, weight: 6},
	}

	frameWidth := baseStyle.GetHorizontalFrameSize()
	// Account for table column spacing (bubble tea table adds spaces between columns)
	// Approximate: 3 spaces between each column
	columnSpacing := (len(specs) - 1) * 3

	availableWidth := m.width - frameWidth - columnSpacing
	if availableWidth < 0 {
		availableWidth = 0
	}

	totalMin := 0
	totalWeight := 0
	for _, spec := range specs {
		totalMin += spec.min
		totalWeight += spec.weight
	}

	if availableWidth < totalMin {
		availableWidth = totalMin
	}

	extraWidth := availableWidth - totalMin
	if extraWidth < 0 {
		extraWidth = 0
	}

	if extraWidth > 0 {
		scaled := int(float64(extraWidth) * 0.8)
		if scaled < 0 {
			scaled = 0
		}
		if scaled > extraWidth {
			scaled = extraWidth
		}
		extraWidth = scaled
	}
	shares := make([]int, len(specs))

	if extraWidth > 0 && totalWeight > 0 {
		assigned := 0
		for i, spec := range specs {
			share := extraWidth * spec.weight / totalWeight
			shares[i] = share
			assigned += share
		}

		leftover := extraWidth - assigned
		for i := 0; leftover > 0 && i < len(specs); i++ {
			if specs[i].weight == 0 {
				continue
			}
			shares[i]++
			leftover--
		}
	}

	columns := make([]table.Column, len(specs))
	totalWidth := 0

	for i, spec := range specs {
		width := spec.min + shares[i]
		if width < spec.min {
			width = spec.min
		}
		columns[i] = table.Column{Title: spec.title, Width: width}
		totalWidth += width
	}

	m.table.SetColumns(columns)
	m.tableViewWidth = totalWidth + frameWidth + columnSpacing
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
		m.horizontalScroll = 0 // Reset horizontal scroll on resize
		m.updateTableSize()

	case tea.KeyMsg:
		if m.confirmKill {
			switch msg.String() {
			case "y", "enter":
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				pid := m.confirmTarget.PID
				name := m.confirmTarget.Name
				m.confirmKill = false
				defer cancel()
				if err := m.pm.KillProcess(ctx, pid); err != nil {
					m.setStatusMessage(fmt.Sprintf("Kill failed: %v", err), statusKindError)
					return m, nil
				}
				if name != "" {
					m.setStatusMessage(fmt.Sprintf("Killed %s (%d)", name, pid), statusKindInfo)
				} else {
					m.setStatusMessage(fmt.Sprintf("Killed PID %d", pid), statusKindInfo)
				}
				return m, nil
			case "n", "esc":
				m.confirmKill = false
				m.setStatusMessage("Kill cancelled", statusKindInfo)
				return m, nil
			}
		}

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

		switch msg.String() {
		case "/":
			m.showSearch = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "t":
			m.filters.toggleTCP()
			return m, nil
		case "u":
			m.filters.toggleUDP()
			return m, nil
		case "l":
			m.filters.toggleListen()
			return m, nil
		case "e":
			m.filters.toggleEstablished()
			return m, nil
		case "x":
			m.filters.clear()
			return m, nil
		case "k":
			row := m.table.SelectedRow()
			if len(row) == 0 {
				m.setStatusMessage("No process selected", statusKindError)
				return m, nil
			}
			pid, err := strconv.Atoi(row[0])
			if err != nil {
				m.setStatusMessage("Invalid PID", statusKindError)
				return m, nil
			}
			target := Process{PID: pid}
			if len(row) > 5 {
				target.Name = row[5]
			}
			if len(row) > 3 {
				target.Status = row[3]
			}
			if len(row) > 4 {
				target.LocalAddr = row[4]
			}
			m.confirmKill = true
			m.confirmTarget = target
			return m, nil

		case "shift+left":
			if m.horizontalScroll > 0 {
				m.horizontalScroll--
			}
			return m, nil

		case "shift+right":
			m.horizontalScroll++
			return m, nil

		case "esc":
			if m.horizontalScroll > 0 {
				m.horizontalScroll = 0
				return m, nil
			}
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
	m.filteredRowCount = len(filteredProcesses)

	rows := make([]table.Row, 0, len(filteredProcesses))

	// Get process column width for scrolling
	cols := m.table.Columns()
	processColWidth := 15 // default
	if len(cols) > 5 {
		processColWidth = cols[5].Width
	}

	for _, process := range filteredProcesses {
		// Apply horizontal scroll to process name
		processName := scrollText(process.Name, m.horizontalScroll, processColWidth)

		rows = append(rows, table.Row{
			strconv.Itoa(process.PID),
			process.Protocol,
			strconv.Itoa(process.Port),
			process.Status,
			process.LocalAddr,
			processName,
		})
	}

	m.table.SetRows(rows)

	// Build the main view
	rawTableView := m.table.View()
	frameWidth := baseStyle.GetHorizontalFrameSize()
	tableBodyWidth := lipgloss.Width(rawTableView)

	tableWidth := m.tableViewWidth
	if tableWidth == 0 {
		tableWidth = tableBodyWidth + frameWidth
	}
	if tableWidth < tableBodyWidth+frameWidth {
		tableWidth = tableBodyWidth + frameWidth
	}

	// Ensure table width doesn't exceed terminal width
	if m.width > 0 && tableWidth > m.width {
		tableWidth = m.width
	}

	tableView := baseStyle.Width(tableWidth).Render(rawTableView)

	versionLabel := Version
	if versionLabel == "" {
		versionLabel = "dev"
	}
	if versionLabel != "dev" {
		if lower := strings.ToLower(versionLabel); !strings.HasPrefix(lower, "v") {
			versionLabel = "v" + versionLabel
		}
	}

	shortcuts := "[/] Search  [t] TCP  [u] UDP  [l] LISTEN  [e] EST  [k] Kill"
	left := headerLeftStyle.Render(fmt.Sprintf("%s %s", appName, versionLabel))
	leftWidth := lipgloss.Width(left)
	rightSpace := tableWidth - leftWidth
	if rightSpace < 0 {
		rightSpace = 0
	}
	rightRaw := headerRightStyle.Render(shortcuts)
	rightTrimmed := lipgloss.NewStyle().MaxWidth(rightSpace).Render(rightRaw)
	right := lipgloss.PlaceHorizontal(rightSpace, lipgloss.Right, rightTrimmed)
	headerView := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	headerView = lipgloss.NewStyle().MaxWidth(tableWidth).Render(headerView)

	sections := []string{headerView}
	if m.showSearch {
		m.searchInput.SetWidth(tableWidth)
		sections = append(sections, m.searchInput.View())
	}

	tableContent := tableView
	if m.confirmKill {
		tableContent = overlayConfirmBox(tableWidth, tableView, m.confirmTarget)
	}
	sections = append(sections, tableContent)

	mainView := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add status bar
	statusBar := m.renderStatusBar()
	inner := lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)
	mainView = lipgloss.JoinVertical(lipgloss.Left, "", inner)

	return mainView
}

func (m *tableModel) filterProcesses(processes []Process) []Process {
	tokens := m.searchTokens()
	filtered := make([]Process, 0, len(processes))

	for _, process := range processes {
		if !m.filters.allows(process) {
			continue
		}
		if !matchesTokens(process, tokens) {
			continue
		}
		filtered = append(filtered, process)
	}

	return filtered
}

func (m *tableModel) searchTokens() []string {
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	if query == "" {
		return nil
	}
	return strings.Fields(query)
}

func matchesTokens(process Process, tokens []string) bool {
	if len(tokens) == 0 {
		return true
	}

	fieldValues := []string{
		strings.ToLower(process.Name),
		strconv.Itoa(process.PID),
		strings.ToLower(process.Protocol),
		strconv.Itoa(process.Port),
		strings.ToLower(process.Status),
		strings.ToLower(process.LocalAddr),
	}

	for _, token := range tokens {
		found := false
		for _, field := range fieldValues {
			if strings.Contains(field, token) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// scrollText applies horizontal scrolling to text, showing a window of maxWidth characters
// starting at offset position. Always returns exactly maxWidth characters to prevent wrapping.
func scrollText(text string, offset int, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// If text fits, pad it to maxWidth
	if len(text) <= maxWidth {
		if len(text) < maxWidth {
			return text + strings.Repeat(" ", maxWidth-len(text))
		}
		return text
	}

	// Apply offset
	if offset >= len(text) {
		offset = len(text) - maxWidth
		if offset < 0 {
			offset = 0
		}
	}

	end := offset + maxWidth
	if end > len(text) {
		// If we're at the end, show the last maxWidth characters
		result := text[len(text)-maxWidth:]
		return result
	}

	return text[offset:end]
}

func (m *tableModel) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	if m.confirmKill {
		prompt := fmt.Sprintf("Kill %s (%d)? [y/N]", displayName(m.confirmTarget.Name), m.confirmTarget.PID)
		return statusStyle.Render(prompt)
	}

	if m.statusMessage != "" && time.Now().Before(m.statusExpires) {
		style := statusStyle
		if m.statusKind == statusKindError {
			style = style.Foreground(lipgloss.Color("203"))
		} else {
			style = style.Foreground(lipgloss.Color("205"))
		}

		return style.Render(m.statusMessage)
	}

	status := "[q] Quit :: [x] Clear :: [Shift+←/→] Scroll"
	if labels := m.filters.activeLabels(); len(labels) > 0 {
		status += "  |  " + strings.Join(labels, ", ")
	}

	// Add horizontal scroll position indicator
	if m.horizontalScroll > 0 {
		status += fmt.Sprintf("  |  ←→ (%d)", m.horizontalScroll)
	}

	// Add scroll indicators if there are more rows than visible
	tableHeight := m.table.Height()
	if m.filteredRowCount > tableHeight {
		cursor := m.table.Cursor()

		// Determine scroll position
		var scrollIndicator string
		if cursor > 0 && cursor < m.filteredRowCount-1 {
			scrollIndicator = "↑↓"
		} else if cursor == 0 && m.filteredRowCount > tableHeight {
			scrollIndicator = "↓"
		} else if cursor >= m.filteredRowCount-1 {
			scrollIndicator = "↑"
		}

		if scrollIndicator != "" {
			status += fmt.Sprintf("  |  ↑↓ (%d/%d)",
				cursor+1, m.filteredRowCount)
		}
	}

	return statusStyle.Render(status)
}

func overlayConfirmBox(width int, tableView string, target Process) string {
	box := renderConfirmBox(target)

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	background := dim.Render(tableView)

	overlay := lipgloss.PlaceHorizontal(width, lipgloss.Center, box)

	return lipgloss.JoinVertical(lipgloss.Center, overlay, background)
}
