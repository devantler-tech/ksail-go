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

// TestManager_addFlagFromField tests the addFlagFromField method with different field types.
//
//nolint:funlen // Comprehensive test requires multiple test cases for coverage
func TestManager_addFlagFromField(t *testing.T) {
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := ksail.NewManager(testCase.fieldSelector)
			cmd := &cobra.Command{Use: "test"}

			manager.AddFlagsFromFields(cmd)

			// Check that the flag was added
			flag := cmd.Flags().Lookup(testCase.expectedFlag)
			require.NotNil(t, flag, "flag %s should exist", testCase.expectedFlag)
			assert.Equal(t, testCase.fieldSelector.Description, flag.Usage)

			// Check flag type
			assert.Equal(t, testCase.expectedType, flag.Value.Type())
		})
	}
}

// TestManager_GenerateFlagName tests the GenerateFlagName method.
//
//nolint:funlen // Comprehensive flag name generation test requires multiple test cases
func TestManager_GenerateFlagName(t *testing.T) {
	t.Parallel()

	manager := ksail.NewManager()

	tests := []struct {
		name     string
		fieldPtr any
		expected string
	}{
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
		{
			name:     "ReconciliationTool field",
			fieldPtr: &manager.Config.Spec.ReconciliationTool,
			expected: "reconciliation-tool",
		},
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Call the public method directly
			result := manager.GenerateFlagName(testCase.fieldPtr)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

// TestManager_GenerateShorthand tests the GenerateShorthand method.
func TestManager_GenerateShorthand(t *testing.T) {
	t.Parallel()

	manager := ksail.NewManager()

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
func TestManager_addFlagFromField_ErrorPaths(t *testing.T) {
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

			manager := ksail.NewManager(testCase.fieldSelector)
			cmd := &cobra.Command{Use: "test"}

			manager.AddFlagsFromFields(cmd)

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

// TestManager_addFlagFromField_AllFieldTypes tests all supported field types.
//
//nolint:funlen // Testing all field types requires comprehensive test cases
func TestManager_addFlagFromField_AllFieldTypes(t *testing.T) {
	t.Parallel()

	// Test all field types with nil default values to test conditional logic
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			manager := ksail.NewManager(testCase.fieldSelector)
			cmd := &cobra.Command{Use: "test"}

			manager.AddFlagsFromFields(cmd)

			// Should have one flag
			assert.True(t, cmd.Flags().HasFlags())

			// Check flag type
			var flagFound bool

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Value.Type() == testCase.expectedType {
					flagFound = true
				}
			})
			assert.True(t, flagFound, "Expected flag type %s not found", testCase.expectedType)
		})
	}
}
