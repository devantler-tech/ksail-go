package ksail_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/validator"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/validator/ksail"
	k3dtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	eksctl "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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
		contextName = "default" // EKS doesn't use prefix pattern
	case v1alpha1.DistributionTind:
		distributionConfigFile = "tind.yaml"
		contextName = "tind-default" // No distribution config provided, use "default"
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

// TestKSailValidatorEKSConsistency tests EKS distribution name consistency validation.
func TestKSailValidatorEKSConsistency(t *testing.T) {
	t.Parallel()

	t.Run("matching_names", func(t *testing.T) {
		t.Parallel()

		config := createValidKSailConfig(v1alpha1.DistributionEKS)
		config.Spec.Connection.Context = "ksail" // Set context to match the provided EKS config name (no prefix for EKS)

		// Create an EKS config with matching name
		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{
				Name: "ksail", // Matches expected cluster name
			},
		}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Matching EKS config names should pass validation")
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
