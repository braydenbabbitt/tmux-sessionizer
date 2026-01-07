package ui

import (
	"fmt"
	"strings"

	config "github.com/Haptic-Labs/tmux-sessionizer/config"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfigUIMode represents the current mode of the config UI
type ConfigUIMode int

const (
	ModeList ConfigUIMode = iota // List all windows
	ModeEdit                     // Edit a window
	ModeAdd                      // Add new window
)

// ConfigModel represents the configuration UI state
type ConfigModel struct {
	Config       *config.Config
	Mode         ConfigUIMode
	Cursor       int
	EditingIndex int
	EditField    int // 0 for name, 1 for command
	NameInput    string
	CommandInput string
	Message      string
	Error        string
	Saved        bool
}

// InitializeConfigModel initializes the config UI model
func InitializeConfigModel(cfg *config.Config) ConfigModel {
	return ConfigModel{
		Config:       cfg,
		Mode:         ModeList,
		Cursor:       0,
		EditingIndex: -1,
		EditField:    0,
		NameInput:    "",
		CommandInput: "",
		Message:      "",
		Error:        "",
		Saved:        false,
	}
}

// Init is the bubbletea initialization function
func (m *ConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Mode {
		case ModeList:
			return m.updateList(msg)
		case ModeEdit, ModeAdd:
			return m.updateEditAdd(msg)
		}
	}
	return m, nil
}

// updateList handles key input in List mode
func (m *ConfigModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		// Quit without saving
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Config.Windows)-1 {
			m.Cursor++
		}
	case "a":
		// Add new window
		m.Mode = ModeAdd
		m.NameInput = ""
		m.CommandInput = ""
		m.EditField = 0
		m.Message = ""
		m.Error = ""
	case "e", "enter", " ":
		// Edit current window
		if len(m.Config.Windows) > 0 && m.Cursor < len(m.Config.Windows) {
			m.Mode = ModeEdit
			m.EditingIndex = m.Cursor
			m.NameInput = m.Config.Windows[m.Cursor].Name
			m.CommandInput = m.Config.Windows[m.Cursor].Command
			m.EditField = 0
			m.Message = ""
			m.Error = ""
		}
	case "d":
		// Delete current window
		if len(m.Config.Windows) > 1 && m.Cursor < len(m.Config.Windows) {
			// Remove the window at cursor position
			m.Config.Windows = append(m.Config.Windows[:m.Cursor], m.Config.Windows[m.Cursor+1:]...)
			// Adjust cursor if needed
			if m.Cursor >= len(m.Config.Windows) {
				m.Cursor = len(m.Config.Windows) - 1
			}
			m.Message = "Window deleted"
		} else if len(m.Config.Windows) == 1 {
			m.Error = "Cannot delete the last window"
		}
	case "ctrl+up", "K":
		// Move window up
		if m.Cursor > 0 && m.Cursor < len(m.Config.Windows) {
			m.Config.Windows[m.Cursor], m.Config.Windows[m.Cursor-1] = m.Config.Windows[m.Cursor-1], m.Config.Windows[m.Cursor]
			m.Cursor--
			m.Message = "Window moved up"
		}
	case "ctrl+down", "J":
		// Move window down
		if m.Cursor >= 0 && m.Cursor < len(m.Config.Windows)-1 {
			m.Config.Windows[m.Cursor], m.Config.Windows[m.Cursor+1] = m.Config.Windows[m.Cursor+1], m.Config.Windows[m.Cursor]
			m.Cursor++
			m.Message = "Window moved down"
		}
	case "s":
		// Save and exit
		if err := config.SaveConfig(m.Config); err != nil {
			m.Error = fmt.Sprintf("Error saving config: %v", err)
		} else {
			m.Saved = true
			m.Message = "Configuration saved!"
			return m, tea.Quit
		}
	}
	return m, nil
}

