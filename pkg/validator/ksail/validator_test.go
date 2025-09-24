package ksail_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/validator"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/validator/ksail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		expectErrors: []string{"spec.distribution"},
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
				Distribution: v1alpha1.DistributionKind,
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
			if err.Field == "spec.distribution" {
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
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.dev/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution: distribution,
		},
	}
}
