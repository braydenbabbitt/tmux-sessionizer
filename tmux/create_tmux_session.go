package tmux

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CreateTmuxSession creates a new tmux session with the specified name and directory
// If forceAttach is true and a session exists, it will attach to the existing session without prompting
// If forceRecreate is true and a session exists, it will kill and recreate the session without prompting
func CreateTmuxSession(name string, directory string, forceAttach bool, forceRecreate bool) error {
	// Check if session already exists
	checkCmd := exec.Command("tmux", "has-session", "-t", name)
	err := checkCmd.Run()

	if err == nil {
		// Session exists, determine what to do
		if forceAttach {
			// Automatically attach to existing session
			attachCmd := exec.Command("tmux", "attach", "-t", name)
			attachCmd.Stdin = os.Stdin
			attachCmd.Stdout = os.Stdout
			attachCmd.Stderr = os.Stderr
			return attachCmd.Run()
		} else if forceRecreate {
			// Automatically kill existing session
			killCmd := exec.Command("tmux", "kill-session", "-t", name)
			if err := killCmd.Run(); err != nil {
				return fmt.Errorf("failed to kill existing session: %w", err)
			}
			// Continue to create new session below
		} else {
			// Prompt user for action
			fmt.Printf("Session '%s' already exists. Choose an option:\n", name)
			fmt.Println("[a/y] Attach")
			fmt.Println("[k/n] Kill and recreate")
			fmt.Println("[q/c] Cancel")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			switch input {
			case "a", "y":
				// Attach to existing session
				attachCmd := exec.Command("tmux", "attach", "-t", name)
				attachCmd.Stdin = os.Stdin
				attachCmd.Stdout = os.Stdout
				attachCmd.Stderr = os.Stderr
				return attachCmd.Run()
			case "k", "n":
				// Kill existing session
				killCmd := exec.Command("tmux", "kill-session", "-t", name)
				if err := killCmd.Run(); err != nil {
					return fmt.Errorf("failed to kill existing session: %w", err)
				}
				// Continue to create new session below
			case "q", "c", "":
				// Cancel operation
				return nil
			default:
				fmt.Println("Invalid option, canceling operation")
				return nil
			}
		}
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config, using defaults: %v\n", err)
		config = DefaultConfig()
	}

	// Create a new detached session with the first window
	firstWindowName := "tmux" // Default name if no windows are configured
	if len(config.Windows) > 0 {
		firstWindowName = config.Windows[0].Name
	}

	// Create new session with the first window
	createCmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", directory, "-n", firstWindowName)
	if err := createCmd.Run(); err != nil {
		return err
	}

	// Configure all windows based on the configuration
	for i, window := range config.Windows {
		if i == 0 {
			// The first window is already created, just send the command if specified
			if window.Command != "" {
				sendKeysCmd := exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:0", name), window.Command, "Enter")
				if err := sendKeysCmd.Run(); err != nil {
					return fmt.Errorf("failed to send keys to first window: %w", err)
				}
			}
		} else {
			// Create additional windows
			newWindowCmd := exec.Command("tmux", "new-window", "-t", fmt.Sprintf("%s:%d", name, i), "-n", window.Name, "-c", directory)
			if err := newWindowCmd.Run(); err != nil {
				return fmt.Errorf("failed to create window %s: %w", window.Name, err)
			}

			// Run the command in the window if specified
			if window.Command != "" {
				cmdStr := window.Command
				sendKeysCmd := exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:%d", name, i), cmdStr, "Enter")
				if err := sendKeysCmd.Run(); err != nil {
					return fmt.Errorf("failed to send keys to window %s: %w", window.Name, err)
				}
			}
		}
	}

	// Select the initial active window
	initialWindow := 0
	if config.InitialActiveWindow >= 0 && config.InitialActiveWindow < len(config.Windows) {
		initialWindow = config.InitialActiveWindow
	}
	selectCmd := exec.Command("tmux", "select-window", "-t", fmt.Sprintf("%s:%d", name, initialWindow))
	if err := selectCmd.Run(); err != nil {
		return fmt.Errorf("failed to select initial window: %w", err)
	}

	// Attach to the session
	attachCmd := exec.Command("tmux", "attach", "-t", name)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
