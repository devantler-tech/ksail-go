// Package ksail provides validation for KSail cluster configurations.
package ksail

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	"github.com/devantler-tech/ksail-go/pkg/io/validator/metadata"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Validator validates KSail cluster configurations for semantic correctness and cross-configuration consistency.
type Validator struct {
	kindConfig *kindv1alpha4.Cluster
	k3dConfig  *k3dapi.SimpleConfig
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
			Message:       "configuration is nil",
			FixSuggestion: "Provide a valid KSail cluster configuration",
		})

		return result
	}

	// Validate required metadata fields
	metadata.ValidateMetadata(
		config.Kind,
		config.APIVersion,
		"Cluster",
		"ksail.dev/v1alpha1",
		result,
	)

	// Validate distribution field
	v.validateDistribution(config, result)

	// Perform cross-configuration validation
	v.validateContextName(config, result)

	// Validate CNI alignment with distribution config
	v.validateCNIAlignment(config, result)

	return result
}

// validateContextName validates the context name pattern matches the distribution and cluster name.
func (v *Validator) validateContextName(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	if config.Spec.Connection.Context == "" {
		// Context is optional, no validation needed if empty
		return
	}

	expectedContext := v.getExpectedContextName(config)
	if expectedContext == "" {
		// For EKS or unknown distributions, skip context validation
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
			fixSuggestion = "Use a supported distribution: Kind, K3d, or EKS"
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

// getExpectedContextName returns the expected context name for the given configuration.
// Context name follows the pattern: {distribution}-{cluster_name}, where cluster_name is extracted
// from the distribution config. If no cluster name is found, "ksail-default" is used as the ultimate fallback.
func (v *Validator) getExpectedContextName(config *v1alpha1.Cluster) string {
	distributionName := v.getDistributionConfigName(config.Spec.Distribution)

	switch config.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return "kind-" + distributionName
	case v1alpha1.DistributionK3d:
		return "k3d-" + distributionName
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

// validateCNIAlignment validates that the distribution configuration aligns with the CNI setting.
// When Cilium CNI is requested, the distribution config must have CNI disabled.
func (v *Validator) validateCNIAlignment(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	// Only validate when Cilium CNI is explicitly requested
	if config.Spec.CNI != v1alpha1.CNICilium {
		return
	}

	switch config.Spec.Distribution {
	case v1alpha1.DistributionKind:
		v.validateKindCNIAlignment(result)
	case v1alpha1.DistributionK3d:
		v.validateK3dCNIAlignment(result)
	}
}

// validateKindCNIAlignment validates that Kind configuration has CNI disabled when Cilium is requested.
func (v *Validator) validateKindCNIAlignment(result *validator.ValidationResult) {
	if v.kindConfig == nil {
		// No Kind config provided for validation, skip
		return
	}

	if !v.kindConfig.Networking.DisableDefaultCNI {
		result.AddError(validator.ValidationError{
			Field:         "spec.cni",
			Message:       "Cilium CNI requires disableDefaultCNI to be true in Kind configuration",
			FixSuggestion: "Add 'networking.disableDefaultCNI: true' to your kind.yaml configuration file",
		})
	}
}

// validateK3dCNIAlignment validates that K3d configuration has Flannel disabled when Cilium is requested.
func (v *Validator) validateK3dCNIAlignment(result *validator.ValidationResult) {
	if v.k3dConfig == nil {
		// No K3d config provided for validation, skip
		return
	}

	// Check if --flannel-backend=none is set in K3s extra args
	hasFlannelDisabled := false

	for _, arg := range v.k3dConfig.Options.K3sOptions.ExtraArgs {
		if arg.Arg == "--flannel-backend=none" {
			hasFlannelDisabled = true

			break
		}
	}

	if !hasFlannelDisabled {
		result.AddError(validator.ValidationError{
			Field:         "spec.cni",
			Message:       "Cilium CNI requires Flannel to be disabled in K3d configuration",
			FixSuggestion: "Add '--flannel-backend=none' to the K3s extra args in your k3d.yaml configuration file",
		})
	}
}
