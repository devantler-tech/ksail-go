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

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected string
	}{
		{
			name: "single error without fix suggestion",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required"},
				},
			},
			expected: "name: is required",
		},
		{
			name: "single error with fix suggestion",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required", FixSuggestion: "add name field"},
				},
			},
			expected: "name: is required (add name field)",
		},
		{
			name: "multiple errors",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required", FixSuggestion: "add name field"},
					{Field: "version", Message: "is invalid"},
				},
			},
			expected: "name: is required (add name field); version: is invalid",
		},
		{
			name: "no errors",
			result: &validator.ValidationResult{
				Valid:  true,
				Errors: []validator.ValidationError{},
			},
			expected: "unknown validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := helpers.FormatValidationErrors(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidationErrorsMultiline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected string
	}{
		{
			name: "single error",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required"},
				},
			},
			expected: "  - name: is required\n",
		},
		{
			name: "multiple errors",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required"},
					{Field: "version", Message: "is invalid"},
				},
			},
			expected: "  - name: is required\n  - version: is invalid\n",
		},
		{
			name: "no errors",
			result: &validator.ValidationResult{
				Valid:  true,
				Errors: []validator.ValidationError{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := helpers.FormatValidationErrorsMultiline(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidationFixSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected []string
	}{
		{
			name: "single fix suggestion",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required", FixSuggestion: "add name field"},
				},
			},
			expected: []string{"    Fix: add name field"},
		},
		{
			name: "multiple fix suggestions",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required", FixSuggestion: "add name field"},
					{Field: "version", Message: "is invalid", FixSuggestion: "use valid version"},
				},
			},
			expected: []string{"    Fix: add name field", "    Fix: use valid version"},
		},
		{
			name: "mixed errors with and without fix suggestions",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required", FixSuggestion: "add name field"},
					{Field: "version", Message: "is invalid"},
				},
			},
			expected: []string{"    Fix: add name field"},
		},
		{
			name: "no fix suggestions",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required"},
				},
			},
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

// Mock validator for testing ValidateConfig
type mockValidator struct {
	shouldReturnValid bool
	validationErrors  []validator.ValidationError
}

func (m *mockValidator) Validate(config *testConfig) *validator.ValidationResult {
	result := validator.NewValidationResult("test.yaml")

	if !m.shouldReturnValid {
		for _, err := range m.validationErrors {
			result.AddError(err)
		}
	}

	return result
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config passes validation", func(t *testing.T) {
		t.Parallel()

		config := &testConfig{
			Name:       "test-cluster",
			APIVersion: "test/v1",
			Kind:       "TestCluster",
		}

		mockVal := &mockValidator{shouldReturnValid: true}

		err := helpers.ValidateConfig(config, mockVal)
		assert.NoError(t, err)
	})

	t.Run("invalid config fails validation", func(t *testing.T) {
		t.Parallel()

		config := &testConfig{
			Name:       "", // Invalid empty name
			APIVersion: "test/v1",
			Kind:       "TestCluster",
		}

		mockVal := &mockValidator{
			shouldReturnValid: false,
			validationErrors: []validator.ValidationError{
				{
					Field:         "name",
					Message:       "name is required",
					FixSuggestion: "provide a valid name",
				},
			},
		}

		err := helpers.ValidateConfig(config, mockVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		t.Parallel()

		config := &testConfig{
			Name:       "", // Invalid empty name
			APIVersion: "", // Invalid empty API version
			Kind:       "TestCluster",
		}

		mockVal := &mockValidator{
			shouldReturnValid: false,
			validationErrors: []validator.ValidationError{
				{
					Field:         "name",
					Message:       "name is required",
					FixSuggestion: "provide a valid name",
				},
				{
					Field:         "apiVersion",
					Message:       "apiVersion is required",
					FixSuggestion: "provide a valid API version",
				},
			},
		}

		err := helpers.ValidateConfig(config, mockVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
		assert.Contains(t, err.Error(), "name is required")
		assert.Contains(t, err.Error(), "apiVersion is required")
	})
}
