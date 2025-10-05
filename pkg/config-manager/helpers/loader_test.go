package helpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name       string `yaml:"name"`
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// createDefaultConfig creates a default test configuration.
func createDefaultConfig() *testConfig {
	return &testConfig{Name: "default", APIVersion: "test/v1", Kind: "TestCluster"}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Parallel()

	t.Run("file exists", testLoadConfigFileExists)
	t.Run("file doesn't exist returns default", testLoadConfigFileNotExists)
	t.Run("invalid YAML returns error", testLoadConfigInvalidYAML)
}

func testLoadConfigFileExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	configContent := "name: test-cluster\napiVersion: test/v1\nkind: TestCluster"
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	config, err := helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.NoError(t, err)
	assert.Equal(t, "test-cluster", config.Name)
	assert.Equal(t, "test/v1", config.APIVersion)
	assert.Equal(t, "TestCluster", config.Kind)
}

func testLoadConfigFileNotExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent.yaml")

	config, err := helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.NoError(t, err)
	assert.Equal(t, "default", config.Name)
	assert.Equal(t, "test/v1", config.APIVersion)
	assert.Equal(t, "TestCluster", config.Kind)
}

func testLoadConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0o600)
	require.NoError(t, err)

	_, err = helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestFormatValidationErrors(t *testing.T) {
	t.Parallel()

	commonResults := createCommonValidationResults()

	tests := []TestCase{
		{
			Name:     "single error without fix suggestion",
			Result:   commonResults["single_error"],
			Expected: "name: is required",
		},
		{
			Name:     "single error with fix suggestion",
			Result:   commonResults["single_error_with_fix"],
			Expected: "name: is required (add name field)",
		},
		{
			Name:     "multiple errors",
			Result:   commonResults["multiple_errors"],
			Expected: "name: is required (add name field); version: is invalid",
		},
		{
			Name:     "no errors",
			Result:   commonResults["no_errors"],
			Expected: "",
		},
	}

	runFormattingTest(t, tests, helpers.FormatValidationErrors)
}

// ---- Test helpers (reintroduced after refactor) ----

// TestCase represents a validation formatting test case.
type TestCase struct {
	Name     string
	Result   *validator.ValidationResult
	Expected string
}

// createCommonValidationResults builds a reusable map of validation results to avoid repetition.
func createCommonValidationResults() map[string]*validator.ValidationResult {
	results := make(map[string]*validator.ValidationResult)

	// Single error
	singleErr := validator.NewValidationResult("test.yaml")
	singleErr.AddError(validator.ValidationError{Field: "name", Message: "is required"})
	results["single_error"] = singleErr

	// Single error with fix
	singleErrWithFix := validator.NewValidationResult("test.yaml")
	singleErrWithFix.AddError(
		validator.ValidationError{
			Field:         "name",
			Message:       "is required",
			FixSuggestion: "add name field",
		},
	)
	results["single_error_with_fix"] = singleErrWithFix

	// Multiple errors (one with fix, one without)
	multi := validator.NewValidationResult("test.yaml")
	multi.AddError(
		validator.ValidationError{
			Field:         "name",
			Message:       "is required",
			FixSuggestion: "add name field",
		},
	)
	multi.AddError(validator.ValidationError{Field: "version", Message: "is invalid"})
	results["multiple_errors"] = multi

	// Multiple errors with fixes
	multiWithFixes := validator.NewValidationResult("test.yaml")
	multiWithFixes.AddError(
		validator.ValidationError{
			Field:         "name",
			Message:       "is required",
			FixSuggestion: "add name field",
		},
	)
	multiWithFixes.AddError(
		validator.ValidationError{
			Field:         "version",
			Message:       "is invalid",
			FixSuggestion: "use valid version",
		},
	)
	results["multiple_errors_with_fixes"] = multiWithFixes

	// No errors
	none := validator.NewValidationResult("test.yaml")
	results["no_errors"] = none

	return results
}

// runFormattingTest executes a series of formatting test cases.
func runFormattingTest(
	t *testing.T,
	tests []TestCase,
	formatFunc func(*validator.ValidationResult) string,
) {
	t.Helper()

	for _, testCaseEntry := range tests {
		// loop variable copy not needed in Go 1.22+
		t.Run(testCaseEntry.Name, func(t *testing.T) {
			t.Parallel()

			output := formatFunc(testCaseEntry.Result)
			assert.Equal(t, testCaseEntry.Expected, output)
		})
	}
}

// assertValidationError asserts that the error contains all expected substrings.
func assertValidationError(t *testing.T, err error, expectedSubstrings ...string) {
	t.Helper()
	require.Error(t, err)

	for _, substr := range expectedSubstrings {
		assert.Contains(t, err.Error(), substr, "expected validation error to contain substring")
	}
}

