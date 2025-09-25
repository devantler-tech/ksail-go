package metadata_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	"github.com/devantler-tech/ksail-go/pkg/validator/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		kind               string
		apiVersion         string
		expectedKind       string
		expectedAPIVersion string
		expectErrors       int
		expectedFields     []string
	}{
		{
			name:               "valid_metadata",
			kind:               "Cluster",
			apiVersion:         "kind.x-k8s.io/v1alpha4",
			expectedKind:       "Cluster",
			expectedAPIVersion: "kind.x-k8s.io/v1alpha4",
			expectErrors:       0,
			expectedFields:     nil,
		},
		{
			name:               "missing_kind",
			kind:               "",
			apiVersion:         "kind.x-k8s.io/v1alpha4",
			expectedKind:       "Cluster",
			expectedAPIVersion: "kind.x-k8s.io/v1alpha4",
			expectErrors:       1,
			expectedFields:     []string{"kind"},
		},
		{
			name:               "missing_api_version",
			kind:               "Cluster",
			apiVersion:         "",
			expectedKind:       "Cluster",
			expectedAPIVersion: "kind.x-k8s.io/v1alpha4",
			expectErrors:       1,
			expectedFields:     []string{"apiVersion"},
		},
		{
			name:               "missing_both",
			kind:               "",
			apiVersion:         "",
			expectedKind:       "Cluster",
			expectedAPIVersion: "kind.x-k8s.io/v1alpha4",
			expectErrors:       2,
			expectedFields:     []string{"kind", "apiVersion"},
		},
		{
			name:               "empty_expected_values",
			kind:               "",
			apiVersion:         "",
			expectedKind:       "",
			expectedAPIVersion: "",
			expectErrors:       2,
			expectedFields:     []string{"kind", "apiVersion"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := &validator.ValidationResult{}

			metadata.ValidateMetadata(
				test.kind,
				test.apiVersion,
				test.expectedKind,
				test.expectedAPIVersion,
				result,
			)

			require.Len(t, result.Errors, test.expectErrors,
				"Expected %d errors, got %d", test.expectErrors, len(result.Errors))

			for index, expectedField := range test.expectedFields {
				assert.Equal(t, expectedField, result.Errors[index].Field,
					"Error %d field mismatch", index)
				assert.NotEmpty(t, result.Errors[index].Message,
					"Error %d should have a message", index)
				assert.NotEmpty(t, result.Errors[index].FixSuggestion,
					"Error %d should have a fix suggestion", index)
			}

			// Validate specific error content for missing kind
			if test.kind == "" {
				kindError := findErrorByField(result.Errors, "kind")
				require.NotNil(t, kindError, "Should have kind error")
				assert.Equal(t, "kind is required", kindError.Message)
				assert.Equal(t, test.expectedKind, kindError.ExpectedValue)
				assert.Equal(t, "Set kind to '"+test.expectedKind+"'", kindError.FixSuggestion)
			}

			// Validate specific error content for missing apiVersion
			if test.apiVersion == "" {
				apiVersionError := findErrorByField(result.Errors, "apiVersion")
				require.NotNil(t, apiVersionError, "Should have apiVersion error")
				assert.Equal(t, "apiVersion is required", apiVersionError.Message)
				assert.Equal(t, test.expectedAPIVersion, apiVersionError.ExpectedValue)
				assert.Equal(
					t,
					"Set apiVersion to '"+test.expectedAPIVersion+"'",
					apiVersionError.FixSuggestion,
				)
			}
		})
	}
}

func TestValidateNilConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     interface{}
		configType string
		expectNil  bool
	}{
		{
			name:       "nil_config",
			config:     nil,
			configType: "Kind",
			expectNil:  true,
		},
		{
			name:       "valid_string_config",
			config:     "test",
			configType: "Kind",
			expectNil:  false,
		},
		{
			name:       "valid_struct_config",
			config:     struct{ Name string }{Name: "test"},
			configType: "K3d",
			expectNil:  false,
		},
		{
			name:       "valid_pointer_config",
			config:     &struct{ Name string }{Name: "test"},
			configType: "EKS",
			expectNil:  false,
		},
		{
			name:       "empty_string_config_type",
			config:     nil,
			configType: "",
			expectNil:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := &validator.ValidationResult{}

			isNil := metadata.ValidateNilConfig(test.config, test.configType, result)

			assert.Equal(t, test.expectNil, isNil,
				"Expected isNil=%v, got %v", test.expectNil, isNil)

			if test.expectNil {
				require.Len(t, result.Errors, 1, "Expected 1 error for nil config")
				assert.Equal(t, "config", result.Errors[0].Field)
				assert.Equal(t, "configuration is nil", result.Errors[0].Message)
				assert.Contains(t, result.Errors[0].FixSuggestion, test.configType)
			} else {
				assert.Empty(t, result.Errors, "Expected no errors for non-nil config")
			}
		})
	}
}

// Helper function to find an error by field name.
func findErrorByField(errors []validator.ValidationError, field string) *validator.ValidationError {
	for i := range errors {
		if errors[i].Field == field {
			return &errors[i]
		}
	}

	return nil
}
