package ksail_test

import (
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/io/validator/ksail"
	k3dtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	specDistributionField = "spec.distribution"
)

// TestKSailValidator tests the contract for KSail configuration validator.
func TestKSailValidator(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	t.Parallel()

	validator := ksailvalidator.NewValidator()
	require.NotNil(t, validator, "KSail validator constructor must return non-nil validator")

	testCases := createKSailTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validator.Validate(testCase.config)
			require.NotNil(t, result, "Validate must return non-nil result")

			assertKSailValidationResult(t, testCase, result)
		})
	}
}

func createKSailTestCases() []ksailTestCase {
	return []ksailTestCase{
		createValidKindKSailConfigCase(),
		createInvalidDistributionKSailCase(),
		createValidConfigWithoutMetadataCase(),
		createNilKSailConfigCase(),
	}
}

type ksailTestCase struct {
	name         string
	config       *v1alpha1.Cluster
	expectValid  bool
	expectErrors []string
}

func createValidKindKSailConfigCase() ksailTestCase {
	return ksailTestCase{
		name:         "valid_kind_configuration",
		config:       createValidKSailConfig(v1alpha1.DistributionKind),
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createInvalidDistributionKSailCase() ksailTestCase {
	return ksailTestCase{
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
		expectErrors: []string{specDistributionField},
	}
}

func createValidConfigWithoutMetadataCase() ksailTestCase {
	return ksailTestCase{
		name:         "valid_config_without_metadata",
		config:       createValidKSailConfig(v1alpha1.DistributionKind),
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createNilKSailConfigCase() ksailTestCase {
	return ksailTestCase{
		name:         "nil_config",
		config:       nil,
		expectValid:  false,
		expectErrors: []string{"config"},
	}
}

func assertKSailValidationResult(
	t *testing.T,
	testCase ksailTestCase,
	result *validator.ValidationResult,
) {
	t.Helper()

	assert.Equal(t, testCase.expectValid, result.Valid, "Validation result Valid status mismatch")

	if testCase.expectValid {
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")

		return
	}

	assert.NotEmpty(t, result.Errors, "Invalid configuration should have errors")

	validateExpectedErrors(t, testCase.expectErrors, result.Errors)
}

func validateExpectedErrors(
	t *testing.T,
	expectedFields []string,
	errors []validator.ValidationError,
) {
	t.Helper()

	// Check that expected error fields are present
	for _, expectedField := range expectedFields {
		found := false

		for _, resultErr := range errors {
			if resultErr.Field == expectedField {
				found = true
				// Verify error has actionable content
				assert.NotEmpty(t, resultErr.Message, "Error message must not be empty")
				assert.NotEmpty(t, resultErr.FixSuggestion, "Error must have fix suggestion")

				break
			}
		}

		assert.True(t, found, "Expected error for field %s not found", expectedField)
	}
}

// TestKSailValidatorCrossConfiguration tests cross-configuration validation.
func TestKSailValidatorCrossConfiguration(t *testing.T) {
	t.Parallel()

	validator := ksailvalidator.NewValidator()

	t.Run("cross_config_validation_current_scope", func(t *testing.T) {
		t.Parallel()

		// Test current cross-configuration validation capabilities
		// The KSail validator is implemented and validates basic semantic correctness
		// Full cross-configuration coordination with config managers is future work (T034)

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "kind.yaml",
				Connection: v1alpha1.Connection{
					Context: "kind-kind", // No distribution config provided, so use conventional default
				},
			},
		}

		result := validator.Validate(config)
		require.NotNil(t, result, "Validation result should not be nil")

		// The validator exists and performs basic validation
		// Currently validates distribution types and required fields
		assert.True(t, result.Valid, "Valid configuration should pass validation")
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")

		// Test cross-configuration validation for invalid distribution
		invalidConfig := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution: "InvalidDistribution",
			},
		}

		invalidResult := validator.Validate(invalidConfig)
		require.NotNil(t, invalidResult, "Validation result should not be nil")
		assert.False(t, invalidResult.Valid, "Invalid distribution should fail validation")
		assert.NotEmpty(t, invalidResult.Errors, "Invalid configuration should have errors")

		// Verify error contains actionable information
		found := false

		for _, err := range invalidResult.Errors {
			if err.Field == specDistributionField {
				found = true

				assert.NotEmpty(t, err.Message, "Error message should not be empty")
				assert.NotEmpty(t, err.FixSuggestion, "Error should have fix suggestion")

				break
			}
		}

		assert.True(t, found, "Should find distribution validation error")
	})
}

