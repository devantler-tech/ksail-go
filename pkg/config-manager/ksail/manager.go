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
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/spf13/viper"
)

// ConfigManager implements the ConfigManager interface for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
	marshaller     marshaller.Marshaller[v1alpha1.Cluster]
	configLoaded   bool // Track if config has been actually loaded from file/env
}

// Verify that Manager implements the ConfigManager interface.
var _ configmanager.ConfigManager[v1alpha1.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager with the specified field selectors.
func NewConfigManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *ConfigManager {
	return &ConfigManager{
		viper:          InitializeViper(),
		fieldSelectors: fieldSelectors,
		Config:         v1alpha1.NewCluster(),
		marshaller:     yamlmarshaller.NewMarshaller[v1alpha1.Cluster](),
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

	// First try to find config file with directory traversal (like the custom loader)
	configPath, found := m.findConfigFile()
	if found {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("read ksail config: %w", err)
		}
		cfg := v1alpha1.Cluster{}
		if err := m.marshaller.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal ksail config: %w", err)
		}
		notify.Activityf(os.Stdout, "'%s' found", configPath)
		notify.Successln(os.Stdout, "config loaded")
		fmt.Println()
		m.Config = &cfg
		m.configLoaded = true
		return m.Config, nil
	}

	// Fallback to Viper-based loading for compatibility
	notify.Activityln(os.Stdout, "'./ksail.yaml' not found, trying additional paths")

	// Try to read from configuration files using Viper
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
		notify.Activityln(os.Stdout, "using default configuration")
	} else {
		notify.Activityf(os.Stdout, "'%s' found", m.viper.ConfigFileUsed())
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

// findConfigFile searches for ksail.yaml file starting from the current directory
// and walking up the directory tree until found or reaching the root.
// Returns the path to the config file and whether it was found.
func (m *ConfigManager) findConfigFile() (string, bool) {
	// Get absolute path of current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", false
	}

	for dir := currentDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, "ksail.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", false
}
