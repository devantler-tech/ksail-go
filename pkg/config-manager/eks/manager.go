// Package eks provides configuration management for EKS cluster configurations.
// This file contains the core Manager implementation for loading EKS configs from files.
package eks

import (
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	eksvalidator "github.com/devantler-tech/ksail-go/pkg/validator/eks"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigManager implements configuration management for EKS cluster configurations.
// It provides file-based configuration loading without Viper dependency.
type ConfigManager struct {
	configPath   string
	config       *eksctlapi.ClusterConfig
	configLoaded bool
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[eksctlapi.ClusterConfig].
var _ configmanager.ConfigManager[eksctlapi.ClusterConfig] = (*ConfigManager)(nil)

// NewEKSClusterConfig creates a new eksctlapi.ClusterConfig with the specified name and region.
// This function provides a canonical way to create EKS clusters with proper field initialization.
// Use empty string for name to create a cluster without a specific name.
func NewEKSClusterConfig(name, region, apiVersion, kind string) *eksctlapi.ClusterConfig {
	// Set default name if empty
	if name == "" {
		name = "eks-default"
	}

	// Set default region if empty
	if region == "" {
		region = "eu-north-1"
	}

	if apiVersion == "" {
		apiVersion = "eksctl.io/v1alpha5"
	}

	if kind == "" {
		kind = "ClusterConfig"
	}

	config := &eksctlapi.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		Metadata: &eksctlapi.ClusterMeta{
			Name:   name,
			Region: region,
		},
	}

	return config
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
func (m *ConfigManager) LoadConfig() (*eksctlapi.ClusterConfig, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.config, nil
	}

	config, err := helpers.LoadConfigFromFile(
		m.configPath,
		func() *eksctlapi.ClusterConfig {
			// Create default with proper APIVersion and Kind
			config := NewEKSClusterConfig(
				"eks-default",
				"eu-north-1",
				"eksctl.io/v1alpha5",
				"ClusterConfig",
			)

			return config
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Apply EKS defaults to the loaded config
	eksctlapi.SetClusterConfigDefaults(config)

	// Validate the loaded configuration
	validator := eksvalidator.NewValidator()

	err = helpers.ValidateConfig(config, validator)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	m.config = config
	m.configLoaded = true

	return m.config, nil
}
