package validator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidationError_Construction tests that ValidationError can be created with all fields.
func TestValidationError_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		message  string
		fix      string
		location FileLocation
		expected ValidationError
	}{
		{
			name:    "Complete validation error",
			field:   "metadata.name",
			message: "missing required field",
			fix:     "provide a valid name",
			location: FileLocation{
				FilePath: "config.yaml",
				Line:     5,
				Column:   10,
			},
			expected: ValidationError{
				Field:         "metadata.name",
				Message:       "missing required field",
				FixSuggestion: "provide a valid name",
				Location: FileLocation{
					FilePath: "config.yaml",
					Line:     5,
					Column:   10,
				},
			},
		},
		{
			name:    "Minimal validation error",
			field:   "",
			message: "general error",
			fix:     "fix the configuration",
			location: FileLocation{
				FilePath: "test.yaml",
				Line:     1,
				Column:   1,
			},
			expected: ValidationError{
				Field:         "",
				Message:       "general error",
				FixSuggestion: "fix the configuration",
				Location: FileLocation{
					FilePath: "test.yaml",
					Line:     1,
					Column:   1,
				},
			},
		},
		{
			name:     "Error with empty location",
			field:    "spec.invalid",
			message:  "invalid value",
			fix:      "use valid value",
			location: FileLocation{},
			expected: ValidationError{
				Field:         "spec.invalid",
				Message:       "invalid value",
				FixSuggestion: "use valid value",
				Location:      FileLocation{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidationError{
				Field:         tt.field,
				Message:       tt.message,
				FixSuggestion: tt.fix,
				Location:      tt.location,
			}

			assert.Equal(t, tt.expected, err)
			assert.Equal(t, tt.field, err.Field)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.fix, err.FixSuggestion)
			assert.Equal(t, tt.location, err.Location)
		})
	}
}

// TestValidationError_Error tests the Error() method of ValidationError.
func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name: "Error with field",
			err: ValidationError{
				Field:   "metadata.name",
				Message: "missing required field",
			},
			expected: "validation error in field 'metadata.name': missing required field",
		},
		{
			name: "Error without field",
			err: ValidationError{
				Field:   "",
				Message: "general error",
			},
			expected: "validation error: general error",
		},
		{
			name: "Error with empty message",
			err: ValidationError{
				Field:   "spec.distribution",
				Message: "",
			},
			expected: "validation error in field 'spec.distribution': ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidationError_ImplementsError tests that ValidationError implements the error interface.
func TestValidationError_ImplementsError(t *testing.T) {
	t.Parallel()

	err := ValidationError{
		Field:         "metadata.name",
		Message:       "missing required field",
		FixSuggestion: "provide a valid name",
		Location: FileLocation{
			FilePath: "config.yaml",
			Line:     5,
		},
	}

	// Should implement error interface
	var _ error = err

	// Error() method should return formatted message
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "metadata.name")
	assert.Contains(t, errorMsg, "missing required field")
}

// TestValidationError_JSONSerialization tests JSON marshaling and unmarshaling of ValidationError.
func TestValidationError_JSONSerialization(t *testing.T) {
	t.Parallel()

	original := ValidationError{
		Field:         "metadata.name",
		Message:       "missing required field",
		FixSuggestion: "provide a valid name",
		Location: FileLocation{
			FilePath: "config.yaml",
			Line:     5,
			Column:   10,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal from JSON
	var unmarshaled ValidationError
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Should be identical
	assert.Equal(t, original, unmarshaled)
	assert.Equal(t, original.Field, unmarshaled.Field)
	assert.Equal(t, original.Message, unmarshaled.Message)
	assert.Equal(t, original.FixSuggestion, unmarshaled.FixSuggestion)
	assert.Equal(t, original.Location.FilePath, unmarshaled.Location.FilePath)
	assert.Equal(t, original.Location.Line, unmarshaled.Location.Line)
	assert.Equal(t, original.Location.Column, unmarshaled.Location.Column)
}

// TestValidationResult_Construction tests that ValidationResult can be created with different states.
func TestValidationResult_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		valid    bool
		errors   []ValidationError
		expected ValidationResult
	}{
		{
			name:   "Valid result with no errors",
			valid:  true,
			errors: nil,
			expected: ValidationResult{
				Valid:  true,
				Errors: nil,
			},
		},
		{
			name:   "Valid result with empty errors slice",
			valid:  true,
			errors: []ValidationError{},
			expected: ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			},
		},
		{
			name:  "Invalid result with errors",
			valid: false,
			errors: []ValidationError{
				{
					Field:         "metadata.name",
					Message:       "missing required field",
					FixSuggestion: "provide a valid name",
					Location:      FileLocation{FilePath: "config.yaml", Line: 5},
				},
				{
					Field:         "spec.distribution",
					Message:       "invalid distribution",
					FixSuggestion: "use valid distribution",
					Location:      FileLocation{FilePath: "config.yaml", Line: 10},
				},
			},
			expected: ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{
						Field:         "metadata.name",
						Message:       "missing required field",
						FixSuggestion: "provide a valid name",
						Location:      FileLocation{FilePath: "config.yaml", Line: 5},
					},
					{
						Field:         "spec.distribution",
						Message:       "invalid distribution",
						FixSuggestion: "use valid distribution",
						Location:      FileLocation{FilePath: "config.yaml", Line: 10},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ValidationResult{
				Valid:  tt.valid,
				Errors: tt.errors,
			}

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.valid, result.Valid)
			assert.Equal(t, len(tt.errors), len(result.Errors))

			if tt.errors != nil {
				for i, expectedErr := range tt.errors {
					assert.Equal(t, expectedErr, result.Errors[i])
				}
			}
		})
	}
}

