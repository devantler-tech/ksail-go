// Package eks provides EKS configuration validation functionality.
package eks

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/jinzhu/copier"
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

	// Validate required fields
	v.validateTypeMetaFields(config, result)
	v.validateMetadataFields(config, result)

	// Run comprehensive eksctl validation if essential validation passes and it's safe to do so
	if !result.HasErrors() {
		err := v.validateWithUpstreamEksctl(config)
		if err != nil {
			result.AddError(validator.ValidationError{
				Field:   "config",
				Message: err.Error(),
				FixSuggestion: "Check the EKS cluster configuration schema at " +
					"https://schema.eksctl.io for complete requirements and examples",
			})
		}
	}

	return result
}

// validateWithUpstreamEksctl attempts to run comprehensive eksctl validation.
// We apply eksctl defaults before validation to prevent panics from missing required fields.
func (v *Validator) validateWithUpstreamEksctl(config *eksctlapi.ClusterConfig) error {
	// Create a deep copy to avoid modifying the original config using marshalling/unmarshalling
	configCopy, err := v.deepCopyConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create deep copy of config: %w", err)
	}

	// Apply eksctl defaults to prevent panics from missing required fields
	// This is what eksctl CLI does before validation
	eksctlapi.SetClusterConfigDefaults(configCopy)

	// Run comprehensive eksctl validation
	// Note: This returns a single error, but that error may contain multiple validation failures
	// wrapped together using Go's error wrapping patterns (errors.Join, fmt.Errorf with %w, etc.)
	validationErr := eksctlapi.ValidateClusterConfig(configCopy)
	if validationErr != nil {
		return fmt.Errorf("eksctl validation failed: %w", validationErr)
	}

	return nil
}

// deepCopyConfig creates a deep copy of the EKS cluster configuration using the copier library.
// This ensures that upstream validation operations cannot modify the original configuration object.
// Using copier is more efficient than JSON marshalling/unmarshalling for frequently called validation.
func (v *Validator) deepCopyConfig(
	config *eksctlapi.ClusterConfig,
) (*eksctlapi.ClusterConfig, error) {
	var configCopy eksctlapi.ClusterConfig

	err := copier.Copy(&configCopy, config)
	if err != nil {
		return nil, fmt.Errorf("failed to deep copy config: %w", err)
	}

	return &configCopy, nil
}

// validateTypeMetaFields validates required TypeMeta fields.
func (v *Validator) validateTypeMetaFields(
	config *eksctlapi.ClusterConfig,
	result *validator.ValidationResult,
) {
	if config.Kind == "" {
		result.AddError(validator.ValidationError{
			Field:         "kind",
			Message:       "kind is required",
			ExpectedValue: "ClusterConfig",
			FixSuggestion: "Set kind to 'ClusterConfig'",
		})
	}

	if config.APIVersion == "" {
		result.AddError(validator.ValidationError{
			Field:         "apiVersion",
			Message:       "apiVersion is required",
			ExpectedValue: "eksctl.io/v1alpha5",
			FixSuggestion: "Set apiVersion to 'eksctl.io/v1alpha5'",
		})
	}
}

// validateMetadataFields validates required metadata fields.
func (v *Validator) validateMetadataFields(
	config *eksctlapi.ClusterConfig,
	result *validator.ValidationResult,
) {
	// Validate metadata is present (required field per eksctl API)
	if config.Metadata == nil {
		result.AddError(validator.ValidationError{
			Field:         "metadata",
			Message:       "metadata is required",
			CurrentValue:  nil,
			ExpectedValue: "ClusterMeta object with name and region",
			FixSuggestion: "Add metadata section with name and region fields",
		})

		return
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
}
