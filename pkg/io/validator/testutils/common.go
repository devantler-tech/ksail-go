// Package testutils provides common test utilities for validator tests to eliminate duplication.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	"github.com/stretchr/testify/require"
)

// ValidatorTestCase represents a common test case structure for validator tests.
type ValidatorTestCase[T any] struct {
	Name           string
	Config         T
	ExpectedValid  bool
	ExpectedErrors []validator.ValidationError
}

// RunValidatorTests runs a common test pattern for validator implementations.
// This eliminates duplication between different validator test files.
func RunValidatorTests[T any](
	t *testing.T,
	validatorInstance validator.Validator[T],
	testCases []ValidatorTestCase[T],
	assertFunc func(*testing.T, ValidatorTestCase[T], *validator.ValidationResult),
) {
	t.Helper()

	require.NotNil(t, validatorInstance, "Validator constructor must return non-nil validator")

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			result := validatorInstance.Validate(testCase.Config)
			require.NotNil(t, result, "Validation result cannot be nil")

			assertFunc(t, testCase, result)
		})
	}
}

// AssertValidationResult provides a common assertion pattern for validation results.
func AssertValidationResult[T any](
	t *testing.T,
	testCase ValidatorTestCase[T],
	result *validator.ValidationResult,
) {
	t.Helper()

	if testCase.ExpectedValid {
		require.True(
			t,
			result.Valid,
			"Expected validation to pass but it failed: %v",
			result.Errors,
		)
		require.Empty(t, result.Errors, "Expected no validation errors")
	} else {
		require.False(t, result.Valid, "Expected validation to fail but it passed")
		require.NotEmpty(t, result.Errors, "Expected validation errors")

		// If specific errors are expected, validate them
		if len(testCase.ExpectedErrors) > 0 {
			require.Len(t, result.Errors, len(testCase.ExpectedErrors),
				"Expected %d validation errors, got %d", len(testCase.ExpectedErrors), len(result.Errors))

			for i, expectedError := range testCase.ExpectedErrors {
				require.Equal(t, expectedError.Field, result.Errors[i].Field,
					"Error %d field mismatch", i)
				require.Contains(t, result.Errors[i].Message, expectedError.Message,
					"Error %d message should contain expected message", i)
			}
		}
	}
}

// CreateNilConfigTestCase creates a standard test case for nil configuration validation.
// This only works correctly for pointer types where the zero value is nil.
func CreateNilConfigTestCase[T any]() ValidatorTestCase[T] {
	var nilConfig T // For pointer types, this will be nil

	return ValidatorTestCase[T]{
		Name:          "nil_config",
		Config:        nilConfig,
		ExpectedValid: false,
		ExpectedErrors: []validator.ValidationError{
			{Field: "config", Message: "configuration is nil"},
		},
	}
}

// MetadataTestCaseConfig contains configuration for metadata validation test cases.
type MetadataTestCaseConfig struct {
	ExpectedKind       string
	ExpectedAPIVersion string
}

// CreateMetadataValidationTestCases creates common metadata validation test cases.
// This eliminates duplication of Kind and APIVersion validation tests across validators.
// The configFactory function should create a new instance each time it's called.
func CreateMetadataValidationTestCases[T any](
	configFactory func() T,
	setKind func(T, string),
	setAPIVersion func(T, string),
	config MetadataTestCaseConfig,
) []ValidatorTestCase[T] {
	return []ValidatorTestCase[T]{
		createMissingKindTestCase(configFactory, setKind, setAPIVersion, config),
		createMissingAPIVersionTestCase(configFactory, setKind, setAPIVersion, config),
		createMissingBothTestCase(configFactory, setKind, setAPIVersion, config),
	}
}

