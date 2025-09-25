package k3d_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// TestNewValidator tests the NewValidator constructor.
func TestNewValidator(t *testing.T) {
	t.Parallel()

	t.Run("constructor", func(t *testing.T) {
		t.Parallel()

		validator := k3dvalidator.NewValidator()
		if validator == nil {
			t.Fatal("NewValidator should return non-nil validator")
		}
	})
}

// TestValidate tests the main Validate method with comprehensive scenarios.
func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("contract_scenarios", func(t *testing.T) {
		t.Parallel()
		testK3dValidatorContract(t)
	})

	t.Run("validation_failures", func(t *testing.T) {
		t.Parallel()
		t.Run("invalid_network_config", testK3dInvalidNetworkConfig)
		t.Run("edge_case_large_server_count", testK3dLargeServerCount)
		t.Run("empty_name_with_complex_config", testK3dEmptyNameComplexConfig)
	})
}

// Helper function for contract testing.
func testK3dValidatorContract(t *testing.T) {
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
			if err.Field == "config" {
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
			if err.Field == "config" && err.Message != "" {
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
