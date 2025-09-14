// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains the core Manager implementation.
package ksail

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/viper"
)

// ConfigManager implements the ConfigManager interface for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
}

// Verify that Manager implements the ConfigManager interface.
var _ configmanager.ConfigManager[v1alpha1.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager with the specified field selectors.
func NewConfigManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *ConfigManager {
	return &ConfigManager{
		viper:          InitializeViper(),
		fieldSelectors: fieldSelectors,
		Config:         v1alpha1.NewCluster(),
	}
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
func (m *ConfigManager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded and populated, return it
	if m.Config != nil && !isEmptyCluster(m.Config) {
		return m.Config, nil
	}

	// Initialize with defaults from field selectors
	m.applyDefaults()

	// Try to read from configuration files
	m.viper.SetConfigName(DefaultConfigFileName)
	m.viper.SetConfigType("yaml")
	m.viper.AddConfigPath(".")
	m.viper.AddConfigPath("$HOME/.config/ksail")
	m.viper.AddConfigPath("/etc/ksail")

	// Read configuration file if it exists
	err := m.viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Set environment variable prefix and bind environment variables
	m.viper.SetEnvPrefix(EnvPrefix)
	m.viper.AutomaticEnv()
	bindEnvironmentVariables(m.viper)

	// Unmarshal into our cluster config
	err = m.viper.Unmarshal(m.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return m.Config, nil
}

// isEmptyCluster checks if the cluster configuration is empty/default.
func isEmptyCluster(config *v1alpha1.Cluster) bool {
	emptyCluster := v1alpha1.NewCluster()

	return reflect.DeepEqual(config, emptyCluster)
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *ConfigManager) GetViper() *viper.Viper {
	return m.viper
}

// applyDefaults applies default values from field selectors to the config.
func (m *ConfigManager) applyDefaults() {
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}
}
