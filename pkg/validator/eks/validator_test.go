package eks_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	eksvalidator "github.com/devantler-tech/ksail-go/pkg/validator/eks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// TestEKSValidatorContract tests the contract for EKS configuration validator.
func TestEKSValidatorContract(t *testing.T) {
	t.Parallel()

	// This test MUST FAIL initially to follow TDD approach
	validator := eksvalidator.NewValidator()
	require.NotNil(t, validator, "EKS validator constructor must return non-nil validator")

	testCases := createEKSTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validator.Validate(testCase.config)
			require.NotNil(t, result, "Validation result cannot be nil")

			assertEKSValidationResult(t, testCase, result)
		})
	}
}

func createEKSTestCases() []struct {
	name         string
	config       *eksctlapi.ClusterConfig
	expectValid  bool
	expectErrors []string
} {
	return []struct {
		name         string
		config       *eksctlapi.ClusterConfig
		expectValid  bool
		expectErrors []string
	}{
		{
			name: "valid_eks_config",
			config: &eksctlapi.ClusterConfig{
				Metadata: &eksctlapi.ClusterMeta{
					Name:   "test-cluster",
					Region: "us-west-2",
				},
			},
			expectValid:  true,
			expectErrors: []string{},
		},
		{
			name: "invalid_eks_config_missing_name",
			config: &eksctlapi.ClusterConfig{
				Metadata: &eksctlapi.ClusterMeta{
					Region: "us-west-2",
				},
			},
			expectValid:  false,
			expectErrors: []string{"cluster name is required"},
		},
		{
			name: "invalid_eks_config_missing_region",
			config: &eksctlapi.ClusterConfig{
				Metadata: &eksctlapi.ClusterMeta{
					Name: "test-cluster",
				},
			},
			expectValid:  false,
			expectErrors: []string{"region is required"},
		},
		{
			name:   "invalid_eks_config_missing_metadata",
			config: &eksctlapi.ClusterConfig{
				// No metadata
			},
			expectValid:  false,
			expectErrors: []string{"metadata is required"},
		},
		{
			name:         "nil_config",
			config:       nil,
			expectValid:  false,
			expectErrors: []string{"configuration cannot be nil"},
		},
	}
}

func assertEKSValidationResult(t *testing.T, testCase struct {
	name         string
	config       *eksctlapi.ClusterConfig
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
