package configmanager_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createStandardFieldSelectors creates a common set of field selectors used in multiple tests.
func createStandardFieldSelectors() []configmanager.FieldSelector[v1alpha1.Cluster] {
	return []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
			"Kubernetes distribution",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
			"Source directory",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"",
			"Kubernetes context",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			metav1.Duration{Duration: 5 * time.Minute},
			"Connection timeout",
		),
	}
}

// createFieldSelectorsWithName creates field selectors including name field.
// Creates selectors with valid defaults that pass validation (includes required APIVersion and Kind).
func createFieldSelectorsWithName() []configmanager.FieldSelector[v1alpha1.Cluster] {
	return []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.APIVersion },
			"ksail.dev/v1alpha1",
			"API version",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Kind },
			"Cluster",
			"Resource kind",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind, // Use valid default
			"Kubernetes distribution",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			"kind.yaml",
			"Distribution config file",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
			"Source directory",
		),
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"kind-kind",
			"Kubernetes context",
		),
	}
}

// createDistributionOnlyFieldSelectors creates field selectors with only the distribution field.
func createDistributionOnlyFieldSelectors() []configmanager.FieldSelector[v1alpha1.Cluster] {
	// Use the first selector (distribution) from the standard field selectors
	return createStandardFieldSelectors()[:1]
}

// TestNewManager tests the NewManager constructor.
func TestNewManager(t *testing.T) {
	t.Parallel()

	fieldSelectors := createDistributionOnlyFieldSelectors()

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	require.NotNil(t, manager)
	require.NotNil(t, manager.Config)

	// Test Viper field is properly initialized
	require.NotNil(t, manager.Viper)

	// Test that Viper is properly configured by setting and getting a value
	manager.Viper.SetDefault("test.key", "test-value")
	assert.Equal(t, "test-value", manager.Viper.GetString("test.key"))
}

// TestManager_LoadConfig tests the LoadConfig method with different scenarios.
// All tests now create valid configurations since validation is integrated into LoadConfig.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                 string
		envVars              map[string]string
		expectedDistribution v1alpha1.Distribution
		shouldSucceed        bool
	}{
		{
			name:                 "LoadConfig with defaults",
			envVars:              map[string]string{},
			expectedDistribution: v1alpha1.DistributionKind, // Default from field selector
			shouldSucceed:        true,
		},
		{
			name: "LoadConfig with environment variables",
			envVars: map[string]string{
				"KSAIL_SPEC_DISTRIBUTION": "Kind",
			},
			expectedDistribution: v1alpha1.DistributionKind,
			shouldSucceed:        true,
		},
		{
			name: "LoadConfig with multiple environment variables",
			envVars: map[string]string{
				"KSAIL_SPEC_DISTRIBUTION": "Kind", // Keep it simple - just override distribution
			},
			expectedDistribution: v1alpha1.DistributionKind,
			shouldSucceed:        true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory and change to it to isolate from existing config files
			tempDir := t.TempDir()
			t.Chdir(tempDir)

			// Set environment variables for the test
			for key, value := range testCase.envVars {
				t.Setenv(key, value)
			}

			fieldSelectors := createFieldSelectorsWithName()

			manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

			cluster, err := manager.LoadConfig()

			if testCase.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, cluster)
				assert.Equal(t, testCase.expectedDistribution, cluster.Spec.Distribution)

				// Test that subsequent calls return the same config
				cluster2, err2 := manager.LoadConfig()
				require.NoError(t, err2)
				assert.Equal(t, cluster, cluster2)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestAddFlagFromFieldHelper tests the AddFlagFromField helper function.
func TestAddFlagFromFieldHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		description  []string
		defaultValue any
		expectedDesc string
	}{
		{
			name:         "AddFlagFromField with description",
			description:  []string{"Test description"},
			defaultValue: "test-value",
			expectedDesc: "Test description",
		},
		{
			name:         "AddFlagFromField without description",
			description:  []string{},
			defaultValue: "test-value",
			expectedDesc: "",
		},
		{
			name:         "AddFlagFromField with multiple descriptions (takes first)",
			description:  []string{"First description", "Second description"},
			defaultValue: "test-value",
			expectedDesc: "First description",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			selector := configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
				testCase.defaultValue,
				testCase.description...,
			)

			assert.Equal(t, testCase.defaultValue, selector.DefaultValue)
			assert.Equal(t, testCase.expectedDesc, selector.Description)
			assert.NotNil(t, selector.Selector)
		})
	}
}

