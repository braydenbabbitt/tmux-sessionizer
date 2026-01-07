package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigChoiceModel represents the choice between global and repo config
type ConfigChoiceModel struct {
	Options  []string
	Cursor   int
	Selected int
}

// InitializeConfigChoiceModel creates a new config choice model
func InitializeConfigChoiceModel() ConfigChoiceModel {
	return ConfigChoiceModel{
		Options:  []string{"Global configuration", "Repo-level configuration"},
		Cursor:   0,
		Selected: -1,
	}
}

// Init is the bubbletea initialization function
func (m *ConfigChoiceModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m *ConfigChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Selected = -1
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Options)-1 {
				m.Cursor++
			}
		case "enter", " ":
			m.Selected = m.Cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *ConfigChoiceModel) View() string {
	var s strings.Builder

	s.WriteString("Choose configuration type:\n\n")

	for i, option := range m.Options {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}
		s.WriteString(cursor + " " + option + "\n")
	}

	s.WriteString("\nPress Enter to select, q to quit.\n")

	return s.String()
}
