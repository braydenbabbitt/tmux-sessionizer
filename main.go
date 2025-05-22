package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Haptic-Labs/tmux-sessionizer/git"
	"github.com/Haptic-Labs/tmux-sessionizer/tmux"
	"github.com/Haptic-Labs/tmux-sessionizer/ui"
	"github.com/Haptic-Labs/tmux-sessionizer/utils"
)

func main() {
	// Variables for command-line options
	var forceAttach bool
	var forceRecreate bool
	var editDefaultConfig bool
	var searchDir string

	// Process all arguments to handle flags in any position
	args := os.Args[1:]
	var nonFlagArgs []string

	for _, arg := range args {
		switch {
		case arg == "-a" || arg == "--attach":
			forceAttach = true
		case arg == "-k" || arg == "--kill":
			forceRecreate = true
		case arg == "--config":
			editDefaultConfig = true
		case !strings.HasPrefix(arg, "-"):
			// This is not a flag, so it's probably a directory
			nonFlagArgs = append(nonFlagArgs, arg)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			fmt.Println("Usage: tmux-sessionizer [directory] [-a|--attach] [-k|--kill] [--config]")
			os.Exit(1)
		}
	}

	if editDefaultConfig {
		newConfig, err := tmux.RunDefaultConfigEditor()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error editing default config: %v\n", err)
			os.Exit(1)
		}

		tmux.SaveConfig(newConfig)
		os.Exit(0)
	}

	if len(nonFlagArgs) > 0 {
		// Use the provided directory
		providedDir := nonFlagArgs[0]
		absDir, err := filepath.Abs(providedDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}
		searchDir = absDir
	} else {
		// Use current directory as default
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		searchDir = currentDir
	}

	fmt.Printf("Searching for git repositories in: %s\n", searchDir)

	// Find git repositories
	repos, err := git.FindGitRepos(searchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding git repositories: %v\n", err)
		os.Exit(1)
	}

	if len(repos) == 0 {
		fmt.Println("No git repositories found.")
		os.Exit(0)
	}

	// Get directory names and convert to options
	options := utils.ConvertPathsToOptions(repos)

	// Sort the options by last modified time (most recent first)
	sort.Slice(options, func(i, j int) bool {
		timeI := utils.GetLastModifiedTime(options[i].Value)
		timeJ := utils.GetLastModifiedTime(options[j].Value)
		return timeI.After(timeJ) // Sort in descending order (newest first)
	})

	selected := ui.GetSelectionFromList("Select a git repository:", options, false)
	selectedPath := selected.Value

	// Create tmux session
	err = tmux.CreateTmuxSession(selected.Label, selectedPath, forceAttach, forceRecreate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
		os.Exit(1)
	}
}