// TestValidationResult_Methods tests the methods of ValidationResult.
func TestValidationResult_Methods(t *testing.T) {
	t.Parallel()

	t.Run("HasErrors", func(t *testing.T) {
		t.Parallel()

		// Result with no errors
		resultWithoutErrors := ValidationResult{Valid: true, Errors: nil}
		assert.False(t, resultWithoutErrors.HasErrors())

		// Result with empty errors slice
		resultWithEmptyErrors := ValidationResult{Valid: true, Errors: []ValidationError{}}
		assert.False(t, resultWithEmptyErrors.HasErrors())

		// Result with errors
		resultWithErrors := ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Message: "error"}},
		}
		assert.True(t, resultWithErrors.HasErrors())
	})

	t.Run("HasWarnings", func(t *testing.T) {
		t.Parallel()

		// Result with no warnings
		resultWithoutWarnings := ValidationResult{Valid: true, Warnings: nil}
		assert.False(t, resultWithoutWarnings.HasWarnings())

		// Result with warnings
		resultWithWarnings := ValidationResult{
			Valid:    true,
			Warnings: []ValidationError{{Message: "warning"}},
		}
		assert.True(t, resultWithWarnings.HasWarnings())
	})

	t.Run("AddError", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{Valid: true}
		err := ValidationError{Field: "test", Message: "test error"}

		result.AddError(err)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
	})

	t.Run("AddWarning", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{Valid: true}
		warning := ValidationError{Field: "test", Message: "test warning"}

		result.AddWarning(warning)

		assert.True(t, result.Valid) // Should remain valid
		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, warning, result.Warnings[0])
	})
}

// TestValidationResult_JSONSerialization tests JSON marshaling and unmarshaling of ValidationResult.
func TestValidationResult_JSONSerialization(t *testing.T) {
	t.Parallel()

	original := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Field:         "metadata.name",
				Message:       "missing required field",
				FixSuggestion: "provide a valid name",
				Location:      FileLocation{FilePath: "config.yaml", Line: 5, Column: 10},
			},
			{
				Field:         "spec.distribution",
				Message:       "invalid distribution",
				FixSuggestion: "use valid distribution",
				Location:      FileLocation{FilePath: "config.yaml", Line: 10, Column: 15},
			},
		},
		Warnings: []ValidationError{
			{
				Field:         "metadata.labels",
				Message:       "recommended field missing",
				FixSuggestion: "consider adding labels",
				Location:      FileLocation{FilePath: "config.yaml", Line: 3},
			},
		},
		ConfigFile: "config.yaml",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal from JSON
	var unmarshaled ValidationResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Should be identical
	assert.Equal(t, original.Valid, unmarshaled.Valid)
	assert.Equal(t, original.ConfigFile, unmarshaled.ConfigFile)
	assert.Equal(t, len(original.Errors), len(unmarshaled.Errors))
	assert.Equal(t, len(original.Warnings), len(unmarshaled.Warnings))

	for i, originalErr := range original.Errors {
		unmarshaledErr := unmarshaled.Errors[i]
		assert.Equal(t, originalErr.Field, unmarshaledErr.Field)
		assert.Equal(t, originalErr.Message, unmarshaledErr.Message)
		assert.Equal(t, originalErr.FixSuggestion, unmarshaledErr.FixSuggestion)
		assert.Equal(t, originalErr.Location.FilePath, unmarshaledErr.Location.FilePath)
		assert.Equal(t, originalErr.Location.Line, unmarshaledErr.Location.Line)
		assert.Equal(t, originalErr.Location.Column, unmarshaledErr.Location.Column)
	}
}

// TestFileLocation_Construction tests that FileLocation can be created with different values.
func TestFileLocation_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		line     int
		column   int
		expected FileLocation
	}{
		{
			name:     "Complete file location",
			filePath: "/path/to/config.yaml",
			line:     42,
			column:   15,
			expected: FileLocation{
				FilePath: "/path/to/config.yaml",
				Line:     42,
				Column:   15,
			},
		},
		{
			name:     "Minimal file location",
			filePath: "config.yaml",
			line:     1,
			column:   1,
			expected: FileLocation{
				FilePath: "config.yaml",
				Line:     1,
				Column:   1,
			},
		},
		{
			name:     "Relative path",
			filePath: "./configs/cluster.yaml",
			line:     10,
			column:   5,
			expected: FileLocation{
				FilePath: "./configs/cluster.yaml",
				Line:     10,
				Column:   5,
			},
		},
		{
			name:     "Zero column",
			filePath: "test.yaml",
			line:     5,
			column:   0,
			expected: FileLocation{
				FilePath: "test.yaml",
				Line:     5,
				Column:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			location := FileLocation{
				FilePath: tt.filePath,
				Line:     tt.line,
				Column:   tt.column,
			}

			assert.Equal(t, tt.expected, location)
			assert.Equal(t, tt.filePath, location.FilePath)
			assert.Equal(t, tt.line, location.Line)
			assert.Equal(t, tt.column, location.Column)
		})
	}
}

