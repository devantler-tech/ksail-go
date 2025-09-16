package ksail_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
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

// TestAddFlagFromField tests the addFlagFromField method with various field types and scenarios.
func TestAddFlagFromField(t *testing.T) {
	t.Parallel()

	t.Run("basic fields", func(t *testing.T) {
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
	})

	t.Run("connection fields", func(t *testing.T) {
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
	})

	t.Run("networking fields", func(t *testing.T) {
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
	})

	t.Run("error handling", func(t *testing.T) {
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
	})
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

// TestGenerateFlagName tests flag name generation for various field types.
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
