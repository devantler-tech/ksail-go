// Package ksail provides validation for KSail cluster configurations.
package ksail

import (
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// Validator validates KSail cluster configurations for semantic correctness and cross-configuration consistency.
type Validator struct{}

// NewValidator creates a new KSail configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded KSail cluster configuration.
func (v *Validator) Validate(config *v1alpha1.Cluster) *validator.ValidationResult {
	result := validator.NewValidationResult("ksail.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration cannot be nil",
			FixSuggestion: "Provide a valid KSail cluster configuration",
		})
		return result
	}

	// Validate distribution field
	if config.Spec.Distribution == "" {
		result.AddError(validator.ValidationError{
			Field:         "spec.distribution",
			Message:       "distribution is required",
			CurrentValue:  config.Spec.Distribution,
			ExpectedValue: "one of: Kind, K3d, EKS",
			FixSuggestion: "Set spec.distribution to a supported distribution type",
		})
	} else if !isValidDistribution(config.Spec.Distribution) {
		result.AddError(validator.ValidationError{
			Field:         "spec.distribution",
			Message:       "invalid distribution value",
			CurrentValue:  config.Spec.Distribution,
			ExpectedValue: "one of: Kind, K3d, EKS",
			FixSuggestion: "Use a valid distribution type: Kind, K3d, or EKS",
		})
	}

	return result
}

// isValidDistribution checks if the distribution value is supported.
func isValidDistribution(distribution v1alpha1.Distribution) bool {
	switch distribution {
	case v1alpha1.DistributionKind, v1alpha1.DistributionK3d, v1alpha1.DistributionEKS:
		return true
	default:
		return false
	}
}
