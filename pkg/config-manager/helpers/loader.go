// Package helpers provides common functionality for config managers to eliminate duplication.
package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// ErrConfigurationValidationFailed is returned when configuration validation fails.
var ErrConfigurationValidationFailed = errors.New("configuration validation failed")

// LoadConfigFromFile loads a configuration from a file with common error handling and path resolution.
// This function eliminates duplication between different config managers.
//
// Parameters:
//   - configPath: The path to the configuration file
//   - createDefault: Function to create a default configuration
//
// Returns the loaded configuration or an error.
//
//nolint:ireturn // Generic function must return interface type
func LoadConfigFromFile[T any](
	configPath string,
	createDefault func() T,
) (T, error) {
	// Resolve the config path (traverse up from current dir if relative)
	resolvedPath, err := io.FindFile(configPath)
	if err != nil {
		var zero T

		return zero, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if config file exists
	_, err = os.Stat(resolvedPath)
	if os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		return createDefault(), nil
	}

	// Read file contents safely
	// Since we've resolved the path through traversal, we use the directory containing the file as the base
	cleaned := filepath.Clean(resolvedPath)
	baseDir := filepath.Dir(cleaned)

	data, err := io.ReadFileSafe(baseDir, cleaned)
	if err != nil {
		var zero T

		return zero, fmt.Errorf("failed to read config file %s: %w", cleaned, err)
	}

	// Parse YAML into the default config (which will overwrite defaults with file values)
	config := createDefault()
	marshaller := yamlmarshaller.YAMLMarshaller[T]{}

	err = marshaller.Unmarshal(data, &config)
	if err != nil {
		var zero T

		return zero, fmt.Errorf("failed to unmarshal config from %s: %w", cleaned, err)
	}

	return config, nil
}

// FormatValidationErrors formats validation errors into a readable string.
// This function eliminates duplication between different config managers.
func FormatValidationErrors(result *validator.ValidationResult) string {
	if len(result.Errors) == 0 {
		return "unknown validation error"
	}

	var errorMsg string

	for i, err := range result.Errors {
		if i > 0 {
			errorMsg += "; "
		}

		errorMsg += fmt.Sprintf("%s: %s", err.Field, err.Message)
		if err.FixSuggestion != "" {
			errorMsg += fmt.Sprintf(" (%s)", err.FixSuggestion)
		}
	}

	return errorMsg
}

// ValidateConfig validates a configuration and returns an error if validation fails.
// This function eliminates duplication between different config managers.
func ValidateConfig[T any](config T, validatorInstance validator.Validator[T]) error {
	result := validatorInstance.Validate(config)
	if !result.Valid {
		return fmt.Errorf(
			"%w: %s",
			ErrConfigurationValidationFailed,
			FormatValidationErrors(result),
		)
	}

	return nil
}
