package tunnel

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/AYM1607/godig/types"
)

const configFileName = "godig-tunnel.yaml"

// loadTunnelConfig loads the tunnel configuration from the YAML file.
// Returns nil if the file doesn't exist.
func loadTunnelConfig() (*types.TunnelConfig, error) {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config types.TunnelConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveTunnelConfig saves the tunnel configuration to a YAML file.
func saveTunnelConfig(config *types.TunnelConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFileName, data, 0600)
}

// configExists checks if the configuration file exists.
func configExists() bool {
	_, err := os.Stat(configFileName)
	return !os.IsNotExist(err)
}
