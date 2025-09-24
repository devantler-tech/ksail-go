package k3d_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestK3dValidatorContract tests the contract for K3d configuration validator.
func TestK3dValidatorContract(t *testing.T) {
	t.Parallel()

	// This test MUST FAIL initially to follow TDD approach
	validator := k3dvalidator.NewValidator()
	require.NotNil(t, validator, "K3d validator constructor must return non-nil validator")

	testCases := createK3dTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validator.Validate(testCase.config)
			require.NotNil(t, result, "Validation result cannot be nil")

			assertK3dValidationResult(t, testCase, result)
		})
	}
}

func createK3dTestCases() []struct {
	name         string
	config       *k3dapi.SimpleConfig
	expectValid  bool
	expectErrors []string
} {
	return []struct {
		name         string
		config       *k3dapi.SimpleConfig
		expectValid  bool
		expectErrors []string
	}{
		{
			name: "valid_k3d_config",
			config: &k3dapi.SimpleConfig{
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "valid_k3d_config_zero_servers",
			config: &k3dapi.SimpleConfig{
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster-zero",
				},
				Servers: 0,
				Agents:  2,
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "valid_k3d_config_no_name",
			config: &k3dapi.SimpleConfig{
				Servers: 1,
				Agents:  2,
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
}

func assertK3dValidationResult(t *testing.T, testCase struct {
	name         string
	config       *k3dapi.SimpleConfig
	expectValid  bool
	expectErrors []string
}, result *validator.ValidationResult,
) {
	t.Helper()

	assert.Equal(t, testCase.expectValid, result.Valid, "Expected validation to pass")

	if testCase.expectValid {
		assert.Empty(t, result.Errors, "Expected no validation errors")
	} else {
		// Check that expected error messages are found
		for _, expectedError := range testCase.expectErrors {
			found := false

			for _, resultErr := range result.Errors {
				if resultErr.Message == expectedError {
					found = true

					break
				}
			}

			assert.True(t, found, "Expected error message '%s' not found in validation errors", expectedError)
		}
	}
}
