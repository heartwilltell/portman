package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type filterPickerModel struct {
	list list.Model
}

func newFilterPickerModel() *filterPickerModel {
	items := []list.Item{
		item{title: "None", desc: "Remove all filters"},
		item{title: "Only listening ports", desc: "Show only processes that are listening on a port"},
		item{title: "Only established connections", desc: "Show only processes with established connections"},
		item{title: "Show only TCP connections", desc: "Show only TCP connections"},
		item{title: "Show only UDP connections", desc: "Show only UDP connections"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a filter"
	l.KeyMap.Filter.SetEnabled(false)
	l.KeyMap.Quit.SetEnabled(false)
	l.KeyMap.ShowFullHelp.SetEnabled(false)
	l.SetShowStatusBar(false)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "close"),
			),
		}
	}

	return &filterPickerModel{list: l}
}

func (m *filterPickerModel) Init() tea.Cmd {
	return nil
}

func (m *filterPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *filterPickerModel) View() string {
	return m.list.View()
}

func (m *filterPickerModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// item implements list.Item
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
