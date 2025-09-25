package ksail_test

import (
	"strings"
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

// TestKSailValidatorUnsupportedDistribution tests handling of unsupported distributions.
func TestKSailValidatorUnsupportedDistribution(t *testing.T) {
	t.Parallel()

	t.Run("tind_distribution", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionTind,
				DistributionConfig: "tind.yaml",
				Connection: v1alpha1.Connection{
					Context: "tind-default",
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Tind distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		// Check for Tind-specific error
		found := false

		for _, err := range result.Errors {
			if err.Field == specDistributionField &&
				strings.Contains(err.Message, "Tind distribution is not yet supported") {
				found = true

				assert.Contains(t, err.FixSuggestion, "Use a supported distribution")

				break
			}
		}

		assert.True(t, found, "Should have Tind-specific unsupported distribution error")
	})

	t.Run("unknown_distribution", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.Distribution("UnknownDistribution"),
				DistributionConfig: "unknown.yaml",
				Connection: v1alpha1.Connection{
					Context: "unknown-context",
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Unknown distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors")

		// Check for unknown distribution error
		found := false

		for _, err := range result.Errors {
			if err.Field == specDistributionField &&
				strings.Contains(err.Message, "unknown distribution") {
				found = true

				assert.Contains(t, err.FixSuggestion, "Use a supported distribution")

				break
			}
		}

		assert.True(t, found, "Should have unknown distribution error")
	})

	t.Run("test_supported_distribution_error_paths", func(t *testing.T) {
		t.Parallel()

		// Test error handling for supported distributions that shouldn't reach unsupported error logic
		// This tests the Kind, K3d, and EKS cases in addUnsupportedDistributionError
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
			{
				name:         "eks_unexpected_error",
				distribution: v1alpha1.DistributionEKS,
				expectedMsg:  "unexpected error in EKS distribution validation",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				// Create a config that would normally be valid but with mismatched context
				// to potentially trigger the unsupported distribution error path
				config := &v1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "ksail.dev/v1alpha1",
						Kind:       "Cluster",
					},
					Spec: v1alpha1.Spec{
						Distribution:       testCase.distribution,
						DistributionConfig: "config.yaml",
						Connection: v1alpha1.Connection{
							Context: "some-context",
						},
					},
				}

				validator := ksailvalidator.NewValidator()
				result := validator.Validate(config)

				// These should normally pass validation or have different errors
				// The unexpected error cases are defensive code paths
				if testCase.distribution == v1alpha1.DistributionEKS {
					// EKS skips context validation entirely
					assert.True(t, result.Valid, "EKS should skip context validation")
				} else {
					// Other distributions may have context validation errors but not the "unexpected" ones
					// This tests the normal validation flow
					if !result.Valid {
						for _, err := range result.Errors {
							assert.NotContains(t, err.Message, testCase.expectedMsg,
								"Should not have unexpected error message in normal validation")
						}
					}
				}
			})
		}
	})
}

// TestKSailValidatorEKSConfigName tests EKS configuration name extraction.
func TestKSailValidatorEKSConfigName(t *testing.T) {
	t.Parallel()

	t.Run("eks_with_metadata_name", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "test-cluster",
				},
			},
		}

		// Create EKS config with metadata name
		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{
				Name: "test-cluster",
			},
		}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "EKS config with metadata should pass validation")
		assert.Empty(t, result.Errors, "Should have no errors")
	})

	t.Run("eks_without_metadata", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "eks-default",
				},
			},
		}

		// Create EKS config without metadata (should use default)
		eksConfig := &eksctl.ClusterConfig{}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "EKS config without metadata should use default name")
		assert.Empty(t, result.Errors, "Should have no errors")
	})

	t.Run("eks_with_empty_metadata_name", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "eks-default",
				},
			},
		}

		// Create EKS config with empty metadata name
		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{
				Name: "",
			},
		}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "EKS config with empty metadata name should use default")
		assert.Empty(t, result.Errors, "Should have no errors")
	})

	t.Run("eks_no_config_uses_default", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "eks-default",
				},
			},
		}

		// No EKS config provided - should use default
		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "EKS without config should use default name")
		assert.Empty(t, result.Errors, "Should have no errors")
	})
}

// TestKSailValidatorContextPatterns tests different context name patterns.
func TestKSailValidatorContextPatterns(t *testing.T) {
	t.Parallel()

	t.Run("eks_skip_context_validation", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "any-context-name", // EKS should skip context validation
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.True(t, result.Valid, "EKS should skip context validation")
		assert.Empty(t, result.Errors, "EKS context validation should be skipped")
	})

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

	t.Run("tind_expected_context_pattern", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionTind,
				DistributionConfig: "tind.yaml",
				Connection: v1alpha1.Connection{
					Context: "tind-cluster",
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Tind distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors for Tind")
	})
}

