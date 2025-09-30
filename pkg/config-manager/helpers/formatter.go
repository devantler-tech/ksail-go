// Package helpers provides common functionality for config managers to eliminate duplication.
package helpers

import (
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// ErrConfigurationValidationFailed is returned when configuration validation fails.
var ErrConfigurationValidationFailed = errors.New("config validation failed")

// FormatValidationErrors formats validation errors into a structured string format.
// This function eliminates duplication between different config managers.
//
// Returns a slice of formatted error strings, where each error includes:
//   - Field name in quotes
//   - Error message
//   - Location information (if available)
//   - Fix suggestion (if available)
func FormatValidationErrors(result *validator.ValidationResult) []string {
	errors := make([]string, 0)

	for _, error := range result.Errors {
		// Start msg
		msg := fmt.Sprintf("error: %s", error.Message)

		// Add location information
		if error.Location.FilePath != "" || error.Field != "" {
			msg += "\n  in: "
			if error.Location.FilePath != "" {
				msg += error.Location.FilePath
				if error.Location.Line > 0 {
					msg += fmt.Sprintf(":%d ", error.Location.Line)
				} else {
					msg += " "
				}
			}
			if error.Field != "" {
				msg += fmt.Sprintf("'%s'", error.Field)
			}
		}

		// Add fix suggestion
		if error.FixSuggestion != "" {
			msg += fmt.Sprintf("\n  fix: %s", error.FixSuggestion)
		}

		errors = append(errors, msg)
	}

	return errors
}

// FormatValidationWarnings formats validation warnings for CLI display.
// This function provides a standardized way to format validation warnings.
//
// Returns a slice of formatted warning strings, where each warning includes:
//   - Field name in quotes
//   - Warning message
//   - Location information (if available)
//   - Fix suggestion (if available)
func FormatValidationWarnings(result *validator.ValidationResult) []string {
	warnings := make([]string, 0)

	for _, warning := range result.Warnings {
		// Start msg
		msg := fmt.Sprintf("warning: %s", warning.Message)

		// Add location information
		if warning.Location.FilePath != "" || warning.Field != "" {
			msg += "\n  in: "
			if warning.Location.FilePath != "" {
				msg += warning.Location.FilePath
				if warning.Location.Line > 0 {
					msg += fmt.Sprintf(":%d ", warning.Location.Line)
				} else {
					msg += " "
				}
			}
			if warning.Field != "" {
				msg += fmt.Sprintf("'%s'", warning.Field)
			}
		}

		// Add fix suggestion
		if warning.FixSuggestion != "" {
			msg += fmt.Sprintf("\n  fix: %s", warning.FixSuggestion)
		}

		warnings = append(warnings, msg)
	}

	return warnings
}

// ValidateConfig validates a configuration and returns an error if validation fails.
// This function eliminates duplication between different config managers.
//
// Returns nil if the configuration is valid, or an error containing formatted
// validation errors if the configuration is invalid.
func ValidateConfig[T any](config T, validatorInstance validator.Validator[T]) error {
	result := validatorInstance.Validate(config)
	if !result.Valid {
		errors := FormatValidationErrors(result)
		// Combine all errors into a single message
		errorMsg := ""
		for i, err := range errors {
			if i > 0 {
				errorMsg += "; "
			}
			errorMsg += err
		}
		return fmt.Errorf("%w: %s", ErrConfigurationValidationFailed, errorMsg)
	}

	return nil
}
