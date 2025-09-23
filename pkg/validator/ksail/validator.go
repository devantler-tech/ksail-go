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

	// Validate required metadata.name
	if config.Metadata.Name == "" {
		result.AddError(validator.ValidationError{
			Field:         "metadata.name",
			Message:       "cluster name is required",
			CurrentValue:  config.Metadata.Name,
			ExpectedValue: "non-empty string",
			FixSuggestion: "Set metadata.name to a valid cluster name",
		})
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

	// Validate context naming patterns if connection context is specified
	if config.Spec.Connection.Context != "" {
		if !isValidContextPattern(
			config.Spec.Distribution,
			config.Spec.Connection.Context,
			config.Metadata.Name,
		) {
			result.AddError(validator.ValidationError{
				Field:        "spec.connection.context",
				Message:      "context name does not match expected pattern for distribution",
				CurrentValue: config.Spec.Connection.Context,
				ExpectedValue: getExpectedContextPattern(
					config.Spec.Distribution,
					config.Metadata.Name,
				),
				FixSuggestion: getContextFixSuggestion(
					config.Spec.Distribution,
					config.Metadata.Name,
				),
			})
		}
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

// isValidContextPattern validates context naming patterns based on distribution.
func isValidContextPattern(
	distribution v1alpha1.Distribution,
	context string,
	clusterName string,
) bool {
	switch distribution {
	case v1alpha1.DistributionKind:
		// Kind contexts should follow pattern: kind-{cluster-name}
		expectedContext := "kind-" + clusterName
		return context == expectedContext
	case v1alpha1.DistributionK3d:
		// K3d contexts should follow pattern: k3d-{cluster-name}
		expectedContext := "k3d-" + clusterName
		return context == expectedContext
	case v1alpha1.DistributionEKS:
		// EKS contexts can be ARN format or simple name
		// ARN format: arn:aws:eks:region:account:cluster/cluster-name
		// Simple format: cluster-name (same as metadata.name)
		if context == clusterName {
			return true
		}
		// Basic ARN pattern validation (simplified)
		if len(context) > 20 && context[:12] == "arn:aws:eks:" {
			return true
		}
		return false
	default:
		return false
	}
}

// getExpectedContextPattern returns the expected context pattern for a distribution.
func getExpectedContextPattern(distribution v1alpha1.Distribution, clusterName string) string {
	switch distribution {
	case v1alpha1.DistributionKind:
		return "kind-" + clusterName
	case v1alpha1.DistributionK3d:
		return "k3d-" + clusterName
	case v1alpha1.DistributionEKS:
		return clusterName + " or arn:aws:eks:region:account:cluster/" + clusterName
	default:
		return "unknown"
	}
}

// getContextFixSuggestion returns a fix suggestion for context validation errors.
func getContextFixSuggestion(distribution v1alpha1.Distribution, clusterName string) string {
	switch distribution {
	case v1alpha1.DistributionKind:
		return "Set spec.connection.context to 'kind-" + clusterName + "'"
	case v1alpha1.DistributionK3d:
		return "Set spec.connection.context to 'k3d-" + clusterName + "'"
	case v1alpha1.DistributionEKS:
		return "Set spec.connection.context to '" + clusterName + "' or the full EKS cluster ARN"
	default:
		return "Use the correct context pattern for your distribution"
	}
}
