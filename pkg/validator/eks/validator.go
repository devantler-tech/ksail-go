package eks

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// EKSClusterConfig represents an EKS cluster configuration.
// TODO: Replace with actual eksctl configuration type in T017.
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
// This is a placeholder implementation that will fail tests initially (TDD approach).
func (v *Validator) Validate(config *EKSClusterConfig) *validator.ValidationResult {
	// Placeholder implementation - INTENTIONALLY INCOMPLETE to make tests fail
	// This follows TDD approach where tests must fail before real implementation
	result := validator.NewValidationResult("eks.yaml")

	// TODO: Implement actual EKS validation using upstream eksctl APIs in T017
	// For now, add a placeholder error to make tests fail
	result.AddError(validator.ValidationError{
		Field:         "placeholder",
		Message:       "EKS validator not implemented yet",
		FixSuggestion: "Implement EKS validator logic using github.com/weaveworks/eksctl APIs in T017",
	})

	return result
}
