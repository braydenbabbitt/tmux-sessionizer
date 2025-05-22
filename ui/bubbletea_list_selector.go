package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func GetSelectionFromList(title string, options []Option, disableSearch bool) Option {
	model := InitializeModel(title, options, disableSearch)
	p := tea.NewProgram(&model)

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running bubbletea program: %v\n", err)
		os.Exit(1)
	}

	m, ok := result.(*BubbleteaModel)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: could not get model from program\n")
		os.Exit(1)
	}

	if m.Selected == -1 {
		fmt.Println("No option selected.")
		os.Exit(0)
	}

	return options[m.Selected]
}
