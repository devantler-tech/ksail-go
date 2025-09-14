package ksail_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCase represents a test case structure for AddFlagFromField tests.
type testCase struct {
	name         string
	selector     func(*v1alpha1.Cluster) any
	defaultValue any
	description  []string
	expectedDesc string
}

// runAddFlagFromFieldTests is a helper function to run multiple test cases.
func runAddFlagFromFieldTests(t *testing.T, tests []testCase) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testAddFlagFromFieldScenario(
				t,
				testCase.selector,
				testCase.defaultValue,
				testCase.description,
				testCase.expectedDesc,
			)
		})
	}
}

// TestFieldSelector_StructureAndTypes tests the FieldSelector struct and types.
func TestFieldSelector_StructureAndTypes(t *testing.T) {
	t.Parallel()

	// Test that FieldSelector has the expected structure
	selector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
		Description:  "Test description",
		DefaultValue: "test-value",
	}

	require.NotNil(t, selector.Selector)
	assert.Equal(t, "Test description", selector.Description)
	assert.Equal(t, "test-value", selector.DefaultValue)

	// Test that the selector function works
	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	require.NotNil(t, result)

	// Verify it returns a pointer to the correct field
	namePtr, ok := result.(*string)
	require.True(t, ok, "Selector should return a pointer to string")
	assert.Equal(t, &cluster.Metadata.Name, namePtr)
}

// TestAddFlagFromField_MetadataAndBasicFields tests AddFlagFromField with metadata and basic spec fields.
func TestAddFlagFromField_MetadataAndBasicFields(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:         "Metadata.Name field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			defaultValue: "test-cluster",
			description:  []string{"Cluster name"},
			expectedDesc: "Cluster name",
		},
		{
			name:         "Spec.Distribution field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			defaultValue: v1alpha1.DistributionKind,
			description:  []string{"Kubernetes distribution"},
			expectedDesc: "Kubernetes distribution",
		},
		{
			name:         "Spec.SourceDirectory field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			defaultValue: "k8s",
			description:  []string{"Source directory"},
			expectedDesc: "Source directory",
		},
		{
			name:         "Spec.ReconciliationTool field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
			defaultValue: v1alpha1.ReconciliationToolFlux,
			description:  []string{"Reconciliation tool"},
			expectedDesc: "Reconciliation tool",
		},
	}

	runAddFlagFromFieldTests(t, tests)
}

// TestAddFlagFromField_ConnectionFields tests AddFlagFromField with connection fields.
func TestAddFlagFromField_ConnectionFields(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:         "Spec.Connection.Context field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			defaultValue: "my-context",
			description:  []string{"Kubernetes context"},
			expectedDesc: "Kubernetes context",
		},
		{
			name:         "Spec.Connection.Kubeconfig field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			defaultValue: "~/.kube/config",
			description:  []string{"Kubeconfig path"},
			expectedDesc: "Kubeconfig path",
		},
	}

	runAddFlagFromFieldTests(t, tests)
}

// TestAddFlagFromField_NetworkingComponents tests AddFlagFromField with networking components.
func TestAddFlagFromField_NetworkingComponents(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:         "Spec.CNI field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
			defaultValue: v1alpha1.CNICilium,
			description:  []string{"CNI plugin"},
			expectedDesc: "CNI plugin",
		},
		{
			name:         "Spec.CSI field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
			defaultValue: v1alpha1.CSILocalPathStorage,
			description:  []string{"CSI driver"},
			expectedDesc: "CSI driver",
		},
		{
			name:         "Spec.IngressController field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
			defaultValue: v1alpha1.IngressControllerTraefik,
			description:  []string{"Ingress controller"},
			expectedDesc: "Ingress controller",
		},
		{
			name:         "Spec.GatewayController field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
			defaultValue: v1alpha1.GatewayControllerTraefik,
			description:  []string{"Gateway controller"},
			expectedDesc: "Gateway controller",
		},
	}

	runAddFlagFromFieldTests(t, tests)
}

