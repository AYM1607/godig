package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ConfigKey represents a configuration key.
type ConfigKey string

const (
	// KeyAPIKey is the configuration key for the API key.
	KeyAPIKey ConfigKey = "api-key"
	// KeyServer is the configuration key for the server address.
	KeyServer ConfigKey = "server"
)

// GlobalConfig represents the user's global configuration.
type GlobalConfig struct {
	APIKey string `yaml:"api_key,omitempty"`
	Server string `yaml:"server,omitempty"`
}

// getConfigDir returns the path to the config directory.
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "godig"), nil
}

// getConfigPath returns the path to the global config file.
func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// ensureConfigDir creates the config directory if it doesn't exist.
func ensureConfigDir() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// LoadGlobalConfig loads the global configuration from ~/.config/godig/config.yaml.
// Returns an empty config if the file doesn't exist.
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration to ~/.config/godig/config.yaml.
func SaveGlobalConfig(config *GlobalConfig) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetConfigValue sets a specific configuration value and saves it.
func SetConfigValue(key ConfigKey, value string) error {
	config, err := LoadGlobalConfig()
	if err != nil {
		return err
	}

	switch key {
	case KeyAPIKey:
		config.APIKey = value
	case KeyServer:
		config.Server = value
	default:
		return fmt.Errorf("unknown config key: %s (valid keys: %s, %s)", key, KeyAPIKey, KeyServer)
	}

	return SaveGlobalConfig(config)
}

// GetConfigValue retrieves a specific configuration value.
func GetConfigValue(key ConfigKey) (string, error) {
	config, err := LoadGlobalConfig()
	if err != nil {
		return "", err
	}

	switch key {
	case KeyAPIKey:
		return config.APIKey, nil
	case KeyServer:
		return config.Server, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
