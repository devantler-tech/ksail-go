package helpers_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Common test cases used by multiple formatter tests to eliminate duplication.

// createCommonValidationResults creates standard ValidationResult instances for testing.
func createCommonValidationResults() map[string]*validator.ValidationResult {
	return map[string]*validator.ValidationResult{
		"no_errors": {
			Valid:  true,
			Errors: []validator.ValidationError{},
		},
		"single_error": {
			Valid: false,
			Errors: []validator.ValidationError{
				{Field: "name", Message: "is required"},
			},
		},
		"single_error_with_fix": {
			Valid: false,
			Errors: []validator.ValidationError{
				{Field: "name", Message: "is required", FixSuggestion: "add name field"},
			},
		},
		"multiple_errors": {
			Valid: false,
			Errors: []validator.ValidationError{
				{Field: "name", Message: "is required", FixSuggestion: "add name field"},
				{Field: "version", Message: "is invalid"},
			},
		},
		"multiple_errors_with_fixes": {
			Valid: false,
			Errors: []validator.ValidationError{
				{Field: "name", Message: "is required", FixSuggestion: "add name field"},
				{Field: "version", Message: "is invalid", FixSuggestion: "use valid version"},
			},
		},
	}
}

// TestCase represents a test case for formatting functions.
type TestCase struct {
	Name     string
	Result   *validator.ValidationResult
	Expected string
}

// runFormattingTest runs test cases for formatting functions to eliminate test loop duplication.
func runFormattingTest(
	t *testing.T,
	testCases []TestCase,
	formatter func(*validator.ValidationResult) string,
) {
	t.Helper()

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result := formatter(tt.Result)
			assert.Equal(t, tt.Expected, result)
		})
	}
}

// assertValidationError asserts that validation failed with expected error messages.
func assertValidationError(t *testing.T, err error, expectedMessages ...string) {
	t.Helper()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")

	for _, msg := range expectedMessages {
		assert.Contains(t, err.Error(), msg)
	}
}
