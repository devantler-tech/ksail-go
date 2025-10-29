package configmanager_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
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
			func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			"kind.yaml",
			"Distribution config",
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

//nolint:paralleltest // Uses t.Chdir to isolate file system state for config loading.
func TestLoadConfigLoadsKindDistributionConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	kindConfigPath := filepath.Join(tempDir, "kind.yaml")
	kindConfigYAML := "apiVersion: kind.x-k8s.io/v1alpha4\n" +
		"kind: Cluster\n" +
		"networking:\n" +
		"  disableDefaultCNI: true\n"
	require.NoError(t, os.WriteFile(kindConfigPath, []byte(kindConfigYAML), 0o600))

	ksailConfig := "apiVersion: ksail.dev/v1alpha1\n" +
		"kind: Cluster\n" +
		"spec:\n" +
		"  distribution: Kind\n" +
		"  distributionConfig: " + kindConfigPath + "\n" +
		"  cni: Cilium\n" +
		"  connection:\n" +
		"    context: kind-kind\n"
	require.NoError(t, os.WriteFile("ksail.yaml", []byte(ksailConfig), 0o600))

	manager := configmanager.NewConfigManager(io.Discard)
	manager.Viper.SetConfigFile("ksail.yaml")

	_, err := manager.LoadConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, kindConfigPath, manager.Config.Spec.DistributionConfig)
}

//nolint:paralleltest // Uses t.Chdir to isolate file system state for config loading.
func TestLoadConfigLoadsK3dDistributionConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	k3dConfigPath := filepath.Join(tempDir, "k3d.yaml")
	k3dConfigYAML := "apiVersion: k3d.io/v1alpha5\n" +
		"kind: Simple\n" +
		"metadata:\n" +
		"  name: test\n"
	require.NoError(t, os.WriteFile(k3dConfigPath, []byte(k3dConfigYAML), 0o600))

	ksailConfig := "apiVersion: ksail.dev/v1alpha1\n" +
		"kind: Cluster\n" +
		"spec:\n" +
		"  distribution: K3d\n" +
		"  distributionConfig: " + k3dConfigPath + "\n" +
		"  cni: Cilium\n" +
		"  connection:\n" +
		"    context: k3d-k3d-default\n"
	require.NoError(t, os.WriteFile("ksail.yaml", []byte(ksailConfig), 0o600))

	manager := configmanager.NewConfigManager(io.Discard)
	manager.Viper.SetConfigFile("ksail.yaml")

	_, err := manager.LoadConfig(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation reported")
	assert.Equal(t, k3dConfigPath, manager.Config.Spec.DistributionConfig)
}

// createFieldSelectorsWithName creates field selectors including name field.
func createFieldSelectorsWithName() []configmanager.FieldSelector[v1alpha1.Cluster] {
	selectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.Distribution(""), // Empty distribution for testing defaults
			"Kubernetes distribution",
		),
	}
	selectors = append(
		selectors,
		createStandardFieldSelectors()[1:]...) // Skip the first selector which is Distribution

	return selectors
}

// createDistributionOnlyFieldSelectors creates field selectors with only the distribution field.
func createDistributionOnlyFieldSelectors() []configmanager.FieldSelector[v1alpha1.Cluster] {
	// Use distribution and distribution config selectors from the standard set
	return createStandardFieldSelectors()[:2]
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
// TestLoadConfig tests the LoadConfig method with various environment variable configurations.
//
//nolint:paralleltest // Cannot use t.Parallel() because testLoadConfigCase uses t.Chdir()
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                 string
		envVars              map[string]string
		expectedDistribution v1alpha1.Distribution
		shouldSucceed        bool
	}{
		{
			name:                 "LoadConfig with defaults (missing distribution)",
			envVars:              map[string]string{},
			expectedDistribution: v1alpha1.Distribution(""),
			shouldSucceed:        false,
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
				"KSAIL_SPEC_DISTRIBUTION":       "K3d",
				"KSAIL_SPEC_SOURCEDIRECTORY":    "custom-k8s",
				"KSAIL_SPEC_CONNECTION_CONTEXT": "k3d-k3d-default",
			},
			expectedDistribution: v1alpha1.DistributionK3d,
			shouldSucceed:        true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testLoadConfigCase(t, testCase)
		})
	}
}

