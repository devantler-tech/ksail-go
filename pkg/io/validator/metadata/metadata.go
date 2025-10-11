// Package metadata provides shared metadata validation utilities used across multiple validators.
package metadata

import "github.com/devantler-tech/ksail-go/pkg/io/validator"

// ValidateMetadata validates Kind and APIVersion fields using provided expected values.
func ValidateMetadata(
	kind, apiVersion, expectedKind, expectedAPIVersion string,
	result *validator.ValidationResult,
) {
	// Validate Kind field
	if kind == "" {
		result.AddError(validator.ValidationError{
			Field:         "kind",
			Message:       "kind is required",
			ExpectedValue: expectedKind,
			FixSuggestion: "Set kind to '" + expectedKind + "'",
		})
	}

	// Validate APIVersion field
	if apiVersion == "" {
		result.AddError(validator.ValidationError{
			Field:         "apiVersion",
			Message:       "apiVersion is required",
			ExpectedValue: expectedAPIVersion,
			FixSuggestion: "Set apiVersion to '" + expectedAPIVersion + "'",
		})
	}
}

// ValidateNilConfig checks if config is nil and adds appropriate error.
func ValidateNilConfig(
	config any,
	configType string,
	result *validator.ValidationResult,
) bool {
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration is nil",
			FixSuggestion: "Provide a valid " + configType + " configuration",
		})

		return true
	}

	return false
}
