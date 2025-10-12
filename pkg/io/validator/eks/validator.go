// Package eks provides EKS configuration validation functionality.
package eks

import (
	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	"github.com/devantler-tech/ksail-go/pkg/io/validator/metadata"
	ekstypes "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// Validator validates EKS cluster configurations using upstream eksctl APIs.
type Validator struct{}

// NewValidator creates a new EKS configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded EKS cluster configuration.
func (v *Validator) Validate(config *ekstypes.ClusterConfig) *validator.ValidationResult {
	result := validator.NewValidationResult("eks.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration is nil",
			FixSuggestion: "Provide a valid EKS cluster configuration",
		})

		return result
	}

	// Validate required metadata fields
	metadata.ValidateMetadata(
		config.Kind,
		config.APIVersion,
		"ClusterConfig",
		"eksctl.io/v1alpha5",
		result,
	)

	return result
}