// createMissingKindTestCase creates a test case for missing kind field.
func createMissingKindTestCase[T any](
	configFactory func() T,
	setKind func(T, string),
	setAPIVersion func(T, string),
	config MetadataTestCaseConfig,
) ValidatorTestCase[T] {
	missingKindConfig := configFactory()
	setKind(missingKindConfig, "")
	setAPIVersion(missingKindConfig, config.ExpectedAPIVersion)

	return ValidatorTestCase[T]{
		Name:          "missing_kind",
		Config:        missingKindConfig,
		ExpectedValid: false,
		ExpectedErrors: []validator.ValidationError{
			{
				Field:         "kind",
				Message:       "kind is required",
				ExpectedValue: config.ExpectedKind,
				FixSuggestion: "Set kind to '" + config.ExpectedKind + "'",
			},
		},
	}
}

// createMissingAPIVersionTestCase creates a test case for missing apiVersion field.
func createMissingAPIVersionTestCase[T any](
	configFactory func() T,
	setKind func(T, string),
	setAPIVersion func(T, string),
	config MetadataTestCaseConfig,
) ValidatorTestCase[T] {
	missingAPIVersionConfig := configFactory()
	setKind(missingAPIVersionConfig, config.ExpectedKind)
	setAPIVersion(missingAPIVersionConfig, "")

	return ValidatorTestCase[T]{
		Name:          "missing_api_version",
		Config:        missingAPIVersionConfig,
		ExpectedValid: false,
		ExpectedErrors: []validator.ValidationError{
			{
				Field:         "apiVersion",
				Message:       "apiVersion is required",
				ExpectedValue: config.ExpectedAPIVersion,
				FixSuggestion: "Set apiVersion to '" + config.ExpectedAPIVersion + "'",
			},
		},
	}
}

// createMissingBothTestCase creates a test case for missing both kind and apiVersion fields.
func createMissingBothTestCase[T any](
	configFactory func() T,
	setKind func(T, string),
	setAPIVersion func(T, string),
	config MetadataTestCaseConfig,
) ValidatorTestCase[T] {
	missingBothConfig := configFactory()
	setKind(missingBothConfig, "")
	setAPIVersion(missingBothConfig, "")

	return ValidatorTestCase[T]{
		Name:          "missing_both",
		Config:        missingBothConfig,
		ExpectedValid: false,
		ExpectedErrors: []validator.ValidationError{
			{
				Field:         "kind",
				Message:       "kind is required",
				ExpectedValue: config.ExpectedKind,
				FixSuggestion: "Set kind to '" + config.ExpectedKind + "'",
			},
			{
				Field:         "apiVersion",
				Message:       "apiVersion is required",
				ExpectedValue: config.ExpectedAPIVersion,
				FixSuggestion: "Set apiVersion to '" + config.ExpectedAPIVersion + "'",
			},
		},
	}
}

// RunNewValidatorConstructorTest runs a common test pattern for validator constructors.
// This eliminates duplication of NewValidator constructor tests across validator test files.
func RunNewValidatorConstructorTest[T any](
	t *testing.T,
	constructorFunc func() validator.Validator[T],
) {
	t.Helper()

	t.Run("constructor", func(t *testing.T) {
		t.Parallel()

		validatorInstance := constructorFunc()
		if validatorInstance == nil {
			t.Fatal("NewValidator should return non-nil validator")
		}
	})
}

// RunValidateTest runs a common test pattern for the main Validate method.
// This eliminates duplication of TestValidate function structure across validator test files.
func RunValidateTest[T any](
	t *testing.T,
	contractTestFunc func(*testing.T),
	edgeTestFuncs ...func(*testing.T),
) {
	t.Helper()

	t.Run("contract_scenarios", func(t *testing.T) {
		t.Parallel()
		contractTestFunc(t)
	})

	// Run edge case tests if provided
	if len(edgeTestFuncs) > 0 {
		t.Run("edge_cases", func(t *testing.T) {
			t.Parallel()

			for _, testFunc := range edgeTestFuncs {
				testFunc(t)
			}
		})
	}
}
