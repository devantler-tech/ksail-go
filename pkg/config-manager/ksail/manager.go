// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains the core Manager implementation.
package ksail

import (
	"errors"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/viper"
)

// ConfigManager implements the ConfigManager interface for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
	configLoaded   bool              // Track if config has been actually loaded
}

// Verify that Manager implements the ConfigManager interface.
var _ configmanager.ConfigManager[v1alpha1.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager with the specified field selectors.
func NewConfigManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *ConfigManager {
	return &ConfigManager{
		viper:          InitializeViper(),
		fieldSelectors: fieldSelectors,
		Config:         v1alpha1.NewCluster(),
		configLoaded:   false,
	}
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
func (m *ConfigManager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.Config, nil
	}

	notify.Activityln(os.Stdout, "Loading KSail config")

	// Delegate initialization steps to specialized methods
	m.initializeDefaults()
	m.setupDirectoryTraversal()

	// Try to read configuration file using Viper
	config, err := m.readConfigurationFile()
	if err != nil {
		return nil, err
	}

	// Complete the environment setup and unmarshal configuration
	err = m.finalizeConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to finalize configuration: %w", err)
	}

	notify.Successln(os.Stdout, "config loaded")

	m.configLoaded = true

	return config, nil
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *ConfigManager) GetViper() *viper.Viper {
	return m.viper
}

// initializeDefaults applies default values from field selectors to the config.
func (m *ConfigManager) initializeDefaults() {
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}
}

// setupDirectoryTraversal configures Viper to search parent directories for config files.
func (m *ConfigManager) setupDirectoryTraversal() {
	addParentDirectoriesToViperPaths(m.viper)
}

// readConfigurationFile attempts to read the configuration file and provides user feedback.
func (m *ConfigManager) readConfigurationFile() (*v1alpha1.Cluster, error) {
	err := m.viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		notify.Activityln(os.Stdout, "using default configuration")
	} else {
		notify.Activityf(os.Stdout, "'%s' found", m.viper.ConfigFileUsed())
	}

	return m.Config, nil
}

// finalizeConfiguration completes the environment setup and unmarshals the configuration.
func (m *ConfigManager) finalizeConfiguration() error {
	// Bind environment variables to complete the configuration
	bindEnvironmentVariables(m.viper)

	// Unmarshal into our cluster config using Viper
	err := m.viper.Unmarshal(m.Config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return nil
}
