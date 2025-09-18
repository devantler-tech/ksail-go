// Package k3d provides configuration management for K3d v1alpha5.SimpleConfig configurations.
// This file contains the core Manager implementation for loading K3d configurations from files.
package k3d

import (
	"fmt"
	"os"
	"path/filepath"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// ConfigManager implements configuration management for K3d v1alpha5.SimpleConfig configurations.
// It provides file-based configuration loading without Viper dependency.
type ConfigManager struct {
	marshaller   yamlmarshaller.YAMLMarshaller[*v1alpha5.SimpleConfig]
	configPath   string
	config       *v1alpha5.SimpleConfig
	configLoaded bool
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[v1alpha5.SimpleConfig].
var _ configmanager.ConfigManager[v1alpha5.SimpleConfig] = (*ConfigManager)(nil)

// NewK3dSimpleConfig creates a new v1alpha5.SimpleConfig with the specified name and TypeMeta.
// This function provides a canonical way to create K3d clusters with proper field initialization.
// Use empty string for name to create a cluster without a specific name.
func NewK3dSimpleConfig(name, apiVersion, kind string) *v1alpha5.SimpleConfig {
	return &v1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		ObjectMeta: types.ObjectMeta{
			Name: name,
		},
		Servers:      0,
		Agents:       0,
		ExposeAPI:    v1alpha5.SimpleExposureOpts{},
		Image:        "",
		Network:      "",
		Subnet:       "",
		ClusterToken: "",
		Volumes:      nil,
		Ports:        nil,
		Options:      v1alpha5.SimpleConfigOptions{},
		Env:          nil,
		Registries:   v1alpha5.SimpleConfigRegistries{},
		HostAliases:  nil,
		Files:        nil,
	}
}

// newK3dSimpleConfig creates a new v1alpha5.SimpleConfig with all required fields properly initialized.
// This satisfies exhaustruct requirements by providing explicit values for all struct fields.
func newK3dSimpleConfig() *v1alpha5.SimpleConfig {
	return NewK3dSimpleConfig("", "k3d.io/v1alpha5", "Simple")
}

// newEmptyK3dSimpleConfig creates a new empty v1alpha5.SimpleConfig for unmarshaling.
// This satisfies exhaustruct requirements by providing explicit values for all struct fields.
func newEmptyK3dSimpleConfig() *v1alpha5.SimpleConfig {
	return NewK3dSimpleConfig("", "", "")
}

// NewConfigManager creates a new configuration manager for K3d cluster configurations.
// configPath specifies the path to the K3d configuration file to load.
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		marshaller:   yamlmarshaller.YAMLMarshaller[*v1alpha5.SimpleConfig]{},
		configPath:   configPath,
		config:       nil,
		configLoaded: false,
	}
}

// LoadConfig loads the K3d configuration from the specified file.
// Returns the previously loaded config if already loaded.
// If the file doesn't exist, returns a default K3d cluster configuration.
func (m *ConfigManager) LoadConfig() (*v1alpha5.SimpleConfig, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.config, nil
	}

	// Resolve the config path (traverse up from current dir if relative)
	configPath, err := m.resolveConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if config file exists
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		m.config = newK3dSimpleConfig()
		m.configLoaded = true

		return m.config, nil
	}

	// Read file contents safely
	// Since we've resolved the path through traversal, we use the directory containing the file as the base
	baseDir := filepath.Dir(configPath)

	data, err := io.ReadFileSafe(baseDir, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML into K3d cluster config
	m.config = newEmptyK3dSimpleConfig()

	err = m.marshaller.Unmarshal(data, &m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal K3d config from %s: %w", configPath, err)
	}

	// Ensure APIVersion and Kind are set
	if m.config.APIVersion == "" {
		m.config.APIVersion = "k3d.io/v1alpha5"
	}

	if m.config.Kind == "" {
		m.config.Kind = "Simple"
	}

	m.configLoaded = true

	return m.config, nil
}

// resolveConfigPath resolves the configuration file path.
// For absolute paths, returns the path as-is.
// For relative paths or filenames, traverses up from current directory to find the file.
func (m *ConfigManager) resolveConfigPath() (string, error) {
	// If absolute path, return as-is
	if filepath.IsAbs(m.configPath) {
		return m.configPath, nil
	}

	// For relative paths, start from current directory and traverse up
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Traverse up the directory tree looking for the config file
	for {
		candidatePath := filepath.Join(currentDir, m.configPath)

		_, err := os.Stat(candidatePath)
		if err == nil {
			return candidatePath, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		// Stop if we've reached the root directory
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// If not found during traversal, return the original relative path
	// This allows the caller to handle the file-not-found case appropriately
	return m.configPath, nil
}
