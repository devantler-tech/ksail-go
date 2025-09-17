// Package kind provides configuration management for Kind cluster configurations.
// This file contains the core Manager implementation for loading Kind configs from files.
package kind

import (
	"fmt"
	"os"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ConfigManager implements configuration management for Kind cluster configurations.
// It provides file-based configuration loading without Viper dependency.
type ConfigManager struct {
	marshaller   yamlmarshaller.YAMLMarshaller[*v1alpha4.Cluster]
	configPath   string
	config       *v1alpha4.Cluster
	configLoaded bool
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[v1alpha4.Cluster].
var _ configmanager.ConfigManager[v1alpha4.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager for Kind cluster configurations.
// configPath specifies the path to the Kind configuration file to load.
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		marshaller:   yamlmarshaller.YAMLMarshaller[*v1alpha4.Cluster]{},
		configPath:   configPath,
		config:       nil,
		configLoaded: false,
	}
}

// LoadConfig loads the Kind configuration from the specified file.
// Returns the previously loaded config if already loaded.
// If the file doesn't exist, returns a default Kind cluster configuration.
func (m *ConfigManager) LoadConfig() (*v1alpha4.Cluster, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.config, nil
	}

	// Check if config file exists
	_, err := os.Stat(m.configPath)
	if os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		//nolint:exhaustruct // Kind defaults are applied via SetDefaultsCluster
		m.config = &v1alpha4.Cluster{
			TypeMeta: v1alpha4.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
		}
		// Apply Kind defaults
		v1alpha4.SetDefaultsCluster(m.config)
		m.configLoaded = true

		return m.config, nil
	}

	// Read file contents
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", m.configPath, err)
	}

	// Parse YAML into Kind cluster config
	//nolint:exhaustruct // Kind defaults are applied via SetDefaultsCluster
	m.config = &v1alpha4.Cluster{}

	err = m.marshaller.Unmarshal(data, &m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kind config from %s: %w", m.configPath, err)
	}

	// Ensure APIVersion and Kind are set
	if m.config.APIVersion == "" {
		m.config.APIVersion = "kind.x-k8s.io/v1alpha4"
	}

	if m.config.Kind == "" {
		m.config.Kind = "Cluster"
	}

	// Apply Kind defaults
	v1alpha4.SetDefaultsCluster(m.config)

	m.configLoaded = true

	return m.config, nil
}