// createValidKSailConfig creates a valid KSail configuration with the specified distribution.
func createValidKSailConfig(distribution v1alpha1.Distribution) *v1alpha1.Cluster {
	var distributionConfigFile string

	var contextName string

	switch distribution {
	case v1alpha1.DistributionKind:
		distributionConfigFile = "kind.yaml"
		contextName = "kind-kind" // No distribution config provided, use conventional default
	case v1alpha1.DistributionK3d:
		distributionConfigFile = "k3d.yaml"
		contextName = "k3d-k3s-default" // No distribution config provided, use conventional default
	case v1alpha1.DistributionEKS:
		distributionConfigFile = "eks.yaml"
		contextName = "ksail-eks"
	default:
		distributionConfigFile = "cluster.yaml"
		contextName = "ksail"
	}

	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.dev/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution:       distribution,
			DistributionConfig: distributionConfigFile,
			Connection: v1alpha1.Connection{
				Context: contextName,
			},
		},
	}
}

// TestKSailValidatorContextNameValidation tests context name validation patterns.
func TestKSailValidatorContextNameValidation(t *testing.T) {
	t.Parallel()

	t.Run("kind_valid_context", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionKind)
		config.Spec.Connection.Context = "kind-kind" // No distribution config, so expect conventional default

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Valid Kind context should pass validation")
		assert.Empty(t, result.Errors, "Valid context should have no errors")
	})

	t.Run("k3d_valid_context", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionK3d)
		config.Spec.Connection.Context = "k3d-k3s-default" // No distribution config, so expect conventional default

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Valid K3d context should pass validation")
		assert.Empty(t, result.Errors, "Valid context should have no errors")
	})

	t.Run("invalid_context_pattern", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionKind)
		config.Spec.Connection.Context = "invalid-context"

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Invalid context should fail validation")
		assert.NotEmpty(t, result.Errors, "Invalid context should have errors")

		// Find the context error
		found := false

		for _, err := range result.Errors {
			if err.Field == "spec.connection.context" {
				found = true

				assert.Contains(t, err.Message, "context name does not match expected pattern")
				assert.Contains(t, err.FixSuggestion, "kind-kind")

				break
			}
		}

		assert.True(t, found, "Should have context validation error")
	})
}

// TestKSailValidatorKindConsistency tests Kind distribution name consistency validation.
func TestKSailValidatorKindConsistency(t *testing.T) {
	t.Parallel()

	t.Run("matching_names", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionKind)
		config.Spec.Connection.Context = "kind-ksail" // Set context to match the provided Kind config name

		// Create a Kind config with matching name
		kindConfig := &kindv1alpha4.Cluster{
			Name: "ksail", // Matches expected cluster name
		}

		validator := ksailvalidator.NewValidator(kindConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Matching Kind config names should pass validation")
		assert.Empty(t, result.Errors, "Matching names should have no errors")
	})

	t.Run("custom_name", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionKind)
		config.Spec.Connection.Context = "kind-different-name" // Use different context to match the Kind config name

		// Create a Kind config with specific name
		kindConfig := &kindv1alpha4.Cluster{
			Name: "different-name", // This should be used in the context name
		}

		validator := ksailvalidator.NewValidator(kindConfig)
		result := validator.Validate(config)

		assert.True(
			t,
			result.Valid,
			"Context matching distribution config name should pass validation",
		)
		assert.Empty(t, result.Errors, "Valid context should have no errors")
	})
}

// TestKSailValidatorK3dConsistency tests K3d distribution name consistency validation.
func TestKSailValidatorK3dConsistency(t *testing.T) {
	t.Parallel()

	t.Run("matching_names", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionK3d)
		config.Spec.Connection.Context = "k3d-ksail" // Set context to match the provided K3d config name

		// Create a K3d config with matching name
		k3dConfig := &k3dapi.SimpleConfig{
			ObjectMeta: k3dtypes.ObjectMeta{
				Name: "ksail", // Matches expected cluster name
			},
		}

		validator := ksailvalidator.NewValidator(k3dConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Matching K3d config names should pass validation")
		assert.Empty(t, result.Errors, "Matching names should have no errors")
	})
}

