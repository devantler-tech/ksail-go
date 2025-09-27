package configmanager_test

import (
	"io"
	"testing"
	"time"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
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
	manager *configmanager.ConfigManager,
	tests []flagNameTestCase,
) {
	t.Helper()
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
	fieldSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	manager := configmanager.NewConfigManager(io.Discard, fieldSelectors...)
	cmd := &cobra.Command{Use: "test"}
	manager.AddFlagsFromFields(cmd)

	return cmd
}

// TestAddFlagFromField tests the addFlagFromField method with various field types and scenarios.
// getBasicFieldTests returns test cases for basic field testing.
func getBasicFieldTests() []struct {
	name          string
	fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
	expectedFlag  string
	expectedType  string
} {
	return []struct {
		name          string
		fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "Distribution field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
				v1alpha1.DistributionKind,
				"Kubernetes distribution",
			),
			expectedFlag: "distribution",
			expectedType: "Distribution",
		},
		{
			name: "SourceDirectory field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
				"k8s",
				"Source directory",
			),
			expectedFlag: "source-directory",
			expectedType: "string",
		},
		{
			name: "ReconciliationTool field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
				v1alpha1.ReconciliationToolFlux,
				"Reconciliation tool",
			),
			expectedFlag: "reconciliation-tool",
			expectedType: "ReconciliationTool",
		},
	}
}

func TestAddFlagFromField(t *testing.T) {
	t.Parallel()

	t.Run("basic fields", func(t *testing.T) {
		t.Parallel()
		testAddFlagFromFieldCases(t, getBasicFieldTests())
	})

	t.Run("connection fields", func(t *testing.T) {
		t.Parallel()
		testAddFlagFromFieldCases(t, getConnectionFieldTests())
	})

	t.Run("networking fields", func(t *testing.T) {
		t.Parallel()
		testAddFlagFromFieldCases(t, getNetworkingFieldTests())
	})

	t.Run("error handling", func(t *testing.T) {
		t.Parallel()
		testAddFlagFromFieldErrorHandling(t)
	})
}

// getConnectionFieldTests returns test cases for connection field testing.
func getConnectionFieldTests() []struct {
	name          string
	fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
	expectedFlag  string
	expectedType  string
} {
	return []struct {
		name          string
		fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "Context field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
				"",
				"Kubernetes context",
			),
			expectedFlag: "context",
			expectedType: "string",
		},
		{
			name: "Timeout field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
				metav1.Duration{Duration: 5 * time.Minute},
				"Connection timeout",
			),
			expectedFlag: "timeout",
			expectedType: "duration",
		},
	}
}

// getNetworkingFieldTests returns test cases for networking field testing.
func getNetworkingFieldTests() []struct {
	name          string
	fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
	expectedFlag  string
	expectedType  string
} {
	return []struct {
		name          string
		fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
		expectedFlag  string
		expectedType  string
	}{
		{
			name: "CNI field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
				v1alpha1.CNICilium,
				"CNI plugin",
			),
			expectedFlag: "cni",
			expectedType: "CNI",
		},
		{
			name: "CSI field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
				v1alpha1.CSILocalPathStorage,
				"CSI driver",
			),
			expectedFlag: "csi",
			expectedType: "CSI",
		},
		{
			name: "IngressController field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
				v1alpha1.IngressControllerTraefik,
				"Ingress controller",
			),
			expectedFlag: "ingress-controller",
			expectedType: "IngressController",
		},
		{
			name: "GatewayController field",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
				v1alpha1.GatewayControllerTraefik,
				"Gateway controller",
			),
			expectedFlag: "gateway-controller",
			expectedType: "GatewayController",
		},
	}
}

// testAddFlagFromFieldErrorHandling tests error handling scenarios for AddFlagFromField.
func testAddFlagFromFieldErrorHandling(t *testing.T) {
	t.Helper()

	tests := []struct {
		name          string
		fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
		expectSkip    bool
	}{
		{
			name: "Nil field selector",
			fieldSelector: configmanager.FieldSelector[v1alpha1.Cluster]{
				Selector: func(_ *v1alpha1.Cluster) any { return nil },
			},
			expectSkip: true,
		},
		{
			name: "Valid field selector",
			fieldSelector: configmanager.AddFlagFromField(
				func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
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

// testAddFlagFromFieldCases is a helper function to test field selector functionality.
func testAddFlagFromFieldCases(t *testing.T, tests []struct {
	name          string
	fieldSelector configmanager.FieldSelector[v1alpha1.Cluster]
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

// TestGenerateFlagName tests flag name generation for various field types.
func TestGenerateFlagName(t *testing.T) {
	t.Parallel()

	manager := configmanager.NewConfigManager(io.Discard)

	tests := []flagNameTestCase{
		{"Distribution field", &manager.Config.Spec.Distribution, "distribution"},
		{
			"DistributionConfig field",
			&manager.Config.Spec.DistributionConfig,
			"distribution-config",
		},
		{"SourceDirectory field", &manager.Config.Spec.SourceDirectory, "source-directory"},
		{
			"ReconciliationTool field",
			&manager.Config.Spec.ReconciliationTool,
			"reconciliation-tool",
		},
		{"Context field", &manager.Config.Spec.Connection.Context, "context"},
		{"Kubeconfig field", &manager.Config.Spec.Connection.Kubeconfig, "kubeconfig"},
		{"Timeout field", &manager.Config.Spec.Connection.Timeout, "timeout"},
		{"CNI field", &manager.Config.Spec.CNI, "cni"},
		{"CSI field", &manager.Config.Spec.CSI, "csi"},
		{
			"IngressController field",
			&manager.Config.Spec.IngressController,
			"ingress-controller",
		},
		{
			"GatewayController field",
			&manager.Config.Spec.GatewayController,
			"gateway-controller",
		},
	}

	runFlagNameGenerationTests(t, manager, tests)
}

// testFlagNameGeneration is a helper function to test flag name generation.
func testFlagNameGeneration(
	t *testing.T,
	manager *configmanager.ConfigManager,
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

	manager := configmanager.NewConfigManager(io.Discard)

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
