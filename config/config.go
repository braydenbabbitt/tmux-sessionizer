package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WindowConfig represents a single window configuration
type WindowConfig struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Config represents the complete configuration
type Config struct {
	Version string         `json:"version"`
	Windows []WindowConfig `json:"windows"`
}

// GetDefaultConfig returns the default configuration matching current hardcoded behavior
func GetDefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Windows: []WindowConfig{
			{Name: "nvim", Command: "nvim"},
			{Name: "server", Command: ""},
			{Name: "term", Command: ""},
		},
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "tmux-sessionizer")
	configPath := filepath.Join(configDir, "config.json")

	return configPath, nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// LoadConfig loads configuration from file or returns defaults on any error
// This implements graceful degradation for backward compatibility
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		// Can't determine config path, use defaults
		return GetDefaultConfig(), nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return GetDefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Can't read config file, use defaults
		return GetDefaultConfig(), nil
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		// Invalid JSON, use defaults
		return GetDefaultConfig(), nil
	}

	// Validate config
	if len(config.Windows) == 0 {
		// No windows configured, use defaults
		return GetDefaultConfig(), nil
	}

	// Validate window names
	for i, window := range config.Windows {
		if window.Name == "" {
			// Invalid window name, use defaults
			return GetDefaultConfig(), nil
		}
		// Sanitize window name to avoid tmux issues
		config.Windows[i].Name = window.Name
	}

	return &config, nil
}

// SaveConfig saves configuration to file with atomic write
func SaveConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate config
	if len(config.Windows) == 0 {
		return fmt.Errorf("config must have at least one window")
	}

	for _, window := range config.Windows {
		if window.Name == "" {
			return fmt.Errorf("window name cannot be empty")
		}
	}

	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first (atomic write)
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Rename temp file to actual config file (atomic operation)
	if err := os.Rename(tempPath, configPath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetRepoConfigPath returns the path to a repository's local config file
// repoDir must be a git repository root directory
func GetRepoConfigPath(repoDir string) (string, error) {
	if repoDir == "" {
		return "", fmt.Errorf("repoDir cannot be empty")
	}

	configPath := filepath.Join(repoDir, ".git", "x-tmux-sessionizer", "config.json")
	return configPath, nil
}

// HasRepoConfig checks if a repository has a local configuration
func HasRepoConfig(repoDir string) bool {
	configPath, err := GetRepoConfigPath(repoDir)
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// LoadRepoConfig loads configuration from a repository's local config file
// Returns error if repo config doesn't exist or can't be read
func LoadRepoConfig(repoDir string) (*Config, error) {
	configPath, err := GetRepoConfigPath(repoDir)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no repo config found at %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	// Validate config
	if len(config.Windows) == 0 {
		return nil, fmt.Errorf("repo config must have at least one window")
	}

	// Validate window names
	for i, window := range config.Windows {
		if window.Name == "" {
			return nil, fmt.Errorf("window name cannot be empty")
		}
		config.Windows[i].Name = window.Name
	}

	return &config, nil
}

// LoadConfigWithFallback loads config with priority: repo-level -> global -> defaults
// repoDir can be empty string to skip repo config check
func LoadConfigWithFallback(repoDir string) (*Config, error) {
	// Try repo-level config first if repoDir is provided
	if repoDir != "" {
		cfg, err := LoadRepoConfig(repoDir)
		if err == nil {
			return cfg, nil
		}
		// If repo config fails, fall through to global config
	}

	// Fall back to global config (which has its own fallback to defaults)
	return LoadConfig()
}

// SaveRepoConfig saves configuration to a repository's local config file
func SaveRepoConfig(repoDir string, config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate config
	if len(config.Windows) == 0 {
		return fmt.Errorf("config must have at least one window")
	}

	for _, window := range config.Windows {
		if window.Name == "" {
			return fmt.Errorf("window name cannot be empty")
		}
	}

	configPath, err := GetRepoConfigPath(repoDir)
	if err != nil {
		return err
	}

	// Create directory structure
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory (check .git permissions): %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first (atomic write)
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Rename temp file to actual config file (atomic operation)
	if err := os.Rename(tempPath, configPath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