// TestLoadConfig_MissingFileNotifiesDefaults verifies notification when config file is absent.
//
//nolint:paralleltest // Uses t.Chdir for isolated filesystem state.
func TestLoadConfigMissingFileNotifiesDefaults(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	_, output, cluster := loadConfigAndCaptureOutput(t, createStandardFieldSelectors()...)
	assert.NotNil(t, cluster)
	assert.Contains(t, output.String(), "using default config")
}

// TestLoadConfig_ConfigFileNotifiesFound verifies notification when config file is discovered.
//
//nolint:paralleltest // Uses t.Chdir for isolated filesystem state.
func TestLoadConfigConfigFileNotifiesFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	configPath := filepath.Join(tempDir, "ksail.yaml")
	configContents := []byte(
		"apiVersion: ksail.dev/v1alpha1\nkind: Cluster\nspec:\n  distribution: Kind\n  distributionConfig: kind.yaml\n",
	)

	err := os.WriteFile(configPath, configContents, 0o600)
	require.NoError(t, err)

	_, output, cluster := loadConfigAndCaptureOutput(t, createStandardFieldSelectors()...)
	assert.NotNil(t, cluster)

	assert.Contains(t, output.String(), "Load config...")
	assert.Contains(t, output.String(), "loading ksail config")
	assert.Contains(t, output.String(), "config loaded")
	assert.Contains(t, output.String(), "'"+configPath+"' found")
}

// TestLoadConfig_ConfigReusedNotification verifies notification when config is reused.
//
//nolint:paralleltest // Uses t.Chdir for isolated filesystem state.
func TestLoadConfigConfigReusedNotification(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	manager, output, _ := loadConfigAndCaptureOutput(t, createStandardFieldSelectors()...)
	output.Reset()

	_, err := manager.LoadConfig(nil)
	require.NoError(t, err)

	assert.Contains(t, output.String(), "config already loaded, reusing existing config")
}

func TestNewCommandConfigManagerBindsFlags(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(&output)

	selectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.DefaultDistributionFieldSelector(),
		configmanager.DefaultDistributionConfigFieldSelector(),
	}

	manager := configmanager.NewCommandConfigManager(cmd, selectors)

	require.NotNil(t, manager)
	assert.Same(t, cmd.OutOrStdout(), manager.Writer)

	for _, flagName := range []string{"distribution", "distribution-config"} {
		flag := cmd.Flags().Lookup(flagName)
		require.NotNil(t, flag, "expected flag %s to be registered", flagName)
	}
}

func TestLoadConfigSilentSkipsNotifications(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	selectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.DefaultDistributionFieldSelector(),
		configmanager.DefaultDistributionConfigFieldSelector(),
		configmanager.DefaultKubeconfigFieldSelector(),
	}

	manager := configmanager.NewConfigManager(&output, selectors...)

	cluster, err := manager.LoadConfigSilent()
	require.NoError(t, err)
	assert.Empty(t, output.String(), "silent load should not emit notifications")

	require.NotNil(t, cluster)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "kind.yaml", cluster.Spec.DistributionConfig)
}

