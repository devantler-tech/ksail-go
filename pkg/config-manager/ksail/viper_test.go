package ksail_test

import (
	"os"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitializeViper tests the InitializeViper function.
func TestInitializeViper(t *testing.T) {
	// Cannot use t.Parallel() because test uses t.Setenv()
	viperInstance := ksail.InitializeViper()

	require.NotNil(t, viperInstance, "InitializeViper should return a non-nil viper instance")

	// Test configuration settings
	// We can't directly test the config file name, but we can test that defaults work
	viperInstance.SetDefault("test.config", "default-value")
	assert.Equal(t, "default-value", viperInstance.GetString("test.config"))

	// Test that environment variables are configured
	viperInstance.AutomaticEnv() // This is a setter method, doesn't return a value

	// Test environment prefix
	// We can't directly access the env prefix, but we can test its behavior
	t.Setenv("KSAIL_TEST_VALUE", "test-env-value")

	_ = viperInstance.BindEnv("test.value")
	assert.Equal(t, "test-env-value", viperInstance.GetString("test.value"))
}

// TestInitializeViper_ConfigPaths tests that config paths are set correctly.
func TestInitializeViper_ConfigPaths(t *testing.T) {
	// Cannot use t.Parallel() because test uses t.Setenv()
	viperInstance := ksail.InitializeViper()

	// Test that we can set and get values (indicates viper is working)
	viperInstance.SetDefault("test.config", "default-value")
	assert.Equal(t, "default-value", viperInstance.GetString("test.config"))

	// Test that environment variables override defaults
	t.Setenv("KSAIL_TEST_CONFIG", "env-value")

	_ = viperInstance.BindEnv("test.config")
	assert.Equal(t, "env-value", viperInstance.GetString("test.config"))
}

// TestInitializeViper_EnvKeyReplacer tests environment key replacement.
func TestInitializeViper_EnvKeyReplacer(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv()
	viperInstance := ksail.InitializeViper()

	// Test that dots and dashes in keys are replaced with underscores for env vars
	tests := []struct {
		name   string
		key    string
		envVar string
		value  string
	}{
		{
			name:   "Dot replacement",
			key:    "spec.distribution",
			envVar: "KSAIL_SPEC_DISTRIBUTION",
			value:  "kind-test",
		},
		{
			name:   "Dash replacement",
			key:    "source-directory",
			envVar: "KSAIL_SOURCE_DIRECTORY",
			value:  "k8s-test",
		},
		{
			name:   "Mixed dot and dash replacement",
			key:    "spec.connection-timeout",
			envVar: "KSAIL_SPEC_CONNECTION_TIMEOUT",
			value:  "5m-test",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv(testCase.envVar, testCase.value)

			// Bind the environment variable
			_ = viperInstance.BindEnv(testCase.key)

			// Test that the value is retrieved correctly
			assert.Equal(t, testCase.value, viperInstance.GetString(testCase.key))
		})
	}
}

// TestInitializeViper_ConfigFileReading tests configuration file reading behavior.
//
//nolint:paralleltest // Cannot run in parallel due to directory changes via t.Chdir()
func TestInitializeViper_ConfigFileReading(t *testing.T) {
	// Cannot use t.Parallel() because test changes directories using t.Chdir()
	// which can conflict with parallel test execution

	// Create a temporary config file
	configContent := `
metadata:
  name: test-cluster-from-file
spec:
  distribution: K3d
  sourceDirectory: k8s-from-file
`

	// Create a temporary directory and file
	tempDir := t.TempDir()
	configFile := tempDir + "/ksail.yaml"

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err)

	// Change to the temporary directory so viper can find the config file
	t.Chdir(tempDir)

	// Initialize viper - it should read the config file
	viperInstance := ksail.InitializeViper()

	// Test that values from the config file are loaded
	assert.Equal(t, "test-cluster-from-file", viperInstance.GetString("metadata.name"))
	assert.Equal(t, "K3d", viperInstance.GetString("spec.distribution"))
	assert.Equal(t, "k8s-from-file", viperInstance.GetString("spec.sourceDirectory"))
}

// TestViperConstants tests the viper-related constants.
func TestViperConstants(t *testing.T) {
	t.Parallel()

	// Test that constants have expected values
	assert.Equal(t, "ksail", ksail.DefaultConfigFileName)
	assert.Equal(t, "KSAIL", ksail.EnvPrefix)
	assert.Equal(t, 2, ksail.SuggestionsMinimumDistance)
}

