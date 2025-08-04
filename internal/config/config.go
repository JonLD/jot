package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Editor           string `json:"editor,omitempty"`
	EditorBackground bool   `json:"editor_background,omitempty"`
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Config{}, nil // Return default config if can't get home dir
	}

	configPath := filepath.Join(homeDir, ".jot", "config.json")
	
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{}, nil // Return default config on error
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return &Config{}, nil // Return default config on parse error
	}

	return &config, nil
}

func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".jot")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}