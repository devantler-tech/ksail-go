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

// flagNameTestCase represents a test case for flag name generation.
type flagNameTestCase struct {
	name     string
	fieldPtr any
	expected string
}

// runFlagNameGenerationTests is a helper function to run multiple flag name generation test cases.
func runFlagNameGenerationTests(
	t *testing.T,
	manager *ksail.ConfigManager,
	tests []flagNameTestCase,
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testFlagNameGeneration(t, manager, testCase.fieldPtr, testCase.expected)
		})
	}
}

// setupFlagBindingTest creates a command for testing flag binding.
func setupFlagBindingTest(
	fieldSelectors ...ksail.FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	manager := ksail.NewConfigManager(fieldSelectors...)
	cmd := &cobra.Command{Use: "test"}
	manager.AddFlagsFromFields(cmd)

	return cmd
}

// TestManager_addFlagFromField_BasicFields tests basic field selectors.
func TestManageraddFlagFromFieldBasicFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "Distribution field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
				v1alpha1.DistributionKind,
				"Kubernetes distribution",
			),
			expectedFlag: "distribution",
			expectedType: "Distribution",
		},
		{
			name: "SourceDirectory field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
				"k8s",
				"Source directory",
			),
			expectedFlag: "source-directory",
			expectedType: "string",
		},
		{
			name: "ReconciliationTool field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
				v1alpha1.ReconciliationToolFlux,
				"Reconciliation tool",
			),
			expectedFlag: "reconciliation-tool",
			expectedType: "ReconciliationTool",
		},
	}

	testAddFlagFromFieldCases(t, tests)
}

// TestManager_addFlagFromField_ConnectionFields tests connection-related field selectors.
func TestManageraddFlagFromFieldConnectionFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "Context field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
				"",
				"Kubernetes context",
			),
			expectedFlag: "context",
			expectedType: "string",
		},
		{
			name: "Timeout field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
				metav1.Duration{Duration: 5 * time.Minute},
				"Connection timeout",
			),
			expectedFlag: "timeout",
			expectedType: "duration",
		},
	}

	testAddFlagFromFieldCases(t, tests)
}

// TestManager_addFlagFromField_NetworkingFields tests networking-related field selectors.
func TestManageraddFlagFromFieldNetworkingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "CNI field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
				v1alpha1.CNICilium,
				"CNI plugin",
			),
			expectedFlag: "cni",
			expectedType: "CNI",
		},
		{
			name: "CSI field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
				v1alpha1.CSILocalPathStorage,
				"CSI driver",
			),
			expectedFlag: "csi",
			expectedType: "CSI",
		},
		{
			name: "IngressController field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
				v1alpha1.IngressControllerTraefik,
				"Ingress controller",
			),
			expectedFlag: "ingress-controller",
			expectedType: "IngressController",
		},
		{
			name: "GatewayController field",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
				v1alpha1.GatewayControllerTraefik,
				"Gateway controller",
			),
			expectedFlag: "gateway-controller",
			expectedType: "GatewayController",
		},
	}

	testAddFlagFromFieldCases(t, tests)
}

// testAddFlagFromFieldCases is a helper function to test field selector functionality.
func testAddFlagFromFieldCases(t *testing.T, tests []struct {
	name          string
	fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
	expectedFlag  string
	expectedType  string
},
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := setupFlagBindingTest(testCase.fieldSelector)

			// Check that the flag was added
			flag := cmd.Flags().Lookup(testCase.expectedFlag)
			require.NotNil(t, flag, "flag %s should exist", testCase.expectedFlag)
			assert.Equal(t, testCase.fieldSelector.Description, flag.Usage)

			// Check flag type
			assert.Equal(t, testCase.expectedType, flag.Value.Type())
		})
	}
}

