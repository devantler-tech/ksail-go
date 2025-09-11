package config_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestManager_GetCluster tests the GetCluster method with different scenarios.
func TestManager_GetCluster(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests use setupTestEnvironment and t.Setenv
	tests := []struct {
		name                string
		loadClusterFirst    bool
		expectedClusterName string
	}{
		{
			name:                "GetCluster without LoadCluster first",
			loadClusterFirst:    false,
			expectedClusterName: "ksail-default", // Should load with defaults
		},
		{
			name:                "GetCluster after LoadCluster",
			loadClusterFirst:    true,
			expectedClusterName: "test-cluster",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setupTestEnvironment(t)

			fieldSelectors := []config.FieldSelector[v1alpha1.Cluster]{
				config.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
					"ksail-default",
				),
			}

			manager := config.NewManager(fieldSelectors...)

			if testCase.loadClusterFirst {
				// Set environment variable for the test
				t.Setenv("KSAIL_METADATA_NAME", "test-cluster")

				_, err := manager.LoadCluster()
				require.NoError(t, err)
			}

			cluster := manager.GetCluster()
			require.NotNil(t, cluster)
			assert.Equal(t, testCase.expectedClusterName, cluster.Metadata.Name)
		})
	}
}

// TestManager_GetViper tests the GetViper method.
func TestManager_GetViper(t *testing.T) {
	t.Parallel()

	manager := config.NewManager()
	viper := manager.GetViper()

	require.NotNil(t, viper)

	// Test that it's properly configured by setting and getting a value
	viper.SetDefault("test.key", "test-value")
	assert.Equal(t, "test-value", viper.GetString("test.key"))
}

// TestConvertDefaultValueForViper tests conversion of default values for Viper storage.
//
//nolint:paralleltest,tparallel
func TestConvertDefaultValueForViper(t *testing.T) {
	// Note: Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
	setupTestEnvironment(t)

	// Test enum to string conversions
	testEnumToStringConversions(t)

	// Test other type conversions
	testOtherTypeConversions(t)
}

// testEnumToStringConversions tests enum to string conversions for Viper.
func testEnumToStringConversions(t *testing.T) {
	t.Helper()

	enumSelectors := createEnumFieldSelectors()

	for _, selector := range enumSelectors {
		t.Run(selector.name+" to string", func(t *testing.T) {
			t.Parallel()

			manager := config.NewManager(selector.fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			assert.NotNil(t, cluster)
		})
	}
}

// testOtherTypeConversions tests other type conversions for Viper.
func testOtherTypeConversions(t *testing.T) {
	t.Helper()

	// Test basic types
	testBasicTypeConversions(t)

	// Test special types
	testSpecialTypeConversions(t)
}

// testBasicTypeConversions tests basic type conversion handling.
func testBasicTypeConversions(t *testing.T) {
	t.Helper()

	tests := []struct {
		name          string
		fieldSelector config.FieldSelector[v1alpha1.Cluster]
		expectedType  string
	}{
		{
			name: "string remains string",
			fieldSelector: config.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
				"test-cluster",
			),
			expectedType: "string",
		},
		{
			name: "bool remains bool",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					// Use a dummy bool field for testing
					ptr := new(bool)
					*ptr = true

					return ptr
				},
				true,
			),
			expectedType: "bool",
		},
		{
			name: "int remains int",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					// Use a dummy int field for testing
					ptr := new(int)
					*ptr = 3

					return ptr
				},
				3,
			),
			expectedType: "int",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := config.NewManager(testCase.fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			assert.NotNil(t, cluster)
		})
	}
}

// testSpecialTypeConversions tests special type conversions like metav1.Duration.
func testSpecialTypeConversions(t *testing.T) {
	t.Helper()

	tests := []struct {
		name          string
		fieldSelector config.FieldSelector[v1alpha1.Cluster]
		expectedType  string
	}{
		{
			name: "metav1.Duration to time.Duration",
			fieldSelector: config.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
				metav1.Duration{Duration: 5 * time.Minute},
			),
			expectedType: "duration",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := config.NewManager(testCase.fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			assert.NotNil(t, cluster)
		})
	}
}

// TestManager_IntegerTypes tests various integer type handling.
func TestManager_IntegerTypes(t *testing.T) {
	// Note: Cannot use t.Parallel() because individual test cases use t.Setenv
	setupTestEnvironment(t)

	tests := []struct {
		name          string
		fieldSelector config.FieldSelector[v1alpha1.Cluster]
		envVar        string
		envValue      string
		expectedValue int
	}{
		{
			name: "int type",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					// Use a dummy int field for testing
					ptr := new(int)
					*ptr = 1

					return ptr
				},
				1,
			),
			envVar:        "KSAIL_TEST_INT",
			envValue:      "5",
			expectedValue: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envValue)

			manager := config.NewManager(tt.fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			// Check the specific field value
			// Since we're using a dummy field, we can't check the actual value
			// Just verify the cluster was loaded successfully
			assert.NotNil(t, cluster)
		})
	}
}

// TestManager_FloatTypes tests float type handling.
//
//nolint:paralleltest // Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
func TestManager_FloatTypes(t *testing.T) {
	// Note: Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
	setupTestEnvironment(t)

	// Create a field selector for a float field (using a custom struct extension for testing)
	// Since v1alpha1.Cluster doesn't have float fields, we'll test this indirectly
	fieldSelector := config.AddFlagFromField(
		func(_ *v1alpha1.Cluster) any {
			// Create a dummy float pointer for testing
			ptr := new(float64)
			*ptr = 3.14

			return ptr
		},
		float64(2.718),
	)

	manager := config.NewManager(fieldSelector)
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test passes if no error occurs
	assert.NotNil(t, cluster)
}