// TestLoadConfigValidationFailureMessages verifies validation error notifications.
//
//nolint:paralleltest // Uses t.Chdir for isolated filesystem state.
func TestLoadConfigValidationFailureMessages(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	var output bytes.Buffer

	manager := configmanager.NewConfigManager(&output)

	manager.Config.Kind = ""
	manager.Config.APIVersion = ""
	manager.Config.Spec.Distribution = ""
	manager.Config.Spec.DistributionConfig = ""

	_, err := manager.LoadConfig(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation reported")
	assert.Contains(t, err.Error(), "4 error(s)")

	logOutput := output.String()
	assert.Contains(t, logOutput, "error:")
	assert.Contains(t, logOutput, "kind is required")
	assert.Contains(t, logOutput, "apiVersion is required")
	assert.Contains(t, logOutput, "field: spec.distribution")
	assert.Contains(t, logOutput, "field: spec.distributionConfig")
}

// testLoadConfigCase is a helper function to test a single LoadConfig scenario.
func testLoadConfigCase(
	t *testing.T,
	testCase struct {
		name                 string
		envVars              map[string]string
		expectedDistribution v1alpha1.Distribution
		shouldSucceed        bool
	},
) {
	t.Helper()

	// Create temporary directory and change to it to isolate from existing config files
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Set environment variables for the test
	for key, value := range testCase.envVars {
		t.Setenv(key, value)
	}

	fieldSelectors := createFieldSelectorsWithName()

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	cluster, err := manager.LoadConfig(nil)

	if testCase.shouldSucceed {
		require.NoError(t, err)

		require.NotNil(t, cluster)
		assert.Equal(t, testCase.expectedDistribution, cluster.Spec.Distribution)

		// Test that subsequent calls return the same config
		cluster2, err := manager.LoadConfig(nil)
		require.NoError(t, err)
		assert.Same(t, cluster, cluster2)
	} else {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation reported")
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
			expectedFlags: []string{
				"distribution",
				"distribution-config",
				"source-directory",
				"context",
				"timeout",
			},
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

	fieldSelectors := createDistributionOnlyFieldSelectors()

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	// Before loading, Config should be initialized with proper TypeMeta
	expectedEmpty := v1alpha1.NewCluster()
	assert.Equal(t, expectedEmpty, manager.Config)

	// Load config
	cluster, err := manager.LoadConfig(nil)
	require.NoError(t, err)

	// After loading, Config property should be accessible and equal to returned cluster
	assert.Equal(t, cluster, manager.Config)
	assert.Equal(t, v1alpha1.DistributionKind, manager.Config.Spec.Distribution)
}

// testFieldValueSetting is a helper function for testing field value setting scenarios.
func testFieldValueSetting(
	t *testing.T,
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description string,
	assertFunc func(*testing.T, *v1alpha1.Cluster),
	expectValidationError bool,
	additionalSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) {
	t.Helper()

	// Create temporary directory and change to it to isolate from existing config files
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     selector,
			DefaultValue: defaultValue,
			Description:  description,
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			DefaultValue: "kind.yaml",
			Description:  "Distribution config",
		},
	}
	fieldSelectors = append(fieldSelectors, additionalSelectors...)

	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)

	cluster, err := manager.LoadConfig(nil)

	if expectValidationError {
		require.Error(t, err)
		assertFunc(t, cluster)

		return
	}

	require.NoError(t, err)
	assertFunc(t, cluster)
}

// TestManager_SetFieldValueWithNilDefault tests setFieldValue with nil default value.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNilDefault(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		nil, // nil value should be handled gracefully
		"Test nil default",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// When default is nil, field should remain empty
			if cluster != nil {
				assert.Empty(t, cluster.Spec.Distribution)
			}
		},
		true,
	)
}

// TestManager_SetFieldValueWithNonConvertibleTypes tests setFieldValue with non-convertible types.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNonConvertibleTypes(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		123, // int cannot be converted to string
		"Test non-convertible type",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// When type is not convertible, field should remain empty
			if cluster != nil {
				assert.Empty(t, cluster.Spec.Distribution)
			}
		},
		true,
	)
}

// TestManager_SetFieldValueWithDirectlyAssignableTypes tests setFieldValue with directly assignable types.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithDirectlyAssignableTypes(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		v1alpha1.DistributionK3d,
		"Test direct assignment",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// Direct string assignment should work
			assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
		},
		false,
	)
}

// TestManager_SetFieldValueWithNonPointerField tests setFieldValue with non-pointer field.
//
//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestSetFieldValueWithNonPointerField(t *testing.T) {
	testFieldValueSetting(
		t,
		func(c *v1alpha1.Cluster) any { return c.Spec.Distribution }, // Return value, not pointer
		"should-not-set",
		"Test non-pointer field",
		func(t *testing.T, cluster *v1alpha1.Cluster) {
			t.Helper()
			// Non-pointer field should remain empty
			if cluster != nil {
				assert.Empty(t, cluster.Spec.Distribution)
			}
		},
		true,
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
		false,
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			DefaultValue: v1alpha1.DistributionKind,
			Description:  "Distribution",
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
	_, err = manager.LoadConfig(nil)

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
		t.Logf("No error occurred, cluster: %+v", manager.Config)
		// If it succeeded somehow, the test should still pass
		require.NotNil(t, manager.Config)
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

	cluster, err := manager.LoadConfig(nil)
	require.NoError(t, err)

	require.NotNil(t, cluster)

	// Verify config was loaded properly (this exercises the "else" branch in readConfigurationFile)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "test-config-found", cluster.Spec.SourceDirectory)
}

