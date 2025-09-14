package ksail_test

import (
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

// createStandardFieldSelectors creates a common set of field selectors used in multiple tests.
func createStandardFieldSelectors() []ksail.FieldSelector[v1alpha1.Cluster] {
	return []ksail.FieldSelector[v1alpha1.Cluster]{
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
	}
}

// createFieldSelectorsWithName creates field selectors including name field.
func createFieldSelectorsWithName() []ksail.FieldSelector[v1alpha1.Cluster] {
	selectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"ksail-default",
			"Name of the cluster",
		),
	}
	selectors = append(selectors, createStandardFieldSelectors()...)

	return selectors
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

	manager := ksail.NewConfigManager(fieldSelectors...)

	require.NotNil(t, manager)
	require.NotNil(t, manager.Config)
	assert.NotNil(t, manager.GetViper())
}

// TestManager_LoadConfig tests the LoadConfig method with different scenarios.
func TestManager_LoadConfig(t *testing.T) {
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
			// Set environment variables for the test
			for key, value := range testCase.envVars {
				t.Setenv(key, value)
			}

			fieldSelectors := createFieldSelectorsWithName()

			manager := ksail.NewConfigManager(fieldSelectors...)

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

	manager := ksail.NewConfigManager()
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
			name:           "AddFlagsFromFields with multiple selectors",
			fieldSelectors: createStandardFieldSelectors(),
			expectedFlags:  []string{"distribution", "source-directory", "context", "timeout"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := ksail.NewConfigManager(testCase.fieldSelectors...)
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

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"test-cluster",
			"Name of the cluster",
		),
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	// Before loading, Config should be initialized with proper TypeMeta
	expectedEmpty := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.Kind,
			APIVersion: v1alpha1.APIVersion,
		},
		Metadata: metav1.ObjectMeta{},
		Spec:     v1alpha1.Spec{},
	}
	assert.Equal(t, expectedEmpty, manager.Config)

	// Load config
	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// After loading, Config property should be accessible and equal to returned cluster
	assert.Equal(t, cluster, manager.Config)
	assert.Equal(t, "test-cluster", manager.Config.Metadata.Name)
}

// TestManager_SetFieldValueWithNilDefault tests setFieldValue with nil default value.
func TestManager_SetFieldValueWithNilDefault(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			DefaultValue: nil, // nil value should be handled gracefully
			Description:  "Test nil default",
		},
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// When default is nil, field should remain empty
	assert.Empty(t, cluster.Metadata.Name)
}

// TestManager_SetFieldValueWithNonConvertibleTypes tests setFieldValue with non-convertible types.
func TestManager_SetFieldValueWithNonConvertibleTypes(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			DefaultValue: 123, // int cannot be converted to string
			Description:  "Test non-convertible type",
		},
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// When type is not convertible, field should remain empty
	assert.Empty(t, cluster.Metadata.Name)
}

// TestManager_SetFieldValueWithDirectlyAssignableTypes tests setFieldValue with directly assignable types.
func TestManager_SetFieldValueWithDirectlyAssignableTypes(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			DefaultValue: "direct-assignment",
			Description:  "Test direct assignment",
		},
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// Direct string assignment should work
	assert.Equal(t, "direct-assignment", cluster.Metadata.Name)
}

// TestManager_SetFieldValueWithNonPointerField tests setFieldValue with non-pointer field.
func TestManager_SetFieldValueWithNonPointerField(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return c.Metadata.Name }, // Return value, not pointer
			DefaultValue: "should-not-set",
			Description:  "Test non-pointer field",
		},
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// Non-pointer field should remain empty
	assert.Empty(t, cluster.Metadata.Name)
}

// TestManager_SetFieldValueWithConvertibleTypes tests setFieldValue with convertible types.
func TestManager_SetFieldValueWithConvertibleTypes(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector: func(c *v1alpha1.Cluster) any {
				// Use the timeout field which accepts time.Duration
				return &c.Spec.Connection.Timeout.Duration
			},
			DefaultValue: int64(5000000000), // 5 seconds as nanoseconds
			Description:  "Test convertible types",
		},
	}

	manager := ksail.NewConfigManager(fieldSelectors...)

	cluster, err := manager.LoadConfig()
	require.NoError(t, err)

	// Converted value should be set
	assert.Equal(t, time.Duration(5000000000), cluster.Spec.Connection.Timeout.Duration)
}