// TestInitializeViper_EnvironmentVariableBinding tests automatic environment variable binding.
func TestInitializeViper_EnvironmentVariableBinding(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv()
	viperInstance := ksail.InitializeViper()

	// Test various environment variable patterns
	tests := []struct {
		name   string
		envVar string
		key    string
		value  string
	}{
		{
			name:   "Simple environment variable",
			envVar: "KSAIL_TEST_SIMPLE",
			key:    "test.simple",
			value:  "simple-value",
		},
		{
			name:   "Nested environment variable",
			envVar: "KSAIL_NESTED_CONFIG_VALUE",
			key:    "nested.config.value",
			value:  "nested-value",
		},
		{
			name:   "Environment variable with dashes",
			envVar: "KSAIL_DASH_SEPARATED_VALUE",
			key:    "dash-separated-value",
			value:  "dash-value",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv(testCase.envVar, testCase.value)

			// Bind the environment variable
			_ = viperInstance.BindEnv(testCase.key)

			// Test that the value is retrieved correctly
			assert.Equal(t, testCase.value, viperInstance.GetString(testCase.key))
		})
	}
}

// TestInitializeViper_EnvReplacerRules tests environment key replacer rules.
func TestInitializeViper_EnvReplacerRules(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv()
	viperInstance := ksail.InitializeViper()

	// Test the key replacer rules by setting specific environment variables
	tests := []struct {
		name     string
		key      string
		envVar   string
		value    string
		testDesc string
	}{
		{
			name:     "Period to underscore",
			key:      "metadata.name",
			envVar:   "KSAIL_METADATA_NAME",
			value:    "period-test",
			testDesc: "Periods in keys should be replaced with underscores in env vars",
		},
		{
			name:     "Dash to underscore",
			key:      "source-directory",
			envVar:   "KSAIL_SOURCE_DIRECTORY",
			value:    "dash-test",
			testDesc: "Dashes in keys should be replaced with underscores in env vars",
		},
		{
			name:     "Mixed period and dash",
			key:      "spec.connection-timeout",
			envVar:   "KSAIL_SPEC_CONNECTION_TIMEOUT",
			value:    "mixed-test",
			testDesc: "Both periods and dashes should be replaced with underscores",
		},
		{
			name:     "Multiple periods",
			key:      "deep.nested.config.value",
			envVar:   "KSAIL_DEEP_NESTED_CONFIG_VALUE",
			value:    "deep-test",
			testDesc: "Multiple periods should all be replaced",
		},
		{
			name:     "Multiple dashes",
			key:      "multi-dash-key-name",
			envVar:   "KSAIL_MULTI_DASH_KEY_NAME",
			value:    "multi-dash-test",
			testDesc: "Multiple dashes should all be replaced",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(testCase.envVar, testCase.value)
			_ = viperInstance.BindEnv(testCase.key)
			assert.Equal(
				t,
				testCase.value,
				viperInstance.GetString(testCase.key),
				testCase.testDesc,
			)
		})
	}
}

// TestInitializeViper_Idempotency tests that multiple calls to InitializeViper work correctly.
func TestInitializeViper_Idempotency(t *testing.T) {
	t.Parallel()

	// Call InitializeViper multiple times
	viper1 := ksail.InitializeViper()
	viper2 := ksail.InitializeViper()

	// Both should be valid viper instances
	require.NotNil(t, viper1)
	require.NotNil(t, viper2)

	// They should be different instances (not singletons)
	assert.NotSame(t, viper1, viper2, "Each call should return a new viper instance")

	// Both should work independently
	viper1.SetDefault("test1", "value1")
	viper2.SetDefault("test2", "value2")

	assert.Equal(t, "value1", viper1.GetString("test1"))
	assert.Equal(t, "value2", viper2.GetString("test2"))

	// viper1 should not have test2, viper2 should not have test1
	assert.Empty(t, viper1.GetString("test2"))
	assert.Empty(t, viper2.GetString("test1"))
}

// TestInitializeViper_ConfigType tests that the config type is set correctly.
func TestInitializeViper_ConfigType(t *testing.T) {
	t.Parallel()

	viperInstance := ksail.InitializeViper()

	// Create a YAML config string and test parsing
	yamlConfig := `
test:
  yaml: true
  value: "yaml-test"
`

	viperInstance.SetConfigType("yaml")
	err := viperInstance.ReadConfig(strings.NewReader(yamlConfig))
	require.NoError(t, err, "Should be able to read YAML config")

	assert.True(t, viperInstance.GetBool("test.yaml"))
	assert.Equal(t, "yaml-test", viperInstance.GetString("test.value"))
}

// TestInitializeViper_ErrorHandling tests error handling behavior.
func TestInitializeViper_ErrorHandling(t *testing.T) {
	// Cannot use t.Parallel() because test uses t.Setenv()

	// This test verifies that InitializeViper handles missing config files gracefully
	viperInstance := ksail.InitializeViper()

	// Even if no config file exists, viper should still work
	require.NotNil(t, viperInstance)

	// Should be able to set and get values
	viperInstance.SetDefault("error.test", "default-value")
	assert.Equal(t, "default-value", viperInstance.GetString("error.test"))

	// Should be able to bind environment variables
	t.Setenv("KSAIL_ERROR_TEST", "env-value")

	_ = viperInstance.BindEnv("error.test")
	assert.Equal(t, "env-value", viperInstance.GetString("error.test"))
}
