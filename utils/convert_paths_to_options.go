package utils

import (
	"path/filepath"

	"github.com/Haptic-Labs/tmux-sessionizer/ui"
)

// ConvertPathsToOptions extracts just the directory names (not full paths) from a list of paths
func ConvertPathsToOptions(paths []string) []ui.Option {
	opts := []ui.Option{}

	for _, path := range paths {
		opts = append(opts, ui.Option{
			Label: filepath.Base(path),
			Value: path,
		})
	}

	return opts
}
