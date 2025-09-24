// Package ksail provides validation for KSail cluster configurations.
package ksail

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	eksctl "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Validator validates KSail cluster configurations for semantic correctness and cross-configuration consistency.
type Validator struct {
	kindConfig *kindv1alpha4.Cluster
	k3dConfig  *k3dapi.SimpleConfig
	eksConfig  *eksctl.ClusterConfig
}

// NewValidator creates a new KSail configuration validator with optional distribution configurations.
// Distribution configs are used for cross-configuration validation (name consistency, context patterns).
func NewValidator(distributionConfigs ...any) *Validator {
	validator := &Validator{}

	// Accept distribution configurations for cross-configuration validation
	for _, config := range distributionConfigs {
		switch cfg := config.(type) {
		case *kindv1alpha4.Cluster:
			validator.kindConfig = cfg
		case *k3dapi.SimpleConfig:
			validator.k3dConfig = cfg
		case *eksctl.ClusterConfig:
			validator.eksConfig = cfg
		}
	}

	return validator
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
			ExpectedValue: "ksail.dev/v1alpha1",
			FixSuggestion: "Set apiVersion to 'ksail.dev/v1alpha1'",
		})
	}

	// Validate distribution field
	v.validateDistribution(config, result)

	// Perform cross-configuration validation
	v.validateContextName(config, result)

	return result
}

// validateContextName validates the context name pattern matches the distribution and cluster name.
func (v *Validator) validateContextName(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	// Skip context validation for EKS as it doesn't rely on context names
	if config.Spec.Distribution == v1alpha1.DistributionEKS {
		return
	}

	if config.Spec.Connection.Context == "" {
		// Context is optional, no validation needed if empty
		return
	}

	expectedContext := v.getExpectedContextName(config)
	if expectedContext == "" {
		// Add error for unsupported distributions
		v.addUnsupportedDistributionError(config, result)
		return
	}

	// Check for unsupported distributions that return invalid context patterns
	if v.isUnsupportedDistribution(config.Spec.Distribution) {
		v.addUnsupportedDistributionError(config, result)
		return
	}

	if config.Spec.Connection.Context != expectedContext {
		result.AddError(validator.ValidationError{
			Field:         "spec.connection.context",
			Message:       "context name does not match expected pattern for distribution",
			CurrentValue:  config.Spec.Connection.Context,
			ExpectedValue: expectedContext,
			FixSuggestion: fmt.Sprintf(
				"Set context to '%s' to match the %s distribution pattern",
				expectedContext,
				config.Spec.Distribution,
			),
		})
	}
}

// validateDistribution validates the distribution field for emptiness and validity.
func (v *Validator) validateDistribution(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	distribution := config.Spec.Distribution

	// Check if distribution is empty or invalid
	if distribution == "" || !distribution.IsValid() {
		var message, fixSuggestion string

		if distribution == "" {
			message = "distribution is required"
			fixSuggestion = "Set spec.distribution to a supported distribution type"
		} else {
			message = "invalid distribution value"
			fixSuggestion = "Use a valid distribution type: Kind, K3d, or EKS"
		}

		result.AddError(validator.ValidationError{
			Field:         "spec.distribution",
			Message:       message,
			CurrentValue:  distribution,
			ExpectedValue: "one of: Kind, K3d, EKS",
			FixSuggestion: fixSuggestion,
		})
	}

	// Validate distributionConfig field
	if config.Spec.DistributionConfig == "" {
		result.AddError(validator.ValidationError{
			Field:         "spec.distributionConfig",
			Message:       "distributionConfig is required",
			FixSuggestion: "Set spec.distributionConfig to the distribution configuration file path",
		})
	}
}

// isUnsupportedDistribution checks if the distribution is not supported for context validation.
func (v *Validator) isUnsupportedDistribution(distribution v1alpha1.Distribution) bool {
	return distribution == v1alpha1.DistributionTind
}

// addUnsupportedDistributionError adds validation errors for unsupported distributions.
func (v *Validator) addUnsupportedDistributionError(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	distribution := config.Spec.Distribution
	switch distribution {
	case v1alpha1.DistributionTind:
		result.AddError(validator.ValidationError{
			Field:         "spec.distribution",
			Message:       "Tind distribution is not yet supported for context validation",
			CurrentValue:  distribution,
			FixSuggestion: "Use a supported distribution: Kind, K3d, or EKS",
		})
	default:
		result.AddError(validator.ValidationError{
			Field:         "spec.distribution",
			Message:       "unknown distribution for context validation",
			CurrentValue:  distribution,
			FixSuggestion: "Use a supported distribution: Kind, K3d, or EKS",
		})
	}
}

// getExpectedContextName returns the expected context name for the given configuration.
// Context name follows the pattern: {distribution}-{cluster_name}, where cluster_name is extracted from the distribution config.
// If no cluster name is found, "ksail-default" is used as the ultimate fallback.
func (v *Validator) getExpectedContextName(config *v1alpha1.Cluster) string {
	distributionName := v.getDistributionConfigName(config.Spec.Distribution)

	switch config.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return "kind-" + distributionName
	case v1alpha1.DistributionK3d:
		return "k3d-" + distributionName
	case v1alpha1.DistributionEKS:
		// EKS context pattern is more flexible (cluster name or ARN)
		return distributionName
	case v1alpha1.DistributionTind:
		// Tind context pattern would be similar to k3d
		return "tind-" + distributionName
	default:
		return ""
	}
}

// getDistributionConfigName extracts the cluster name from the distribution configuration.
func (v *Validator) getDistributionConfigName(distribution v1alpha1.Distribution) string {
	switch distribution {
	case v1alpha1.DistributionKind:
		return v.getKindConfigName()
	case v1alpha1.DistributionK3d:
		return v.getK3dConfigName()
	case v1alpha1.DistributionEKS:
		return v.getEKSConfigName()
	case v1alpha1.DistributionTind:
		// Tind configuration name extraction would go here when implemented
		return ""
	default:
		return ""
	}
}

// getKindConfigName returns the Kind configuration name if available.
func (v *Validator) getKindConfigName() string {
	if v.kindConfig != nil && v.kindConfig.Name != "" {
		return v.kindConfig.Name
	}

	// Return default Kind cluster name when no config is provided
	return "kind"
}

// getK3dConfigName returns the K3d configuration name if available.
func (v *Validator) getK3dConfigName() string {
	if v.k3dConfig != nil && v.k3dConfig.Name != "" {
		return v.k3dConfig.Name
	}

	// Return default K3d cluster name when no config is provided
	return "k3s-default"
}

// getEKSConfigName returns the EKS configuration name if available.
func (v *Validator) getEKSConfigName() string {
	if v.eksConfig != nil && v.eksConfig.Metadata != nil && v.eksConfig.Metadata.Name != "" {
		return v.eksConfig.Metadata.Name
	}

	// Return default EKS cluster name when no config is provided
	return "eks-default"
}