// TestKSailValidatorCrossConfigurationValidation tests validation with distribution configs
// This indirectly tests the getDistributionConfigName method through the public API.
func TestKSailValidatorCrossConfigurationValidation(t *testing.T) {
	t.Parallel()

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

	t.Run("eks_cross_validation_with_config_name", func(t *testing.T) {
		t.Parallel()

		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{Name: "custom-eks-cluster"},
		}
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "", // EKS doesn't use context validation
				},
			},
		}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass for EKS config")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})

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

	t.Run("eks_default_fallback", func(t *testing.T) {
		t.Parallel()

		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{Name: ""}, // Empty name gets default behavior
		}
		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionEKS,
				DistributionConfig: "eks.yaml",
				Connection: v1alpha1.Connection{
					Context: "", // EKS skips context validation
				},
			},
		}

		validator := ksailvalidator.NewValidator(eksConfig)
		result := validator.Validate(config)

		assert.True(t, result.Valid, "Validation should pass for EKS with empty metadata name")
		assert.Empty(t, result.Errors, "Should have no validation errors")
	})

	t.Run("tind_distribution_handling", func(t *testing.T) {
		t.Parallel()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution:       v1alpha1.DistributionTind,
				DistributionConfig: "tind.yaml",
				Connection: v1alpha1.Connection{
					Context: "tind-cluster",
				},
			},
		}

		validator := ksailvalidator.NewValidator()
		result := validator.Validate(config)

		assert.False(t, result.Valid, "Tind distribution should fail validation")
		assert.NotEmpty(t, result.Errors, "Should have validation errors for Tind")
	})

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

				assert.Contains(t, err.FixSuggestion, "Use a valid distribution type")

				break
			}
		}

		assert.True(t, found, "Should have invalid distribution error")
	})
}

