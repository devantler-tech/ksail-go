package configmanager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ConfigManager implements configuration management for KSail v1alpha1.Cluster configurations.
type ConfigManager struct {
	Viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
	configLoaded   bool              // Track if config has been actually loaded
	Writer         io.Writer         // Writer for output notifications
	command        *cobra.Command    // Associated Cobra command for flag introspection
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
	manager.command = cmd
	manager.AddFlagsFromFields(cmd)

	return manager
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the loaded config (either freshly loaded or previously cached) and an error if loading failed.
// Returns nil config on error.
// Configuration priority: defaults < config files < environment variables < flags.
// If timer is provided, timing information will be included in the success notification.
func (m *ConfigManager) LoadConfig(tmr timer.Timer) (*v1alpha1.Cluster, error) {
	return m.loadConfigWithOptions(tmr, false)
}

// LoadConfigSilent loads the configuration without outputting notifications.
// Returns the loaded config, either freshly loaded or previously cached.
func (m *ConfigManager) LoadConfigSilent() (*v1alpha1.Cluster, error) {
	return m.loadConfigWithOptions(nil, true)
}

// loadConfigWithOptions is the internal implementation with silent option.
func (m *ConfigManager) loadConfigWithOptions(
	tmr timer.Timer,
	silent bool,
) (*v1alpha1.Cluster, error) {
	if !silent {
		m.notifyLoadingStart()
	}

	if m.configLoaded {
		if !silent {
			m.notifyConfigReused()
		}

		return m.Config, nil
	}

	if !silent {
		m.notifyLoadingConfig()
	}

	// Use native Viper API to read configuration
	err := m.readConfig(silent)
	if err != nil {
		return nil, err
	}

	// Unmarshal and apply defaults
	flagOverrides := m.captureChangedFlagValues()

	err = m.unmarshalAndApplyDefaults()
	if err != nil {
		return nil, err
	}

	err = m.applyFlagOverrides(flagOverrides)
	if err != nil {
		return nil, err
	}

	err = m.validateConfig()
	if err != nil {
		return nil, err
	}

	if !silent {
		m.notifyLoadingComplete(tmr)
	}

	m.configLoaded = true

	return m.Config, nil
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

func (m *ConfigManager) captureChangedFlagValues() map[string]string {
	if m.command == nil {
		return nil
	}

	flags := m.command.Flags()
	overrides := make(map[string]string)

	flags.Visit(func(f *pflag.Flag) {
		overrides[f.Name] = f.Value.String()
	})

	return overrides
}

func (m *ConfigManager) applyFlagOverrides(overrides map[string]string) error {
	if overrides == nil {
		return nil
	}

	for _, selector := range m.fieldSelectors {
		fieldPtr := selector.Selector(m.Config)
		if fieldPtr == nil {
			continue
		}

		flagName := m.GenerateFlagName(fieldPtr)

		value, ok := overrides[flagName]
		if !ok {
			continue
		}

		err := setFieldValueFromFlag(fieldPtr, value)
		if err != nil {
			return fmt.Errorf("failed to apply flag override for %s: %w", flagName, err)
		}
	}

	return nil
}

func setFieldValueFromFlag(fieldPtr any, raw string) error {
	if setter, ok := fieldPtr.(interface{ Set(value string) error }); ok {
		err := setter.Set(raw)
		if err != nil {
			return fmt.Errorf("set field value via setter: %w", err)
		}

		return nil
	}

	switch ptr := fieldPtr.(type) {
	case *string:
		*ptr = raw

		return nil
	case *metav1.Duration:
		if raw == "" {
			ptr.Duration = 0

			return nil
		}

		dur, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("parse duration %q: %w", raw, err)
		}

		ptr.Duration = dur

		return nil
	default:
		return nil
	}
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
	// Create validator with distribution config for cross-validation
	validator := m.createValidatorForDistribution()
	result := validator.Validate(m.Config)

	if !result.Valid {
		errorMessages := helpers.FormatValidationErrorsMultiline(result)
		notify.WriteMessage(notify.Message{
			Type:    notify.ErrorType,
			Content: "%s",
			Args:    []any{errorMessages},
			Writer:  m.Writer,
		})

		warnings := helpers.FormatValidationWarnings(result)
		for _, warning := range warnings {
			notify.WriteMessage(notify.Message{
				Type:    notify.WarningType,
				Content: warning,
				Writer:  m.Writer,
			})
		}

		// Return validation summary error instead of full error stack
		return helpers.NewValidationSummaryError(len(result.Errors), len(result.Warnings))
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

// createValidatorForDistribution creates a validator with the appropriate distribution config.
// Only loads distribution config when custom CNI (Cilium or Istio) is requested for validation.
func (m *ConfigManager) createValidatorForDistribution() *ksailvalidator.Validator {
	// Only load distribution config for custom CNI validation
	if m.Config.Spec.DistributionConfig == "" ||
		(m.Config.Spec.CNI != v1alpha1.CNICilium && m.Config.Spec.CNI != v1alpha1.CNIIstio) {
		return ksailvalidator.NewValidator()
	}

	// Create distribution-specific validator based on configured distribution
	switch m.Config.Spec.Distribution {
	case v1alpha1.DistributionKind:
		kindConfig := m.loadKindConfig()
		if kindConfig != nil {
			return ksailvalidator.NewValidatorForKind(kindConfig)
		}
	case v1alpha1.DistributionK3d:
		k3dConfig := m.loadK3dConfig()
		if k3dConfig != nil {
			return ksailvalidator.NewValidatorForK3d(k3dConfig)
		}
	}

	return ksailvalidator.NewValidator()
}

// loadKindConfig loads the Kind distribution configuration if it exists.
// Returns nil if the config doesn't exist or cannot be loaded (non-critical for validation).
func (m *ConfigManager) loadKindConfig() *kindv1alpha4.Cluster {
	if m.Config.Spec.DistributionConfig == "" {
		return nil
	}

	// Check if the file actually exists before trying to load it
	// This prevents validation against default configs during init
	_, err := os.Stat(m.Config.Spec.DistributionConfig)
	if os.IsNotExist(err) {
		return nil
	}

	kindManager := kindconfigmanager.NewConfigManager(m.Config.Spec.DistributionConfig)

	config, err := kindManager.LoadConfig(nil)
	if err != nil {
		// Config not found or invalid, return nil for validation to continue
		return nil
	}

	return config
}

// loadK3dConfig loads the K3d distribution configuration if it exists.
// Returns nil if the config doesn't exist or cannot be loaded (non-critical for validation).
func (m *ConfigManager) loadK3dConfig() *k3dv1alpha5.SimpleConfig {
	if m.Config.Spec.DistributionConfig == "" {
		return nil
	}

	// Check if the file actually exists before trying to load it
	// This prevents validation against default configs during init
	_, err := os.Stat(m.Config.Spec.DistributionConfig)
	if os.IsNotExist(err) {
		return nil
	}

	k3dManager := k3dconfigmanager.NewConfigManager(m.Config.Spec.DistributionConfig)

	config, err := k3dManager.LoadConfig(nil)
	if err != nil {
		// Config not found or invalid, return nil for validation to continue
		return nil
	}

	return config
}