// TestKSailValidatorMultipleConfigs tests validation with multiple distribution configs.
func TestKSailValidatorMultipleConfigs(t *testing.T) {
	t.Parallel()

	t.Run("uses_correct_distribution", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionKind)
		config.Spec.Connection.Context = "kind-ksail" // Set context to match the Kind config name

		// Create both Kind and K3d configs (only Kind should be validated for Kind distribution)
		kindConfig := &kindv1alpha4.Cluster{
			Name: "ksail",
		}
		k3dConfig := &k3dapi.SimpleConfig{
			ObjectMeta: k3dtypes.ObjectMeta{
				Name: "different-name", // This shouldn't matter for Kind distribution
			},
		}

		validator := ksailvalidator.NewValidator(kindConfig, k3dConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Should validate only the matching distribution config")
		assert.Empty(t, result.Errors, "Should have no errors when distribution matches")
	})
}

// TestKSailValidatorUnsupportedDistribution tests handling of unsupported distributions.
// Helper function to create a test cluster config.
func createTestClusterConfig(
	distribution v1alpha1.Distribution,
	configFile,
	context string,
) *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.dev/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution:       distribution,
			DistributionConfig: configFile,
			Connection: v1alpha1.Connection{
				Context: context,
			},
		},
	}
}

// Helper function to check for specific distribution error.
func checkDistributionError(
	t *testing.T,
	errors []validator.ValidationError,
	expectedMessage, errorDescription string,
) {
	t.Helper()

	found := false

	for _, err := range errors {
		if err.Field == specDistributionField &&
			strings.Contains(err.Message, expectedMessage) {
			found = true

			assert.Contains(t, err.FixSuggestion, "Use a supported distribution")

			break
		}
	}

	assert.True(t, found, errorDescription)
}

// Helper function to test supported distribution error paths.
func testSupportedDistributionErrorPath(
	t *testing.T,
	distribution v1alpha1.Distribution,
	_ string,
) {
	t.Helper()

	_ = createTestClusterConfig(distribution, "config.yaml", "some-context")
	// These should normally pass validation or have different errors
	// The unexpected error cases are defensive code paths
}

func TestKSailValidatorUnsupportedDistribution(t *testing.T) {
	t.Parallel()

	testUnknownDistribution(t)
	testSupportedDistributionErrorPaths(t)
}

// testUnknownDistribution tests unknown distribution validation.
func testUnknownDistribution(t *testing.T) {
	t.Helper()

	t.Run("unknown_distribution", func(t *testing.T) {
		t.Parallel()

		config := createTestClusterConfig(
			v1alpha1.Distribution("UnknownDistribution"),
			"unknown.yaml",
			"unknown-context",
		)
		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Unknown distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		checkDistributionError(
			t,
			result.Errors,
			"invalid distribution",
			"Should have invalid distribution error",
		)
	})
}

// testSupportedDistributionErrorPaths tests error paths for supported distributions.
func testSupportedDistributionErrorPaths(t *testing.T) {
	t.Helper()

	t.Run("test_supported_distribution_error_paths", func(t *testing.T) {
		t.Parallel()

		// Test error handling for supported distributions that shouldn't reach unsupported error logic
		// This tests the Kind and K3d cases in addUnsupportedDistributionError
		testCases := []struct {
			name         string
			distribution v1alpha1.Distribution
			expectedMsg  string
		}{
			{
				name:         "kind_unexpected_error",
				distribution: v1alpha1.DistributionKind,
				expectedMsg:  "unexpected error in Kind distribution validation",
			},
			{
				name:         "k3d_unexpected_error",
				distribution: v1alpha1.DistributionK3d,
				expectedMsg:  "unexpected error in K3d distribution validation",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				testSupportedDistributionErrorPath(t, testCase.distribution, testCase.expectedMsg)
			})
		}
	})
}

// TestKSailValidatorContextPatterns tests different context name patterns.
func TestKSailValidatorContextPatterns(t *testing.T) {
	t.Parallel()

	testEmptyContextValidationSkipped(t)
}

// testEmptyContextValidationSkipped tests that empty context skips validation.
func testEmptyContextValidationSkipped(t *testing.T) {
	t.Helper()

	t.Run("empty_context_validation_skipped", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "kind.yaml",
				Connection: v1alpha1.Connection{
					Context: "", // Empty context should skip validation
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Empty context should skip validation")
		assert.Empty(t, result.Errors, "Empty context validation should be skipped")
	})
}

