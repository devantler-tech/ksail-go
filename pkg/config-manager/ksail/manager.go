// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains the core Manager implementation.
package ksail

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	// Initialize with defaults from field selectors
	m.applyDefaults()

	// Add all parent directories to Viper's search paths for directory traversal
	m.addParentDirectoriesToConfigPaths()

	// Try to read from configuration files using Viper
	m.viper.SetConfigName(DefaultConfigFileName)
	m.viper.SetConfigType("yaml")

	// Read configuration file if it exists
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

	// Set environment variable prefix and bind environment variables
	m.viper.SetEnvPrefix(EnvPrefix)
	m.viper.AutomaticEnv()
	bindEnvironmentVariables(m.viper)

	// Unmarshal into our cluster config using Viper
	err = m.viper.Unmarshal(m.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	notify.Successln(os.Stdout, "config loaded")
	fmt.Println()
	m.configLoaded = true
	return m.Config, nil
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

// addParentDirectoriesToConfigPaths adds all parent directories to Viper's config search paths
// starting from the current directory and walking up the directory tree.
// This enables directory traversal functionality similar to how Git finds .git directories.
func (m *ConfigManager) addParentDirectoriesToConfigPaths() {
	// Get absolute path of current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		// If we can't get current dir, the default paths in InitializeViper should suffice
		return
	}

	// Track which directories we've added to avoid duplicates
	addedPaths := make(map[string]bool)

	// Get existing config paths from Viper
	// (This is not directly exposed, but we can add our paths safely since Viper handles duplicates)

	// Walk up the directory tree and add each directory to Viper's search paths
	// but only if a ksail.yaml file actually exists in that directory
	for dir := currentDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, "ksail.yaml")
		if _, err := os.Stat(configPath); err == nil {
			// Only add the directory to search path if ksail.yaml exists there
			// and we haven't added it already
			if !addedPaths[dir] {
				m.viper.AddConfigPath(dir)
				addedPaths[dir] = true
			}
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	// Note: Standard system paths are already added in InitializeViper()
}
