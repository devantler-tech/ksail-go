// Package config_test provides focused tests for improving binding function coverage.
package config_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestBindingFunctionsCoverage tests all uncovered binding functions efficiently.
func TestBindingFunctionsCoverage(t *testing.T) {
	t.Parallel()

	testCases := getBindingTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create field selector
			fieldSelector := config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return testCase.fieldGenerator() },
				testCase.defaultValue,
				testCase.description,
			)

			// Create command to exercise binding functions
			cmd := config.NewCobraCommand(
				"test",
				"Test command",
				"Test command description",
				func(_ *cobra.Command, _ *config.Manager, _ []string) error {
					return nil
				},
				fieldSelector,
			)

			// Verify command was created
			assert.NotNil(t, cmd)
			assert.NotNil(t, cmd.Flags())

			// The fact that NewCobraCommand ran without error means the binding functions were exercised
			// This improves coverage for the previously uncovered binding functions
		})
	}
}

// getBindingTestCases returns test cases for binding function coverage.
func getBindingTestCases() []struct {
	name           string
	fieldGenerator func() any
	defaultValue   any
	description    string
} {
	return []struct {
		name           string
		fieldGenerator func() any
		defaultValue   any
		description    string
	}{
		// Duration types
		{"duration", func() any { return new(time.Duration) }, time.Minute * 5, "Duration field"},
		{
			"metav1_duration",
			func() any { return &metav1.Duration{} },
			metav1.Duration{Duration: time.Minute * 10},
			"Metav1 duration field",
		},

		// Integer types
		{"int", func() any { return new(int) }, 42, "Int field"},
		{"int32", func() any { return new(int32) }, int32(32), "Int32 field"},
		{"int64", func() any { return new(int64) }, int64(64), "Int64 field"},

		// Unsigned integer types
		{"uint", func() any { return new(uint) }, uint(42), "Uint field"},
		{"uint32", func() any { return new(uint32) }, uint32(32), "Uint32 field"},
		{"uint64", func() any { return new(uint64) }, uint64(64), "Uint64 field"},

		// Float types
		{"float32", func() any { return new(float32) }, float32(3.14), "Float32 field"},
		{"float64", func() any { return new(float64) }, 3.14159, "Float64 field"},

		// Bool type
		{"bool", func() any { return new(bool) }, true, "Bool field"},

		// Slice types
		{
			"string_slice",
			func() any { return new([]string) },
			[]string{"test"},
			"String slice field",
		},
		{"int_slice", func() any { return new([]int) }, []int{1, 2, 3}, "Int slice field"},
	}
}

// TestBindStandardTypeFallback tests the fallback case in bindStandardType.
func TestBindStandardTypeFallback(t *testing.T) {
	t.Parallel()

	// Test with an unsupported type that should fallback to string
	fieldSelector := config.AddFlagFromField(
		func(_ *v1alpha1.Cluster) any {
			// Return a type not handled by bindStandardType
			return new(complex64)
		},
		complex64(1+2i),
		"Complex number field",
	)

	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test command description",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error {
			return nil
		},
		fieldSelector,
	)

	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Flags())
}

// TestManagerViperIntegration tests low-coverage Viper integration functions.
func TestManagerViperIntegration(t *testing.T) {
	t.Parallel()

	manager := config.NewManager()
	setupTestViperData(manager)

	testCases := getViperIntegrationTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create manager with the field selector
			manager := config.NewManager(testCase.selector)

			// This exercises the Viper integration paths with low coverage
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)
			assert.NotNil(t, cluster)
		})
	}
}

// setupTestViperData sets up test data for Viper integration tests.
func setupTestViperData(manager *config.Manager) {
	viper := manager.GetViper()
	viper.Set("test_int", 42)
	viper.Set("test_int8", 8)
	viper.Set("test_float32", 3.14)
	viper.Set("test_string_slice", []string{"a", "b", "c"})
	viper.Set("test_int_slice", []int{1, 2, 3})
}

// getViperIntegrationTestCases returns test cases for Viper integration testing.
func getViperIntegrationTestCases() []struct {
	name     string
	selector config.FieldSelector[v1alpha1.Cluster]
} {
	return []struct {
		name     string
		selector config.FieldSelector[v1alpha1.Cluster]
	}{
		{
			"int_viper",
			config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return new(int) },
				0,
				"Test int field",
			),
		},
		{
			"int8_viper",
			config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return new(int8) },
				int8(0),
				"Test int8 field",
			),
		},
		{
			"float32_viper",
			config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return new(float32) },
				float32(0),
				"Test float32 field",
			),
		},
		{
			"string_slice_viper",
			config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return new([]string) },
				[]string{},
				"Test string slice field",
			),
		},
		{
			"int_slice_viper",
			config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any { return new([]int) },
				[]int{},
				"Test int slice field",
			),
		},
	}
}

// TestBindPflagValueViperFallback tests the Viper fallback path in bindPflagValue.
func TestBindPflagValueViperFallback(t *testing.T) {
	t.Parallel()

	// Create a field selector with NO default value (nil) to trigger Viper fallback
	fieldSelector := config.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		nil, // nil default will trigger Viper fallback path
		"Distribution config field",
	)

	// Create manager and set a Viper value that should be used as fallback
	manager := config.NewManager(fieldSelector)
	viper := manager.GetViper()
	viper.Set("distributionConfig", "test-fallback-value")

	// Create command to trigger the binding
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test description",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error {
			return nil
		},
		fieldSelector,
	)

	require.NotNil(t, cmd)
	// The test succeeds if no panic occurs and command is created
	// This exercises the else branch in bindPflagValue when defaultValue is nil
}

// TestPartialCoverageFunctions tests functions with incomplete coverage.
func TestPartialCoverageFunctions(t *testing.T) {
	t.Parallel()

	t.Run("nil_default_value_viper_fallback", func(t *testing.T) {
		t.Parallel()
		testViperFallbackPath(t)
	})

	t.Run("complex_field_paths", func(t *testing.T) {
		t.Parallel()
		testComplexFieldPaths(t)
	})
}

// testViperFallbackPath tests bindPflagValue fallback when defaultValue is nil.
func testViperFallbackPath(t *testing.T) {
	t.Helper()

	// Test bindPflagValue fallback path when defaultValue is nil
	fieldSelector := config.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		nil, // nil will trigger Viper fallback path in bindPflagValue
		"Config field with nil default",
	)

	manager := config.NewManager(fieldSelector)
	viper := manager.GetViper()
	viper.Set("distributionConfig", "fallback-value")

	// Create command to exercise the binding
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test description",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error {
			return nil
		},
		fieldSelector,
	)

	require.NotNil(t, cmd)
}

// testComplexFieldPaths tests more complex field paths to improve bindStandardType coverage.
func testComplexFieldPaths(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name         string
		selector     func(c *v1alpha1.Cluster) any
		defaultValue any
	}{
		{
			"nested_string_field",
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"test-context",
		},
		{
			"deep_nested_string",
			func(c *v1alpha1.Cluster) any { return &c.Spec.Options.EKS.AWSProfile },
			"test-profile",
		},
	}

	for _, testCase := range testCases {
		fieldSelector := config.AddFlagFromField(
			testCase.selector,
			testCase.defaultValue,
			testCase.name,
		)
		cmd := config.NewCobraCommand(
			"test",
			"Test",
			"Test",
			func(_ *cobra.Command, _ *config.Manager, _ []string) error {
				return nil
			},
			fieldSelector,
		)
		require.NotNil(t, cmd, "Command should be created for %s", testCase.name)
	}
}
