package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var searchInputStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 2)

type searchInputModel struct {
	input textinput.Model
	width int
}

func newSearchInputModel() *searchInputModel {
	input := textinput.New()
	input.Placeholder = "Search processes, ports, or addresses..."
	input.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
	input.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	return &searchInputModel{input: input}
}

func (m *searchInputModel) Init() tea.Cmd { return textinput.Blink }

func (m *searchInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m *searchInputModel) View() string {
	if m.width > 0 {
		m.input.Width = m.width - searchInputStyle.GetHorizontalFrameSize()

		return searchInputStyle.Width(m.width).
			Render(m.input.View())
	}

	return searchInputStyle.Render(m.input.View())
}

func (m *searchInputModel) Focus() tea.Cmd    { return m.input.Focus() }
func (m *searchInputModel) Value() string     { return m.input.Value() }
func (m *searchInputModel) SetValue(s string) { m.input.SetValue(s) }
func (m *searchInputModel) SetWidth(w int)    { m.width = w }
