package ksail_test

import (
	"os"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setupTestEnvironment sets up a clean test environment.
func setupTestEnvironment(t *testing.T) {
	t.Helper()
	// Clean up any existing environment variables that might affect tests
	os.Unsetenv("KSAIL_METADATA_NAME")
	os.Unsetenv("KSAIL_SPEC_DISTRIBUTION")
	os.Unsetenv("KSAIL_SPEC_SOURCEDIRECTORY")
	os.Unsetenv("KSAIL_SPEC_CONNECTION_CONTEXT")
	os.Unsetenv("KSAIL_SPEC_CONNECTION_KUBECONFIG")
	os.Unsetenv("KSAIL_SPEC_CONNECTION_TIMEOUT")
}

// TestNewManager tests the NewManager constructor.
func TestNewManager(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"test-cluster",
			"Name of the cluster",
		),
	}

	manager := ksail.NewManager(fieldSelectors...)

	require.NotNil(t, manager)
	require.NotNil(t, manager.Config)
	assert.NotNil(t, manager.GetViper())
}

// TestManager_LoadConfig tests the LoadConfig method with different scenarios.
func TestManager_LoadConfig(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests use setupTestEnvironment and t.Setenv
	tests := []struct {
		name                string
		envVars             map[string]string
		expectedClusterName string
		shouldSucceed       bool
	}{
		{
			name:                "LoadConfig with defaults",
			envVars:             map[string]string{},
			expectedClusterName: "ksail-default",
			shouldSucceed:       true,
		},
		{
			name: "LoadConfig with environment variables",
			envVars: map[string]string{
				"KSAIL_METADATA_NAME": "test-cluster",
			},
			expectedClusterName: "test-cluster",
			shouldSucceed:       true,
		},
		{
			name: "LoadConfig with multiple environment variables",
			envVars: map[string]string{
				"KSAIL_METADATA_NAME":           "env-cluster",
				"KSAIL_SPEC_DISTRIBUTION":       "K3d",
				"KSAIL_SPEC_SOURCEDIRECTORY":    "custom-k8s",
				"KSAIL_SPEC_CONNECTION_CONTEXT": "custom-context",
			},
			expectedClusterName: "env-cluster",
			shouldSucceed:       true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setupTestEnvironment(t)

			// Set environment variables for the test
			for key, value := range testCase.envVars {
				t.Setenv(key, value)
			}

			fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
					"ksail-default",
					"Name of the cluster",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
					v1alpha1.DistributionKind,
					"Kubernetes distribution",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
					"k8s",
					"Source directory for workloads",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
					"",
					"Kubernetes context",
				),
			}

			manager := ksail.NewManager(fieldSelectors...)

			cluster, err := manager.LoadConfig()

			if testCase.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, cluster)
				assert.Equal(t, testCase.expectedClusterName, cluster.Metadata.Name)

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

// TestManager_GetViper tests the GetViper method.
func TestManager_GetViper(t *testing.T) {
	t.Parallel()

	manager := ksail.NewManager()
	viper := manager.GetViper()

	require.NotNil(t, viper)

	// Test that it's properly configured by setting and getting a value
	viper.SetDefault("test.key", "test-value")
	assert.Equal(t, "test-value", viper.GetString("test.key"))
}

// TestAddFlagFromField tests the AddFlagFromField function.
func TestAddFlagFromField(t *testing.T) {
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

			selector := ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
				testCase.defaultValue,
				testCase.description...,
			)

			assert.Equal(t, testCase.defaultValue, selector.DefaultValue)
			assert.Equal(t, testCase.expectedDesc, selector.Description)
			assert.NotNil(t, selector.Selector)
		})
	}
}

// TestNewCobraCommand tests the NewCobraCommand function.
func TestNewCobraCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		use            string
		short          string
		long           string
		fieldSelectors []ksail.FieldSelector[v1alpha1.Cluster]
		expectFlags    bool
	}{
		{
			name:           "NewCobraCommand without field selectors",
			use:            "test",
			short:          "Test command",
			long:           "Test command description",
			fieldSelectors: []ksail.FieldSelector[v1alpha1.Cluster]{},
			expectFlags:    false,
		},
		{
			name:  "NewCobraCommand with field selectors",
			use:   "test",
			short: "Test command",
			long:  "Test command description",
			fieldSelectors: []ksail.FieldSelector[v1alpha1.Cluster]{
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
					v1alpha1.DistributionKind,
					"Kubernetes distribution",
				),
			},
			expectFlags: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var calledRunE bool
			runE := func(_ *cobra.Command, manager *ksail.Manager, args []string) error {
				calledRunE = true
				require.NotNil(t, manager)
				return nil
			}

			cmd := ksail.NewCobraCommand(
				testCase.use,
				testCase.short,
				testCase.long,
				runE,
				testCase.fieldSelectors...,
			)

			require.NotNil(t, cmd)
			assert.Equal(t, testCase.use, cmd.Use)
			assert.Equal(t, testCase.short, cmd.Short)
			assert.Equal(t, testCase.long, cmd.Long)
			assert.Equal(t, ksail.SuggestionsMinimumDistance, cmd.SuggestionsMinimumDistance)

			// Test RunE function
			err := cmd.RunE(cmd, []string{})
			require.NoError(t, err)
			assert.True(t, calledRunE)

			// Check flags
			if testCase.expectFlags {
				assert.True(t, cmd.Flags().HasFlags())
				// Should have distribution flag
				flag := cmd.Flags().Lookup("distribution")
				require.NotNil(t, flag, "distribution flag should exist")
				assert.Equal(t, "d", flag.Shorthand)
			} else {
				assert.False(t, cmd.Flags().HasFlags())
			}
		})
	}
}

// TestManager_AddFlagsFromFields tests the AddFlagsFromFields method.
func TestManager_AddFlagsFromFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fieldSelectors []ksail.FieldSelector[v1alpha1.Cluster]
		expectedFlags  []string
	}{
		{
			name:           "AddFlagsFromFields with no selectors",
			fieldSelectors: []ksail.FieldSelector[v1alpha1.Cluster]{},
			expectedFlags:  []string{},
		},
		{
			name: "AddFlagsFromFields with distribution selector",
			fieldSelectors: []ksail.FieldSelector[v1alpha1.Cluster]{
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
					v1alpha1.DistributionKind,
					"Kubernetes distribution",
				),
			},
			expectedFlags: []string{"distribution"},
		},
		{
			name: "AddFlagsFromFields with multiple selectors",
			fieldSelectors: []ksail.FieldSelector[v1alpha1.Cluster]{
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
					v1alpha1.DistributionKind,
					"Kubernetes distribution",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
					"k8s",
					"Source directory",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
					"",
					"Kubernetes context",
				),
				ksail.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
					metav1.Duration{Duration: 5 * time.Minute},
					"Connection timeout",
				),
			},
			expectedFlags: []string{"distribution", "source-directory", "context", "timeout"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := ksail.NewManager(testCase.fieldSelectors...)
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
func TestManager_LoadConfig_ConfigProperty(t *testing.T) {
	t.Parallel()

	setupTestEnvironment(t)

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"test-cluster",
			"Name of the cluster",
		),
	}

	manager := ksail.NewManager(fieldSelectors...)

	// Before loading, Config should be empty
	assert.Equal(t, &v1alpha1.Cluster{}, manager.Config)

	// Load config
	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// After loading, Config property should be accessible and equal to returned cluster
	assert.Equal(t, cluster, manager.Config)
	assert.Equal(t, "test-cluster", manager.Config.Metadata.Name)
}
