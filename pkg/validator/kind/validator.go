package kind

import (
	"github.com/devantler-tech/ksail-go/pkg/validator"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Validator validates Kind cluster configurations using upstream Kind APIs.
type Validator struct{}

// NewValidator creates a new Kind configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded Kind cluster configuration.
func (v *Validator) Validate(config *kindapi.Cluster) *validator.ValidationResult {
	result := validator.NewValidationResult("kind.yaml")

	// Handle nil config
	if config == nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       "configuration cannot be nil",
			FixSuggestion: "Provide a valid Kind cluster configuration",
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
			FixSuggestion: "Set the name field to a valid cluster name",
		})
	}

	// Validate that at least one control-plane node exists
	hasControlPlane := false
	for _, node := range config.Nodes {
		if node.Role == kindapi.ControlPlaneRole {
			hasControlPlane = true
			break
		}
	}

	if !hasControlPlane {
		result.AddError(validator.ValidationError{
			Field:         "nodes",
			Message:       "at least one control-plane node is required",
			CurrentValue:  len(config.Nodes),
			ExpectedValue: "at least one node with role: control-plane",
			FixSuggestion: "Add at least one node with role set to 'control-plane'",
		})
	}

	return result
}
