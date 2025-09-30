package configmanager

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanagerinterface "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/validator/ksail"
	"github.com/spf13/viper"
)

// ConfigManager implements configuration management for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	Viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
	configLoaded   bool              // Track if config has been actually loaded
	Writer         io.Writer         // Writer for output notifications
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanagerinterface.ConfigManager[v1alpha1.Cluster].
var _ configmanagerinterface.ConfigManager[v1alpha1.Cluster] = (*ConfigManager)(nil)

// NewConfigManager creates a new configuration manager with the specified field selectors.
// Initializes Viper with all configuration including paths and environment handling.
func NewConfigManager(
	writer io.Writer,
	fieldSelectors ...FieldSelector[v1alpha1.Cluster],
) *ConfigManager {
	viperInstance := InitializeViper()
	config := v1alpha1.NewCluster()

	manager := &ConfigManager{
		Viper:          viperInstance,
		fieldSelectors: fieldSelectors,
		Config:         config,
		configLoaded:   false,
		Writer:         writer,
	}

	return manager
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
// Configuration priority: defaults < config files < environment variables < flags.
// Validates the configuration after loading and returns detailed error messages for validation failures.
func (m *ConfigManager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded, return it
	notify.TitleMessage(m.Writer, "⏳", notify.NewMessage("Loading configuration..."))

	if m.configLoaded {
		notify.SuccessMessage(
			m.Writer,
			notify.NewMessage("config already loaded, reusing existing config"),
		)

		return m.Config, nil
	}

	// Use native Viper API to read configuration
	// All paths and environment handling are already configured in constructor
	err := m.Viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and environment/flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		notify.ActivityMessage(m.Writer, notify.NewMessage("using default config"))
	} else {
		notify.ActivityMessage(m.Writer, notify.NewMessage(fmt.Sprintf("'%s' found", m.Viper.ConfigFileUsed())))
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

	// Validate the loaded configuration
	validator := ksailvalidator.NewValidator()
	validationResult := validator.Validate(m.Config)
	if !validationResult.Valid {
		formattedWarnings := helpers.FormatValidationWarnings(validationResult)
		for _, warning := range formattedWarnings {
			notify.WarnMessage(m.Writer, notify.NewMessage(warning))
		}
		formattedErrors := helpers.FormatValidationErrors(validationResult)
		for _, errMsg := range formattedErrors {
			notify.ErrorMessage(m.Writer, notify.NewMessage(errMsg))
		}

		warningLength := len(formattedWarnings)
		errorLength := len(formattedErrors)
		return nil, fmt.Errorf(
			"%w: %s",
			helpers.ErrConfigurationValidationFailed,
			fmt.Sprintf("found %d warning(s) and %d error(s)", warningLength, errorLength),
		)
	}

	notify.SuccessMessage(m.Writer, notify.NewMessage("config loaded"))

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
