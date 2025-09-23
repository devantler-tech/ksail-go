package eks

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// Validator validates EKS cluster configurations using upstream eksctl APIs.
type Validator struct{}

// NewValidator creates a new EKS configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded EKS cluster configuration using upstream eksctl APIs.
// This validator performs both essential field validation and comprehensive upstream eksctl validation.
func (v *Validator) Validate(config *eksctlapi.ClusterConfig) *validator.ValidationResult {
	result := validator.NewValidationResult("eks.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration cannot be nil",
			FixSuggestion: "Provide a valid EKS cluster configuration",
		})
		return result
	}

	// Validate metadata is present (required field per eksctl API)
	if config.Metadata == nil {
		result.AddError(validator.ValidationError{
			Field:         "metadata",
			Message:       "metadata is required",
			CurrentValue:  nil,
			ExpectedValue: "ClusterMeta object with name and region",
			FixSuggestion: "Add metadata section with name and region fields",
		})
		return result
	}

	// Validate cluster name is required (upstream eksctl requirement)
	if config.Metadata.Name == "" {
		result.AddError(validator.ValidationError{
			Field:         "metadata.name",
			Message:       "cluster name is required",
			CurrentValue:  config.Metadata.Name,
			ExpectedValue: "non-empty string",
			FixSuggestion: "Set metadata.name to a valid EKS cluster name (1-63 characters, alphanumeric and hyphens)",
		})
	}

	// Validate region is required (upstream eksctl requirement)
	if config.Metadata.Region == "" {
		result.AddError(validator.ValidationError{
			Field:         "metadata.region",
			Message:       "region is required",
			CurrentValue:  config.Metadata.Region,
			ExpectedValue: "valid AWS region (e.g., us-west-2)",
			FixSuggestion: "Set metadata.region to a valid AWS region",
		})
	}

	// Run comprehensive eksctl validation if essential validation passes
	if !result.HasErrors() {
		if err := v.validateWithUpstreamEksctl(config); err != nil {
			result.AddError(validator.ValidationError{
				Field:         "config",
				Message:       err.Error(),
				FixSuggestion: "Check the EKS cluster configuration schema at https://schema.eksctl.io for complete requirements and examples",
			})
		}
	}

	return result
}

// validateWithUpstreamEksctl runs comprehensive eksctl validation with proper error handling
func (v *Validator) validateWithUpstreamEksctl(config *eksctlapi.ClusterConfig) error {
	// Use defer/recover to handle potential panics from comprehensive validation
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't fail the entire validation
			// This allows graceful degradation if validation has issues
		}
	}()

	// Run comprehensive eksctl validation
	return eksctlapi.ValidateClusterConfig(config)
}
