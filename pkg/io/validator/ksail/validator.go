// Package ksail provides validation for KSail cluster configurations.
package ksail

import (
	"fmt"
	"strings"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	"github.com/devantler-tech/ksail-go/pkg/io/validator/metadata"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const requiredCiliumArgs = 2

// Validator validates KSail cluster configurations for semantic correctness and cross-configuration consistency.
type Validator struct {
	kindConfig *kindv1alpha4.Cluster
	k3dConfig  *k3dapi.SimpleConfig
}

// NewValidator creates a new KSail configuration validator without distribution configuration.
// Use NewValidatorForKind or NewValidatorForK3d for distribution-specific validation.
func NewValidator() *Validator {
	return &Validator{}
}

// NewValidatorForKind creates a new KSail configuration validator with Kind distribution configuration.
// The Kind config is used for cross-configuration validation (name consistency, CNI alignment).
func NewValidatorForKind(kindConfig *kindv1alpha4.Cluster) *Validator {
	return &Validator{
		kindConfig: kindConfig,
	}
}

// NewValidatorForK3d creates a new KSail configuration validator with K3d distribution configuration.
// The K3d config is used for cross-configuration validation (name consistency, CNI alignment).
func NewValidatorForK3d(k3dConfig *k3dapi.SimpleConfig) *Validator {
	return &Validator{
		k3dConfig: k3dConfig,
	}
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
// Only validates when a distribution config is provided to the validator.
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
		// For EKS or unknown distributions, or when no distribution config is provided, skip context validation
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
// from the distribution config. Returns empty string if no distribution config is available.
func (v *Validator) getExpectedContextName(config *v1alpha1.Cluster) string {
	distributionName := v.getDistributionConfigName(config.Spec.Distribution)
	if distributionName == "" {
		// No distribution config available, skip context validation
		return ""
	}

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
// Returns empty string if no Kind config is provided to the validator.
func (v *Validator) getKindConfigName() string {
	if v.kindConfig != nil && v.kindConfig.Name != "" {
		return v.kindConfig.Name
	}

	// No Kind config provided, return empty to skip validation
	return ""
}

// getK3dConfigName returns the K3d configuration name if available.
// Returns empty string if no K3d config is provided to the validator.
func (v *Validator) getK3dConfigName() string {
	if v.k3dConfig != nil && v.k3dConfig.Name != "" {
		return v.k3dConfig.Name
	}

	// No K3d config provided, return empty to skip validation
	return ""
}

// validateCNIAlignment validates that the distribution configuration aligns with the CNI setting.
// When Cilium CNI is requested, the distribution config must have CNI disabled.
// When Istio is requested, the default CNI should remain enabled (Istio is a service mesh, not a CNI replacement).
// When Default CNI is used, the distribution config must NOT have CNI disabled.
func (v *Validator) validateCNIAlignment(
	config *v1alpha1.Cluster,
	result *validator.ValidationResult,
) {
	// Validate Cilium CNI alignment - requires disabling default CNI
	if config.Spec.CNI == v1alpha1.CNICilium {
		switch config.Spec.Distribution {
		case v1alpha1.DistributionKind:
			v.validateKindCustomCNIAlignment(config.Spec.CNI, result)
		case v1alpha1.DistributionK3d:
			v.validateK3dCustomCNIAlignment(config.Spec.CNI, result)
		}

		return
	}

	// Istio is a service mesh, not a CNI replacement - it works on top of the default CNI
	// No special CNI configuration required for Istio

	// Validate Default CNI alignment (empty string or explicit "Default")
	if config.Spec.CNI == "" || config.Spec.CNI == v1alpha1.CNIDefault {
		switch config.Spec.Distribution {
		case v1alpha1.DistributionKind:
			v.validateKindDefaultCNIAlignment(result)
		case v1alpha1.DistributionK3d:
			v.validateK3dDefaultCNIAlignment(result)
		}
	}
}

// validateKindCustomCNIAlignment validates that Kind configuration has CNI disabled when custom CNI is requested.
func (v *Validator) validateKindCustomCNIAlignment(
	cni v1alpha1.CNI,
	result *validator.ValidationResult,
) {
	if v.kindConfig == nil {
		// No Kind config provided for validation, skip
		return
	}

	if !v.kindConfig.Networking.DisableDefaultCNI {
		result.AddError(validator.ValidationError{
			Field: "spec.cni",
			Message: fmt.Sprintf(
				"%s CNI requires disableDefaultCNI to be true in Kind configuration",
				cni,
			),
			FixSuggestion: "Add 'networking.disableDefaultCNI: true' to your kind.yaml configuration file",
		})
	}
}

// validateKindDefaultCNIAlignment validates that Kind configuration does NOT have CNI disabled when Default is used.
func (v *Validator) validateKindDefaultCNIAlignment(result *validator.ValidationResult) {
	if v.kindConfig == nil {
		// No Kind config provided for validation, skip
		return
	}

	if v.kindConfig.Networking.DisableDefaultCNI {
		result.AddError(validator.ValidationError{
			Field:         "spec.cni",
			Message:       "Default CNI requires disableDefaultCNI to be false in Kind configuration",
			CurrentValue:  "disableDefaultCNI: true",
			ExpectedValue: "disableDefaultCNI: false (or omit the field)",
			FixSuggestion: "Remove 'networking.disableDefaultCNI: true' from your kind.yaml " +
				"configuration file or set CNI to Cilium",
		})
	}
}

// checkK3dFlannelAndNetworkPolicyStatus checks if Flannel and network policy are disabled in K3d configuration.
// Returns (hasFlannelDisabled, hasNetworkPolicyDisabled).
func (v *Validator) checkK3dFlannelAndNetworkPolicyStatus() (bool, bool) {
	var (
		hasFlannelDisabled       bool
		hasNetworkPolicyDisabled bool
	)

	for _, arg := range v.k3dConfig.Options.K3sOptions.ExtraArgs {
		switch arg.Arg {
		case "--flannel-backend=none":
			hasFlannelDisabled = true
		case "--disable-network-policy":
			hasNetworkPolicyDisabled = true
		}
	}

	return hasFlannelDisabled, hasNetworkPolicyDisabled
}

// validateK3dCustomCNIAlignment validates that K3d configuration has Flannel disabled when custom CNI is requested.
func (v *Validator) validateK3dCustomCNIAlignment(
	cni v1alpha1.CNI,
	result *validator.ValidationResult,
) {
	if v.k3dConfig == nil {
		// No K3d config provided for validation, skip
		return
	}

	hasFlannelDisabled, hasNetworkPolicyDisabled := v.checkK3dFlannelAndNetworkPolicyStatus()

	missingArgs := make([]string, 0, requiredCiliumArgs)
	if !hasFlannelDisabled {
		missingArgs = append(missingArgs, "'--flannel-backend=none'")
	}

	if !hasNetworkPolicyDisabled {
		missingArgs = append(missingArgs, "'--disable-network-policy'")
	}

	if len(missingArgs) == 0 {
		return
	}

	result.AddError(validator.ValidationError{
		Field: "spec.cni",
		Message: fmt.Sprintf(
			"%s CNI requires %s in K3d configuration",
			cni,
			strings.Join(missingArgs, " and "),
		),
		FixSuggestion: fmt.Sprintf(
			"Add %s to the K3s extra args in your k3d.yaml configuration file",
			strings.Join(missingArgs, " and "),
		),
	})
}

// validateK3dDefaultCNIAlignment validates that K3d configuration does NOT have Flannel disabled when Default is used.
func (v *Validator) validateK3dDefaultCNIAlignment(result *validator.ValidationResult) {
	if v.k3dConfig == nil {
		// No K3d config provided for validation, skip
		return
	}

	hasFlannelDisabled, hasNetworkPolicyDisabled := v.checkK3dFlannelAndNetworkPolicyStatus()

	problematicArgs := make([]string, 0, requiredCiliumArgs)
	if hasFlannelDisabled {
		problematicArgs = append(problematicArgs, "'--flannel-backend=none'")
	}

	if hasNetworkPolicyDisabled {
		problematicArgs = append(problematicArgs, "'--disable-network-policy'")
	}

	if len(problematicArgs) == 0 {
		return
	}

	result.AddError(validator.ValidationError{
		Field: "spec.cni",
		Message: fmt.Sprintf(
			"Default CNI requires Flannel to be enabled, but found %s in K3d configuration",
			strings.Join(problematicArgs, " and "),
		),
		FixSuggestion: fmt.Sprintf(
			"Remove %s from the K3s extra args in your k3d.yaml configuration file or set CNI to Cilium",
			strings.Join(problematicArgs, " and "),
		),
	})
}
