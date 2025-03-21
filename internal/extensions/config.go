package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the configuration for the extensions system
type Config struct {
	// Enable or disable the extensions system
	Enabled bool `json:"enabled"`
	
	// ExtensionSettings contains specific settings for each extension
	ExtensionSettings map[string]map[string]interface{} `json:"extension_settings"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Enabled: false, // Disabled by default for security
		ExtensionSettings: make(map[string]map[string]interface{}),
	}
}

// LoadConfig loads the extension configuration
func LoadConfig() (Config, error) {
	config := DefaultConfig()
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".mcgraph")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return config, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configFile := filepath.Join(configDir, "extensions.json")
	
	// Check if the file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config file
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, fmt.Errorf("failed to marshal default config: %w", err)
		}
		
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return config, fmt.Errorf("failed to write default config: %w", err)
		}
		
		return config, nil
	}
	
	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse the config
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return config, nil
}

// SaveConfig saves the extension configuration
func SaveConfig(config Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".mcgraph")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configFile := filepath.Join(configDir, "extensions.json")
	
	// Marshal the config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write the config file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}