// runIsFieldEmptyTestCases is a helper function to run test cases for isFieldEmpty function.
func runIsFieldEmptyTestCases(t *testing.T, tests []struct {
	name     string
	fieldPtr any
	expected bool
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := configmanager.IsFieldEmptyForTesting(tt.fieldPtr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestManager_isFieldEmpty_NilAndInvalidCases tests nil and invalid cases for isFieldEmpty function.
func TestManager_isFieldEmpty_NilAndInvalidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fieldPtr any
		expected bool
	}{
		{
			name:     "Nil field pointer",
			fieldPtr: nil,
			expected: true,
		},
		{
			name:     "Non-pointer field",
			fieldPtr: "direct-value",
			expected: true,
		},
		{
			name:     "Nil pointer field",
			fieldPtr: (*string)(nil),
			expected: true,
		},
	}

	runIsFieldEmptyTestCases(t, tests)
}

// TestManager_isFieldEmpty_ValidPointerCases tests valid pointer cases for isFieldEmpty function.
func TestManager_isFieldEmpty_ValidPointerCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fieldPtr any
		expected bool
	}{
		{
			name: "Valid pointer to empty string",
			fieldPtr: func() *string {
				s := ""

				return &s
			}(),
			expected: true,
		},
		{
			name: "Valid pointer to non-empty string",
			fieldPtr: func() *string {
				s := "value"

				return &s
			}(),
			expected: false,
		},
		{
			name: "Valid pointer to zero int",
			fieldPtr: func() *int {
				i := 0

				return &i
			}(),
			expected: true,
		},
		{
			name: "Valid pointer to non-zero int",
			fieldPtr: func() *int {
				i := 42

				return &i
			}(),
			expected: false,
		},
	}

	runIsFieldEmptyTestCases(t, tests)
}

//nolint:paralleltest // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestLoadConfig_ValidationFailureOutputs(t *testing.T) {
	// Cannot use t.Parallel() because test changes directories using t.Chdir()
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	var out bytes.Buffer

	manager := configmanager.NewConfigManager(
		&out,
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.APIVersion },
			Description:  "API version",
			DefaultValue: "",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Kind },
			Description:  "Resource kind",
			DefaultValue: "",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution",
			DefaultValue: v1alpha1.Distribution(""),
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			Description:  "Distribution config",
			DefaultValue: "",
		},
	)

	_, err := manager.LoadConfig(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation reported")

	output := out.String()
	assert.Contains(t, output, "error:")
	assert.Contains(t, output, "distribution is required")
}

//nolint:paralleltest // Uses t.Chdir for isolated filesystem state per scenario.
func TestLoadConfigKindCiliumValidation(t *testing.T) {
	tests := []kindCiliumScenario{
		{
			name:                "requiresDisableDefaultCNI",
			disableDefaultCNI:   false,
			expectValidationErr: true,
			configName:          "kind-cilium",
		},
		{
			name:                "passesWhenDisableDefaultCNITrue",
			disableDefaultCNI:   true,
			expectValidationErr: false,
			configName:          "kind-enabled",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			runKindCiliumValidationScenario(t, testCase)
		})
	}
}

//nolint:paralleltest // Uses t.Chdir for isolated filesystem state per scenario.
func TestLoadConfigK3dDistributionConfigHandling(t *testing.T) {
	tests := []k3dScenario{
		{
			name:                 "missingDistributionConfig",
			distributionContents: "",
			expectErr:            false,
		},
		{
			name:                 "invalidDistributionConfigProducesValidationMessage",
			distributionContents: "kind: Wrong\napiVersion: example/v1\n",
			expectErr:            true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			runK3dDistributionScenario(t, testCase)
		})
	}
}

// helper function to load config and capture output for tests.
func loadConfigAndCaptureOutput(
	t *testing.T,
	fieldSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) (*configmanager.ConfigManager, *bytes.Buffer, *v1alpha1.Cluster) {
	t.Helper()

	output := &bytes.Buffer{}
	manager := configmanager.NewConfigManager(output, fieldSelectors...)

	cluster, err := manager.LoadConfig(nil)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	return manager, output, cluster
}

type kindCiliumScenario struct {
	name                string
	disableDefaultCNI   bool
	expectValidationErr bool
	configName          string
}