// TestFileLocation_String tests the string representation of FileLocation.
func TestFileLocation_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		location FileLocation
		expected string
	}{
		{
			name: "Complete location",
			location: FileLocation{
				FilePath: "/path/to/config.yaml",
				Line:     42,
				Column:   15,
			},
			expected: "/path/to/config.yaml:42:15",
		},
		{
			name: "Location without column",
			location: FileLocation{
				FilePath: "config.yaml",
				Line:     5,
				Column:   0,
			},
			expected: "config.yaml:5",
		},
		{
			name: "Location without line or column",
			location: FileLocation{
				FilePath: "test.yaml",
				Line:     0,
				Column:   0,
			},
			expected: "test.yaml",
		},
		{
			name: "Large line numbers",
			location: FileLocation{
				FilePath: "large.yaml",
				Line:     999999,
				Column:   123,
			},
			expected: "large.yaml:999999:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.location.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFileLocation_JSONSerialization tests JSON marshaling and unmarshaling of FileLocation.
func TestFileLocation_JSONSerialization(t *testing.T) {
	t.Parallel()

	original := FileLocation{
		FilePath: "/path/to/config.yaml",
		Line:     42,
		Column:   15,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal from JSON
	var unmarshaled FileLocation
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Should be identical
	assert.Equal(t, original, unmarshaled)
	assert.Equal(t, original.FilePath, unmarshaled.FilePath)
	assert.Equal(t, original.Line, unmarshaled.Line)
	assert.Equal(t, original.Column, unmarshaled.Column)
}

// TestTypes_EdgeCases tests edge cases and boundary conditions.
func TestTypes_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Empty string fields", func(t *testing.T) {
		t.Parallel()

		err := ValidationError{
			Field:         "",
			Message:       "",
			FixSuggestion: "",
			Location: FileLocation{
				FilePath: "",
				Line:     0,
				Column:   0,
			},
		}

		// Should handle empty strings gracefully
		assert.NotPanics(t, func() {
			errMsg := err.Error()
			assert.NotEmpty(t, errMsg) // Should still produce some output
		})

		// Location String should handle empty path
		assert.NotPanics(t, func() {
			locationStr := err.Location.String()
			assert.Equal(t, "", locationStr) // Empty path, no line/column
		})
	})

	t.Run("Very long field names and messages", func(t *testing.T) {
		t.Parallel()

		longString := make([]byte, 10000)
		for i := range longString {
			longString[i] = byte('a' + (i % 26))
		}
		longStr := string(longString)

		err := ValidationError{
			Field:         longStr,
			Message:       longStr,
			FixSuggestion: longStr,
			Location: FileLocation{
				FilePath: longStr,
				Line:     999999999,
				Column:   999999999,
			},
		}

		// Should handle very long strings without panic
		assert.NotPanics(t, func() {
			_ = err.Error()
			_, jsonErr := json.Marshal(err)
			assert.NoError(t, jsonErr)
		})
	})

	t.Run("Negative line and column numbers", func(t *testing.T) {
		t.Parallel()

		location := FileLocation{
			FilePath: "test.yaml",
			Line:     -1,
			Column:   -5,
		}

		// Should handle negative numbers gracefully
		assert.NotPanics(t, func() {
			str := location.String()
			// With negative line, should just return file path
			assert.Equal(t, "test.yaml", str)
		})
	})

	t.Run("Zero values", func(t *testing.T) {
		t.Parallel()

		// Zero value ValidationError
		var zeroErr ValidationError
		assert.NotPanics(t, func() {
			_ = zeroErr.Error()
		})

		// Zero value ValidationResult
		var zeroResult ValidationResult
		assert.False(t, zeroResult.HasErrors())
		assert.False(t, zeroResult.HasWarnings())

		// Zero value FileLocation
		var zeroLocation FileLocation
		assert.NotPanics(t, func() {
			str := zeroLocation.String()
			assert.Equal(t, "", str)
		})
	})

	t.Run("Large slices", func(t *testing.T) {
		t.Parallel()

		// Create result with many errors
		result := ValidationResult{Valid: false}
		for i := range 10000 {
			result.AddError(ValidationError{
				Field:   fmt.Sprintf("field%d", i),
				Message: fmt.Sprintf("error %d", i),
			})
		}

		assert.False(t, result.Valid)
		assert.Equal(t, 10000, len(result.Errors))
		assert.True(t, result.HasErrors())

		// Should handle JSON serialization of large slices
		assert.NotPanics(t, func() {
			_, err := json.Marshal(result)
			assert.NoError(t, err)
		})
	})
}
