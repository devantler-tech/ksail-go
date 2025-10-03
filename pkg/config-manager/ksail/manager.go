package configmanager

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanagerinterface "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
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
// If timer is provided, timing information will be included in the success notification.
func (m *ConfigManager) LoadConfig(tmr timer.Timer) (*v1alpha1.Cluster, error) {
	m.notifyLoadingStart()

	if m.configLoaded {
		m.notifyConfigReused()

		return m.Config, nil
	}

	m.notifyLoadingConfig()

	// Use native Viper API to read configuration
	err := m.readConfig()
	if err != nil {
		return nil, err
	}

	// Unmarshal and apply defaults
	err = m.unmarshalAndApplyDefaults()
	if err != nil {
		return nil, err
	}

	m.notifyLoadingComplete(tmr)
	m.configLoaded = true

	return m.Config, nil
}

func (m *ConfigManager) readConfig() error {
	err := m.Viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		m.notifyUsingDefaults()
	} else {
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
		Content: "Loading configuration...",
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
