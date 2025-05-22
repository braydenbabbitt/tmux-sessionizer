package tmux

import (
	"fmt"
	"strconv"

	"github.com/Haptic-Labs/tmux-sessionizer/ui"
	"github.com/Haptic-Labs/tmux-sessionizer/utils"
)

func RunDefaultConfigEditor() (SessionConfig, error) {
	currentConfig, err := LoadConfig()
	if err != nil {
		return currentConfig, fmt.Errorf("failed to find config to edit: %w", err)
	}

	windowOpts := make([]ui.Option, 0)
	for i, window := range currentConfig.Windows {
		windowOpts = append(windowOpts, ui.Option{
			Label: fmt.Sprintf("%d: %s", i, window.Name),
			Value: strconv.Itoa(i),
		})
	}

	selWin := ui.GetSelectionFromList("Select a window in the default config to edit:", windowOpts, false)

	// Clear current view to remove the previous selection UI without clearing terminal history
	utils.ClearCurrentView()

	editOpts := []ui.Option{
		{
			Value: "edit_name",
			Label: "Edit window name",
		},
		{
			Value: "remove",
			Label: "Remove window",
		},
		{
			Value: "edit_cmd",
			Label: "Edit window command",
		},
		{
			Value: "save",
			Label: "Save changes",
		},
	}
	selEdit := ui.GetSelectionFromList("Selected Window "+selWin.Label+":\nSelect an action to perform:", editOpts, true)
	fmt.Println("Selected action:", selEdit)

	return currentConfig, nil
}