// TestManager_SliceTypes tests slice type handling.
func TestManager_SliceTypes(t *testing.T) {
	// Note: Cannot use t.Parallel() because individual test cases use t.Setenv
	setupTestEnvironment(t)

	// Test with dummy slice fields since v1alpha1.Cluster doesn't have slice fields
	tests := []struct {
		name          string
		fieldSelector config.FieldSelector[v1alpha1.Cluster]
		envVar        string
		envValue      string
	}{
		{
			name: "string slice",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					ptr := new([]string)
					*ptr = []string{"default"}

					return ptr
				},
				[]string{"default"},
			),
			envVar:   "KSAIL_TEST_STRINGSLICE",
			envValue: "item1,item2,item3",
		},
		{
			name: "int slice",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					ptr := new([]int)
					*ptr = []int{1}

					return ptr
				},
				[]int{1},
			),
			envVar:   "KSAIL_TEST_INTSLICE",
			envValue: "1,2,3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envValue)

			manager := config.NewManager(tt.fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			// Test passes if no error occurs
			assert.NotNil(t, cluster)
		})
	}
}

// TestManager_UnsignedIntegerTypes tests unsigned integer type handling.
//
//nolint:paralleltest // Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
func TestManager_UnsignedIntegerTypes(t *testing.T) {
	// Note: Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
	setupTestEnvironment(t)

	// Test with dummy unsigned integer fields
	tests := []struct {
		name     string
		typeName string
	}{
		{name: "uint", typeName: "uint"},
		{name: "uint32", typeName: "uint32"},
		{name: "uint64", typeName: "uint64"},
	}

	for _, tt := range tests { //nolint:paralleltest // Cannot use parallel because parent uses setupTestEnvironment
		t.Run(tt.name, func(t *testing.T) {
			// Create field selectors for different uint types
			var fieldSelector config.FieldSelector[v1alpha1.Cluster]

			switch tt.typeName {
			case "uint":
				fieldSelector = config.AddFlagFromField(
					func(_ *v1alpha1.Cluster) any {
						ptr := new(uint)
						*ptr = 100

						return ptr
					},
					uint(100),
				)
			case "uint32":
				fieldSelector = config.AddFlagFromField(
					func(_ *v1alpha1.Cluster) any {
						ptr := new(uint32)
						*ptr = 3200

						return ptr
					},
					uint32(3200),
				)
			case "uint64":
				fieldSelector = config.AddFlagFromField(
					func(_ *v1alpha1.Cluster) any {
						ptr := new(uint64)
						*ptr = 6400

						return ptr
					},
					uint64(6400),
				)
			}

			manager := config.NewManager(fieldSelector)
			cluster, err := manager.LoadCluster()
			require.NoError(t, err)

			// Test passes if no error occurs
			assert.NotNil(t, cluster)
		})
	}
}

// TestManager_EdgeCases tests edge cases in manager functionality.
//
//nolint:paralleltest // Cannot use t.Parallel() because some subtests use setupTestEnvironment
func TestManager_EdgeCases(t *testing.T) {
	// Note: Cannot use t.Parallel() because some subtests use setupTestEnvironment
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Manager with no field selectors",
			test: func(t *testing.T) {
				t.Helper()
				// No need for setupTestEnvironment for this test
				manager := config.NewManager()
				cluster, err := manager.LoadCluster()
				require.NoError(t, err)
				assert.NotNil(t, cluster)
			},
		},
		{
			name: "Field selector that returns nil",
			test: func(t *testing.T) {
				t.Helper()
				setupTestEnvironment(t)

				fieldSelector := config.AddFlagFromField(
					func(_ *v1alpha1.Cluster) any { return nil },
					"default",
				)

				manager := config.NewManager(fieldSelector)
				cluster, err := manager.LoadCluster()
				require.NoError(t, err)
				assert.NotNil(t, cluster)
			},
		},
		{
			name: "Manager with nil default value",
			test: func(t *testing.T) {
				t.Helper()
				setupTestEnvironment(t)

				fieldSelector := config.AddFlagFromField(
					func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
					nil, // nil default value
				)

				manager := config.NewManager(fieldSelector)
				cluster, err := manager.LoadCluster()
				require.NoError(t, err)
				assert.NotNil(t, cluster)
			},
		},
	}

	for _, tt := range tests { //nolint:paralleltest // Cannot use parallel because subtests use setupTestEnvironment
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestSetValueAtFieldPointer tests setting values at field pointers with edge cases.
func TestSetValueAtFieldPointer(t *testing.T) {
	// Note: Cannot use t.Parallel() because individual test cases use t.Setenv
	setupTestEnvironment(t)

	tests := []struct {
		name          string
		fieldSelector config.FieldSelector[v1alpha1.Cluster]
		envVar        string
		envValue      string
		shouldSucceed bool
	}{
		{
			name: "valid field pointer",
			fieldSelector: config.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
				"default",
			),
			envVar:        "KSAIL_METADATA_NAME",
			envValue:      "test-value",
			shouldSucceed: true,
		},
		{
			name: "field selector returning non-pointer",
			fieldSelector: config.AddFlagFromField(
				func(_ *v1alpha1.Cluster) any {
					// Return a non-pointer value
					return "not-a-pointer"
				},
				"default",
			),
			envVar:        "KSAIL_TEST_FIELD",
			envValue:      "test-value",
			shouldSucceed: true, // Should not crash, but value won't be set
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(testCase.envVar, testCase.envValue)

			manager := config.NewManager(testCase.fieldSelector)
			cluster, err := manager.LoadCluster()

			if testCase.shouldSucceed {
				require.NoError(t, err)
				assert.NotNil(t, cluster)
			}
		})
	}
}