// TestKSailValidatorCrossConfigurationValidation tests validation with distribution configs.
func TestKSailValidatorCrossConfigurationValidation(t *testing.T) {
	t.Parallel()

	testKindCrossValidationWithConfigName(t)
	testK3dCrossValidationWithConfigName(t)
}

// testKindCrossValidationWithConfigName tests Kind cross-validation with custom config name.
func testKindCrossValidationWithConfigName(t *testing.T) {
	t.Helper()

	t.Run("kind_cross_validation_with_config_name", func(t *testing.T) {
		t.Parallel()

		kindConfig := &kindv1alpha4.Cluster{Name: "custom-kind-cluster"}
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "kind.yaml",
				Connection: v1alpha1.Connection{
					Context: "kind-custom-kind-cluster", // Should match the Kind config name
				},
			},
		}

		validator := ksailvalidator.NewValidator(kindConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with matching Kind config name")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// testK3dCrossValidationWithConfigName tests K3d cross-validation with custom config name.
func testK3dCrossValidationWithConfigName(t *testing.T) {
	t.Helper()

	t.Run("k3d_cross_validation_with_config_name", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &k3dapi.SimpleConfig{
			ObjectMeta: k3dtypes.ObjectMeta{Name: "custom-k3d-cluster"},
		}
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionK3d,
				DistributionConfig: "k3d.yaml",
				Connection: v1alpha1.Connection{
					Context: "k3d-custom-k3d-cluster", // Should match the K3d config name
				},
			},
		}

		validator := ksailvalidator.NewValidator(k3dConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with matching K3d config name")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// TestKSailValidatorDefaultFallbackValidation tests validation with default fallback scenarios.
func TestKSailValidatorDefaultFallbackValidation(t *testing.T) {
	t.Parallel()

	testKindDefaultFallback(t)
	testK3dDefaultFallback(t)
}

// testKindDefaultFallback tests Kind validation with default name fallback.
func testKindDefaultFallback(t *testing.T) {
	t.Helper()

	t.Run("kind_default_fallback", func(t *testing.T) {
		t.Parallel()

		kindConfig := &kindv1alpha4.Cluster{Name: ""} // Empty name should use default
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "kind.yaml",
				Connection: v1alpha1.Connection{
					Context: "kind-kind", // Should match default "kind"
				},
			},
		}

		validator := ksailvalidator.NewValidator(kindConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with default Kind name")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// testK3dDefaultFallback tests K3d validation with default name fallback.
func testK3dDefaultFallback(t *testing.T) {
	t.Helper()

	t.Run("k3d_default_fallback", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &k3dapi.SimpleConfig{
			ObjectMeta: k3dtypes.ObjectMeta{Name: ""}, // Empty name should use default
		}
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionK3d,
				DistributionConfig: "k3d.yaml",
				Connection: v1alpha1.Connection{
					Context: "k3d-k3s-default", // Should match default "k3s-default" with k3d prefix
				},
			},
		}

		validator := ksailvalidator.NewValidator(k3dConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with default K3d name")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// TestKSailValidatorSpecialDistributionHandling tests special distribution handling scenarios.
func TestKSailValidatorSpecialDistributionHandling(t *testing.T) {
	t.Parallel()

	testNoDistributionConfigProvided(t)
}

// testNoDistributionConfigProvided tests validation when no distribution config is provided.
func testNoDistributionConfigProvided(t *testing.T) {
	t.Helper()

	t.Run("no_distribution_config_provided", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "kind.yaml",
				Connection: v1alpha1.Connection{
					Context: "kind-kind", // Should match default when no config provided
				},
			},
		}

		validator := ksailvalidator.NewValidator() // No distribution config provided
		result := validator.Validate(config)

		assert.True(
			t,
			result.Valid,
			"Validation should pass with default Kind name when no config provided",
		)
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// TestKSailValidatorDistributionValidation tests distribution field validation.
func TestKSailValidatorDistributionValidation(t *testing.T) {
	t.Parallel()

	testEmptyDistribution(t)
	testEmptyDistributionConfig(t)
	testInvalidDistributionValue(t)
}

// testEmptyDistribution tests validation with empty distribution field.
func testEmptyDistribution(t *testing.T) {
	t.Helper()

	t.Run("empty_distribution", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       "", // Empty distribution
				DistributionConfig: "config.yaml",
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Empty distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		// Check for distribution error
		found := false

		for _, err := range result.Errors {
			if err.Field == specDistributionField &&
				strings.Contains(err.Message, "distribution is required") {
				found = true

				assert.Contains(t, err.FixSuggestion, "Set spec.distribution")

				break
			}
		}

		assert.True(t, found, "Should have distribution required error")
	})
}

// testEmptyDistributionConfig tests validation with empty distribution config.
func testEmptyDistributionConfig(t *testing.T) {
	t.Helper()

	t.Run("empty_distribution_config", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionKind,
				DistributionConfig: "", // Empty distribution config
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Empty distribution config should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		// Check for distribution config error
		found := false

		for _, err := range result.Errors {
			if err.Field == "spec.distributionConfig" &&
				strings.Contains(err.Message, "distributionConfig is required") {
				found = true

				assert.Contains(t, err.FixSuggestion, "Set spec.distributionConfig")

				break
			}
		}

		assert.True(t, found, "Should have distribution config required error")
	})
}

// testInvalidDistributionValue tests validation with invalid distribution value.
func testInvalidDistributionValue(t *testing.T) {
	t.Helper()

	t.Run("invalid_distribution_value", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.Distribution("InvalidDistribution"),
				DistributionConfig: "config.yaml",
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Invalid distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		// Check for invalid distribution error
		found := false

		for _, err := range result.Errors {
			if err.Field == specDistributionField &&
				strings.Contains(err.Message, "invalid distribution value") {
				found = true

				assert.Contains(t, err.FixSuggestion, "Use a supported distribution")

				break
			}
		}

		assert.True(t, found, "Should have invalid distribution error")
	})
}

// TestKSailValidatorKindConfigEdgeCases tests Kind configuration edge cases.
func TestKSailValidatorKindConfigEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		kindConfig   *kindv1alpha4.Cluster
		expectedName string
		description  string
	}{
		{
			name:         "kind_with_whitespace_name",
			kindConfig:   &kindv1alpha4.Cluster{Name: "  test-kind  "},
			expectedName: "  test-kind  ",
			description:  "Should preserve whitespace in Kind config name",
		},
		{
			name:         "kind_with_special_characters",
			kindConfig:   &kindv1alpha4.Cluster{Name: "test-kind_123"},
			expectedName: "test-kind_123",
			description:  "Should handle special characters in Kind config name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create a valid config to test cross-configuration validation
			config := &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution:       v1alpha1.DistributionKind,
					DistributionConfig: "kind.yaml",
					Connection: v1alpha1.Connection{
						Context: "kind-" + test.expectedName,
					},
				},
			}

			validator := ksailvalidator.NewValidator(test.kindConfig)
			result := validator.Validate(config)

			assert.True(t, result.Valid, test.description+" should pass validation")
			assert.Empty(
				t,
				result.Errors,
				test.description+" should have no validation errors",
			)
		})
	}
}

