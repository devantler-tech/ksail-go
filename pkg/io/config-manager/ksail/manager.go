package configmanager

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanagerinterface "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/helpers"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/io/validator/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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

// NewCommandConfigManager constructs a ConfigManager bound to the provided Cobra command.
// It registers the supplied field selectors, binds flags from struct fields, and writes output
// to the command's standard output writer.
func NewCommandConfigManager(
	cmd *cobra.Command,
	selectors []FieldSelector[v1alpha1.Cluster],
) *ConfigManager {
	manager := NewConfigManager(cmd.OutOrStdout(), selectors...)
	manager.AddFlagsFromFields(cmd)

	return manager
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
// Configuration priority: defaults < config files < environment variables < flags.
// If timer is provided, timing information will be included in the success notification.
func (m *ConfigManager) LoadConfig(tmr timer.Timer) error {
	return m.loadConfigWithOptions(tmr, false)
}

// LoadConfigSilent loads the configuration without outputting notifications.
// Returns the previously loaded config if already loaded.
func (m *ConfigManager) LoadConfigSilent() error {
	return m.loadConfigWithOptions(nil, true)
}

// GetConfig implements configmanager.ConfigManager by returning the loaded cluster configuration.
func (m *ConfigManager) GetConfig() *v1alpha1.Cluster {
	return m.Config
}

// loadConfigWithOptions is the internal implementation with silent option.
func (m *ConfigManager) loadConfigWithOptions(
	tmr timer.Timer,
	silent bool,
) error {
	if !silent {
		m.notifyLoadingStart()
	}

	if m.configLoaded {
		if !silent {
			m.notifyConfigReused()
		}

		return nil
	}

	if !silent {
		m.notifyLoadingConfig()
	}

	// Use native Viper API to read configuration
	err := m.readConfig(silent)
	if err != nil {
		return err
	}

	// Unmarshal and apply defaults
	err = m.unmarshalAndApplyDefaults()
	if err != nil {
		return err
	}

	err = m.validateConfig()
	if err != nil {
		return err
	}

	if !silent {
		m.notifyLoadingComplete(tmr)
	}

	m.configLoaded = true

	return nil
}

func (m *ConfigManager) readConfig(silent bool) error {
	err := m.Viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if !silent {
			m.notifyUsingDefaults()
		}
	} else if !silent {
		m.notifyConfigFound()
	}

	return nil
}

func (m *ConfigManager) unmarshalAndApplyDefaults() error {
	err := m.Viper.Unmarshal(m.Config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Apply field selector defaults for empty fields
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil && isFieldEmpty(fieldPtr) {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}

	return nil
}

func (m *ConfigManager) notifyLoadingStart() {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Load config...",
		Emoji:   "⏳",
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) notifyConfigReused() {
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "config already loaded, reusing existing config",
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) notifyLoadingConfig() {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "loading ksail config",
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) notifyUsingDefaults() {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "using default config",
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) notifyConfigFound() {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "'%s' found",
		Args:    []any{m.Viper.ConfigFileUsed()},
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) notifyLoadingComplete(tmr timer.Timer) {
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "config loaded",
		Timer:   tmr,
		Writer:  m.Writer,
	})
}

func (m *ConfigManager) validateConfig() error {
	// Load distribution configs for cross-validation only if distributionConfig is set
	distributionConfigs := m.loadDistributionConfigsForValidation()

	validator := ksailvalidator.NewValidator(distributionConfigs...)
	result := validator.Validate(m.Config)

	if !result.Valid {
		errorMessages := helpers.FormatValidationErrorsMultiline(result)
		notify.WriteMessage(notify.Message{
			Type:    notify.ErrorType,
			Content: "Configuration validation failed:\n%s",
			Args:    []any{errorMessages},
			Writer:  m.Writer,
		})

		fixSuggestions := helpers.FormatValidationFixSuggestions(result)
		for _, suggestion := range fixSuggestions {
			notify.WriteMessage(notify.Message{
				Type:    notify.ActivityType,
				Content: suggestion,
				Writer:  m.Writer,
			})
		}

		warnings := helpers.FormatValidationWarnings(result)
		for _, warning := range warnings {
			notify.WriteMessage(notify.Message{
				Type:    notify.WarningType,
				Content: warning,
				Writer:  m.Writer,
			})
		}

		errorCount := len(result.Errors)
		warningCount := len(result.Warnings)

		return fmt.Errorf(
			"%w with %d errors and %d warnings",
			helpers.ErrConfigurationValidationFailed,
			errorCount,
			warningCount,
		)
	}

	warnings := helpers.FormatValidationWarnings(result)
	for _, warning := range warnings {
		notify.WriteMessage(notify.Message{
			Type:    notify.WarningType,
			Content: warning,
			Writer:  m.Writer,
		})
	}

	return nil
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

// loadDistributionConfigsForValidation loads distribution configurations for cross-validation.
// Only loads configs when Cilium CNI is requested and distributionConfig is specified.
func (m *ConfigManager) loadDistributionConfigsForValidation() []any {
	var distributionConfigs []any

	// Only attempt to load distribution configs if the config file path is specified
	// This avoids unnecessary file operations during testing or when using defaults
	if m.Config.Spec.DistributionConfig != "" && m.Config.Spec.CNI == v1alpha1.CNICilium {
		// Try to load distribution config for CNI alignment validation
		switch m.Config.Spec.Distribution {
		case v1alpha1.DistributionKind:
			kindConfig := m.loadKindConfig()
			if kindConfig != nil {
				distributionConfigs = append(distributionConfigs, kindConfig)
			}
		case v1alpha1.DistributionK3d:
			k3dConfig := m.loadK3dConfig()
			if k3dConfig != nil {
				distributionConfigs = append(distributionConfigs, k3dConfig)
			}
		}
	}

	return distributionConfigs
}

// loadKindConfig loads the Kind distribution configuration if it exists.
// Returns nil if the config doesn't exist or cannot be loaded (non-critical for validation).
func (m *ConfigManager) loadKindConfig() *kindv1alpha4.Cluster {
	if m.Config.Spec.DistributionConfig == "" {
		return nil
	}

	kindManager := kindconfigmanager.NewConfigManager(m.Config.Spec.DistributionConfig)

	err := kindManager.LoadConfig(nil)
	if err != nil {
		// Config not found or invalid, return nil for validation to continue
		return nil
	}

	return kindManager.GetConfig()
}

// loadK3dConfig loads the K3d distribution configuration if it exists.
// Returns nil if the config doesn't exist or cannot be loaded (non-critical for validation).
func (m *ConfigManager) loadK3dConfig() *k3dv1alpha5.SimpleConfig {
	if m.Config.Spec.DistributionConfig == "" {
		return nil
	}

	k3dManager := k3dconfigmanager.NewConfigManager(m.Config.Spec.DistributionConfig)

	err := k3dManager.LoadConfig(nil)
	if err != nil {
		// Config not found or invalid, return nil for validation to continue
		return nil
	}

	return k3dManager.GetConfig()
}
