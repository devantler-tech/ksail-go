// Package kind provides configuration management for Kind cluster configurations.
// This file contains the core Manager implementation for loading Kind configs from files.
package kind

import (
	"fmt"
	"os"
	"path/filepath"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/io"
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

// NewKindCluster creates a new v1alpha4.Cluster with the specified name and TypeMeta.
// This function provides a canonical way to create Kind clusters with proper field initialization.
// Use empty string for name to create a cluster without a specific name.
func NewKindCluster(name, apiVersion, kind string) *v1alpha4.Cluster {
	return &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		Name:  name,
		Nodes: nil,
		Networking: v1alpha4.Networking{
			IPFamily:          "",
			APIServerPort:     0,
			APIServerAddress:  "",
			PodSubnet:         "",
			ServiceSubnet:     "",
			DisableDefaultCNI: false,
			KubeProxyMode:     "",
			DNSSearch:         nil,
		},
		FeatureGates:                    nil,
		RuntimeConfig:                   nil,
		KubeadmConfigPatches:            nil,
		KubeadmConfigPatchesJSON6902:    nil,
		ContainerdConfigPatches:         nil,
		ContainerdConfigPatchesJSON6902: nil,
	}
}

// newKindCluster creates a new v1alpha4.Cluster with all required fields properly initialized.
// This satisfies exhaustruct requirements by providing explicit values for all struct fields.
func newKindCluster() *v1alpha4.Cluster {
	return NewKindCluster("", "kind.x-k8s.io/v1alpha4", "Cluster")
}

// newEmptyKindCluster creates a new empty v1alpha4.Cluster for unmarshaling.
// This satisfies exhaustruct requirements by providing explicit values for all struct fields.
func newEmptyKindCluster() *v1alpha4.Cluster {
	return NewKindCluster("", "", "")
}

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

	// Resolve the config path (traverse up from current dir if relative)
	configPath, err := io.FindFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if config file exists
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		m.config = newKindCluster()
		// Apply Kind defaults
		v1alpha4.SetDefaultsCluster(m.config)
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

	// Parse YAML into Kind cluster config
	m.config = newEmptyKindCluster()

	err = m.marshaller.Unmarshal(data, &m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kind config from %s: %w", configPath, err)
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