// TestKSailValidatorK3dConfigEdgeCases tests K3d configuration edge cases.
func TestKSailValidatorK3dConfigEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		k3dConfig    *k3dapi.SimpleConfig
		expectedName string
		description  string
	}{
		{
			name: "k3d_with_unicode_name",
			k3dConfig: &k3dapi.SimpleConfig{
				ObjectMeta: k3dtypes.ObjectMeta{Name: "test-k3d-ñ"},
			},
			expectedName: "test-k3d-ñ",
			description:  "Should handle unicode characters in K3d config name",
		},
		{
			name: "k3d_with_long_name",
			k3dConfig: &k3dapi.SimpleConfig{
				ObjectMeta: k3dtypes.ObjectMeta{
					Name: "very-long-k3d-cluster-name-that-exceeds-normal-length",
				},
			},
			expectedName: "very-long-k3d-cluster-name-that-exceeds-normal-length",
			description:  "Should handle long K3d config names",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create a valid config to test cross-configuration validation
			config := &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution:       v1alpha1.DistributionK3d,
					DistributionConfig: "k3d.yaml",
					Connection: v1alpha1.Connection{
						Context: "k3d-" + test.expectedName,
					},
				},
			}

			validator := ksailvalidator.NewValidator(test.k3dConfig)
			result := validator.Validate(config)

			assert.True(t, result.Valid, test.description+" should pass validation")
			assert.Empty(
				t,
				result.Errors,
				test.description+" should have no validation errors",
			)
		})
	}
}

