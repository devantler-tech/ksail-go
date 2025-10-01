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
			expected: []string{"error: is required\n  in: 'name'"},
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
				"error: is required\n  in: 'name'",
				"error: is invalid\n  in: 'version'",
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
				"error: value is invalid\n  in: ksail.yaml:15 'spec.distribution'\n  fix: use one of: Kind, K3d",
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
				"error: contains invalid characters\n  in: config.yaml:3 'metadata.name'",
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
				"error: value is too high\n  in: 'spec.replicas'\n  fix: reduce to 10 or less",
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
			expected: []string{"warning: field is deprecated\n  in: 'deprecated'"},
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
				"warning: field is deprecated\n  in: 'deprecated'",
				"warning: field is unused\n  in: 'unused'",
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
				"warning: using deprecated value\n  in: ksail.yaml:10 'spec.distribution'\n  fix: use 'Kind' instead of 'kind'",
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
				"warning: name is quite long\n  in: config.yaml:5 'metadata.name'",
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
				"warning: value is very high\n  in: 'spec.timeout'\n  fix: reduce to 5m or less",
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
