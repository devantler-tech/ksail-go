package stubs

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
)

// ValidatorStub is a stub implementation of validator.Validator[T] interface.
// It provides configurable behavior for testing without external dependencies.
type ValidatorStub[T any] struct {
	ValidateResult *validator.ValidationResult
	ValidateError  error
}

// NewValidatorStub creates a new ValidatorStub with default success behavior.
func NewValidatorStub[T any]() *ValidatorStub[T] {
	return &ValidatorStub[T]{
		ValidateResult: validator.NewValidationResult("test-config.yaml"),
	}
}

// Validate returns the configured validation result.
func (v *ValidatorStub[T]) Validate(config T) *validator.ValidationResult {
	if v.ValidateResult != nil {
		return v.ValidateResult
	}
	return validator.NewValidationResult("test-config.yaml")
}

// WithValidationError configures the stub to return a validation error.
func (v *ValidatorStub[T]) WithValidationError(field, message string) *ValidatorStub[T] {
	result := validator.NewValidationResult("test-config.yaml")
	err := validator.NewValidationError(
		field,
		message,
		nil, // currentValue
		nil, // expectedValue
		"Fix the configuration",
		validator.NewFileLocation("test-config.yaml", 1, 1),
	)
	result.AddError(err)
	v.ValidateResult = result
	return v
}

// WithValidationSuccess configures the stub to return successful validation.
func (v *ValidatorStub[T]) WithValidationSuccess() *ValidatorStub[T] {
	v.ValidateResult = validator.NewValidationResult("test-config.yaml")
	return v
}