// TestKSailValidatorCoverageEnhancement tests additional scenarios to improve code coverage.
func TestKSailValidatorCoverageEnhancement(t *testing.T) {
	t.Parallel()

	t.Run("distribution_config_name_edge_cases", func(t *testing.T) {
		t.Parallel()

		// Test getKindConfigName with various scenarios
		t.Run("kind_config_edge_cases", func(t *testing.T) {
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
		})

		// Test getK3dConfigName with various scenarios
		t.Run("k3d_config_edge_cases", func(t *testing.T) {
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
		})

		// Test EKS config name handling
		t.Run("eks_config_edge_cases", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				eksConfig   *eksctl.ClusterConfig
				description string
			}{
				{
					name: "eks_with_complex_metadata",
					eksConfig: &eksctl.ClusterConfig{
						Metadata: &eksctl.ClusterMeta{
							Name:    "test-eks-cluster",
							Region:  "us-west-2",
							Version: "1.28",
						},
					},
					description: "Should handle EKS config with complex metadata",
				},
				{
					name: "eks_with_minimal_metadata",
					eksConfig: &eksctl.ClusterConfig{
						Metadata: &eksctl.ClusterMeta{
							Name: "minimal",
						},
					},
					description: "Should handle EKS config with minimal metadata",
				},
			}

			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()

					// Create a valid config - EKS doesn't validate context
					config := &v1alpha1.Cluster{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "ksail.dev/v1alpha1",
							Kind:       "Cluster",
						},
						Spec: v1alpha1.Spec{
							Distribution:       v1alpha1.DistributionEKS,
							DistributionConfig: "eks.yaml",
							Connection: v1alpha1.Connection{
								Context: "", // EKS allows empty context
							},
						},
					}

					validator := ksailvalidator.NewValidator(test.eksConfig)
					result := validator.Validate(config)

					assert.True(t, result.Valid, test.description+" should pass validation")
					assert.Empty(
						t,
						result.Errors,
						test.description+" should have no validation errors",
					)
				})
			}
		})
	})

	t.Run("context_validation_comprehensive", func(t *testing.T) {
		t.Parallel()

		// Test various context validation scenarios to improve coverage
		tests := []struct {
			name         string
			distribution v1alpha1.Distribution
			context      string
			shouldPass   bool
			description  string
		}{
			{
				name:         "kind_with_exact_match",
				distribution: v1alpha1.DistributionKind,
				context:      "kind-kind",
				shouldPass:   true,
				description:  "Kind context should match exactly",
			},
			{
				name:         "k3d_with_exact_match",
				distribution: v1alpha1.DistributionK3d,
				context:      "k3d-k3s-default",
				shouldPass:   true,
				description:  "K3d context should match exactly",
			},
			{
				name:         "eks_with_any_context",
				distribution: v1alpha1.DistributionEKS,
				context:      "any-context-name",
				shouldPass:   true,
				description:  "EKS allows any context",
			},
			{
				name:         "kind_with_case_mismatch",
				distribution: v1alpha1.DistributionKind,
				context:      "KIND-kind",
				shouldPass:   false,
				description:  "Kind context is case sensitive",
			},
			{
				name:         "k3d_with_extra_prefix",
				distribution: v1alpha1.DistributionK3d,
				context:      "prefix-k3d-k3s-default",
				shouldPass:   false,
				description:  "K3d context should not have extra prefix",
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
						Distribution:       test.distribution,
						DistributionConfig: "config.yaml",
						Connection: v1alpha1.Connection{
							Context: test.context,
						},
					},
				}

				validator := ksailvalidator.NewValidator()
				result := validator.Validate(config)

				if test.shouldPass {
					assert.True(t, result.Valid, test.description+" should pass validation")
					assert.Empty(
						t,
						result.Errors,
						test.description+" should have no validation errors",
					)
				} else {
					assert.False(t, result.Valid, test.description+" should fail validation")
					assert.NotEmpty(t, result.Errors, test.description+" should have validation errors")
				}
			})
		}
	})

	t.Run("multiple_distribution_configs", func(t *testing.T) {
		t.Parallel()

		// Test validator with multiple distribution configs
		kindConfig := &kindv1alpha4.Cluster{Name: "test-kind"}
		k3dConfig := &k3dapi.SimpleConfig{
			ObjectMeta: k3dtypes.ObjectMeta{Name: "test-k3d"},
		}
		eksConfig := &eksctl.ClusterConfig{
			Metadata: &eksctl.ClusterMeta{Name: "test-eks"},
		}

		tests := []struct {
			name         string
			distribution v1alpha1.Distribution
			context      string
			shouldPass   bool
		}{
			{
				name:         "kind_with_all_configs",
				distribution: v1alpha1.DistributionKind,
				context:      "kind-test-kind",
				shouldPass:   true,
			},
			{
				name:         "k3d_with_all_configs",
				distribution: v1alpha1.DistributionK3d,
				context:      "k3d-test-k3d",
				shouldPass:   true,
			},
			{
				name:         "eks_with_all_configs",
				distribution: v1alpha1.DistributionEKS,
				context:      "", // EKS allows empty context
				shouldPass:   true,
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
						Distribution:       test.distribution,
						DistributionConfig: "config.yaml",
						Connection: v1alpha1.Connection{
							Context: test.context,
						},
					},
				}

				// Create validator with all distribution configs
				validator := ksailvalidator.NewValidator(kindConfig, k3dConfig, eksConfig)
				result := validator.Validate(config)

				if test.shouldPass {
					assert.True(t, result.Valid, "Validation should pass with multiple configs")
					assert.Empty(t, result.Errors, "Should have no validation errors")
				} else {
					assert.False(t, result.Valid, "Validation should fail")
					assert.NotEmpty(t, result.Errors, "Should have validation errors")
				}
			})
		}
	})

	t.Run("additional_context_patterns", func(t *testing.T) {
		t.Parallel()

		// Test additional context patterns to improve getExpectedContextName coverage
		tests := []struct {
			name         string
			distribution v1alpha1.Distribution
			context      string
			kindConfig   *kindv1alpha4.Cluster
			k3dConfig    *k3dapi.SimpleConfig
			eksConfig    *eksctl.ClusterConfig
			shouldPass   bool
		}{
			{
				name:         "tind_context_pattern_with_empty_name",
				distribution: v1alpha1.DistributionTind,
				context:      "tind-", // Empty name case
				kindConfig:   nil,
				k3dConfig:    nil,
				eksConfig:    nil,
				shouldPass:   false, // Tind is unsupported
			},
			{
				name:         "unknown_distribution_context",
				distribution: v1alpha1.Distribution("UnknownDist"),
				context:      "unknown-context",
				kindConfig:   nil,
				k3dConfig:    nil,
				eksConfig:    nil,
				shouldPass:   false, // Unknown distribution
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
						Distribution:       test.distribution,
						DistributionConfig: "config.yaml",
						Connection: v1alpha1.Connection{
							Context: test.context,
						},
					},
				}

				// Create validator based on available configs
				var validator *ksailvalidator.Validator

				if test.kindConfig != nil || test.k3dConfig != nil || test.eksConfig != nil {
					configs := make([]any, 0)
					if test.kindConfig != nil {
						configs = append(configs, test.kindConfig)
					}

					if test.k3dConfig != nil {
						configs = append(configs, test.k3dConfig)
					}

					if test.eksConfig != nil {
						configs = append(configs, test.eksConfig)
					}

					validator = ksailvalidator.NewValidator(configs...)
				} else {
					validator = ksailvalidator.NewValidator()
				}

				result := validator.Validate(config)

				if test.shouldPass {
					assert.True(t, result.Valid, test.name+" should pass validation")
					assert.Empty(t, result.Errors, test.name+" should have no validation errors")
				} else {
					assert.False(t, result.Valid, test.name+" should fail validation")
					assert.NotEmpty(t, result.Errors, test.name+" should have validation errors")
				}
			})
		}
	})
}
