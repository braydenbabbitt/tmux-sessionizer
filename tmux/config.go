package tmux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WindowConfig represents configuration for a tmux window
type WindowConfig struct {
	Name    string `json:"name"`
	Command string `json:"command,omitempty"`
}

// SessionConfig represents the configuration for tmux sessions
type SessionConfig struct {
	Windows             []WindowConfig `json:"windows"`
	InitialActiveWindow int            `json:"initialActiveWindow"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() SessionConfig {
	return SessionConfig{
		Windows: []WindowConfig{
			{Name: "nvim", Command: "nvim"},
			{Name: "server", Command: ""},
			{Name: "term", Command: ""},
		},
		InitialActiveWindow: 0,
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "tmux-sessionizer")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (SessionConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			return config, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to read config file: %w", err)
	}

	var config SessionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config SessionConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
