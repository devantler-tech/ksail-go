// Package k3d provides configuration management for K3d v1alpha5.SimpleConfig configurations.
// This file contains the core Manager implementation for loading K3d configurations from files.
package k3d

import (
	"fmt"
	"os"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// ConfigManager implements configuration management for K3d v1alpha5.SimpleConfig configurations.
// It provides simple file-based configuration loading without Viper dependency.
type ConfigManager struct {
	filePath     string
	config       *v1alpha5.SimpleConfig
	configLoaded bool
	marshaller   yamlmarshaller.YAMLMarshaller[v1alpha5.SimpleConfig]
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[v1alpha5.SimpleConfig].
var _ configmanager.ConfigManager[v1alpha5.SimpleConfig] = (*ConfigManager)(nil)

// NewConfigManager creates a new K3d configuration manager for loading from the specified file path.
func NewConfigManager(filePath string) *ConfigManager {
	return &ConfigManager{
		filePath:     filePath,
		config:       nil,
		configLoaded: false,
		marshaller:   yamlmarshaller.YAMLMarshaller[v1alpha5.SimpleConfig]{},
	}
}

// LoadConfig loads the K3d configuration from the specified file.
// Returns the previously loaded config if already loaded.
func (m *ConfigManager) LoadConfig() (*v1alpha5.SimpleConfig, error) {
	// If config is already loaded, return it
	if m.configLoaded && m.config != nil {
		return m.config, nil
	}

	// Read the configuration file
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", m.filePath, err)
	}

	// Initialize a new config
	config := v1alpha5.SimpleConfig{}

	// Unmarshal the YAML data
	err = m.marshaller.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal K3d config from '%s': %w", m.filePath, err)
	}

	// Store the loaded config
	m.config = &config
	m.configLoaded = true

	return m.config, nil
}