// updateEditAdd handles key input in Edit/Add mode
func (m *ConfigModel) updateEditAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and return to list
		m.Mode = ModeList
		m.EditingIndex = -1
		m.Message = ""
		m.Error = ""
	case "tab":
		// Switch between fields
		m.EditField = (m.EditField + 1) % 2
	case "enter":
		// Save the window
		if m.NameInput == "" {
			m.Error = "Window name cannot be empty"
			return m, nil
		}

		if m.Mode == ModeEdit {
			// Update existing window
			m.Config.Windows[m.EditingIndex].Name = m.NameInput
			m.Config.Windows[m.EditingIndex].Command = m.CommandInput
			m.Message = "Window updated"
		} else if m.Mode == ModeAdd {
			// Add new window
			newWindow := config.WindowConfig{
				Name:    m.NameInput,
				Command: m.CommandInput,
			}
			m.Config.Windows = append(m.Config.Windows, newWindow)
			m.Cursor = len(m.Config.Windows) - 1
			m.Message = "Window added"
		}

		// Return to list mode
		m.Mode = ModeList
		m.EditingIndex = -1
	case "backspace":
		// Delete character from current field
		if m.EditField == 0 && len(m.NameInput) > 0 {
			m.NameInput = m.NameInput[:len(m.NameInput)-1]
		} else if m.EditField == 1 && len(m.CommandInput) > 0 {
			m.CommandInput = m.CommandInput[:len(m.CommandInput)-1]
		}
	default:
		// Add character to current field
		if len(msg.String()) == 1 {
			if m.EditField == 0 {
				m.NameInput += msg.String()
			} else if m.EditField == 1 {
				m.CommandInput += msg.String()
			}
		}
	}
	return m, nil
}

// View renders the UI
func (m *ConfigModel) View() string {
	var s strings.Builder

	switch m.Mode {
	case ModeList:
		s.WriteString("Configure tmux-sessionizer windows\n")
		s.WriteString("(Use arrows to navigate, Enter to edit, 'a' to add, 'd' to delete, 's' to save)\n\n")

		// Display all windows
		for i, window := range m.Config.Windows {
			cursor := " "
			if m.Cursor == i {
				cursor = ">"
			}

			commandDisplay := "<none>"
			if window.Command != "" {
				commandDisplay = window.Command
			}

			s.WriteString(fmt.Sprintf("%s Window %d: %s (command: %s)\n", cursor, i, window.Name, commandDisplay))
		}

		s.WriteString("\n")

		// Show key bindings
		s.WriteString("[a] Add  [e/Enter] Edit  [d] Delete  [Ctrl+↑/↓ or K/J] Move  [s] Save & Exit  [q] Cancel\n")

		// Show message or error
		if m.Message != "" {
			s.WriteString(fmt.Sprintf("\n%s\n", m.Message))
		}
		if m.Error != "" {
			s.WriteString(fmt.Sprintf("\nError: %s\n", m.Error))
		}

	case ModeEdit:
		s.WriteString(fmt.Sprintf("Editing Window %d\n", m.EditingIndex))
		s.WriteString("(Tab to switch fields, Enter to save, Esc to cancel)\n\n")

		// Name field
		nameCursor := " "
		if m.EditField == 0 {
			nameCursor = "▊"
		}
		s.WriteString(fmt.Sprintf("Name:    %s%s\n", m.NameInput, nameCursor))

		// Command field
		commandCursor := " "
		if m.EditField == 1 {
			commandCursor = "▊"
		}
		s.WriteString(fmt.Sprintf("Command: %s%s\n", m.CommandInput, commandCursor))

		s.WriteString("\n[Tab] Switch field  [Enter] Save  [Esc] Cancel\n")

		if m.Error != "" {
			s.WriteString(fmt.Sprintf("\nError: %s\n", m.Error))
		}

	case ModeAdd:
		s.WriteString("Add New Window\n")
		s.WriteString("(Tab to switch fields, Enter to save, Esc to cancel)\n\n")

		// Name field
		nameCursor := " "
		if m.EditField == 0 {
			nameCursor = "▊"
		}
		s.WriteString(fmt.Sprintf("Name:    %s%s\n", m.NameInput, nameCursor))

		// Command field
		commandCursor := " "
		if m.EditField == 1 {
			commandCursor = "▊"
		}
		s.WriteString(fmt.Sprintf("Command: %s%s\n", m.CommandInput, commandCursor))

		s.WriteString("\n[Tab] Switch field  [Enter] Save  [Esc] Cancel\n")

		if m.Error != "" {
			s.WriteString(fmt.Sprintf("\nError: %s\n", m.Error))
		}
	}

	return s.String()
}