// TestAddFlagFromField_DescriptionHandling tests AddFlagFromField with various description scenarios.
func TestAddFlagFromField_DescriptionHandling(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:         "No description provided",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			defaultValue: "test",
			description:  []string{},
			expectedDesc: "",
		},
		{
			name:         "Empty description provided",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			defaultValue: "test",
			description:  []string{""},
			expectedDesc: "",
		},
		{
			name:         "Multiple descriptions (takes first)",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			defaultValue: "test",
			description:  []string{"First description", "Second description"},
			expectedDesc: "First description",
		},
		{
			name:         "Nil default value",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			defaultValue: nil,
			description:  []string{"Test with nil default"},
			expectedDesc: "Test with nil default",
		},
	}

	runAddFlagFromFieldTests(t, tests)
}

// testAddFlagFromFieldScenario is a helper function to test AddFlagFromField scenarios.
func testAddFlagFromFieldScenario(
	t *testing.T,
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description []string,
	expectedDesc string,
) {
	t.Helper()

	fieldSelector := ksail.AddFlagFromField(
		selector,
		defaultValue,
		description...,
	)

	require.NotNil(t, fieldSelector.Selector)
	assert.Equal(t, expectedDesc, fieldSelector.Description)
	assert.Equal(t, defaultValue, fieldSelector.DefaultValue)

	// Test that the selector function works correctly
	cluster := &v1alpha1.Cluster{}
	result := fieldSelector.Selector(cluster)
	require.NotNil(t, result, "Selector should return a non-nil pointer")
}

// TestAddFlagFromField_BasicTypes tests AddFlagFromField with basic value types.
func TestAddFlagFromField_BasicTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultValue any
		expectedType string
	}{
		{
			name:         "String default",
			defaultValue: "test-string",
			expectedType: "string",
		},
		{
			name:         "Boolean default",
			defaultValue: true,
			expectedType: "bool",
		},
		{
			name:         "Integer default",
			defaultValue: 42,
			expectedType: "int",
		},
		{
			name:         "Nil default",
			defaultValue: nil,
			expectedType: "<nil>",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testAddFlagFromFieldType(t, testCase.defaultValue, testCase.expectedType)
		})
	}
}

// TestAddFlagFromField_EnumTypes tests AddFlagFromField with enum value types.
func TestAddFlagFromField_EnumTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultValue any
		expectedType string
	}{
		{
			name:         "Distribution enum",
			defaultValue: v1alpha1.DistributionKind,
			expectedType: "v1alpha1.Distribution",
		},
		{
			name:         "ReconciliationTool enum",
			defaultValue: v1alpha1.ReconciliationToolFlux,
			expectedType: "v1alpha1.ReconciliationTool",
		},
		{
			name:         "CNI enum",
			defaultValue: v1alpha1.CNICilium,
			expectedType: "v1alpha1.CNI",
		},
		{
			name:         "CSI enum",
			defaultValue: v1alpha1.CSILocalPathStorage,
			expectedType: "v1alpha1.CSI",
		},
		{
			name:         "IngressController enum",
			defaultValue: v1alpha1.IngressControllerTraefik,
			expectedType: "v1alpha1.IngressController",
		},
		{
			name:         "GatewayController enum",
			defaultValue: v1alpha1.GatewayControllerTraefik,
			expectedType: "v1alpha1.GatewayController",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testAddFlagFromFieldType(t, testCase.defaultValue, testCase.expectedType)
		})
	}
}

// testAddFlagFromFieldType is a helper function to test AddFlagFromField with different types.
func testAddFlagFromFieldType(t *testing.T, defaultValue any, expectedType string) {
	t.Helper()

	selector := ksail.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
		defaultValue,
		"Test description",
	)

	assert.Equal(t, defaultValue, selector.DefaultValue)

	// Verify the type of the default value
	if defaultValue != nil {
		actualType := selector.DefaultValue

		switch expectedType {
		case "string":
			_, ok := actualType.(string)
			assert.True(t, ok, "Expected string type")
		case "bool":
			_, ok := actualType.(bool)
			assert.True(t, ok, "Expected bool type")
		case "int":
			_, ok := actualType.(int)
			assert.True(t, ok, "Expected int type")
		default:
			// For enum types, check that it's not nil
			assert.NotNil(t, actualType, "Expected non-nil value for enum type")
		}
	} else {
		assert.Nil(t, selector.DefaultValue, "Expected nil default value")
	}
}

