package configmanager_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
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

type standardFieldSelectorCase struct {
	name            string
	factory         func() configmanager.FieldSelector[v1alpha1.Cluster]
	expectedDesc    string
	expectedDefault any
	assertPointer   func(*testing.T, *v1alpha1.Cluster, any)
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
func TestStandardFieldSelectors(t *testing.T) {
	t.Parallel()

	cases := standardFieldSelectorCases()

	runStandardFieldSelectorTests(t, cases)
}

func standardFieldSelectorCases() []standardFieldSelectorCase {
	return []standardFieldSelectorCase{
		{
			name:            "distribution",
			factory:         configmanager.DefaultDistributionFieldSelector,
			expectedDesc:    "Kubernetes distribution to use",
			expectedDefault: v1alpha1.DistributionKind,
			assertPointer:   assertDistributionSelector,
		},
		{
			name:            "source directory",
			factory:         configmanager.StandardSourceDirectoryFieldSelector,
			expectedDesc:    "Directory containing workloads to deploy",
			expectedDefault: "k8s",
			assertPointer:   assertSourceDirectorySelector,
		},
		{
			name:            "distribution config",
			factory:         configmanager.DefaultDistributionConfigFieldSelector,
			expectedDesc:    "Configuration file for the distribution",
			expectedDefault: "kind.yaml",
			assertPointer:   assertDistributionConfigSelector,
		},
		{
			name:            "context",
			factory:         configmanager.DefaultContextFieldSelector,
			expectedDesc:    "Kubernetes context of cluster",
			expectedDefault: "kind-kind",
			assertPointer:   assertContextSelector,
		},
	}
}

func specFieldTestCases() []testCase {
	return []testCase{
		{
			name:         "Spec.Distribution field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			defaultValue: v1alpha1.DistributionKind,
			description:  []string{"Kubernetes distribution to use"},
			expectedDesc: "Kubernetes distribution to use",
		},
		{
			name:         "Spec.SourceDirectory field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			defaultValue: "k8s",
			description:  []string{"Directory containing workloads to deploy"},
			expectedDesc: "Directory containing workloads to deploy",
		},
		{
			name:         "Spec.DistributionConfig field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			defaultValue: "kind.yaml",
			description:  []string{"Configuration file for the distribution"},
			expectedDesc: "Configuration file for the distribution",
		},
		{
			name:         "Spec.Connection.Context field",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			defaultValue: "kind-kind",
			description:  []string{"Kubernetes context of cluster"},
			expectedDesc: "Kubernetes context of cluster",
		},
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
}

func assertDistributionSelector(t *testing.T, cluster *v1alpha1.Cluster, ptr any) {
	t.Helper()
	assertPointerSame(t, ptr, &cluster.Spec.Distribution)
}

func assertSourceDirectorySelector(t *testing.T, cluster *v1alpha1.Cluster, ptr any) {
	t.Helper()
	assertPointerSame(t, ptr, &cluster.Spec.SourceDirectory)
}

func assertDistributionConfigSelector(t *testing.T, cluster *v1alpha1.Cluster, ptr any) {
	t.Helper()
	assertPointerSame(t, ptr, &cluster.Spec.DistributionConfig)
}

func assertContextSelector(t *testing.T, cluster *v1alpha1.Cluster, ptr any) {
	t.Helper()
	assertPointerSame(t, ptr, &cluster.Spec.Connection.Context)
}

func runStandardFieldSelectorTests(t *testing.T, cases []standardFieldSelectorCase) {
	t.Helper()

	for _, testCase := range cases {
		caseData := testCase
		t.Run(caseData.name, func(t *testing.T) {
			t.Parallel()

			cluster := &v1alpha1.Cluster{}
			selector := caseData.factory()

			assert.Equal(t, caseData.expectedDesc, selector.Description)
			assert.Equal(t, caseData.expectedDefault, selector.DefaultValue)

			pointer := selector.Selector(cluster)
			caseData.assertPointer(t, cluster, pointer)
		})
	}
}

func assertPointerSame[T any](t *testing.T, actual any, expected *T) {
	t.Helper()

	value, ok := actual.(*T)
	require.True(t, ok)
	assert.Same(t, expected, value)
}

// TestAddFlagFromField_SpecFields tests AddFlagFromField with spec fields.
func TestAddFlagFromFieldSpecFields(t *testing.T) {
	t.Parallel()
	runAddFlagFromFieldTests(t, specFieldTestCases())
}

// TestAddFlagFromField_DescriptionHandling tests AddFlagFromField with various description scenarios.
func TestAddFlagFromFieldDescriptionHandling(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:         "No description provided",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			defaultValue: "test",
			description:  []string{},
			expectedDesc: "",
		},
		{
			name:         "Empty description provided",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			defaultValue: "test",
			description:  []string{""},
			expectedDesc: "",
		},
		{
			name:         "Multiple descriptions (takes first)",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			defaultValue: "test",
			description:  []string{"First description", "Second description"},
			expectedDesc: "First description",
		},
		{
			name:         "Nil default value",
			selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
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

	fieldSelector := configmanager.AddFlagFromField(
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
func TestAddFlagFromFieldBasicTypes(t *testing.T) {
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

	runTypeTestCases(t, tests)
}

// TestAddFlagFromField_EnumTypes tests AddFlagFromField with enum value types.
func TestAddFlagFromFieldEnumTypes(t *testing.T) {
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

	runTypeTestCases(t, tests)
}

// runTypeTestCases is a helper function to run type test cases with common loop pattern.
func runTypeTestCases(t *testing.T, tests []struct {
	name         string
	defaultValue any
	expectedType string
},
) {
	t.Helper()

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

	selector := configmanager.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
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

// TestFieldSelector_SpecBasicFields tests compile-time safety for basic spec field selectors.
func TestFieldSelectorSpecBasicFields(t *testing.T) {
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
func TestFieldSelectorConnectionFields(t *testing.T) {
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
func TestFieldSelectorNetworkingFields(t *testing.T) {
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
	fieldSelector := configmanager.AddFlagFromField(
		selector,
		"default-value",
		"Test description",
	)

	require.NotNil(t, fieldSelector.Selector)
	assert.Equal(t, "Test description", fieldSelector.Description)
	assert.Equal(t, "default-value", fieldSelector.DefaultValue)
}

func TestAddFlagFromFieldUsesOptionalDescription(t *testing.T) {
	t.Parallel()

	selector := configmanager.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		v1alpha1.DistributionKind,
	)

	assert.Empty(t, selector.Description)

	withDescription := configmanager.AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		v1alpha1.DistributionKind,
		"Distribution help",
	)

	assert.Equal(t, "Distribution help", withDescription.Description)
}

func TestDefaultClusterFieldSelectorsProvideDefaults(t *testing.T) {
	t.Parallel()

	selectors := configmanager.DefaultClusterFieldSelectors()
	require.Len(t, selectors, 2)

	cluster := v1alpha1.NewCluster()

	for _, selector := range selectors {
		field := selector.Selector(cluster)

		if distribution, ok := field.(*v1alpha1.Distribution); ok {
			assert.Equal(t, v1alpha1.DistributionKind, selector.DefaultValue)

			*distribution = v1alpha1.DistributionK3d
			assert.Equal(t, v1alpha1.DistributionK3d, *distribution)

			continue
		}

		pathPtr, ok := field.(*string)
		require.True(t, ok, "selector did not return supported pointer type")
		assert.Equal(t, "kind.yaml", selector.DefaultValue)

		*pathPtr = "custom.yaml"
		assert.Equal(t, "custom.yaml", *pathPtr)
	}
}

func TestDefaultContextFieldSelector(t *testing.T) {
	t.Parallel()

	selector := configmanager.DefaultContextFieldSelector()
	cluster := v1alpha1.NewCluster()

	ptr, ok := selector.Selector(cluster).(*string)
	require.True(t, ok, "expected selector to return *string")
	assert.Equal(t, "kind-kind", selector.DefaultValue)

	*ptr = "custom"
	assert.Equal(t, "custom", cluster.Spec.Connection.Context)
	assert.NotEmpty(t, selector.Description)
}