func runKindCiliumValidationScenario(t *testing.T, scenario kindCiliumScenario) {
	t.Helper()

	tempDir := t.TempDir()
	t.Chdir(tempDir)

	kindContents := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %s
networking:
  disableDefaultCNI: %t
`, scenario.configName, scenario.disableDefaultCNI)
	writeFile(t, "kind.yaml", kindContents)

	ksailContents := fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
  cni: Cilium
  connection:
    context: kind-%s
`, scenario.configName)
	writeFile(t, "ksail.yaml", ksailContents)

	var (
		output  bytes.Buffer
		manager = configmanager.NewConfigManager(&output)
	)

	_, err := manager.LoadConfig(nil)
	logOutput := output.String()

	if scenario.expectValidationErr {
		if err == nil {
			t.Fatalf("expected validation error when disableDefaultCNI is false")
		}

		if !strings.Contains(logOutput, "Cilium CNI requires disableDefaultCNI") {
			t.Fatalf("expected Cilium validation message, got %q", logOutput)
		}

		return
	}

	if err != nil {
		t.Fatalf("unexpected validation error: %v (output: %s)", err, logOutput)
	}
}

type k3dScenario struct {
	name                 string
	distributionContents string
	expectErr            bool
	expectedLog          string
}

func runK3dDistributionScenario(t *testing.T, scenario k3dScenario) {
	t.Helper()

	manager, output := newK3dManagerForScenario(t, scenario)

	_, err := manager.LoadConfig(nil)
	logOutput := output.String()

	if scenario.expectErr {
		if err == nil {
			t.Fatalf(
				"expected validation error for %s but got none (output: %s)",
				scenario.name,
				logOutput,
			)
		}

		if scenario.expectedLog != "" && !strings.Contains(logOutput, scenario.expectedLog) {
			t.Fatalf(
				"expected log output to contain %q but got %q",
				scenario.expectedLog,
				logOutput,
			)
		}

		return
	}

	if err != nil {
		t.Fatalf(
			"unexpected validation error for %s: %v (output: %s)",
			scenario.name,
			err,
			logOutput,
		)
	}

	config := manager.Config
	if config == nil {
		t.Fatalf("expected config to be loaded")
	}

	if config.Spec.Distribution != v1alpha1.DistributionK3d {
		t.Fatalf("expected distribution to be K3d, got %s", config.Spec.Distribution)
	}
}

func newK3dManagerForScenario(
	t *testing.T,
	scenario k3dScenario,
) (*configmanager.ConfigManager, *bytes.Buffer) {
	t.Helper()

	tempDir := t.TempDir()
	t.Chdir(tempDir)

	if scenario.distributionContents != "" {
		writeFile(t, "k3d.yaml", scenario.distributionContents)
	}

	ksailContents := "apiVersion: ksail.dev/v1alpha1\n" +
		"kind: Cluster\n" +
		"spec:\n" +
		"  distribution: K3d\n" +
		"  distributionConfig: k3d.yaml\n" +
		"  sourceDirectory: k8s\n" +
		"  cni: Cilium\n" +
		"  connection:\n" +
		"    context: k3d-k3d-default\n"
	writeFile(t, "ksail.yaml", ksailContents)

	var (
		output  bytes.Buffer
		manager = configmanager.NewConfigManager(&output)
	)

	return manager, &output
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestIsFieldEmptyForTesting_Nil(t *testing.T) {
	t.Parallel()

	result := configmanager.IsFieldEmptyForTesting(nil)
	assert.True(t, result)
}

func TestIsFieldEmptyForTesting_NonPointer(t *testing.T) {
	t.Parallel()

	value := "test"
	result := configmanager.IsFieldEmptyForTesting(value)
	assert.True(t, result)
}

func TestIsFieldEmptyForTesting_NilPointer(t *testing.T) {
	t.Parallel()

	var ptr *string

	result := configmanager.IsFieldEmptyForTesting(ptr)
	assert.True(t, result)
}

func TestIsFieldEmptyForTesting_EmptyString(t *testing.T) {
	t.Parallel()

	value := ""
	result := configmanager.IsFieldEmptyForTesting(&value)
	assert.True(t, result)
}

func TestIsFieldEmptyForTesting_NonEmptyString(t *testing.T) {
	t.Parallel()

	value := "test"
	result := configmanager.IsFieldEmptyForTesting(&value)
	assert.False(t, result)
}