// TestFieldSelector_MetadataFields tests compile-time safety for metadata field selectors.
func TestFieldSelector_MetadataFields(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{}

	metadataSelectors := []struct {
		name     string
		selector func(*v1alpha1.Cluster) any
	}{
		{
			name:     "Metadata.Name",
			selector: func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
		},
		{
			name:     "Metadata.Namespace",
			selector: func(c *v1alpha1.Cluster) any { return &c.Metadata.Namespace },
		},
	}

	for _, testCase := range metadataSelectors {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testFieldSelector(t, cluster, testCase.selector, testCase.name)
		})
	}
}

// TestFieldSelector_SpecBasicFields tests compile-time safety for basic spec field selectors.
func TestFieldSelector_SpecBasicFields(t *testing.T) {
	t.Parallel()

	specBasicSelectors := []fieldSelectorTestCase{
		{
			name:     "Spec.Distribution",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		},
		{
			name:     "Spec.DistributionConfig",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		},
		{
			name:     "Spec.SourceDirectory",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		},
		{
			name:     "Spec.ReconciliationTool",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
		},
	}

	runFieldSelectorTests(t, specBasicSelectors)
}

// TestFieldSelector_ConnectionFields tests compile-time safety for connection field selectors.
func TestFieldSelector_ConnectionFields(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{}

	connectionSelectors := []struct {
		name     string
		selector func(*v1alpha1.Cluster) any
	}{
		{
			name:     "Spec.Connection.Context",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
		},
		{
			name:     "Spec.Connection.Kubeconfig",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
		},
		{
			name:     "Spec.Connection.Timeout",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
		},
	}

	for _, testCase := range connectionSelectors {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testFieldSelector(t, cluster, testCase.selector, testCase.name)
		})
	}
}

// TestFieldSelector_NetworkingFields tests compile-time safety for networking field selectors.
func TestFieldSelector_NetworkingFields(t *testing.T) {
	t.Parallel()

	networkingSelectors := []fieldSelectorTestCase{
		{
			name:     "Spec.CNI",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
		},
		{
			name:     "Spec.CSI",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
		},
		{
			name:     "Spec.IngressController",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
		},
		{
			name:     "Spec.GatewayController",
			selector: func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
		},
	}

	runFieldSelectorTests(t, networkingSelectors)
}

// fieldSelectorTestCase represents a test case for field selector functionality.
type fieldSelectorTestCase struct {
	name     string
	selector func(*v1alpha1.Cluster) any
}

// runFieldSelectorTests runs a series of field selector tests.
func runFieldSelectorTests(t *testing.T, testCases []fieldSelectorTestCase) {
	t.Helper()

	cluster := v1alpha1.NewCluster()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testFieldSelector(t, cluster, testCase.selector, testCase.name)
		})
	}
}

// testFieldSelector is a helper function to test field selector functionality.
func testFieldSelector(
	t *testing.T,
	cluster *v1alpha1.Cluster,
	selector func(*v1alpha1.Cluster) any,
	name string,
) {
	t.Helper()

	// Test that the selector compiles and returns a non-nil pointer
	result := selector(cluster)
	require.NotNil(t, result, "Selector %s should return non-nil pointer", name)

	// Create a field selector using AddFlagFromField
	fieldSelector := ksail.AddFlagFromField(
		selector,
		"default-value",
		"Test description",
	)

	require.NotNil(t, fieldSelector.Selector)
	assert.Equal(t, "Test description", fieldSelector.Description)
	assert.Equal(t, "default-value", fieldSelector.DefaultValue)
}
