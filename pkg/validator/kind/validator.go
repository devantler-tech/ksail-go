// Package kind provides Kind configuration validation functionality.
package kind

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Validator validates Kind cluster configurations using upstream Kind APIs.
type Validator struct{}

// NewValidator creates a new Kind configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded Kind cluster configuration.
func (v *Validator) Validate(config *kindapi.Cluster) *validator.ValidationResult {
	result := validator.NewValidationResult("kind.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration cannot be nil",
			FixSuggestion: "Provide a valid Kind cluster configuration",
		})

		return result
	}

	// Validate required metadata fields
	if config.Kind == "" {
		result.AddError(validator.ValidationError{
			Field:         "kind",
			Message:       "kind is required",
			ExpectedValue: "Cluster",
			FixSuggestion: "Set kind to 'Cluster'",
		})
	}

	if config.APIVersion == "" {
		result.AddError(validator.ValidationError{
			Field:         "apiVersion",
			Message:       "apiVersion is required",
			ExpectedValue: "kind.x-k8s.io/v1alpha4",
			FixSuggestion: "Set apiVersion to 'kind.x-k8s.io/v1alpha4'",
		})
	}

	return result
}
