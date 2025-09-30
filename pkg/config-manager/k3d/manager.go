// Package k3d provides configuration management for K3d v1alpha5.SimpleConfig configurations.
// This file contains the core Manager implementation for loading K3d configurations from files.
package k3d

import (
	"fmt"
	"io"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// ConfigManager implements configuration management for K3d v1alpha5.SimpleConfig configurations.
// It provides file-based configuration loading without Viper dependency.
type ConfigManager struct {
	configPath   string
	config       *v1alpha5.SimpleConfig
	configLoaded bool
	writer       io.Writer
}

// Compile-time interface compliance verification.
// This ensures ConfigManager properly implements configmanager.ConfigManager[v1alpha5.SimpleConfig].
var _ configmanager.ConfigManager[v1alpha5.SimpleConfig] = (*ConfigManager)(nil)

// NewK3dSimpleConfig creates a new v1alpha5.SimpleConfig with the specified name and TypeMeta.
// This function provides a canonical way to create K3d clusters with proper field initialization.
// Use empty string for name to create a cluster without a specific name.
func NewK3dSimpleConfig(name, apiVersion, kind string) *v1alpha5.SimpleConfig {
	// Set default name if empty
	if name == "" {
		name = "k3d-default"
	}

	if apiVersion == "" {
		apiVersion = "k3d.io/v1alpha5"
	}

	if kind == "" {
		kind = "Simple"
	}

	return &v1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		ObjectMeta: types.ObjectMeta{
			Name: name,
		},
	}
}

// NewConfigManager creates a new configuration manager for K3d cluster configurations.
// configPath specifies the path to the K3d configuration file to load.
func NewConfigManager(configPath string, writer io.Writer) *ConfigManager {
	return &ConfigManager{
		configPath:   configPath,
		config:       nil,
		configLoaded: false,
		writer:       writer,
	}
}

// LoadConfig loads the K3d configuration from the specified file.
// Returns the previously loaded config if already loaded.
// If the file doesn't exist, returns a default K3d cluster configuration.
// Validates the configuration after loading and returns an error if validation fails.
func (m *ConfigManager) LoadConfig() (*v1alpha5.SimpleConfig, error) {
	// If config is already loaded, return it
	if m.configLoaded {
		return m.config, nil
	}

	config, err := helpers.LoadConfigFromFile(
		m.configPath,
		func() *v1alpha5.SimpleConfig {
			// Create default with proper APIVersion and Kind
			config := NewK3dSimpleConfig("", "k3d.io/v1alpha5", "Simple")

			return config
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate the loaded configuration
	validator := k3dvalidator.NewValidator()
	validationResult := validator.Validate(config)
	if !validationResult.Valid {
		formattedWarnings := helpers.FormatValidationWarnings(validationResult)
		for _, warning := range formattedWarnings {
			notify.WarnMessage(m.writer, notify.NewMessage(warning))
		}
		formattedErrors := helpers.FormatValidationErrors(validationResult)
		for _, errMsg := range formattedErrors {
			notify.ErrorMessage(m.writer, notify.NewMessage(errMsg))
		}

		warningLength := len(formattedWarnings)
		errorLength := len(formattedErrors)
		return nil, fmt.Errorf(
			"%w: %s",
			helpers.ErrConfigurationValidationFailed,
			fmt.Sprintf("found %d warning(s) and %d error(s)", warningLength, errorLength),
		)
	}

	m.config = config
	m.configLoaded = true

	return m.config, nil
}
