package ksail

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestKSailValidator tests the contract for KSail configuration validator.
func TestKSailValidator(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	validator := NewValidator()
	require.NotNil(t, validator, "KSail validator constructor must return non-nil validator")

	tests := []struct {
		name           string
		config         *v1alpha1.Cluster
		expectValid    bool
		expectErrors   []string
		expectWarnings []string
	}{
		{
			name: "valid_kind_configuration",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionKind,
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "invalid_distribution",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: "InvalidDistribution",
				},
			},
			expectValid:  false,
			expectErrors: []string{"spec.distribution"},
		},
		{
			name: "valid_config_without_metadata",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionKind,
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name:         "nil_config",
			config:       nil,
			expectValid:  false,
			expectErrors: []string{"config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.config)
			require.NotNil(t, result, "Validate must return non-nil result")

			assert.Equal(t, tt.expectValid, result.Valid, "Validation result Valid status mismatch")

			if tt.expectValid {
				assert.Empty(t, result.Errors, "Valid configuration should have no errors")
			} else {
				assert.NotEmpty(t, result.Errors, "Invalid configuration should have errors")

				// Check that expected error fields are present
				for _, expectedField := range tt.expectErrors {
					found := false
					for _, err := range result.Errors {
						if err.Field == expectedField {
							found = true
							// Verify error has actionable content
							assert.NotEmpty(t, err.Message, "Error message must not be empty")
							assert.NotEmpty(t, err.FixSuggestion, "Error must have fix suggestion")

							break
						}
					}
					assert.True(t, found, "Expected error for field %s not found", expectedField)
				}
			}

			// Verify warnings if expected
			if len(tt.expectWarnings) > 0 {
				assert.Len(t, result.Warnings, len(tt.expectWarnings), "Warning count mismatch")
			}
		})
	}
}

// TestKSailValidatorCrossConfiguration tests cross-configuration validation
func TestKSailValidatorCrossConfiguration(t *testing.T) {
	validator := NewValidator()

	t.Run("cross_config_validation_placeholder", func(t *testing.T) {
		// This test will validate cross-configuration consistency
		// For now, it's a placeholder that will be implemented
		// when the actual cross-configuration logic is added

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution: v1alpha1.DistributionKind,
			},
		}

		result := validator.Validate(config)
		require.NotNil(t, result)

		// This test will fail initially because the validator doesn't exist yet
		// Once implemented, it should validate cross-configuration consistency
		t.Log("Cross-configuration validation test placeholder - implement in T034")
	})
}
