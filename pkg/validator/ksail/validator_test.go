package ksail

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestKSailValidator tests the contract for KSail configuration validator
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
				Metadata: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionKind,
					Connection: v1alpha1.Connection{
						Context: "kind-test-cluster",
					},
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
				Metadata: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: "InvalidDistribution",
				},
			},
			expectValid:  false,
			expectErrors: []string{"spec.distribution"},
		},
		{
			name: "missing_required_name",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionKind,
				},
			},
			expectValid:  false,
			expectErrors: []string{"metadata.name"},
		},
		{
			name: "invalid_context_pattern_kind",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Metadata: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionKind,
					Connection: v1alpha1.Connection{
						Context: "wrong-context-name",
					},
				},
			},
			expectValid:  false,
			expectErrors: []string{"spec.connection.context"},
		},
		{
			name: "invalid_context_pattern_k3d",
			config: &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Metadata: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution: v1alpha1.DistributionK3d,
					Connection: v1alpha1.Connection{
						Context: "kind-test-cluster", // Wrong prefix for K3d
					},
				},
			},
			expectValid:  false,
			expectErrors: []string{"spec.connection.context"},
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
			Metadata: metav1.ObjectMeta{
				Name: "test-cluster",
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

// TestKSailValidatorContextPatterns tests context name validation patterns
func TestKSailValidatorContextPatterns(t *testing.T) {
	validator := NewValidator()

	contextTests := []struct {
		name         string
		distribution v1alpha1.Distribution
		clusterName  string
		context      string
		expectValid  bool
	}{
		{"kind_valid_context", v1alpha1.DistributionKind, "my-cluster", "kind-my-cluster", true},
		{"kind_invalid_context", v1alpha1.DistributionKind, "my-cluster", "k3d-my-cluster", false},
		{"k3d_valid_context", v1alpha1.DistributionK3d, "my-cluster", "k3d-my-cluster", true},
		{"k3d_invalid_context", v1alpha1.DistributionK3d, "my-cluster", "kind-my-cluster", false},
		{
			"eks_valid_arn",
			v1alpha1.DistributionEKS,
			"my-cluster",
			"arn:aws:eks:us-west-2:123456789012:cluster/my-cluster",
			true,
		},
		{"eks_valid_name", v1alpha1.DistributionEKS, "my-cluster", "my-cluster", true},
		{"eks_invalid_context", v1alpha1.DistributionEKS, "my-cluster", "kind-my-cluster", false},
	}

	for _, tt := range contextTests {
		t.Run(tt.name, func(t *testing.T) {
			config := &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Metadata: metav1.ObjectMeta{
					Name: tt.clusterName,
				},
				Spec: v1alpha1.Spec{
					Distribution: tt.distribution,
					Connection: v1alpha1.Connection{
						Context: tt.context,
					},
				},
			}

			result := validator.Validate(config)
			require.NotNil(t, result)

			if tt.expectValid {
				assert.True(t, result.Valid, "Context pattern should be valid for %s", tt.name)
			} else {
				assert.False(t, result.Valid, "Context pattern should be invalid for %s", tt.name)
				// Find the context error
				found := false
				for _, err := range result.Errors {
					if err.Field == "spec.connection.context" {
						found = true
						assert.NotEmpty(t, err.FixSuggestion, "Context error must have fix suggestion")
						break
					}
				}
				assert.True(t, found, "Expected context validation error not found")
			}
		})
	}
}
