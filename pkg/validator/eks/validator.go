package eks

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// EKSClusterConfig represents an EKS cluster configuration.
// TODO: Replace with actual eksctl configuration type.
type EKSClusterConfig struct {
	Name   string
	Region string
}

// Validator validates EKS cluster configurations using upstream eksctl APIs.
type Validator struct{}

// NewValidator creates a new EKS configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded EKS cluster configuration.
func (v *Validator) Validate(config *EKSClusterConfig) *validator.ValidationResult {
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

	// Validate cluster name is required
	if config.Name == "" {
		result.AddError(validator.ValidationError{
			Field:         "name",
			Message:       "cluster name is required",
			CurrentValue:  config.Name,
			ExpectedValue: "non-empty string",
			FixSuggestion: "Set the name field to a valid EKS cluster name",
		})
	}

	// Validate region is required
	if config.Region == "" {
		result.AddError(validator.ValidationError{
			Field:         "region",
			Message:       "region is required",
			CurrentValue:  config.Region,
			ExpectedValue: "valid AWS region (e.g., us-west-2)",
			FixSuggestion: "Set the region field to a valid AWS region",
		})
	}

	return result
}
