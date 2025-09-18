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
		Servers: 0,
		Agents:  0,
		ExposeAPI: v1alpha5.SimpleExposureOpts{
			Host:     "",
			HostIP:   "",
			HostPort: "",
		},
		Image:        "",
		Network:      "",
		Subnet:       "",
		ClusterToken: "",
		Volumes:      nil,
		Ports:        nil,
		Options: v1alpha5.SimpleConfigOptions{
			K3dOptions: v1alpha5.SimpleConfigOptionsK3d{
				Wait:                false,
				Timeout:             0,
				DisableLoadbalancer: false,
				DisableImageVolume:  false,
				NoRollback:          false,
				NodeHookActions:     nil,
				Loadbalancer: v1alpha5.SimpleConfigOptionsK3dLoadbalancer{
					ConfigOverrides: nil,
				},
			},
			K3sOptions: v1alpha5.SimpleConfigOptionsK3s{
				ExtraArgs:  nil,
				NodeLabels: nil,
			},
			KubeconfigOptions: v1alpha5.SimpleConfigOptionsKubeconfig{
				UpdateDefaultKubeconfig: false,
				SwitchCurrentContext:    false,
			},
			Runtime: v1alpha5.SimpleConfigOptionsRuntime{
				GPURequest:    "",
				ServersMemory: "",
				AgentsMemory:  "",
				HostPidMode:   false,
				Labels:        nil,
				Ulimits:       nil,
			},
		},
		Env: nil,
		Registries: v1alpha5.SimpleConfigRegistries{
			Use:    nil,
			Create: nil,
			Config: "",
		},
		HostAliases: nil,
		Files:       nil,
	}
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
	configPath, err := io.FindFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if config file exists
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		m.config = NewK3dSimpleConfig("", "k3d.io/v1alpha5", "Simple")
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
	m.config = NewK3dSimpleConfig("", "", "")

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
