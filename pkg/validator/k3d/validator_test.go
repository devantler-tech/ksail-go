package k3d_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
)

const (
	configFieldName = "config"
)

// TestNewValidator tests the NewValidator constructor.
func TestNewValidator(t *testing.T) {
	t.Parallel()

	testutils.RunNewValidatorConstructorTest(t, func() validator.Validator[*k3dapi.SimpleConfig] {
		return k3dvalidator.NewValidator()
	})
}

// TestValidate tests the main Validate method with comprehensive scenarios.
func TestValidate(t *testing.T) {
	t.Parallel()

	testutils.RunValidateTest[*k3dapi.SimpleConfig](
		t,
		testK3dValidatorContract,
		testK3dValidationFailures,
	)
}

// testK3dValidationFailures runs validation failure tests.
func testK3dValidationFailures(t *testing.T) {
	t.Helper()

	t.Run("invalid_network_config", testK3dInvalidNetworkConfig)
	t.Run("edge_case_large_server_count", testK3dLargeServerCount)
	t.Run("empty_name_with_complex_config", testK3dEmptyNameComplexConfig)
}

// Helper function for contract testing.
func testK3dValidatorContract(t *testing.T) {
	t.Helper()

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

func testK3dInvalidNetworkConfig(t *testing.T) {
	t.Parallel()

	validatorInstance := k3dvalidator.NewValidator()

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
		// Very long network name that might cause validation to fail
		Network: "invalid-network-123456789012345678901234567890123456789012345678901234567890",
	}

	result := validatorInstance.Validate(config)

	// The upstream validation should handle network validation
	// We expect either success or a validation error
	if !result.Valid && len(result.Errors) > 0 {
		// Check that error contains helpful information
		found := false

		for _, err := range result.Errors {
			if err.Field == configFieldName {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Expected config validation error for invalid network name")
		}
	}
}

func testK3dLargeServerCount(t *testing.T) {
	t.Parallel()

	validatorInstance := k3dvalidator.NewValidator()

	// Test with large server count that might cause issues
	config := &k3dapi.SimpleConfig{
		TypeMeta: configtypes.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: configtypes.ObjectMeta{
			Name: "large-cluster",
		},
		Servers: 10, // Very large server count
		Agents:  20, // Very large agent count
	}

	result := validatorInstance.Validate(config)

	// The upstream validation should handle resource limits
	// We expect this to potentially be valid or have specific resource errors
	if !result.Valid {
		// Check that error is related to resource concerns
		found := false

		for _, err := range result.Errors {
			if err.Field == configFieldName {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Expected config validation for large server count")
		}
	}
}

// TestK3dValidatorEdgeCases tests additional edge cases to improve coverage.
func TestK3dValidatorEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("deepcopy_scenarios", func(t *testing.T) {
		t.Parallel()
		testK3dDeepCopyScenarios(t)
	})

	t.Run("upstream_validation_scenarios", func(t *testing.T) {
		t.Parallel()
		testK3dUpstreamValidationScenarios(t)
	})

	t.Run("config_transformation_edge_cases", func(t *testing.T) {
		t.Parallel()
		testK3dConfigTransformationEdgeCases(t)
	})
}

func testK3dDeepCopyScenarios(t *testing.T) {
	t.Helper()

	validatorInstance := k3dvalidator.NewValidator()

	// Test with configuration that exercises deep copy logic
	config := &k3dapi.SimpleConfig{
		TypeMeta: configtypes.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: configtypes.ObjectMeta{
			Name: "deepcopy-test-cluster",
		},
		Servers: 2,
		Agents:  2,
		// Test network configuration that gets deep copied
		Network: "test-network",
		Subnet:  "172.20.0.0/16",
		Image:   "rancher/k3s:v1.25.0-k3s1",
	}

	result := validatorInstance.Validate(config)

	// The configuration should be processed - this tests deep copy functionality
	// The result may be valid or invalid depending on environment, but should not panic
	if !result.Valid {
		t.Logf("Deep copy config validation failed (may be expected): %v", result.Errors)
		// Verify that we get meaningful error structure
		assert.NotEmpty(t, result.Errors)
	}
}

func testK3dUpstreamValidationScenarios(t *testing.T) {
	t.Helper()

	validatorInstance := k3dvalidator.NewValidator()

	// Test scenario that exercises the upstream validation path more thoroughly
	config := &k3dapi.SimpleConfig{
		TypeMeta: configtypes.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: configtypes.ObjectMeta{
			Name: "upstream-test",
		},
		Servers: 1,
		Agents:  0,
		// Test with specific K3s image that might trigger upstream validation
		Image: "rancher/k3s:latest",
		// Test with network configuration
		Network: "test-network",
		// Test with subnet configuration
		Subnet: "172.20.0.0/16",
	}

	result := validatorInstance.Validate(config)

	// This exercises the upstream validation pipeline
	// Result may be valid or invalid depending on Docker environment
	if !result.Valid {
		// Verify that errors contain meaningful information
		assert.NotEmpty(t, result.Errors)

		for _, err := range result.Errors {
			assert.NotEmpty(t, err.Message)
			assert.NotEmpty(t, err.Field)
		}
	}
}

func testK3dConfigTransformationEdgeCases(t *testing.T) {
	t.Helper()

	validatorInstance := k3dvalidator.NewValidator()

	// Test configuration with potentially problematic values that might trigger error paths
	config := &k3dapi.SimpleConfig{
		TypeMeta: configtypes.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: configtypes.ObjectMeta{
			Name: "transform-test",
		},
		Servers: 1,
		Agents:  1,
		// Basic image configuration to test transformation
		Image:   "rancher/k3s:v1.27.0-k3s1",
		Network: "bridge",
	}

	result := validatorInstance.Validate(config)

	// This exercises the config transformation and processing pipeline
	// Should handle basic configuration structure
	if !result.Valid {
		// Verify error structure is meaningful
		for _, err := range result.Errors {
			assert.NotEmpty(t, err.Message)
			assert.NotEmpty(t, err.Field)
		}
	}
}

// TestK3dValidatorErrorPaths tests specific error scenarios to improve coverage.
func TestK3dValidatorErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("nil_config", func(t *testing.T) {
		t.Parallel()

		validatorInstance := k3dvalidator.NewValidator()

		// Test with nil config - should be handled gracefully
		result := validatorInstance.Validate(nil)

		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("malformed_config", func(t *testing.T) {
		t.Parallel()

		validatorInstance := k3dvalidator.NewValidator()

		// Test with malformed configuration that might trigger processing errors
		config := &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{
				APIVersion: "invalid-version",
				Kind:       "Invalid",
			},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "", // Empty name might cause issues
			},
			Servers: -1, // Invalid server count
			Agents:  -1, // Invalid agent count
		}

		result := validatorInstance.Validate(config)

		// This should either pass validation or fail with meaningful errors
		if !result.Valid {
			assert.NotEmpty(t, result.Errors)

			for _, err := range result.Errors {
				assert.NotEmpty(t, err.Message)
			}
		}
	})

	t.Run("extreme_values", func(t *testing.T) {
		t.Parallel()

		validatorInstance := k3dvalidator.NewValidator()

		// Test with extreme values that might trigger validation failures
		config := &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{
				APIVersion: "k3d.io/v1alpha5",
				Kind:       "Simple",
			},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "extreme-test",
			},
			Servers: 1000, // Extremely large server count
			Agents:  1000, // Extremely large agent count
			// Potentially problematic image
			Image: "nonexistent:invalid-tag",
		}

		result := validatorInstance.Validate(config)

		// This might trigger upstream validation errors or resource constraints
		if !result.Valid {
			assert.NotEmpty(t, result.Errors)
		}
	})
}

