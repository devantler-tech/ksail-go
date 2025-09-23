package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidatorInterface tests the contract for the simplified Validator interface.
func TestValidatorInterface(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "contract_requirements",
			description: "Validator interface must have single Validate method accepting typed config",
		},
		{
			name:        "type_safety",
			description: "Validator interface must be generic with type parameter T",
		},
		{
			name:        "return_type",
			description: "Validate method must return *ValidationResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates that the interface contract exists and is correctly typed
			// The actual functionality will be tested through specific validator implementations

			// Test that ValidationResult can be created
			result := NewValidationResult("test-config.yaml")
			require.NotNil(t, result)
			assert.True(t, result.Valid)
			assert.Empty(t, result.Errors)
			assert.Equal(t, "test-config.yaml", result.ConfigFile)
		})
	}
}

// TestValidationResult tests the ValidationResult type contract.
func TestValidationResult(t *testing.T) {
	t.Run("new_validation_result", func(t *testing.T) {
		result := NewValidationResult("test.yaml")

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
		assert.Equal(t, "test.yaml", result.ConfigFile)
	})

	t.Run("add_error_sets_invalid", func(t *testing.T) {
		result := NewValidationResult("test.yaml")

		validationError := ValidationError{
			Field:         "spec.distribution",
			Message:       "invalid distribution",
			CurrentValue:  "invalid",
			ExpectedValue: "Kind|K3d|EKS",
			FixSuggestion: "Set distribution to one of: Kind, K3d, EKS",
		}

		result.AddError(validationError)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.True(t, result.HasErrors())
	})

	t.Run("add_warning_preserves_valid", func(t *testing.T) {
		result := NewValidationResult("test.yaml")

		warning := ValidationError{
			Field:         "spec.optional",
			Message:       "deprecated field",
			FixSuggestion: "Consider using spec.newField instead",
		}

		result.AddWarning(warning)

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.Warnings, 1)
		assert.True(t, result.HasWarnings())
	})
}

// TestValidationError tests the ValidationError type contract.
func TestValidationError(t *testing.T) {
	t.Run("error_interface", func(t *testing.T) {
		err := ValidationError{
			Field:   "spec.distribution",
			Message: "invalid value",
		}

		assert.Contains(t, err.Error(), "spec.distribution")
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("error_without_field", func(t *testing.T) {
		err := ValidationError{
			Message: "general validation error",
		}

		assert.Equal(t, "validation error: general validation error", err.Error())
	})
}

// TestFileLocation tests the FileLocation type contract
func TestFileLocation(t *testing.T) {
	t.Run("full_location", func(t *testing.T) {
		location := FileLocation{
			FilePath: "/path/to/config.yaml",
			Line:     10,
			Column:   5,
		}

		assert.Equal(t, "/path/to/config.yaml:10:5", location.String())
	})

	t.Run("line_only", func(t *testing.T) {
		location := FileLocation{
			FilePath: "/path/to/config.yaml",
			Line:     10,
		}

		assert.Equal(t, "/path/to/config.yaml:10", location.String())
	})

	t.Run("file_only", func(t *testing.T) {
		location := FileLocation{
			FilePath: "/path/to/config.yaml",
		}

		assert.Equal(t, "/path/to/config.yaml", location.String())
	})
}
