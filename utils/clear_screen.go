package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ClearScreen clears the entire terminal screen, including history
func ClearScreen() {
	// Different clear commands based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		// For Unix-like systems (MacOS, Linux)
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	_ = cmd.Run() // Ignoring error as this is not critical
}

// ClearCurrentView clears the current view without clearing terminal history
// This is useful for showing multiple interactive UIs sequentially
func ClearCurrentView() {
	// Use ANSI escape sequences to:
	// 1. Save cursor position
	// 2. Move to top-left corner
	// 3. Clear from cursor to end of screen
	// 4. Restore cursor position
	
	if runtime.GOOS == "windows" {
		// Windows might need cmd /c cls, but this will clear history too
		// Using ANSI escape sequences might work on newer Windows terminals
		fmt.Print("\033[2J\033[0;0H") // Clear screen and move cursor to top-left
	} else {
		// For Unix-like systems (MacOS, Linux) using ANSI escape codes
		// \033[H moves cursor to home position (top-left)
		// \033[J clears from cursor to end of screen
		fmt.Print("\033[H\033[J")
	}
}
