package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	config "github.com/Haptic-Labs/tmux-sessionizer/config"
	git "github.com/Haptic-Labs/tmux-sessionizer/git"
	tmux "github.com/Haptic-Labs/tmux-sessionizer/tmux"
	ui "github.com/Haptic-Labs/tmux-sessionizer/ui"
	utils "github.com/Haptic-Labs/tmux-sessionizer/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Define flags
	var forceAttach bool
	var forceRecreate bool
	var useCurrent bool
	var configMode bool

	flag.BoolVar(&forceAttach, "a", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceAttach, "attach", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceRecreate, "k", false, "Automatically kill and recreate existing session if it exists")
	flag.BoolVar(&forceRecreate, "kill", false, "Automatically kill and recreate existing session if it exists")
	flag.BoolVar(&useCurrent, "c", false, "Use current directory for session (skip directory selection)")
	flag.BoolVar(&useCurrent, "current", false, "Use current directory for session (skip directory selection)")
	flag.BoolVar(&configMode, "config", false, "Open interactive configuration UI")

	// Parse flags
	flag.Parse()

	// If --config flag is set, launch configuration UI
	if configMode {
		if useCurrent {
			// --config -c: Edit current directory's repo config directly
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			// Validate current directory is a git repo
			if !git.IsGitRepo(currentDir) {
				fmt.Fprintf(os.Stderr, "Error: Current directory is not a git repository\n")
				fmt.Fprintf(os.Stderr, "Repo-level config requires a git repository. Use --config without -c for global config.\n")
				os.Exit(1)
			}

			// Load or create repo config
			cfg, err := config.LoadRepoConfig(currentDir)
			if err != nil {
				// No repo config exists, start with global config
				cfg, _ = config.LoadConfig()
			}

			// Launch config UI with repo context
			model := ui.InitializeRepoConfigModel(cfg, currentDir)
			p := tea.NewProgram(&model)
			_, err = p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running config UI: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// --config (without -c): Show global vs repo choice
		choiceModel := ui.InitializeConfigChoiceModel()
		p := tea.NewProgram(&choiceModel)
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running config choice UI: %v\n", err)
			os.Exit(1)
		}

		cm, ok := result.(*ui.ConfigChoiceModel)
		if !ok || cm.Selected == -1 {
			// User cancelled
			return
		}

		if cm.Selected == 0 {
			// Global config selected
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			model := ui.InitializeConfigModel(cfg)
			p := tea.NewProgram(&model)
			_, err = p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running config UI: %v\n", err)
				os.Exit(1)
			}
			return
		} else {
			// Repo-level config selected - show repo picker
			// Get search directory (same logic as normal flow)
			var searchDir string
			args := flag.Args()
			if len(args) > 0 {
				providedDir := args[0]
				absDir, err := filepath.Abs(providedDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
					os.Exit(1)
				}
				searchDir = absDir
			} else {
				currentDir, err := os.Getwd()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				searchDir = currentDir
			}

			fmt.Printf("Searching for git repositories in: %s\n", searchDir)

			// Find repos
			repos, err := git.FindGitRepos(searchDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding git repositories: %v\n", err)
				os.Exit(1)
			}

			if len(repos) == 0 {
				fmt.Println("No git repositories found.")
				os.Exit(0)
			}

			// Create directory mapping and options
			dirMap := utils.GetDirectoryNames(repos)
			var options []string
			for name := range dirMap {
				options = append(options, name)
			}

			// Sort options
			sort.Slice(options, func(i, j int) bool {
				return strings.ToLower(options[i]) < strings.ToLower(options[j])
			})

			// Show repo selector with configured indicators
			repoModel := ui.InitializeRepoSelectorModel(options, dirMap, true)
			p = tea.NewProgram(&repoModel)
			result, err = p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running repo selection: %v\n", err)
				os.Exit(1)
			}

			m, ok := result.(*ui.BubbleteaModel)
			if !ok || m.Selected == -1 {
				fmt.Println("No repository selected.")
				return
			}

			selected := options[m.Selected]
			selectedPath := dirMap[selected]

			// Load or create repo config
			cfg, err := config.LoadRepoConfig(selectedPath)
			if err != nil {
				// No repo config exists, start with global config
				cfg, _ = config.LoadConfig()
			}

			// Launch config UI with repo context
			model := ui.InitializeRepoConfigModel(cfg, selectedPath)
			p = tea.NewProgram(&model)
			_, err = p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running config UI: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// If --current flag is set, use current directory directly
	if useCurrent {
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Get the base name of the current directory for the session name
		sessionName := filepath.Base(currentDir)

		// Load configuration with repo-level fallback
		cfg, _ := config.LoadConfigWithFallback(currentDir)

		// Create tmux session directly
		err = tmux.CreateTmuxSession(sessionName, currentDir, forceAttach, forceRecreate, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
			os.Exit(1)
		}
		return
	}

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

	// Get directory names and create mapping to full paths
	dirMap := utils.GetDirectoryNames(repos)

	// Create a slice of directory names for the selection
	var options []string
	for name := range dirMap {
		options = append(options, name)
	}

	// Sort the options alphabetically (case-insensitive)
	sort.Slice(options, func(i, j int) bool {
		return strings.ToLower(options[i]) < strings.ToLower(options[j])
	})

	// Create bubbletea model for repository selection
	model := ui.InitializeModel(options, dirMap)
	p := tea.NewProgram(&model)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running bubbletea program: %v\n", err)
		os.Exit(1)
	}

	// Get the selected repository from the model
	m, ok := result.(*ui.BubbleteaModel)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: could not get model from program\n")
		os.Exit(1)
	}

	if m.Selected == -1 {
		fmt.Println("No repository selected.")
		os.Exit(0)
	}

	selected := options[m.Selected]
	selectedPath := dirMap[selected]

	// Load configuration with repo-level fallback
	cfg, _ := config.LoadConfigWithFallback(selectedPath)

	// Create tmux session
	err = tmux.CreateTmuxSession(selected, selectedPath, forceAttach, forceRecreate, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
		os.Exit(1)
	}
}
