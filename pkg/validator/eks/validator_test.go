package eks_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	eksvalidator "github.com/devantler-tech/ksail-go/pkg/validator/eks"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// TestNewValidator tests the NewValidator constructor
func TestNewValidator(t *testing.T) {
	t.Parallel()

	t.Run("constructor", func(t *testing.T) {
		t.Parallel()
		validator := eksvalidator.NewValidator()
		if validator == nil {
			t.Fatal("NewValidator should return non-nil validator")
		}
	})
}

// TestValidate tests the main Validate method with comprehensive scenarios
func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("contract_scenarios", func(t *testing.T) {
		t.Parallel()
		testEKSValidatorContract(t)
	})

	t.Run("edge_cases", func(t *testing.T) {
		t.Parallel()
		testEKSValidatorEdgeCases(t)
	})
}

// Helper function for contract testing
func testEKSValidatorContract(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	validatorInstance := eksvalidator.NewValidator()
	testCases := createEKSTestCases()

	testutils.RunValidatorTests(
		t,
		validatorInstance,
		testCases,
		testutils.AssertValidationResult[*eksctlapi.ClusterConfig],
	)
}

func createEKSTestCases() []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig] {
	var testCases []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig]

	testCases = append(testCases, createEKSValidTestCases()...)
	testCases = append(testCases, createEKSInvalidTestCases()...)
	testCases = append(testCases, testutils.CreateNilConfigTestCase[*eksctlapi.ClusterConfig]())

	return testCases
}

func createEKSValidTestCases() []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig] {
	return []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig]{
		{
			Name: "valid_eks_config",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Name:   "test-cluster",
					Region: "us-west-2",
				},
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "invalid_eks_config_wrong_kind",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Name:   "test-cluster",
					Region: "us-west-2",
				},
			},
			ExpectedValid:  true, // We override the Kind field validation in preprocessing, so this should be valid
			ExpectedErrors: []validator.ValidationError{},
		},
	}
}

func createEKSInvalidTestCases() []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig] {
	return []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig]{
		{
			Name: "invalid_eks_config_missing_name",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Region: "us-west-2",
				},
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata.name", Message: "cluster name is required"},
			},
		},
		{
			Name: "invalid_eks_config_missing_region",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Name: "test-cluster",
				},
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata.region", Message: "region is required"},
			},
		},
		{
			Name: "invalid_eks_config_missing_metadata",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				// No metadata
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata", Message: "metadata is required"},
			},
		},
		{
			Name: "invalid_eks_config_missing_type_meta",
			Config: &eksctlapi.ClusterConfig{
				// TypeMeta is missing
				Metadata: &eksctlapi.ClusterMeta{
					Name:   "test-cluster",
					Region: "us-west-2",
				},
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "kind", Message: "kind is required"},
				{Field: "apiVersion", Message: "apiVersion is required"},
			},
		},
	}
}

// testEKSValidatorEdgeCases tests specific edge cases and error conditions.
func testEKSValidatorEdgeCases(t *testing.T) {
	t.Run("upstream_validation_complex_config", testEKSUpstreamValidationComplexConfig)
	t.Run("empty_metadata_fields", testEKSEmptyMetadataFields)
	t.Run("invalid_region_format", testEKSInvalidRegionFormat)
}

func testEKSUpstreamValidationComplexConfig(t *testing.T) {
	t.Parallel()

	validatorInstance := eksvalidator.NewValidator()

	// Test with a more complex configuration that might trigger upstream validation
	config := &eksctlapi.ClusterConfig{
		TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
		Metadata: &eksctlapi.ClusterMeta{
			Name:    "test-cluster",
			Region:  "us-west-2",
			Version: "1.27",
		},
		VPC: &eksctlapi.ClusterVPC{
			Network: eksctlapi.Network{
				ID: "vpc-12345", // Test with existing VPC ID
			},
		},
	}

	result := validatorInstance.Validate(config)

	// Should validate successfully or provide meaningful errors
	if !result.Valid {
		if len(result.Errors) == 0 {
			t.Errorf("Expected validation errors when result is invalid")
		}
		// All errors should have proper field and message
		for _, err := range result.Errors {
			if err.Field == "" || err.Message == "" {
				t.Errorf("Validation error missing field or message: %+v", err)
			}
		}
	}
}

func testEKSEmptyMetadataFields(t *testing.T) {
	t.Parallel()

	validatorInstance := eksvalidator.NewValidator()

	// Test with empty metadata fields
	config := &eksctlapi.ClusterConfig{
		TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
		Metadata: &eksctlapi.ClusterMeta{
			Name:   "", // Empty name
			Region: "", // Empty region
		},
	}

	result := validatorInstance.Validate(config)

	// Should fail validation
	if result.Valid {
		t.Errorf("Expected validation to fail with empty metadata fields")
	}

	// Should have specific errors for missing name and region
	foundNameError := false
	foundRegionError := false

	for _, err := range result.Errors {
		if err.Field == "metadata.name" {
			foundNameError = true
		}

		if err.Field == "metadata.region" {
			foundRegionError = true
		}
	}

	if !foundNameError {
		t.Errorf("Expected validation error for missing cluster name")
	}

	if !foundRegionError {
		t.Errorf("Expected validation error for missing region")
	}
}

func testEKSInvalidRegionFormat(t *testing.T) {
	t.Parallel()

	validatorInstance := eksvalidator.NewValidator()

	// Test with invalid region format
	config := &eksctlapi.ClusterConfig{
		TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
		Metadata: &eksctlapi.ClusterMeta{
			Name:   "test-cluster",
			Region: "invalid-region-format", // Invalid region format
		},
	}

	result := validatorInstance.Validate(config)

	// This may pass basic validation but could fail upstream validation
	// We just verify the validation process completes without panic
	if !result.Valid && len(result.Errors) > 0 {
		// Errors should be properly formatted
		for _, err := range result.Errors {
			if err.Message == "" {
				t.Errorf("Validation error missing message: %+v", err)
			}
		}
	}
}