// TestManager_GenerateFlagName_BasicFields tests flag name generation for basic spec fields.
func TestGenerateFlagName(t *testing.T) {
	t.Parallel()

	manager := ksail.NewConfigManager()

	t.Run("basic fields", func(t *testing.T) {
		t.Parallel()

		tests := []flagNameTestCase{
			{
				name:     "Distribution field",
				fieldPtr: &manager.Config.Spec.Distribution,
				expected: "distribution",
			},
			{
				name:     "DistributionConfig field",
				fieldPtr: &manager.Config.Spec.DistributionConfig,
				expected: "distribution-config",
			},
			{
				name:     "SourceDirectory field",
				fieldPtr: &manager.Config.Spec.SourceDirectory,
				expected: "source-directory",
			},
			{
				name:     "ReconciliationTool field",
				fieldPtr: &manager.Config.Spec.ReconciliationTool,
				expected: "reconciliation-tool",
			},
		}

		runFlagNameGenerationTests(t, manager, tests)
	})

	t.Run("connection fields", func(t *testing.T) {
		t.Parallel()

		tests := []flagNameTestCase{
			{
				name:     "Context field",
				fieldPtr: &manager.Config.Spec.Connection.Context,
				expected: "context",
			},
			{
				name:     "Kubeconfig field",
				fieldPtr: &manager.Config.Spec.Connection.Kubeconfig,
				expected: "kubeconfig",
			},
			{
				name:     "Timeout field",
				fieldPtr: &manager.Config.Spec.Connection.Timeout,
				expected: "timeout",
			},
		}

		runFlagNameGenerationTests(t, manager, tests)
	})

	t.Run("networking fields", func(t *testing.T) {
		t.Parallel()

		tests := []flagNameTestCase{
			{
				name:     "CNI field",
				fieldPtr: &manager.Config.Spec.CNI,
				expected: "cni",
			},
			{
				name:     "CSI field",
				fieldPtr: &manager.Config.Spec.CSI,
				expected: "csi",
			},
			{
				name:     "IngressController field",
				fieldPtr: &manager.Config.Spec.IngressController,
				expected: "ingress-controller",
			},
			{
				name:     "GatewayController field",
				fieldPtr: &manager.Config.Spec.GatewayController,
				expected: "gateway-controller",
			},
		}

		runFlagNameGenerationTests(t, manager, tests)
	})
}

// testFlagNameGeneration is a helper function to test flag name generation.
func testFlagNameGeneration(
	t *testing.T,
	manager *ksail.ConfigManager,
	fieldPtr any,
	expected string,
) {
	t.Helper()

	result := manager.GenerateFlagName(fieldPtr)
	assert.Equal(t, expected, result)
}

// TestManager_GenerateShorthand tests the GenerateShorthand method.
func TestGenerateShorthand(t *testing.T) {
	t.Parallel()

	manager := ksail.NewConfigManager()

	tests := []struct {
		name     string
		flagName string
		expected string
	}{
		{
			name:     "distribution flag",
			flagName: "distribution",
			expected: "d",
		},
		{
			name:     "context flag",
			flagName: "context",
			expected: "c",
		},
		{
			name:     "kubeconfig flag",
			flagName: "kubeconfig",
			expected: "k",
		},
		{
			name:     "timeout flag",
			flagName: "timeout",
			expected: "t",
		},
		{
			name:     "source-directory flag",
			flagName: "source-directory",
			expected: "s",
		},
		{
			name:     "reconciliation-tool flag",
			flagName: "reconciliation-tool",
			expected: "r",
		},
		{
			name:     "distribution-config flag (no shorthand)",
			flagName: "distribution-config",
			expected: "",
		},
		{
			name:     "unknown flag (no shorthand)",
			flagName: "unknown-flag",
			expected: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Call the public method directly
			result := manager.GenerateShorthand(testCase.flagName)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

// TestManager_addFlagFromField_ErrorPaths tests error handling in addFlagFromField.
func TestManageraddFlagFromFieldErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectSkip    bool
	}{
		{
			name: "Nil field selector",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector: func(_ *v1alpha1.Cluster) any { return nil },
			},
			expectSkip: true,
		},
		{
			name: "Valid field selector",
			fieldSelector: ksail.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
				"test",
				"Test field",
			),
			expectSkip: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := setupFlagBindingTest(testCase.fieldSelector)

			if testCase.expectSkip {
				// Should have no flags when selector returns nil
				assert.False(t, cmd.Flags().HasFlags())
			} else {
				// Should have flags when selector is valid
				assert.True(t, cmd.Flags().HasFlags())
			}
		})
	}
}