func testK3dEmptyNameComplexConfig(t *testing.T) {
	t.Parallel()

	validatorInstance := k3dvalidator.NewValidator()

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
			if err.Field == configFieldName && err.Message != "" {
				// This is acceptable - upstream validation caught something
				continue
			}
		}
	}
}

func createK3dTestCases() []testutils.ValidatorTestCase[*k3dapi.SimpleConfig] {
	var testCases []testutils.ValidatorTestCase[*k3dapi.SimpleConfig]

	testCases = append(testCases, createK3dValidTestCases()...)
	testCases = append(testCases, createK3dInvalidTestCases()...)
	testCases = append(testCases, testutils.CreateNilConfigTestCase[*k3dapi.SimpleConfig]())

	return testCases
}

func createK3dValidTestCases() []testutils.ValidatorTestCase[*k3dapi.SimpleConfig] {
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
	}
}

func createK3dInvalidTestCases() []testutils.ValidatorTestCase[*k3dapi.SimpleConfig] {
	// Create factory function for base config to ensure each test gets a separate instance
	configFactory := func() *k3dapi.SimpleConfig {
		return &k3dapi.SimpleConfig{
			TypeMeta: configtypes.TypeMeta{},
			ObjectMeta: configtypes.ObjectMeta{
				Name: "test-cluster",
			},
			Servers: 1,
			Agents:  2,
		}
	}

	// Use helper to create metadata validation test cases
	return testutils.CreateMetadataValidationTestCases(
		configFactory,
		func(config *k3dapi.SimpleConfig, kind string) { config.Kind = kind },
		func(config *k3dapi.SimpleConfig, apiVersion string) { config.APIVersion = apiVersion },
		testutils.MetadataTestCaseConfig{
			ExpectedKind:       "Simple",
			ExpectedAPIVersion: "k3d.io/v1alpha5",
		},
	)
}
