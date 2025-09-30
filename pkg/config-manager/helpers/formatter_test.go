package helpers_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestFormatValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *validator.ValidationResult
		expected []string
	}{
		{
			name: "single error",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{Field: "name", Message: "is required"},
				},
			},
			expected: []string{"'name' is required"},
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
			expected: []string{
				"'name' is required",
				"'version' is invalid",
			},
		},
		{
			name: "no errors",
			result: &validator.ValidationResult{
				Valid:  true,
				Errors: []validator.ValidationError{},
			},
			expected: []string{},
		},
		{
			name: "error with location and fix suggestion",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{
						Field:   "spec.distribution",
						Message: "value is invalid",
						Location: validator.FileLocation{
							FilePath: "ksail.yaml",
							Line:     15,
						},
						FixSuggestion: "use one of: Kind, K3d",
					},
				},
			},
			expected: []string{
				"'spec.distribution' value is invalid\n   in: ksail.yaml:15\n   fix: use one of: Kind, K3d",
			},
		},
		{
			name: "error with location only",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{
						Field:   "metadata.name",
						Message: "contains invalid characters",
						Location: validator.FileLocation{
							FilePath: "config.yaml",
							Line:     3,
						},
					},
				},
			},
			expected: []string{
				"'metadata.name' contains invalid characters\n   in: config.yaml:3",
			},
		},
		{
			name: "error with fix suggestion only",
			result: &validator.ValidationResult{
				Valid: false,
				Errors: []validator.ValidationError{
					{
						Field:         "spec.replicas",
						Message:       "value is too high",
						FixSuggestion: "reduce to 10 or less",
					},
				},
			},
			expected: []string{
				"'spec.replicas' value is too high\n   fix: reduce to 10 or less",
			},
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
			expected: []string{"- warning: 'deprecated' field is deprecated"},
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
				"- warning: 'deprecated' field is deprecated",
				"- warning: 'unused' field is unused",
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
		{
			name: "warning with location and fix suggestion",
			result: &validator.ValidationResult{
				Valid: true,
				Warnings: []validator.ValidationError{
					{
						Field:   "spec.distribution",
						Message: "using deprecated value",
						Location: validator.FileLocation{
							FilePath: "ksail.yaml",
							Line:     10,
						},
						FixSuggestion: "use 'Kind' instead of 'kind'",
					},
				},
			},
			expected: []string{
				"- warning: 'spec.distribution' using deprecated value\n   in: ksail.yaml:10\n   fix: use 'Kind' instead of 'kind'",
			},
		},
		{
			name: "warning with location only",
			result: &validator.ValidationResult{
				Valid: true,
				Warnings: []validator.ValidationError{
					{
						Field:   "metadata.name",
						Message: "name is quite long",
						Location: validator.FileLocation{
							FilePath: "config.yaml",
							Line:     5,
						},
					},
				},
			},
			expected: []string{
				"- warning: 'metadata.name' name is quite long\n   in: config.yaml:5",
			},
		},
		{
			name: "warning with fix suggestion only",
			result: &validator.ValidationResult{
				Valid: true,
				Warnings: []validator.ValidationError{
					{
						Field:         "spec.timeout",
						Message:       "value is very high",
						FixSuggestion: "reduce to 5m or less",
					},
				},
			},
			expected: []string{
				"- warning: 'spec.timeout' value is very high\n   fix: reduce to 5m or less",
			},
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

// Test config struct for ValidateConfig tests.
type testConfig struct {
	Name       string `yaml:"name"`
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// Mock validator for testing ValidateConfig.
type mockValidator struct {
	shouldReturnValid bool
	validationErrors  []validator.ValidationError
}

func (m *mockValidator) Validate(_ *testConfig) *validator.ValidationResult {
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

	tests := []struct {
		name      string
		config    *testConfig
		validator *mockValidator
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid config passes validation",
			config: &testConfig{
				Name:       "test-cluster",
				APIVersion: "test/v1",
				Kind:       "TestCluster",
			},
			validator: &mockValidator{shouldReturnValid: true},
			wantErr:   false,
		},
		{
			name: "invalid config fails validation",
			config: &testConfig{
				Name:       "",
				APIVersion: "test/v1",
				Kind:       "TestCluster",
			},
			validator: &mockValidator{
				shouldReturnValid: false,
				validationErrors: []validator.ValidationError{
					{
						Field:         "name",
						Message:       "name is required",
						FixSuggestion: "provide a valid name",
					},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "multiple validation errors",
			config: &testConfig{
				Name:       "",
				APIVersion: "",
				Kind:       "TestCluster",
			},
			validator: &mockValidator{
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
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := helpers.ValidateConfig(tt.config, tt.validator)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.ErrorIs(t, err, helpers.ErrConfigurationValidationFailed)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
