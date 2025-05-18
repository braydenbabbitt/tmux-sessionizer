package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	git "github.com/Haptic-Labs/tmux-sessionizer/git"
	tmux "github.com/Haptic-Labs/tmux-sessionizer/tmux"
	ui "github.com/Haptic-Labs/tmux-sessionizer/ui"
	utils "github.com/Haptic-Labs/tmux-sessionizer/utils"
)

func main() {
	// Define flags
	var forceAttach bool
	var forceRecreate bool
	var editDefaultConfig bool

	flag.BoolVar(&forceAttach, "a", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceAttach, "attach", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceRecreate, "k", false, "Automatically kill and recreate existing session if it exists")
	flag.BoolVar(&forceRecreate, "kill", false, "Automatically kill and recreate existing session if it exists")
	flag.BoolVar(&editDefaultConfig, "config", false, "Edit the default session config")

	// Parse flags
	flag.Parse()

	var searchDir string

	// Check if a directory argument was provided
	args := flag.Args()
	if len(args) > 0 {
		// Use the provided directory
		providedDir := args[0]
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

	selected := ui.GetSelectionFromList("Select a git repository:", options)
	selectedPath := selected.Value

	// Create tmux session
	err = tmux.CreateTmuxSession(selected.Label, selectedPath, forceAttach, forceRecreate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
		os.Exit(1)
	}
}