func TestFormatValidationErrorsMultiline(t *testing.T) {
	t.Parallel()

	commonResults := createCommonValidationResults()

	tests := []TestCase{
		{
			Name:     "single error",
			Result:   commonResults["single_error"],
			Expected: "  - name: is required\n",
		},
		{
			Name:     "multiple errors with specific validation data",
			Result:   commonResults["multiple_errors"],
			Expected: "  - name: is required\n  - version: is invalid\n",
		},
		{
			Name:     "no errors",
			Result:   commonResults["no_errors"],
			Expected: "",
		},
	}

	runFormattingTest(t, tests, helpers.FormatValidationErrorsMultiline)
}

func TestFormatValidationFixSuggestions(t *testing.T) {
	t.Parallel()

	commonResults := createCommonValidationResults()

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected []string
	}{
		{
			name:     "single fix suggestion",
			result:   commonResults["single_error_with_fix"],
			expected: []string{"    Fix: add name field"},
		},
		{
			name:     "multiple fix suggestions",
			result:   commonResults["multiple_errors_with_fixes"],
			expected: []string{"    Fix: add name field", "    Fix: use valid version"},
		},
		{
			name:     "mixed errors with and without fix suggestions",
			result:   commonResults["multiple_errors"],
			expected: []string{"    Fix: add name field"},
		},
		{
			name:     "no fix suggestions",
			result:   commonResults["single_error"],
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := helpers.FormatValidationFixSuggestions(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidationWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected []string
	}{
		{
			name: "single warning",
			result: &validator.ValidationResult{
				Valid: true,
				Warnings: []validator.ValidationError{
					{Field: "deprecated", Message: "field is deprecated"},
				},
			},
			expected: []string{"Warning - deprecated: field is deprecated"},
		},
		{
			name: "multiple warnings",
			result: &validator.ValidationResult{
				Valid: true,
				Warnings: []validator.ValidationError{
					{Field: "deprecated", Message: "field is deprecated"},
					{Field: "unused", Message: "field is unused"},
				},
			},
			expected: []string{
				"Warning - deprecated: field is deprecated",
				"Warning - unused: field is unused",
			},
		},
		{
			name: "no warnings",
			result: &validator.ValidationResult{
				Valid:    true,
				Warnings: []validator.ValidationError{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := helpers.FormatValidationWarnings(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config passes validation", testValidateConfigValid)
	t.Run("invalid config fails validation", testValidateConfigInvalid)
	t.Run("multiple validation errors", testValidateConfigMultipleErrors)
}

func testValidateConfigValid(t *testing.T) {
	t.Parallel()

	config := &testConfig{
		Name:       "test-cluster",
		APIVersion: "test/v1",
		Kind:       "TestCluster",
	}

	validationResult := validator.NewValidationResult("test.yaml")
	mockVal := validator.NewMockValidator[*testConfig](t)
	mockVal.EXPECT().Validate(config).Return(validationResult)

	err := helpers.ValidateConfig(config, mockVal)
	assert.NoError(t, err)
}

func testValidateConfigInvalid(t *testing.T) {
	t.Parallel()

	config := &testConfig{
		Name:       "", // Invalid empty name
		APIVersion: "test/v1",
		Kind:       "TestCluster",
	}

	validationResult := validator.NewValidationResult("test.yaml")
	validationResult.AddError(validator.ValidationError{
		Field:         "name",
		Message:       "name is required",
		FixSuggestion: "provide a valid name",
	})

	mockVal := validator.NewMockValidator[*testConfig](t)
	mockVal.EXPECT().Validate(config).Return(validationResult)

	err := helpers.ValidateConfig(config, mockVal)
	assertValidationError(t, err, "name is required")
}

func testValidateConfigMultipleErrors(t *testing.T) {
	t.Parallel()

	config := &testConfig{
		Name:       "", // Invalid empty name
		APIVersion: "", // Invalid empty API version
		Kind:       "TestCluster",
	}

	validationResult := validator.NewValidationResult("test.yaml")
	validationResult.AddError(validator.ValidationError{
		Field:         "name",
		Message:       "name is required",
		FixSuggestion: "provide a valid name",
	})
	validationResult.AddError(validator.ValidationError{
		Field:         "apiVersion",
		Message:       "apiVersion is required",
		FixSuggestion: "provide a valid API version",
	})

	mockVal := validator.NewMockValidator[*testConfig](t)
	mockVal.EXPECT().Validate(config).Return(validationResult)

	err := helpers.ValidateConfig(config, mockVal)
	assertValidationError(t, err, "name is required", "apiVersion is required")
}
