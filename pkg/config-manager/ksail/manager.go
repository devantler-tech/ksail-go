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
// Configuration priority: defaults < config files < environment variables < flags.
func (m *ConfigManager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.Config, nil
	}

	notify.Activityln(os.Stdout, "Loading KSail config")

	// Step 1: Apply defaults (lowest priority)
	m.applyDefaults()

	// Step 2: Setup configuration paths including directory traversal
	m.setupConfigurationPaths()

	// Step 3: Load configuration files (higher priority than defaults)
	err := m.loadConfigurationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration files: %w", err)
	}

	// Step 4: Apply final configuration with proper precedence
	err = m.finalizeConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to finalize configuration: %w", err)
	}

	notify.Successln(os.Stdout, "config loaded")

	m.configLoaded = true

	return m.Config, nil
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *ConfigManager) GetViper() *viper.Viper {
	return m.viper
}

// applyDefaults applies default values from field selectors to the configuration.
// This establishes the baseline configuration values (lowest priority).
func (m *ConfigManager) applyDefaults() {
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}
}

// setupConfigurationPaths configures Viper to search parent directories for config files.
// This enables directory traversal functionality for config discovery.
func (m *ConfigManager) setupConfigurationPaths() {
	addParentDirectoriesToViperPaths(m.viper)
}

// loadConfigurationFiles attempts to read configuration files and provides user feedback.
// This handles the config file layer in the priority chain.
func (m *ConfigManager) loadConfigurationFiles() error {
	err := m.viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and environment/flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		notify.Activityln(os.Stdout, "using default configuration")
	} else {
		notify.Activityf(os.Stdout, "'%s' found", m.viper.ConfigFileUsed())
	}

	return nil
}

// finalizeConfiguration completes the configuration loading process.
// This unmarshal the final configuration with proper precedence handling.
func (m *ConfigManager) finalizeConfiguration() error {
	// Unmarshal into our cluster config using Viper's precedence system
	// Viper will automatically handle: defaults < config files < environment variables < flags
	err := m.viper.Unmarshal(m.Config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return nil
}