// TestManager_addFlagFromField_EnumTypesWithNilDefault tests enum field types with nil defaults.
func TestManageraddFlagFromFieldEnumTypesWithNilDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedType  string
	}{
		{
			name: "ReconciliationTool with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
				DefaultValue: nil,
				Description:  "Tool with nil default",
			},
			expectedType: "ReconciliationTool",
		},
		{
			name: "CNI with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
				DefaultValue: nil,
				Description:  "CNI with nil default",
			},
			expectedType: "CNI",
		},
		{
			name: "CSI with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
				DefaultValue: nil,
				Description:  "CSI with nil default",
			},
			expectedType: "CSI",
		},
	}

	runNilDefaultTestCases(t, tests)
}

// TestManager_addFlagFromField_ControllerTypesWithNilDefault tests controller field types with nil defaults.
func TestManageraddFlagFromFieldControllerTypesWithNilDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedType  string
	}{
		{
			name: "IngressController with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
				DefaultValue: nil,
				Description:  "Ingress with nil default",
			},
			expectedType: "IngressController",
		},
		{
			name: "GatewayController with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
				DefaultValue: nil,
				Description:  "Gateway with nil default",
			},
			expectedType: "GatewayController",
		},
	}

	runNilDefaultTestCases(t, tests)
}

// TestManager_addFlagFromField_BasicTypesWithNilDefault tests basic field types with nil defaults.
func TestManageraddFlagFromFieldBasicTypesWithNilDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
		expectedType  string
	}{
		{
			name: "String field with nil default",
			fieldSelector: ksail.FieldSelector[v1alpha1.Cluster]{
				Selector:     func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
				DefaultValue: nil,
				Description:  "String with nil default",
			},
			expectedType: "string",
		},
	}

	runNilDefaultTestCases(t, tests)
}

// runNilDefaultTestCases is a helper function to run nil default test cases with common loop pattern.
func runNilDefaultTestCases(t *testing.T, tests []struct {
	name          string
	fieldSelector ksail.FieldSelector[v1alpha1.Cluster]
	expectedType  string
},
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testFieldTypeWithNilDefault(t, testCase.fieldSelector, testCase.expectedType)
		})
	}
}

// testFieldTypeWithNilDefault is a helper function to test field types with nil defaults.
func testFieldTypeWithNilDefault(
	t *testing.T,
	fieldSelector ksail.FieldSelector[v1alpha1.Cluster],
	expectedType string,
) {
	t.Helper()

	cmd := setupFlagBindingTest(fieldSelector)
	assertFlagTypeExists(t, cmd, expectedType)
}

// TestManager_addFlagFromField_TimeDuration tests time.Duration field type.
func TestManageraddFlagFromFieldTimeDuration(t *testing.T) {
	t.Parallel()

	// Test with metav1.Duration which has a time.Duration field
	fieldSelector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
		DefaultValue: metav1.Duration{Duration: 5 * time.Minute},
		Description:  "Test timeout duration",
	}

	cmd := setupFlagBindingTest(fieldSelector)
	assertFlagTypeExists(t, cmd, "duration")
}

