package metadata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	"github.com/devantler-tech/ksail-go/pkg/io/validator/metadata"
)

func TestValidateMetadata(t *testing.T) {
	t.Parallel()

	testValidMetadata(t)
	testMissingKind(t)
	testMissingAPIVersion(t)
	testMissingBothFields(t)
	testEmptyExpectedValues(t)
}

// testValidMetadata tests validation with valid metadata.
func testValidMetadata(t *testing.T) {
	t.Helper()

	t.Run("valid_metadata", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}

		metadata.ValidateMetadata(
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			result,
		)

		require.Empty(t, result.Errors, "Expected no errors for valid metadata")
	})
}

// testMissingKind tests validation with missing kind field.
func testMissingKind(t *testing.T) {
	t.Helper()

	t.Run("missing_kind", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}

		metadata.ValidateMetadata(
			"",
			"kind.x-k8s.io/v1alpha4",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			result,
		)

		require.Len(t, result.Errors, 1, "Expected 1 error for missing kind")
		validateKindError(t, result.Errors, "Cluster")
	})
}

// testMissingAPIVersion tests validation with missing apiVersion field.
func testMissingAPIVersion(t *testing.T) {
	t.Helper()

	t.Run("missing_api_version", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}

		metadata.ValidateMetadata(
			"Cluster",
			"",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			result,
		)

		require.Len(t, result.Errors, 1, "Expected 1 error for missing apiVersion")
		validateAPIVersionError(t, result.Errors, "kind.x-k8s.io/v1alpha4")
	})
}

// testMissingBothFields tests validation with both fields missing.
func testMissingBothFields(t *testing.T) {
	t.Helper()

	t.Run("missing_both", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}

		metadata.ValidateMetadata(
			"",
			"",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			result,
		)

		require.Len(t, result.Errors, 2, "Expected 2 errors for missing both fields")
		validateKindError(t, result.Errors, "Cluster")
		validateAPIVersionError(t, result.Errors, "kind.x-k8s.io/v1alpha4")
	})
}

// testEmptyExpectedValues tests validation with empty expected values.
func testEmptyExpectedValues(t *testing.T) {
	t.Helper()

	t.Run("empty_expected_values", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}

		metadata.ValidateMetadata("", "", "", "", result)

		require.Len(t, result.Errors, 2, "Expected 2 errors for empty expected values")
		validateKindError(t, result.Errors, "")
		validateAPIVersionError(t, result.Errors, "")
	})
}

// validateKindError validates that a kind error exists with expected content.
func validateKindError(t *testing.T, errors []validator.ValidationError, expectedKind string) {
	t.Helper()

	kindError := findErrorByField(errors, "kind")
	require.NotNil(t, kindError, "Should have kind error")
	assert.Equal(t, "kind is required", kindError.Message)
	assert.Equal(t, expectedKind, kindError.ExpectedValue)
	assert.Equal(t, "Set kind to '"+expectedKind+"'", kindError.FixSuggestion)
}

// validateAPIVersionError validates that an apiVersion error exists with expected content.
func validateAPIVersionError(
	t *testing.T,
	errors []validator.ValidationError,
	expectedAPIVersion string,
) {
	t.Helper()

	apiVersionError := findErrorByField(errors, "apiVersion")
	require.NotNil(t, apiVersionError, "Should have apiVersion error")
	assert.Equal(t, "apiVersion is required", apiVersionError.Message)
	assert.Equal(t, expectedAPIVersion, apiVersionError.ExpectedValue)
	assert.Equal(t, "Set apiVersion to '"+expectedAPIVersion+"'", apiVersionError.FixSuggestion)
}

func TestValidateNilConfig(t *testing.T) {
	t.Parallel()

	testNilConfig(t)
	testValidConfigs(t)
	testEmptyConfigType(t)
}

// testNilConfig tests validation with nil config.
func testNilConfig(t *testing.T) {
	t.Helper()

	t.Run("nil_config", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}
		isNil := metadata.ValidateNilConfig(nil, "Kind", result)

		assert.True(t, isNil, "Expected isNil=true for nil config")
		require.Len(t, result.Errors, 1, "Expected 1 error for nil config")
		assert.Equal(t, "config", result.Errors[0].Field)
		assert.Equal(t, "configuration is nil", result.Errors[0].Message)
		assert.Contains(t, result.Errors[0].FixSuggestion, "Kind")
	})
}

// testValidConfigs tests validation with various valid config types.
func testValidConfigs(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		config     any
		configType string
	}{
		{
			name:       "valid_string_config",
			config:     "test",
			configType: "Kind",
		},
		{
			name:       "valid_struct_config",
			config:     struct{ Name string }{Name: "test"},
			configType: "K3d",
		},
		{
			name:       "valid_pointer_config",
			config:     &struct{ Name string }{Name: "test"},
			configType: "EKS",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := &validator.ValidationResult{}
			isNil := metadata.ValidateNilConfig(test.config, test.configType, result)

			assert.False(t, isNil, "Expected isNil=false for non-nil config")
			assert.Empty(t, result.Errors, "Expected no errors for non-nil config")
		})
	}
}

// testEmptyConfigType tests validation with empty config type.
func testEmptyConfigType(t *testing.T) {
	t.Helper()

	t.Run("empty_string_config_type", func(t *testing.T) {
		t.Parallel()

		result := &validator.ValidationResult{}
		isNil := metadata.ValidateNilConfig(nil, "", result)

		assert.True(t, isNil, "Expected isNil=true for nil config with empty type")
		require.Len(t, result.Errors, 1, "Expected 1 error for nil config")
		assert.Equal(t, "config", result.Errors[0].Field)
		assert.Equal(t, "configuration is nil", result.Errors[0].Message)
		assert.Contains(t, result.Errors[0].FixSuggestion, "")
	})
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
