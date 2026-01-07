package tmux

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	config "github.com/Haptic-Labs/tmux-sessionizer/config"
)

// CreateTmuxSession creates a new tmux session with the specified name and directory
// If forceAttach is true and a session exists, it will attach to the existing session without prompting
// If forceRecreate is true and a session exists, it will kill and recreate the session without prompting
// cfg specifies the window configuration; if nil, defaults will be used
func CreateTmuxSession(name string, directory string, forceAttach bool, forceRecreate bool, cfg *config.Config) error {
	// Check if session already exists
	checkCmd := exec.Command("tmux", "has-session", "-t", "="+name)
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

	// Use defaults if config is nil
	if cfg == nil {
		cfg = config.GetDefaultConfig()
	}

	if len(cfg.Windows) == 0 {
		return fmt.Errorf("no windows configured")
	}

	// Create first window (session creation)
	firstWindow := cfg.Windows[0]
	createCmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", directory, "-n", firstWindow.Name)
	if err := createCmd.Run(); err != nil {
		return err
	}

	// Run command in first window if specified
	if firstWindow.Command != "" {
		cmdParts := strings.Fields(firstWindow.Command)
		args := []string{"send-keys", "-t", name + ":0"}
		args = append(args, cmdParts...)
		args = append(args, "Enter")
		sendCmd := exec.Command("tmux", args...)
		if err := sendCmd.Run(); err != nil {
			return err
		}
	}

	// Create additional windows
	for i := 1; i < len(cfg.Windows); i++ {
		window := cfg.Windows[i]
		windowIndex := fmt.Sprintf("%d", i)

		newWindowCmd := exec.Command("tmux", "new-window", "-t",
			name+":"+windowIndex, "-n", window.Name, "-c", directory)
		if err := newWindowCmd.Run(); err != nil {
			return err
		}

		// Run command if specified
		if window.Command != "" {
			cmdParts := strings.Fields(window.Command)
			args := []string{"send-keys", "-t", name + ":" + windowIndex}
			args = append(args, cmdParts...)
			args = append(args, "Enter")
			sendCmd := exec.Command("tmux", args...)
			if err := sendCmd.Run(); err != nil {
				return err
			}
		}
	}

	// Select the first window
	selectCmd := exec.Command("tmux", "select-window", "-t", name+":0")
	if err := selectCmd.Run(); err != nil {
		return err
	}

	// Attach to the session
	attachCmd := exec.Command("tmux", "attach", "-t", name)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