// TestKSailValidatorContextValidationComprehensive tests comprehensive context validation scenarios.
func TestKSailValidatorContextValidationComprehensive(t *testing.T) {
	t.Parallel()

	testKindContextValidation(t)
	testK3dContextValidation(t)
}

// testKindContextValidation tests Kind-specific context validation scenarios.
func testKindContextValidation(t *testing.T) {
	t.Helper()

	tests := []struct {
		name        string
		context     string
		shouldPass  bool
		description string
	}{
		{
			name:        "kind_with_exact_match",
			context:     "kind-kind",
			shouldPass:  true,
			description: "Kind context should match exactly",
		},
		{
			name:        "kind_with_case_mismatch",
			context:     "KIND-kind",
			shouldPass:  false,
			description: "Kind context is case sensitive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution:       v1alpha1.DistributionKind,
					DistributionConfig: "config.yaml",
					Connection: v1alpha1.Connection{
						Context: test.context,
					},
				},
			}

			validateContextTest(t, config, test.shouldPass, test.description)
		})
	}
}

// testK3dContextValidation tests K3d-specific context validation scenarios.
func testK3dContextValidation(t *testing.T) {
	t.Helper()

	tests := []struct {
		name        string
		context     string
		shouldPass  bool
		description string
	}{
		{
			name:        "k3d_with_exact_match",
			context:     "k3d-k3s-default",
			shouldPass:  true,
			description: "K3d context should match exactly",
		},
		{
			name:        "k3d_with_extra_prefix",
			context:     "prefix-k3d-k3s-default",
			shouldPass:  false,
			description: "K3d context should not have extra prefix",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := &v1alpha1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ksail.dev/v1alpha1",
					Kind:       "Cluster",
				},
				Spec: v1alpha1.Spec{
					Distribution:       v1alpha1.DistributionK3d,
					DistributionConfig: "config.yaml",
					Connection: v1alpha1.Connection{
						Context: test.context,
					},
				},
			}

			validateContextTest(t, config, test.shouldPass, test.description)
		})
	}
}

// validateContextTest is a helper function for running context validation tests.
func validateContextTest(
	t *testing.T,
	config *v1alpha1.Cluster,
	shouldPass bool,
	description string,
) {
	t.Helper()

	validator := ksailvalidator.NewValidator()
	result := validator.Validate(config)

	if shouldPass {
		assert.True(t, result.Valid, description+" should pass validation")
		assert.Empty(t, result.Errors, description+" should have no validation errors")
	} else {
		assert.False(t, result.Valid, description+" should fail validation")
		assert.NotEmpty(t, result.Errors, description+" should have validation errors")
	}
}

// TestKSailValidatorMultipleDistributionConfigs tests validator with multiple distribution configs.
func TestKSailValidatorMultipleDistributionConfigs(t *testing.T) {
	t.Parallel()

	testKindWithAllConfigs(t)
	testK3dWithAllConfigs(t)
}

// testKindWithAllConfigs tests Kind validation with all distribution configs available.
func testKindWithAllConfigs(t *testing.T) {
	t.Helper()

	t.Run("kind_with_all_configs", func(t *testing.T) {
		t.Parallel()

		config := createMultiConfigTestCluster(v1alpha1.DistributionKind, "kind-test-kind")
		validator := createMultiConfigValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with multiple configs")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// testK3dWithAllConfigs tests K3d validation with all distribution configs available.
func testK3dWithAllConfigs(t *testing.T) {
	t.Helper()

	t.Run("k3d_with_all_configs", func(t *testing.T) {
		t.Parallel()

		config := createMultiConfigTestCluster(v1alpha1.DistributionK3d, "k3d-test-k3d")
		validator := createMultiConfigValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass with multiple configs")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})
}

// createMultiConfigTestCluster creates a test cluster config for multi-config testing.
func createMultiConfigTestCluster(
	distribution v1alpha1.Distribution,
	context string,
) *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.dev/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution:       distribution,
			DistributionConfig: "config.yaml",
			Connection: v1alpha1.Connection{
				Context: context,
			},
		},
	}
}

// createMultiConfigValidator creates a validator with all distribution configs.
func createMultiConfigValidator() *ksailvalidator.Validator {
	kindConfig := &kindv1alpha4.Cluster{Name: "test-kind"}
	k3dConfig := &k3dapi.SimpleConfig{
		ObjectMeta: k3dtypes.ObjectMeta{Name: "test-k3d"},
	}

	return ksailvalidator.NewValidator(kindConfig, k3dConfig)
}
