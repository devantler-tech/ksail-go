// Package testutils provides common test utilities for validator tests to eliminate duplication.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
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
func CreateNilConfigTestCase[T any]() ValidatorTestCase[T] {
	var nilConfig T

	return ValidatorTestCase[T]{
		Name:          "nil_config",
		Config:        nilConfig,
		ExpectedValid: false,
		ExpectedErrors: []validator.ValidationError{
			{Field: "config", Message: "configuration cannot be nil"},
		},
	}
}
