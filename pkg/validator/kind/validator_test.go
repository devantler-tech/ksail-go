package kind

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestKindValidatorContract tests the contract for Kind configuration validator
func TestKindValidatorContract(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	validator := NewValidator()
	require.NotNil(t, validator, "Kind validator constructor must return non-nil validator")

	tests := []struct {
		name         string
		config       *kindapi.Cluster
		expectValid  bool
		expectErrors []string
	}{
		{
			name: "valid_kind_config",
			config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
				Name: "test-cluster",
				Nodes: []kindapi.Node{
					{
						Role: kindapi.ControlPlaneRole,
					},
					{
						Role: kindapi.WorkerRole,
					},
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "valid_kind_config_no_name",
			config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
				Nodes: []kindapi.Node{
					{
						Role: kindapi.ControlPlaneRole,
					},
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "valid_kind_config_no_nodes",
			config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
				Name: "test-cluster",
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "valid_kind_config_minimal",
			config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name:         "nil_config",
			config:       nil,
			expectValid:  false,
			expectErrors: []string{"configuration cannot be nil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.config)
			require.NotNil(t, result, "Validation result cannot be nil")

			assert.Equal(t, tt.expectValid, result.Valid, "Expected validation to pass")

			if tt.expectValid {
				assert.Empty(t, result.Errors, "Expected no validation errors")
			} else {
				// Check that expected error messages are found
				for _, expectedError := range tt.expectErrors {
					found := false
					for _, err := range result.Errors {
						if err.Message == expectedError {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message '%s' not found in validation errors", expectedError)
				}
			}
		})
	}
}