// assertFlagTypeExists is a helper function to check if a flag of expected type exists.
func assertFlagTypeExists(t *testing.T, cmd *cobra.Command, expectedType string) {
	t.Helper()

	// Should have one flag
	assert.True(t, cmd.Flags().HasFlags())

	// Check flag type
	var flagFound bool

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() == expectedType {
			flagFound = true
		}
	})
	assert.True(t, flagFound, "Expected flag type %s not found", expectedType)
}

// TestManager_addFlagFromField_EmptyShorthand tests pflag.Value with empty shorthand.
func TestManageraddFlagFromFieldEmptyShorthand(t *testing.T) {
	t.Parallel()

	// Use distribution-config which should have empty shorthand according to GenerateShorthand tests
	fieldSelector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		DefaultValue: nil,
		Description:  "Distribution configuration",
	}

	cmd := setupFlagBindingTest(fieldSelector)

	// Should have one flag with no shorthand
	assert.True(t, cmd.Flags().HasFlags())

	var flagFoundWithoutShorthand bool

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "distribution-config" && flag.Shorthand == "" {
			flagFoundWithoutShorthand = true
		}
	})
	assert.True(t, flagFoundWithoutShorthand, "Expected flag without shorthand not found")
}

// testPflagValue is a custom type implementing pflag.Value for testing the default case.
type testPflagValue struct {
	value string
}

const testPflagValueType = "testPflagValue"

func (v *testPflagValue) Set(val string) error {
	v.value = val

	return nil
}

func (v *testPflagValue) String() string {
	return v.value
}

func (v *testPflagValue) Type() string {
	return testPflagValueType
}

// TestManager_setPflagValueDefault_StringDefaultValue tests the default case in setPflagValueDefault.
func TestManagersetPflagValueDefaultStringDefaultValue(t *testing.T) {
	t.Parallel()

	var testValue testPflagValue

	// Test with custom type using string default value
	fieldSelector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector: func(_ *v1alpha1.Cluster) any {
			return &testValue
		},
		DefaultValue: "test-string-value",
		Description:  "Test string default",
	}

	cmd := setupFlagBindingTest(fieldSelector)

	// Should have one flag
	assert.True(t, cmd.Flags().HasFlags())

	// The custom type should be handled by the default case
	var flagFound bool

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() == testPflagValueType && flag.Value.String() == "test-string-value" {
			flagFound = true
		}
	})
	assert.True(t, flagFound, "Expected custom flag with string default not found")
}

// TestManager_addFlagFromField_NilFieldPtr tests early return when fieldPtr is nil.
func TestManageraddFlagFromFieldNilFieldPtr(t *testing.T) {
	t.Parallel()

	fieldSelector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector: func(_ *v1alpha1.Cluster) any {
			return nil // Return nil to trigger early return
		},
		DefaultValue: "test",
		Description:  "Test nil field pointer",
	}

	cmd := setupFlagBindingTest(fieldSelector)

	// Should have no flags since fieldPtr is nil
	assert.False(t, cmd.Flags().HasFlags())
}

// TestManager_addFlagFromField_NonConvertiblePflagDefaultValue tests pflag.Value with non-string default.
func TestManageraddFlagFromFieldNonConvertiblePflagDefaultValue(t *testing.T) {
	t.Parallel()

	var testValue testPflagValue

	// Test with custom type using non-string/non-enum default value
	fieldSelector := ksail.FieldSelector[v1alpha1.Cluster]{
		Selector: func(_ *v1alpha1.Cluster) any {
			return &testValue
		},
		DefaultValue: 123, // int value should not be handled by any case in setPflagValueDefault
		Description:  "Test non-convertible default",
	}

	cmd := setupFlagBindingTest(fieldSelector)

	// Should have one flag, but the default won't be set since it's not a known type
	assert.True(t, cmd.Flags().HasFlags())

	// The custom type should still be created but with empty default
	var flagFound bool

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() == testPflagValueType {
			flagFound = true
		}
	})
	assert.True(t, flagFound, "Expected custom flag not found")
}
