package configmanager

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanagerinterface "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/viper"
)

// ConfigManager implements configuration management for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	Viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
	configLoaded   bool              // Track if config has been actually loaded
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanagerinterface.ConfigManager[v1alpha1.Cluster].
var _ configmanagerinterface.ConfigManager[v1alpha1.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager with the specified field selectors.
// Initializes Viper with all configuration including paths and environment handling.
func NewConfigManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *ConfigManager {
	viperInstance := InitializeViper()
	config := v1alpha1.NewCluster()

	manager := &ConfigManager{
		Viper:          viperInstance,
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

	notify.Activityln(os.Stdout, "loading ksail config")

	// Use native Viper API to read configuration
	// All paths and environment handling are already configured in constructor
	err := m.Viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and environment/flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		notify.Activityln(os.Stdout, "using default config")
	} else {
		notify.Activityf(os.Stdout, "'%s' found", m.Viper.ConfigFileUsed())
	}

	// Unmarshal configuration using Viper's native precedence handling
	// Viper will handle: config files < environment variables < flags
	err = m.Viper.Unmarshal(m.Config)
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

// IsFieldEmptyForTesting exposes isFieldEmpty for testing purposes.
func IsFieldEmptyForTesting(fieldPtr any) bool {
	return isFieldEmpty(fieldPtr)
}