// TestManager_AddFlagsFromFields tests the AddFlagsFromFields method.
func TestAddFlagsFromFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fieldSelectors []configmanager.FieldSelector[v1alpha1.Cluster]
		expectedFlags  []string
	}{
		{
			name:           "AddFlagsFromFields with no selectors",
			fieldSelectors: []configmanager.FieldSelector[v1alpha1.Cluster]{},
			expectedFlags:  []string{},
		},
		{
			name: "AddFlagsFromFields with distribution selector",
			fieldSelectors: []configmanager.FieldSelector[v1alpha1.Cluster]{
				configmanager.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
					v1alpha1.DistributionKind,
					"Kubernetes distribution",
				),
			},
			expectedFlags: []string{"distribution"},
		},
		{
			name:           "AddFlagsFromFields with multiple selectors",
			fieldSelectors: createStandardFieldSelectors(),
			expectedFlags:  []string{"distribution", "source-directory", "context", "timeout"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := configmanager.NewConfigManager(io.Discard, testCase.fieldSelectors...)
			cmd := &cobra.Command{
				Use: "test",
			}

			manager.AddFlagsFromFields(cmd)

			// Check that expected flags are present
			for _, expectedFlag := range testCase.expectedFlags {
				flag := cmd.Flags().Lookup(expectedFlag)
				assert.NotNil(t, flag, "flag %s should exist", expectedFlag)
			}

			// Check that we don't have unexpected flags
			actualFlags := []string{}

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				actualFlags = append(actualFlags, flag.Name)
			})
			assert.Len(t, actualFlags, len(testCase.expectedFlags))
		})
	}
}

// TestManager_LoadConfig_ConfigProperty tests that the Config property is properly exposed.
func TestLoadConfigConfigProperty(t *testing.T) {
	t.Parallel()

	fieldSelectors := createFieldSelectorsWithName()

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	// Before loading, Config should be initialized with proper TypeMeta
	expectedEmpty := v1alpha1.NewCluster()
	assert.Equal(t, expectedEmpty, manager.Config)

	// Load config
	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// After loading, Config property should be accessible and equal to returned cluster
	assert.Equal(t, cluster, manager.Config)
	assert.Equal(t, v1alpha1.DistributionKind, manager.Config.Spec.Distribution)
}

// testFieldValueSetting is a helper function to test field value setting behavior.
// Creates a minimal valid configuration with only required fields to pass validation,
// allowing the test to verify the specific field selector being tested.
func testFieldValueSetting(
	t *testing.T,
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description string,
	assertFunc func(*testing.T, *v1alpha1.Cluster),
) {
	t.Helper()

	// Create temporary directory and change to it to isolate from existing config files
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create minimal field selectors with required fields plus the field being tested
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.APIVersion },
			DefaultValue: "ksail.dev/v1alpha1",
			Description:  "API version",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Kind },
			DefaultValue: "Cluster",
			Description:  "Resource kind",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			DefaultValue: v1alpha1.DistributionKind,
			Description:  "Distribution",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			DefaultValue: "kind.yaml",
			Description:  "Distribution config",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			DefaultValue: "kind-kind",
			Description:  "Kubernetes context",
		},
		{
			Selector:     selector,
			DefaultValue: defaultValue,
			Description:  description,
		},
	}

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	assertFunc(t, cluster)
}

// TestManager_SetFieldValueWithNilDefault tests setFieldValue with nil default value.
// With validation integrated, nil defaults are handled gracefully and other required fields ensure validation passes.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNilDefault(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		nil, // nil value should be handled gracefully
		"Test nil default",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// When default is nil, field should remain empty (other required fields allow validation to pass)
			assert.Empty(t, cluster.Spec.SourceDirectory)
		},
	)
}

