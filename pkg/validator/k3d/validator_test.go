package k3d

import (
	"testing"

	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestK3dValidatorContract tests the contract for K3d configuration validator
func TestK3dValidatorContract(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	validator := NewValidator()
	require.NotNil(t, validator, "K3d validator constructor must return non-nil validator")

	tests := []struct {
		name         string
		config       *k3dapi.SimpleConfig
		expectValid  bool
		expectErrors []string
	}{
		{
			name: "valid_k3d_config",
			config: &k3dapi.SimpleConfig{
				Servers: 1,
				Agents:  2,
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "invalid_k3d_config_no_servers",
			config: &k3dapi.SimpleConfig{
				Servers: 0,
				Agents:  2,
			},
			expectValid:  false,
			expectErrors: []string{"at least one server node is required"},
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
