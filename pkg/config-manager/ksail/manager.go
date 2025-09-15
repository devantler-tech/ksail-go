// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains the core Manager implementation.
package ksail

import (
	"errors"
	"fmt"
	"os"
	"reflect"

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
// Initializes Viper with all configuration including paths and environment handling.
func NewConfigManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *ConfigManager {
	viperInstance := InitializeViper()
	config := v1alpha1.NewCluster()

	manager := &ConfigManager{
		viper:          viperInstance,
		fieldSelectors: fieldSelectors,
		Config:         config,
		configLoaded:   false,
	}

	return manager
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

	// Use native Viper API to read configuration
	// All paths and environment handling are already configured in constructor
	err := m.viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and environment/flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		notify.Activityln(os.Stdout, "using default configuration")
	} else {
		notify.Activityf(os.Stdout, "'%s' found", m.viper.ConfigFileUsed())
	}

	// Unmarshal configuration using Viper's native precedence handling
	// Viper will handle: config files < environment variables < flags
	err = m.viper.Unmarshal(m.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Apply field selector defaults only for fields that are still empty
	// This ensures defaults are applied with the lowest precedence
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil && isFieldEmpty(fieldPtr) {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}

	notify.Successln(os.Stdout, "config loaded")

	m.configLoaded = true

	return m.Config, nil
}

// getViperKeyFromFieldPtr converts a field pointer to its corresponding Viper configuration key.
// Returns empty string if the field is not mapped.
func (m *ConfigManager) getViperKeyFromFieldPtr(fieldPtr any) string {
	// Map field pointers to their Viper configuration keys
	fieldToViperKey := map[any]string{
		&m.Config.Metadata.Name:              "metadata.name",
		&m.Config.Spec.Distribution:          "spec.distribution",
		&m.Config.Spec.DistributionConfig:    "spec.distributionconfig",
		&m.Config.Spec.SourceDirectory:       "spec.sourcedirectory",
		&m.Config.Spec.Connection.Context:    "spec.connection.context",
		&m.Config.Spec.Connection.Kubeconfig: "spec.connection.kubeconfig",
		&m.Config.Spec.Connection.Timeout:    "spec.connection.timeout",
		&m.Config.Spec.ReconciliationTool:    "spec.reconciliationtool",
		&m.Config.Spec.CNI:                   "spec.cni",
		&m.Config.Spec.CSI:                   "spec.csi",
		&m.Config.Spec.IngressController:     "spec.ingresscontroller",
		&m.Config.Spec.GatewayController:     "spec.gatewaycontroller",
	}

	if viperKey, exists := fieldToViperKey[fieldPtr]; exists {
		return viperKey
	}

	return ""
}

// isFieldEmpty checks if a field pointer points to an empty/zero value.
func isFieldEmpty(fieldPtr any) bool {
	if fieldPtr == nil {
		return true
	}

	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr || fieldVal.IsNil() {
		return true
	}

	fieldVal = fieldVal.Elem()
	return fieldVal.IsZero()
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *ConfigManager) GetViper() *viper.Viper {
	return m.viper
}
