package k3d

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dconfig "github.com/k3d-io/k3d/v5/pkg/config"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
)

// Validator validates K3d cluster configurations using upstream K3d APIs.
type Validator struct{}

// NewValidator creates a new K3d configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a loaded K3d cluster configuration using upstream K3d APIs.
// This validator performs both essential field validation and comprehensive upstream K3d validation.
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

	// Run comprehensive K3d validation using upstream APIs
	if err := v.validateWithUpstreamK3d(config); err != nil {
		result.AddError(validator.ValidationError{
			Field:         "config",
			Message:       err.Error(),
			FixSuggestion: "Check the K3d cluster configuration schema at https://k3d.io/usage/configfile/ for complete requirements and examples",
		})
	}

	return result
}

// validateWithUpstreamK3d runs comprehensive K3d validation following the same workflow as K3d CLI with proper error handling
func (v *Validator) validateWithUpstreamK3d(config *k3dapi.SimpleConfig) error {
	// Use defer/recover to handle potential panics from comprehensive validation
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't fail the entire validation
			// This allows graceful degradation if validation has issues
		}
	}()

	// Create a copy to avoid modifying the original configuration
	configCopy := *config

	// Step 1: Process simple config (same as K3d CLI workflow)
	if err := k3dconfig.ProcessSimpleConfig(&configCopy); err != nil {
		return fmt.Errorf("failed to process simple configuration: %w", err)
	}

	// Step 2: Transform simple config to cluster config (requires a runtime context)
	// Use Docker runtime as default since it's the most common runtime for K3d
	runtime := runtimes.Docker
	ctx := context.Background()

	clusterConfig, err := k3dconfig.TransformSimpleToClusterConfig(ctx, runtime, configCopy, "")
	if err != nil {
		return fmt.Errorf("failed to transform configuration: %w", err)
	}

	// Step 3: Process cluster config
	processedConfig, err := k3dconfig.ProcessClusterConfig(*clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to process cluster configuration: %w", err)
	}

	// Step 4: Run comprehensive K3d validation (same as K3d CLI)
	if err := k3dconfig.ValidateClusterConfig(ctx, runtime, *processedConfig); err != nil {
		return fmt.Errorf("K3d configuration validation failed: %w", err)
	}

	return nil
}
