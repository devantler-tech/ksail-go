package k3d

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// Validator validates K3d cluster configurations using upstream K3d APIs.
type Validator struct{}

// NewValidator creates a new K3d configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded K3d cluster configuration.
func (v *Validator) Validate(config *k3dapi.SimpleConfig) *validator.ValidationResult {
	result := validator.NewValidationResult("k3d.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration cannot be nil",
			FixSuggestion: "Provide a valid K3d cluster configuration",
		})
		return result
	}

	// Validate that at least one server node is required
	if config.Servers <= 0 {
		result.AddError(validator.ValidationError{
			Field:         "servers",
			Message:       "at least one server node is required",
			CurrentValue:  config.Servers,
			ExpectedValue: "integer >= 1",
			FixSuggestion: "Set servers to at least 1 for a functional K3d cluster",
		})
	}

	return result
}