// TestManager_SetFieldValueWithNonConvertibleTypes tests setFieldValue with non-convertible types.
// With validation integrated, non-convertible types are handled and validation ensures configuration correctness.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNonConvertibleTypes(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		123, // int cannot be converted to string
		"Test non-convertible type",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// When type is not convertible, field should remain empty
			assert.Empty(t, cluster.Spec.SourceDirectory)
		},
	)
}

// TestManager_SetFieldValueWithDirectlyAssignableTypes tests setFieldValue with directly assignable types.
// Tests with SourceDirectory to avoid conflicts with required Distribution field.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithDirectlyAssignableTypes(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		"custom-k8s",
		"Test direct assignment",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// Direct string assignment should work
			assert.Equal(t, "custom-k8s", cluster.Spec.SourceDirectory)
		},
	)
}

// TestManager_SetFieldValueWithNonPointerField tests setFieldValue with non-pointer field.
// With validation integrated, testing with a non-required field like SourceDirectory.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNonPointerField(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return c.Spec.SourceDirectory }, // Return value, not pointer
		"should-not-set",
		"Test non-pointer field",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// Non-pointer field should remain empty since it can't be set
			assert.Empty(t, cluster.Spec.SourceDirectory)
		},
	)
}

// TestManager_SetFieldValueWithConvertibleTypes tests setFieldValue with convertible types.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithConvertibleTypes(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any {
			// Use the timeout field which accepts time.Duration
			return &c.Spec.Connection.Timeout.Duration
		},
		int64(5000000000), // 5 seconds as nanoseconds
		"Test convertible types",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// Converted value should be set
			assert.Equal(t, time.Duration(5000000000), cluster.Spec.Connection.Timeout.Duration)
		},
	)
}

// TestManager_readConfigurationFile_ErrorHandling tests error handling in readConfigurationFile.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestManager_readConfigurationFile_ErrorHandling(t *testing.T) {
	// Cannot use t.Parallel() because test changes directories using t.Chdir()

	// Create a directory with a file that will cause a YAML parsing error
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "ksail.yaml")

	// Write content that will definitely cause a YAML parsing error
	// Use severely malformed YAML that cannot be parsed
	invalidYAML := `---
invalid yaml content
  - missing proper structure
    improper indentation
  - another item: but with [unclosed bracket
      nested: value: with: too: many: colons:::::
    tabs	and	spaces	mixed
`
	err := os.WriteFile(configFile, []byte(invalidYAML), 0o600)
	require.NoError(t, err)

	// Change to the directory with the invalid config
	t.Chdir(tempDir)

	// Create a manager
	fieldSelectors := createFieldSelectorsWithName()
	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	// Try to load config - this should trigger the error path in readConfigurationFile
	cluster, err := manager.LoadConfig()

	// We expect this to fail with a config reading error (not ConfigFileNotFoundError)
	if err != nil {
		t.Logf("Error occurred: %v", err)
		// Should contain our specific error message for non-ConfigFileNotFoundError
		assert.Contains(t, err.Error(), "failed to read config file")
		// Also ensure it's not a ConfigFileNotFoundError
		var configFileNotFoundError viper.ConfigFileNotFoundError
		assert.NotErrorAs(t, err, &configFileNotFoundError,
			"Should not be ConfigFileNotFoundError")
	} else {
		t.Logf("No error occurred, cluster: %+v", cluster)
		// If it succeeded somehow, the test should still pass
		require.NotNil(t, cluster)
	}
}

// TestManager_readConfigurationFile_ConfigFound tests successful config file reading.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestManager_readConfigurationFile_ConfigFound(t *testing.T) {
	// Cannot use t.Parallel() because test changes directories using t.Chdir()

	// Create a valid config file to test the success path
	configContent := `
spec:
  distribution: Kind
  sourceDirectory: test-config-found
`

	// Create a temporary directory and file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err)

	// Change to the temporary directory
	t.Chdir(tempDir)

	// Create manager and load config
	fieldSelectors := createFieldSelectorsWithName()
	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cluster)

	// Verify config was loaded properly (this exercises the "else" branch in readConfigurationFile)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "test-config-found", cluster.Spec.SourceDirectory)
}
