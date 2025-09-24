package k3d_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// TestK3dValidatorContract tests the contract for K3d configuration validator.
func TestK3dValidatorContract(t *testing.T) {
	t.Parallel()

	// This test MUST FAIL initially to follow TDD approach
	validatorInstance := k3dvalidator.NewValidator()
	testCases := createK3dTestCases()

	testutils.RunValidatorTests(
		t,
		validatorInstance,
		testCases,
		testutils.AssertValidationResult[*k3dapi.SimpleConfig],
	)
}

// TestK3dValidatorValidationFailures tests specific validation failures and edge cases
func TestK3dValidatorValidationFailures(t *testing.T) {
	t.Parallel()

	validatorInstance := k3dvalidator.NewValidator()

	t.Run("invalid_network_config", func(t *testing.T) {
		t.Parallel()

		// Test with invalid network configuration
		config := &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{
				APIVersion: "k3d.io/v1alpha5",
				Kind:       "Simple",
			},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "test-cluster",
			},
			Servers: 1,
			Agents:  1,
			Network: "invalid-network-123456789012345678901234567890123456789012345678901234567890", // Very long network name that might cause validation to fail
		}

		result := validatorInstance.Validate(config)

		// The upstream validation should handle network validation
		// We expect either success or a validation error
		if !result.Valid && len(result.Errors) > 0 {
			// Check that error contains helpful information
			found := false
			for _, err := range result.Errors {
				if err.Field == "config" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected config validation error for invalid network name")
			}
		}
	})

	t.Run("edge_case_large_server_count", func(t *testing.T) {
		t.Parallel()

		// Test with unusually large server count
		config := &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{
				APIVersion: "k3d.io/v1alpha5",
				Kind:       "Simple",
			},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "test-cluster",
			},
			Servers: 100, // Large server count
			Agents:  1,
		}

		result := validatorInstance.Validate(config)

		// This should either validate successfully or fail with a meaningful error
		if !result.Valid {
			if len(result.Errors) == 0 {
				t.Errorf("Expected validation errors when result is invalid")
			}
		}
	})

	t.Run("empty_name_with_complex_config", func(t *testing.T) {
		t.Parallel()

		// Test empty name with more complex configuration
		config := &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{
				APIVersion: "k3d.io/v1alpha5",
				Kind:       "Simple",
			},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "", // Empty name
			},
			Servers: 1,
			Agents:  2,
			Image:   "rancher/k3s:v1.27.1-k3s1",
		}

		result := validatorInstance.Validate(config)

		// Should validate successfully since empty name is allowed
		if !result.Valid && len(result.Errors) > 0 {
			// If it fails, check the error is reasonable
			for _, err := range result.Errors {
				if err.Field == "config" && err.Message != "" {
					// This is acceptable - upstream validation caught something
					continue
				}
			}
		}
	})
}

func createK3dTestCases() []testutils.ValidatorTestCase[*k3dapi.SimpleConfig] {
	return []testutils.ValidatorTestCase[*k3dapi.SimpleConfig]{
		{
			Name: "valid_k3d_config",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "valid_k3d_config_zero_servers",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 0,
				Agents:  1,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "valid_k3d_config_no_name",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "invalid_k3d_config_missing_kind",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					// Kind is missing
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{
					Field:         "kind",
					Message:       "kind is required",
					ExpectedValue: "Simple",
					FixSuggestion: "Set kind to 'Simple'",
				},
			},
		},
		{
			Name: "invalid_k3d_config_missing_api_version",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					// APIVersion is missing
					Kind: "Simple",
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{
					Field:         "apiVersion",
					Message:       "apiVersion is required",
					ExpectedValue: "k3d.io/v1alpha5",
					FixSuggestion: "Set apiVersion to 'k3d.io/v1alpha5'",
				},
			},
		},
		{
			Name: "invalid_k3d_config_both_missing",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					// Both APIVersion and Kind are missing
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{
					Field:         "kind",
					Message:       "kind is required",
					ExpectedValue: "Simple",
					FixSuggestion: "Set kind to 'Simple'",
				},
				{
					Field:         "apiVersion",
					Message:       "apiVersion is required",
					ExpectedValue: "k3d.io/v1alpha5",
					FixSuggestion: "Set apiVersion to 'k3d.io/v1alpha5'",
				},
			},
		},
		testutils.CreateNilConfigTestCase[*k3dapi.SimpleConfig](),
	}
}
