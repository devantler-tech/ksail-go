// Package eks provides configuration management for EKS cluster configurations.
// This file contains the core Manager implementation for loading EKS configs from files.
package eks

import (
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/helpers"
	eksvalidator "github.com/devantler-tech/ksail-go/pkg/io/validator/eks"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	ekstypes "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigManager implements configuration management for EKS cluster configurations.
// It provides file-based configuration loading without Viper dependency.
type ConfigManager struct {
	configPath   string
	config       *ekstypes.ClusterConfig
	configLoaded bool
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[ekstypes.ClusterConfig].
var _ configmanager.ConfigManager[ekstypes.ClusterConfig] = (*ConfigManager)(nil)

// NewEKSCluster creates a new ClusterConfig with the specified name and TypeMeta.
// This function provides a canonical way to create EKS clusters with proper field initialization.
// Use empty string for name to create a cluster without a specific name.
func NewEKSCluster(name, apiVersion, kind string) *ekstypes.ClusterConfig {
	// Set default name if empty
	if name == "" {
		name = "ksail-eks"
	}

	if apiVersion == "" {
		apiVersion = "eksctl.io/v1alpha5"
	}

	if kind == "" {
		kind = "ClusterConfig"
	}

	return &ekstypes.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		Metadata: &ekstypes.ClusterMeta{
			Name: name,
		},
	}
}

// NewConfigManager creates a new configuration manager for EKS cluster configurations.
// configPath specifies the path to the EKS configuration file to load.
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath:   configPath,
		config:       nil,
		configLoaded: false,
	}
}

// LoadConfig loads the EKS configuration from the specified file.
// Returns the previously loaded config if already loaded.
// If the file doesn't exist, returns a default EKS cluster configuration.
// Validates the configuration after loading and returns an error if validation fails.
// The timer parameter is accepted for interface compliance but not currently used.
func (m *ConfigManager) LoadConfig(_ timer.Timer) error {
	// If config is already loaded, return it
	if m.configLoaded {
		return nil
	}

	config, err := helpers.LoadAndValidateConfig(
		m.configPath,
		func() *ekstypes.ClusterConfig {
			// Create default with proper APIVersion and Kind
			config := NewEKSCluster("", "eksctl.io/v1alpha5", "ClusterConfig")
			// Apply EKS defaults
			ekstypes.SetClusterConfigDefaults(config)

			return config
		},
		eksvalidator.NewValidator(),
	)
	if err != nil {
		return fmt.Errorf("failed to load EKS config: %w", err)
	}

	m.config = config
	m.configLoaded = true

	return nil
}

// GetConfig implements configmanager.ConfigManager.
func (m *ConfigManager) GetConfig() *ekstypes.ClusterConfig {
	return m.config
}
